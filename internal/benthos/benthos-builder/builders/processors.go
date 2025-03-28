package benthosbuilder_builders

import (
	"context"
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	javascript_userland "github.com/nucleuscloud/neosync/internal/javascript/userland"
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	tablesync_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/shared"
	"google.golang.org/protobuf/encoding/protojson"
)

func buildProcessorConfigsByRunType(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	config *runconfigs.RunConfig,
	columnForeignKeysMap map[string][]*bb_internal.ReferenceKey,
	transformedFktoPkMap map[string][]*bb_internal.ReferenceKey,
	jobId, runId string,
	redisConfig *neosync_redis.RedisConfig,
	mappings []*shared.JobTransformationMapping,
	columnInfoMap map[string]*sqlmanager_shared.DatabaseSchemaRow,
	jobSourceOptions *mgmtv1alpha1.JobSourceOptions,
	mappedKeys []string,
) ([]*neosync_benthos.ProcessorConfig, error) {
	if config.RunType() == runconfigs.RunTypeUpdate {
		// sql update processor configs
		processorConfigs, err := buildSqlUpdateProcessorConfigs(
			config,
			redisConfig,
			jobId,
			runId,
			transformedFktoPkMap,
		)
		if err != nil {
			return nil, err
		}
		return processorConfigs, nil
	} else {
		// sql insert processor configs
		fkSourceCols := []string{}
		for col := range columnForeignKeysMap {
			fkSourceCols = append(fkSourceCols, col)
		}
		processorConfigs, err := buildProcessorConfigs(
			ctx,
			transformerclient,
			mappings,
			columnInfoMap,
			transformedFktoPkMap,
			fkSourceCols,
			jobId,
			runId,
			redisConfig,
			config,
			jobSourceOptions,
			mappedKeys,
		)
		if err != nil {
			return nil, err
		}
		return processorConfigs, nil
	}
}

func buildSqlUpdateProcessorConfigs(
	config *runconfigs.RunConfig,
	redisConfig *neosync_redis.RedisConfig,
	jobId, runId string,
	transformedFktoPkMap map[string][]*bb_internal.ReferenceKey,
) ([]*neosync_benthos.ProcessorConfig, error) {
	processorConfigs := []*neosync_benthos.ProcessorConfig{}
	for fkCol, pks := range transformedFktoPkMap {
		for _, pk := range pks {
			if !slices.Contains(config.InsertColumns(), fkCol) {
				continue
			}

			// circular dependent foreign key
			hashedKey := neosync_benthos.HashBenthosCacheKey(jobId, runId, pk.Table, pk.Column)
			requestMap := fmt.Sprintf(
				`root = if this.%q == null { deleted() } else { this }`,
				fkCol,
			)
			argsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, fkCol)
			resultMap := fmt.Sprintf("root.%q = this", fkCol)
			fkBranch, err := buildRedisGetBranchConfig(
				resultMap,
				argsMapping,
				&requestMap,
				redisConfig,
			)
			if err != nil {
				return nil, err
			}
			processorConfigs = append(
				processorConfigs,
				&neosync_benthos.ProcessorConfig{Branch: fkBranch},
			)
		}
	}

	if len(processorConfigs) > 0 {
		for _, pk := range config.PrimaryKeys() {
			// primary key
			hashedKey := neosync_benthos.HashBenthosCacheKey(jobId, runId, config.Table(), pk)
			pkRequestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, pk)
			pkArgsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, pk)
			pkResultMap := fmt.Sprintf("root.%q = this", pk)
			pkBranch, err := buildRedisGetBranchConfig(
				pkResultMap,
				pkArgsMapping,
				&pkRequestMap,
				redisConfig,
			)
			if err != nil {
				return nil, err
			}
			processorConfigs = append(
				processorConfigs,
				&neosync_benthos.ProcessorConfig{Branch: pkBranch},
			)
		}
		// add catch and error processor
		processorConfigs = append(
			processorConfigs,
			&neosync_benthos.ProcessorConfig{Catch: []*neosync_benthos.ProcessorConfig{
				{Error: &neosync_benthos.ErrorProcessorConfig{
					ErrorMsg: `${! error()}`,
				}},
			}},
		)
	}
	return processorConfigs, nil
}

func buildProcessorConfigs(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	cols []*shared.JobTransformationMapping,
	tableColumnInfo map[string]*sqlmanager_shared.DatabaseSchemaRow,
	transformedFktoPkMap map[string][]*bb_internal.ReferenceKey,
	fkSourceCols []string,
	jobId, runId string,
	redisConfig *neosync_redis.RedisConfig,
	runconfig *runconfigs.RunConfig,
	jobSourceOptions *mgmtv1alpha1.JobSourceOptions,
	mappedKeys []string,
) ([]*neosync_benthos.ProcessorConfig, error) {
	// filter columns by config insert cols
	filteredColumnMappings := []*shared.JobTransformationMapping{}
	for _, col := range cols {
		if slices.Contains(runconfig.InsertColumns(), col.Column) {
			filteredColumnMappings = append(filteredColumnMappings, col)
		}
	}
	jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, filteredColumnMappings)
	if err != nil {
		return nil, err
	}

	mutations, err := buildMutationConfigs(
		ctx,
		transformerclient,
		filteredColumnMappings,
		tableColumnInfo,
		runconfig.SplitColumnPaths(),
	)
	if err != nil {
		return nil, err
	}

	cacheBranches, err := buildBranchCacheConfigs(
		filteredColumnMappings,
		transformedFktoPkMap,
		jobId,
		runId,
		redisConfig,
	)
	if err != nil {
		return nil, err
	}

	pkMapping := buildPrimaryKeyMappingConfigs(filteredColumnMappings, fkSourceCols)

	defaultTransformerConfig, err := buildDefaultTransformerConfigs(jobSourceOptions, mappedKeys)
	if err != nil {
		return nil, err
	}
	var processorConfigs []*neosync_benthos.ProcessorConfig
	if pkMapping != "" {
		processorConfigs = append(
			processorConfigs,
			&neosync_benthos.ProcessorConfig{Mapping: &pkMapping},
		)
	}
	if mutations != "" {
		processorConfigs = append(
			processorConfigs,
			&neosync_benthos.ProcessorConfig{Mutation: &mutations},
		)
	}
	if jsCode != "" {
		processorConfigs = append(
			processorConfigs,
			&neosync_benthos.ProcessorConfig{
				NeosyncJavascript: &neosync_benthos.NeosyncJavascriptConfig{Code: jsCode},
			},
		)
	}
	if len(cacheBranches) > 0 {
		for _, config := range cacheBranches {
			processorConfigs = append(
				processorConfigs,
				&neosync_benthos.ProcessorConfig{Branch: config},
			)
		}
	}
	if defaultTransformerConfig != nil {
		processorConfigs = append(
			processorConfigs,
			&neosync_benthos.ProcessorConfig{NeosyncDefaultTransformer: defaultTransformerConfig},
		)
	}

	if len(processorConfigs) > 0 {
		// add catch and error processor
		processorConfigs = append(
			processorConfigs,
			&neosync_benthos.ProcessorConfig{Catch: []*neosync_benthos.ProcessorConfig{
				{Error: &neosync_benthos.ErrorProcessorConfig{
					ErrorMsg: `${! error()}`,
				}},
			}},
		)
	}

	return processorConfigs, err
}

func buildDefaultTransformerConfigs(
	jobSourceOptions *mgmtv1alpha1.JobSourceOptions,
	mappedKeys []string,
) (*neosync_benthos.NeosyncDefaultTransformerConfig, error) {
	// only available for dynamodb source
	if jobSourceOptions == nil || jobSourceOptions.GetDynamodb() == nil {
		return nil, nil
	}

	sourceOptBits, err := protojson.Marshal(jobSourceOptions)
	if err != nil {
		return nil, err
	}
	return &neosync_benthos.NeosyncDefaultTransformerConfig{
		JobSourceOptionsString: string(sourceOptBits),
		MappedKeys:             mappedKeys,
	}, nil
}

func extractJsFunctionsAndOutputs(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	cols []*shared.JobTransformationMapping,
) (string, error) {
	var benthosOutputs []string
	var jsFunctions []string

	for _, col := range cols {
		jmTransformer := col.GetTransformer()
		if shouldProcessStrict(jmTransformer) {
			if jmTransformer.GetConfig().GetUserDefinedTransformerConfig() != nil {
				val, err := convertUserDefinedFunctionConfig(
					ctx,
					transformerclient,
					col.GetTransformer(),
				)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				jmTransformer = val
			}
			switch cfg := jmTransformer.GetConfig().GetConfig().(type) {
			case *mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig:
				code := cfg.TransformJavascriptConfig.GetCode()
				if code != "" {
					jsFunctions = append(jsFunctions, constructJsFunction(code, col.Column, mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT))
					benthosOutputs = append(benthosOutputs, constructBenthosJavascriptObject(col.Column, mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT))
				}
			case *mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig:
				code := cfg.GenerateJavascriptConfig.GetCode()
				if code != "" {
					jsFunctions = append(jsFunctions, constructJsFunction(code, col.Column, mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT))
					benthosOutputs = append(benthosOutputs, constructBenthosJavascriptObject(col.Column, mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT))
				}
			}
		}
	}

	if len(jsFunctions) > 0 {
		return javascript_userland.GetFunction(jsFunctions, benthosOutputs), nil
	} else {
		return "", nil
	}
}

// Checks if it is a gen or transform javascript
func isJavascriptTransformer(jmt *mgmtv1alpha1.JobMappingTransformer) bool {
	if jmt == nil {
		return false
	}

	isConfig := jmt.GetConfig().GetTransformJavascriptConfig() != nil ||
		jmt.GetConfig().GetGenerateJavascriptConfig() != nil
	return isConfig
}

func buildIdentityCursors(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	cols []*shared.JobTransformationMapping,
) (map[string]*tablesync_shared.IdentityCursor, error) {
	cursors := map[string]*tablesync_shared.IdentityCursor{}

	for _, col := range cols {
		transformer := col.GetTransformer()

		if transformer.GetConfig().GetUserDefinedTransformerConfig() != nil {
			val, err := convertUserDefinedFunctionConfig(ctx, transformerclient, transformer)
			if err != nil {
				return nil, fmt.Errorf(
					"unable to look up user defined transformer config by id: %w",
					err,
				)
			}
			transformer = val
		}

		scrambleConfig := transformer.GetConfig().GetTransformScrambleIdentityConfig()
		if scrambleConfig != nil {
			cursors[buildScrambleIdentityToken(col)] = tablesync_shared.NewDefaultIdentityCursor()
		}
	}

	return cursors, nil
}

func buildMutationConfigs(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	cols []*shared.JobTransformationMapping,
	tableColumnInfo map[string]*sqlmanager_shared.DatabaseSchemaRow,
	splitColumnPaths bool,
) (string, error) {
	mutations := []string{}

	for _, col := range cols {
		colInfo := tableColumnInfo[col.GetColumn()]
		if shouldProcessColumn(col.GetTransformer()) {
			if col.GetTransformer().GetConfig().GetUserDefinedTransformerConfig() != nil {
				// handle user defined transformer -> get the user defined transformer configs using the id
				val, err := convertUserDefinedFunctionConfig(
					ctx,
					transformerclient,
					col.GetTransformer(),
				)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				col.Transformer = val
			}
			if !isJavascriptTransformer(col.GetTransformer()) {
				mutation, err := computeMutationFunction(col, colInfo, splitColumnPaths)
				if err != nil {
					return "", fmt.Errorf(
						"%s is not a supported transformer: %w",
						col.GetTransformer(),
						err,
					)
				}
				mutations = append(
					mutations,
					fmt.Sprintf(
						"root.%s = %s",
						getBenthosColumnKey(col.GetColumn(), splitColumnPaths),
						mutation,
					),
				)
			}
		}
	}
	return strings.Join(mutations, "\n"), nil
}

const pathSeparator = "."

func getBenthosColumnKey(column string, shouldSplitPath bool) string {
	if shouldSplitPath {
		segments := strings.Split(column, pathSeparator)
		quotedSegments := make([]string, 0, len(segments))
		for _, segment := range segments {
			quotedSegments = append(quotedSegments, fmt.Sprintf("%q", segment))
		}
		return strings.Join(quotedSegments, pathSeparator)
	}
	return fmt.Sprintf("%q", column)
}

func buildPrimaryKeyMappingConfigs(cols []*shared.JobTransformationMapping, primaryKeys []string) string {
	mappings := []string{}
	for _, col := range cols {
		if shouldProcessColumn(col.Transformer) && slices.Contains(primaryKeys, col.Column) {
			mappings = append(
				mappings,
				fmt.Sprintf(
					"meta %s = this.%q",
					hashPrimaryKeyMetaKey(col.Schema, col.Table, col.Column),
					col.Column,
				),
			)
		}
	}
	return strings.Join(mappings, "\n")
}

func hashPrimaryKeyMetaKey(schema, table, column string) string {
	return generateSha1Hash(fmt.Sprintf("neosync_%s_%s_%s", schema, table, column))
}

func generateSha1Hash(input string) string {
	hasher := sha1.New() //nolint:gosec
	hasher.Write([]byte(input))
	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes)
}

func buildBranchCacheConfigs(
	cols []*shared.JobTransformationMapping,
	transformedFktoPkMap map[string][]*bb_internal.ReferenceKey,
	jobId, runId string,
	redisConfig *neosync_redis.RedisConfig,
) ([]*neosync_benthos.BranchConfig, error) {
	branchConfigs := []*neosync_benthos.BranchConfig{}
	for _, col := range cols {
		fks, ok := transformedFktoPkMap[col.Column]
		if ok {
			for _, fk := range fks {
				// skip self referencing cols
				if fk.Table == neosync_benthos.BuildBenthosTable(col.Schema, col.Table) {
					continue
				}

				hashedKey := neosync_benthos.HashBenthosCacheKey(jobId, runId, fk.Table, fk.Column)
				requestMap := fmt.Sprintf(
					`root = if this.%q == null { deleted() } else { this }`,
					col.Column,
				)
				argsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, col.Column)
				resultMap := fmt.Sprintf("root.%q = this", col.Column)
				br, err := buildRedisGetBranchConfig(
					resultMap,
					argsMapping,
					&requestMap,
					redisConfig,
				)
				if err != nil {
					return nil, err
				}
				branchConfigs = append(branchConfigs, br)
			}
		}
	}
	return branchConfigs, nil
}

func buildRedisGetBranchConfig(
	resultMap, argsMapping string,
	requestMap *string,
	redisConfig *neosync_redis.RedisConfig,
) (*neosync_benthos.BranchConfig, error) {
	if redisConfig == nil {
		return nil, fmt.Errorf("missing redis config. this operation requires redis")
	}
	return &neosync_benthos.BranchConfig{
		RequestMap: requestMap,
		Processors: []neosync_benthos.ProcessorConfig{
			{
				Redis: &neosync_benthos.RedisProcessorConfig{
					Url:         redisConfig.Url,
					Command:     "hget",
					ArgsMapping: argsMapping,
					Kind:        &redisConfig.Kind,
					Master:      redisConfig.Master,
					Tls:         shared.BuildBenthosRedisTlsConfig(redisConfig),
				},
			},
		},
		ResultMap: &resultMap,
	}, nil
}

func constructJsFunction(jsCode, col string, source mgmtv1alpha1.TransformerSource) string {
	switch source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT:
		return javascript_userland.GetTransformJavascriptFunction(jsCode, col, true)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return javascript_userland.GetGenerateJavascriptFunction(jsCode, col)
	default:
		return ""
	}
}

func constructBenthosJavascriptObject(col string, source mgmtv1alpha1.TransformerSource) string {
	switch source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT:
		return javascript_userland.BuildOutputSetter(col, true, true)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return javascript_userland.BuildOutputSetter(col, false, false)
	default:
		return ""
	}
}

// takes in an user defined config with just an id field and return the right transformer config for that user defined function id
func convertUserDefinedFunctionConfig(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	t *mgmtv1alpha1.JobMappingTransformer,
) (*mgmtv1alpha1.JobMappingTransformer, error) {
	transformerResp, err := transformerclient.GetUserDefinedTransformerById(
		ctx,
		connect.NewRequest(
			&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
				TransformerId: t.Config.GetUserDefinedTransformerConfig().Id,
			},
		),
	)
	if err != nil {
		return nil, err
	}
	transformer := transformerResp.Msg.GetTransformer()

	return &mgmtv1alpha1.JobMappingTransformer{
		Config: transformer.GetConfig(),
	}, nil
}

func computeMutationFunction(
	col *shared.JobTransformationMapping,
	colInfo *sqlmanager_shared.DatabaseSchemaRow,
	splitColumnPath bool,
) (string, error) {
	var maxLen int64 = 10000
	if colInfo != nil && colInfo.CharacterMaximumLength > 0 {
		maxLen = int64(colInfo.CharacterMaximumLength)
	}

	formattedColPath := getBenthosColumnKey(col.Column, splitColumnPath)
	config := col.GetTransformer().GetConfig()

	switch cfg := config.GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig:
		opts, err := transformers.NewGenerateCategoricalOptsFromConfig(cfg.GenerateCategoricalConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateEmailConfig:
		opts, err := transformers.NewGenerateEmailOptsFromConfig(cfg.GenerateEmailConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_TransformEmailConfig:
		opts, err := transformers.NewTransformEmailOptsFromConfig(cfg.TransformEmailConfig, &maxLen)
		if err != nil {
			return "", nil
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateBoolConfig:
		opts, err := transformers.NewGenerateBoolOptsFromConfig(cfg.GenerateBoolConfig)
		if err != nil {
			return "", nil
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig:
		opts, err := transformers.NewGenerateCardNumberOptsFromConfig(cfg.GenerateCardNumberConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateCityConfig:
		opts, err := transformers.NewGenerateCityOptsFromConfig(cfg.GenerateCityConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig:
		opts, err := transformers.NewGenerateInternationalPhoneNumberOptsFromConfig(cfg.GenerateE164PhoneNumberConfig)
		if err != nil {
			return "", nil
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig:
		opts, err := transformers.NewGenerateFirstNameOptsFromConfig(cfg.GenerateFirstNameConfig, &maxLen)
		if err != nil {
			return "", nil
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateFloat64Config:
		var precision *int64
		if cfg.GenerateFloat64Config.GetPrecision() > 0 {
			userDefinedPrecision := cfg.GenerateFloat64Config.GetPrecision()
			precision = &userDefinedPrecision
			cfg.GenerateFloat64Config.Precision = &userDefinedPrecision
		}
		if colInfo != nil && colInfo.NumericPrecision > 0 {
			newPrecision := transformer_utils.Ceil(*precision, int64(colInfo.NumericPrecision))
			precision = &newPrecision
		}
		if precision != nil {
			config.GetGenerateFloat64Config().Precision = precision
		}

		var scale *int64
		if colInfo != nil && colInfo.NumericScale >= 0 {
			newScale := int64(colInfo.NumericScale)
			scale = &newScale
		}

		opts, err := transformers.NewGenerateFloat64OptsFromConfig(cfg.GenerateFloat64Config, scale)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig:
		opts, err := transformers.NewGenerateFullAddressOptsFromConfig(cfg.GenerateFullAddressConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig:
		opts, err := transformers.NewGenerateFullNameOptsFromConfig(cfg.GenerateFullNameConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateGenderConfig:
		opts, err := transformers.NewGenerateGenderOptsFromConfig(cfg.GenerateGenderConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig:
		opts, err := transformers.NewGenerateInt64PhoneNumberOptsFromConfig(cfg.GenerateInt64PhoneNumberConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateInt64Config:
		opts, err := transformers.NewGenerateInt64OptsFromConfig(cfg.GenerateInt64Config)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig:
		opts, err := transformers.NewGenerateLastNameOptsFromConfig(cfg.GenerateLastNameConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig:
		opts, err := transformers.NewGenerateSHA256HashOptsFromConfig(cfg.GenerateSha256HashConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateSsnConfig:
		opts, err := transformers.NewGenerateSSNOptsFromConfig(cfg.GenerateSsnConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateStateConfig:
		opts, err := transformers.NewGenerateStateOptsFromConfig(cfg.GenerateStateConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig:
		opts, err := transformers.NewGenerateStreetAddressOptsFromConfig(cfg.GenerateStreetAddressConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig:
		opts, err := transformers.NewGenerateStringPhoneNumberOptsFromConfig(cfg.GenerateStringPhoneNumberConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateStringConfig:
		// todo: we need to pull in the min from the database schema
		opts, err := transformers.NewGenerateRandomStringOptsFromConfig(cfg.GenerateStringConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig:
		opts, err := transformers.NewGenerateUnixTimestampOptsFromConfig(cfg.GenerateUnixtimestampConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig:
		opts, err := transformers.NewGenerateUsernameOptsFromConfig(cfg.GenerateUsernameConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig:
		opts, err := transformers.NewGenerateUTCTimestampOptsFromConfig(cfg.GenerateUtctimestampConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateUuidConfig:
		opts, err := transformers.NewGenerateUUIDOptsFromConfig(cfg.GenerateUuidConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig:
		opts, err := transformers.NewGenerateZipcodeOptsFromConfig(cfg.GenerateZipcodeConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig:
		opts, err := transformers.NewTransformE164PhoneNumberOptsFromConfig(cfg.TransformE164PhoneNumberConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig:
		opts, err := transformers.NewTransformFirstNameOptsFromConfig(cfg.TransformFirstNameConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformFloat64Config:
		var precision *int64
		if colInfo != nil && colInfo.NumericPrecision > 0 {
			newPrecision := int64(colInfo.NumericPrecision)
			precision = &newPrecision
		}
		var scale *int64
		if colInfo != nil && colInfo.NumericScale >= 0 {
			newScale := int64(colInfo.NumericScale)
			scale = &newScale
		}
		opts, err := transformers.NewTransformFloat64OptsFromConfig(cfg.TransformFloat64Config, scale, precision)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformFullNameConfig:
		opts, err := transformers.NewTransformFullNameOptsFromConfig(cfg.TransformFullNameConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig:
		opts, err := transformers.NewTransformInt64PhoneNumberOptsFromConfig(cfg.TransformInt64PhoneNumberConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformInt64Config:
		opts, err := transformers.NewTransformInt64OptsFromConfig(cfg.TransformInt64Config)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformLastNameConfig:
		opts, err := transformers.NewTransformLastNameOptsFromConfig(cfg.TransformLastNameConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig:
		opts, err := transformers.NewTransformStringPhoneNumberOptsFromConfig(cfg.TransformPhoneNumberConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformStringConfig:
		minLength := int64(3) // todo: we need to pull in this value from the database schema
		opts, err := transformers.NewTransformStringOptsFromConfig(cfg.TransformStringConfig, &minLength, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_Nullconfig:
		return shared.NullString, nil
	case *mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig:
		return `"DEFAULT"`, nil
	case *mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig:
		opts, err := transformers.NewTransformCharacterScrambleOptsFromConfig(cfg.TransformCharacterScrambleConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateCountryConfig:
		opts, err := transformers.NewGenerateCountryOptsFromConfig(cfg.GenerateCountryConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateBusinessNameConfig:
		opts, err := transformers.NewGenerateBusinessNameOptsFromConfig(cfg.GenerateBusinessNameConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_GenerateIpAddressConfig:
		opts, err := transformers.NewGenerateIpAddressOptsFromConfig(cfg.GenerateIpAddressConfig, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case *mgmtv1alpha1.TransformerConfig_TransformUuidConfig:
		opts, err := transformers.NewTransformUuidOptsFromConfig(cfg.TransformUuidConfig)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case *mgmtv1alpha1.TransformerConfig_TransformScrambleIdentityConfig:
		token := buildScrambleIdentityToken(col)
		opts, err := transformers.NewTransformIdentityScrambleOptsFromConfigWithToken(token)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	default:
		return "", fmt.Errorf("unsupported transformer: %T", cfg)
	}
}

func buildScrambleIdentityToken(col *shared.JobTransformationMapping) string {
	return neosync_benthos.ToSha256(
		fmt.Sprintf("%s.%s.%s", col.GetSchema(), col.GetTable(), col.GetColumn()),
	)
}

func shouldProcessColumn(t *mgmtv1alpha1.JobMappingTransformer) bool {
	switch t.GetConfig().GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig,
		nil:
		return false
	default:
		return true
	}
}

func shouldProcessStrict(t *mgmtv1alpha1.JobMappingTransformer) bool {
	switch t.GetConfig().GetConfig().(type) {
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig,
		*mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig,
		*mgmtv1alpha1.TransformerConfig_Nullconfig,
		nil:
		return false
	default:
		return true
	}
}
