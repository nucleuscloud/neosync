package benthosbuilder_builders

import (
	"context"
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"

	"connectrpc.com/connect"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"google.golang.org/protobuf/encoding/protojson"
)

func buildProcessorConfigsByRunType(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	config *tabledependency.RunConfig,
	columnForeignKeysMap map[string][]*bb_internal.ReferenceKey,
	transformedFktoPkMap map[string][]*bb_internal.ReferenceKey,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
	mappings []*mgmtv1alpha1.JobMapping,
	columnInfoMap map[string]*sqlmanager_shared.ColumnInfo,
	jobSourceOptions *mgmtv1alpha1.JobSourceOptions,
	mappedKeys []string,
) ([]*neosync_benthos.ProcessorConfig, error) {
	if config.RunType() == tabledependency.RunTypeUpdate {
		// sql update processor configs
		processorConfigs, err := buildSqlUpdateProcessorConfigs(config, redisConfig, jobId, runId, transformedFktoPkMap)
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
	config *tabledependency.RunConfig,
	redisConfig *shared.RedisConfig,
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
			requestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, fkCol)
			argsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, fkCol)
			resultMap := fmt.Sprintf("root.%q = this", fkCol)
			fkBranch, err := buildRedisGetBranchConfig(resultMap, argsMapping, &requestMap, redisConfig)
			if err != nil {
				return nil, err
			}
			processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Branch: fkBranch})
		}
	}

	if len(processorConfigs) > 0 {
		for _, pk := range config.PrimaryKeys() {
			// primary key
			hashedKey := neosync_benthos.HashBenthosCacheKey(jobId, runId, config.Table(), pk)
			pkRequestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, pk)
			pkArgsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, pk)
			pkResultMap := fmt.Sprintf("root.%q = this", pk)
			pkBranch, err := buildRedisGetBranchConfig(pkResultMap, pkArgsMapping, &pkRequestMap, redisConfig)
			if err != nil {
				return nil, err
			}
			processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Branch: pkBranch})
		}
		// add catch and error processor
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Catch: []*neosync_benthos.ProcessorConfig{
			{Error: &neosync_benthos.ErrorProcessorConfig{
				ErrorMsg: `${! error()}`,
			}},
		}})
	}
	return processorConfigs, nil
}

func buildProcessorConfigs(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	cols []*mgmtv1alpha1.JobMapping,
	tableColumnInfo map[string]*sqlmanager_shared.ColumnInfo,
	transformedFktoPkMap map[string][]*bb_internal.ReferenceKey,
	fkSourceCols []string,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
	runconfig *tabledependency.RunConfig,
	jobSourceOptions *mgmtv1alpha1.JobSourceOptions,
	mappedKeys []string,
) ([]*neosync_benthos.ProcessorConfig, error) {
	// filter columns by config insert cols
	filteredCols := []*mgmtv1alpha1.JobMapping{}
	for _, col := range cols {
		if slices.Contains(runconfig.InsertColumns(), col.Column) {
			filteredCols = append(filteredCols, col)
		}
	}
	jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, filteredCols)
	if err != nil {
		return nil, err
	}

	mutations, err := buildMutationConfigs(ctx, transformerclient, filteredCols, tableColumnInfo, runconfig.SplitColumnPaths())
	if err != nil {
		return nil, err
	}

	cacheBranches, err := buildBranchCacheConfigs(filteredCols, transformedFktoPkMap, jobId, runId, redisConfig)
	if err != nil {
		return nil, err
	}

	pkMapping := buildPrimaryKeyMappingConfigs(filteredCols, fkSourceCols)

	defaultTransformerConfig, err := buildDefaultTransformerConfigs(jobSourceOptions, mappedKeys)
	if err != nil {
		return nil, err
	}
	var processorConfigs []*neosync_benthos.ProcessorConfig
	if pkMapping != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Mapping: &pkMapping})
	}
	if mutations != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Mutation: &mutations})
	}
	if jsCode != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{NeosyncJavascript: &neosync_benthos.NeosyncJavascriptConfig{Code: jsCode}})
	}
	if len(cacheBranches) > 0 {
		for _, config := range cacheBranches {
			processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Branch: config})
		}
	}
	if defaultTransformerConfig != nil {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{NeosyncDefaultTransformer: defaultTransformerConfig})
	}

	if len(processorConfigs) > 0 {
		// add catch and error processor
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Catch: []*neosync_benthos.ProcessorConfig{
			{Error: &neosync_benthos.ErrorProcessorConfig{
				ErrorMsg: `${! error()}`,
			}},
		}})
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

func extractJsFunctionsAndOutputs(ctx context.Context, transformerclient mgmtv1alpha1connect.TransformersServiceClient, cols []*mgmtv1alpha1.JobMapping) (string, error) {
	var benthosOutputs []string
	var jsFunctions []string

	for _, col := range cols {
		if shouldProcessStrict(col.Transformer) {
			if _, ok := col.Transformer.Config.Config.(*mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig); ok {
				val, err := convertUserDefinedFunctionConfig(ctx, transformerclient, col.Transformer)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				col.Transformer = val
			}
			switch col.Transformer.Source {
			case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT:
				code := col.Transformer.Config.GetTransformJavascriptConfig().Code
				if code != "" {
					jsFunctions = append(jsFunctions, constructJsFunction(code, col.Column, col.Transformer.Source))
					benthosOutputs = append(benthosOutputs, constructBenthosJavascriptObject(col.Column, col.Transformer.Source))
				}
			case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
				code := col.Transformer.Config.GetGenerateJavascriptConfig().Code
				if code != "" {
					jsFunctions = append(jsFunctions, constructJsFunction(code, col.Column, col.Transformer.Source))
					benthosOutputs = append(benthosOutputs, constructBenthosJavascriptObject(col.Column, col.Transformer.Source))
				}
			}
		}
	}

	if len(jsFunctions) > 0 {
		return constructBenthosJsProcessor(jsFunctions, benthosOutputs), nil
	} else {
		return "", nil
	}
}

func buildMutationConfigs(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	cols []*mgmtv1alpha1.JobMapping,
	tableColumnInfo map[string]*sqlmanager_shared.ColumnInfo,
	splitColumnPaths bool,
) (string, error) {
	mutations := []string{}

	for _, col := range cols {
		colInfo := tableColumnInfo[col.Column]
		if shouldProcessColumn(col.Transformer) {
			if _, ok := col.Transformer.Config.Config.(*mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig); ok {
				// handle user defined transformer -> get the user defined transformer configs using the id
				val, err := convertUserDefinedFunctionConfig(ctx, transformerclient, col.Transformer)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				col.Transformer = val
			}
			if col.Transformer.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT && col.Transformer.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT {
				mutation, err := computeMutationFunction(col, colInfo, splitColumnPaths)
				if err != nil {
					return "", fmt.Errorf("%s is not a supported transformer: %w", col.Transformer, err)
				}
				mutations = append(mutations, fmt.Sprintf("root.%s = %s", getBenthosColumnKey(col.Column, splitColumnPaths), mutation))
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

func buildPrimaryKeyMappingConfigs(cols []*mgmtv1alpha1.JobMapping, primaryKeys []string) string {
	mappings := []string{}
	for _, col := range cols {
		if shouldProcessColumn(col.Transformer) && slices.Contains(primaryKeys, col.Column) {
			mappings = append(mappings, fmt.Sprintf("meta %s = this.%q", hashPrimaryKeyMetaKey(col.Schema, col.Table, col.Column), col.Column))
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
	cols []*mgmtv1alpha1.JobMapping,
	transformedFktoPkMap map[string][]*bb_internal.ReferenceKey,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
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
				requestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, col.Column)
				argsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, col.Column)
				resultMap := fmt.Sprintf("root.%q = this", col.Column)
				br, err := buildRedisGetBranchConfig(resultMap, argsMapping, &requestMap, redisConfig)
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
	redisConfig *shared.RedisConfig,
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
		return fmt.Sprintf(`
function fn_%s(value, input){
  %s
};
`, sanitizeJsFunctionName(col), jsCode)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return fmt.Sprintf(`
function fn_%s(){
  %s
};
`, sanitizeJsFunctionName(col), jsCode)
	default:
		return ""
	}
}

func sanitizeJsFunctionName(input string) string {
	var result strings.Builder

	for i, r := range input {
		if unicode.IsLetter(r) || r == '_' || r == '$' || (unicode.IsDigit(r) && i > 0) {
			result.WriteRune(r)
		} else if unicode.IsDigit(r) && i == 0 {
			result.WriteRune('_')
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	return result.String()
}

func constructBenthosJsProcessor(jsFunctions, benthosOutputs []string) string {
	jsFunctionStrings := strings.Join(jsFunctions, "\n")

	benthosOutputString := strings.Join(benthosOutputs, "\n")

	jsCode := fmt.Sprintf(`
(() => {
%s
const input = benthos.v0_msg_as_structured();
const updatedValues = {}
%s
neosync.patchStructuredMessage(updatedValues)
})();`, jsFunctionStrings, benthosOutputString)
	return jsCode
}

func constructBenthosJavascriptObject(col string, source mgmtv1alpha1.TransformerSource) string {
	switch source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT:
		return fmt.Sprintf(
			`updatedValues[%q] = fn_%s(%s, input)`,
			col,
			sanitizeJsFunctionName(col),
			convertJsObjPathToOptionalChain(fmt.Sprintf("input.%s", col)),
		)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return fmt.Sprintf(
			`updatedValues[%q] = fn_%s()`,
			col,
			sanitizeJsFunctionName(col),
		)
	default:
		return ""
	}
}

func convertJsObjPathToOptionalChain(inputPath string) string {
	parts := strings.Split(inputPath, ".")
	for i := 1; i < len(parts); i++ {
		parts[i] = fmt.Sprintf("['%s']", parts[i])
	}
	return strings.Join(parts, "?.")
}

// takes in an user defined config with just an id field and return the right transformer config for that user defined function id
func convertUserDefinedFunctionConfig(
	ctx context.Context,
	transformerclient mgmtv1alpha1connect.TransformersServiceClient,
	t *mgmtv1alpha1.JobMappingTransformer,
) (*mgmtv1alpha1.JobMappingTransformer, error) {
	transformer, err := transformerclient.GetUserDefinedTransformerById(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{TransformerId: t.Config.GetUserDefinedTransformerConfig().Id}))
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.JobMappingTransformer{
		Source: transformer.Msg.Transformer.Source,
		Config: transformer.Msg.Transformer.Config,
	}, nil
}

/*
function transformers
root.{destination_col} = transformerfunction(args)
*/

func computeMutationFunction(col *mgmtv1alpha1.JobMapping, colInfo *sqlmanager_shared.ColumnInfo, splitColumnPath bool) (string, error) {
	var maxLen int64 = 10000
	if colInfo != nil && colInfo.CharacterMaximumLength != nil && *colInfo.CharacterMaximumLength > 0 {
		maxLen = int64(*colInfo.CharacterMaximumLength)
	}

	formattedColPath := getBenthosColumnKey(col.Column, splitColumnPath)
	config := col.GetTransformer().GetConfig()

	switch col.Transformer.Source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL:
		opts, err := transformers.NewGenerateCategoricalOptsFromConfig(config.GetGenerateCategoricalConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL:
		opts, err := transformers.NewGenerateEmailOptsFromConfig(config.GetGenerateEmailConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL:
		opts, err := transformers.NewTransformEmailOptsFromConfig(config.GetTransformEmailConfig(), &maxLen)
		if err != nil {
			return "", nil
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL:
		opts, err := transformers.NewGenerateBoolOptsFromConfig(config.GetGenerateBoolConfig())
		if err != nil {
			return "", nil
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER:
		opts, err := transformers.NewGenerateCardNumberOptsFromConfig(config.GetGenerateCardNumberConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY:
		opts, err := transformers.NewGenerateCityOptsFromConfig(config.GetGenerateCityConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER:
		opts, err := transformers.NewGenerateInternationalPhoneNumberOptsFromConfig(config.GetGenerateE164PhoneNumberConfig())
		if err != nil {
			return "", nil
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME:
		opts, err := transformers.NewGenerateFirstNameOptsFromConfig(config.GetGenerateFirstNameConfig(), &maxLen)
		if err != nil {
			return "", nil
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64:
		var precision *int64
		if config != nil && config.GetGenerateFloat64Config() != nil && config.GetGenerateFloat64Config().GetPrecision() > 0 {
			userDefinedPrecision := config.GetGenerateFloat64Config().GetPrecision()
			precision = &userDefinedPrecision
			config.GetGenerateFloat64Config().Precision = &userDefinedPrecision
		}
		if colInfo != nil && colInfo.NumericPrecision != nil && *colInfo.NumericPrecision > 0 {
			newPrecision := transformer_utils.Ceil(*precision, int64(*colInfo.NumericPrecision))
			precision = &newPrecision
		}
		if config != nil && config.GetGenerateFloat64Config() != nil && precision != nil {
			config.GetGenerateFloat64Config().Precision = precision
		}

		var scale *int64
		if colInfo != nil && colInfo.NumericScale != nil && *colInfo.NumericScale >= 0 {
			newScale := int64(*colInfo.NumericScale)
			scale = &newScale
		}

		opts, err := transformers.NewGenerateFloat64OptsFromConfig(col.Transformer.Config.GetGenerateFloat64Config(), scale)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS:
		opts, err := transformers.NewGenerateFullAddressOptsFromConfig(config.GetGenerateFullAddressConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME:
		opts, err := transformers.NewGenerateFullNameOptsFromConfig(config.GetGenerateFullNameConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_GENDER:
		opts, err := transformers.NewGenerateGenderOptsFromConfig(config.GetGenerateGenderConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER:
		opts, err := transformers.NewGenerateInt64PhoneNumberOptsFromConfig(config.GetGenerateInt64PhoneNumberConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64:
		opts, err := transformers.NewGenerateInt64OptsFromConfig(config.GetGenerateInt64Config())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME:
		opts, err := transformers.NewGenerateLastNameOptsFromConfig(config.GetGenerateLastNameConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH:
		opts, err := transformers.NewGenerateSHA256HashOptsFromConfig(config.GetGenerateSha256HashConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN:
		opts, err := transformers.NewGenerateSSNOptsFromConfig(config.GetGenerateSsnConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE:
		opts, err := transformers.NewGenerateStateOptsFromConfig(config.GetGenerateStateConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS:
		opts, err := transformers.NewGenerateStreetAddressOptsFromConfig(config.GetGenerateStreetAddressConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER:
		opts, err := transformers.NewGenerateStringPhoneNumberOptsFromConfig(config.GetGenerateStringPhoneNumberConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING:
		// todo: we need to pull in the min from the database schema
		opts, err := transformers.NewGenerateRandomStringOptsFromConfig(config.GetGenerateStringConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP:
		opts, err := transformers.NewGenerateUnixTimestampOptsFromConfig(config.GetGenerateUnixtimestampConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_USERNAME:
		opts, err := transformers.NewGenerateUsernameOptsFromConfig(config.GetGenerateUsernameConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP:
		opts, err := transformers.NewGenerateUTCTimestampOptsFromConfig(config.GetGenerateUtctimestampConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID:
		opts, err := transformers.NewGenerateUUIDOptsFromConfig(config.GetGenerateUuidConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_ZIPCODE:
		opts, err := transformers.NewGenerateZipcodeOptsFromConfig(config.GetGenerateZipcodeConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER:
		opts, err := transformers.NewTransformE164PhoneNumberOptsFromConfig(config.GetTransformE164PhoneNumberConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME:
		opts, err := transformers.NewTransformFirstNameOptsFromConfig(config.GetTransformFirstNameConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64:
		var precision *int64
		if colInfo != nil && colInfo.NumericPrecision != nil && *colInfo.NumericPrecision > 0 {
			newPrecision := int64(*colInfo.NumericPrecision)
			precision = &newPrecision
		}
		var scale *int64
		if colInfo != nil && colInfo.NumericScale != nil && *colInfo.NumericScale >= 0 {
			newScale := int64(*colInfo.NumericScale)
			scale = &newScale
		}
		opts, err := transformers.NewTransformFloat64OptsFromConfig(col.Transformer.Config.GetTransformFloat64Config(), scale, precision)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME:
		opts, err := transformers.NewTransformFullNameOptsFromConfig(config.GetTransformFullNameConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER:
		opts, err := transformers.NewTransformInt64PhoneNumberOptsFromConfig(config.GetTransformInt64PhoneNumberConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64:
		opts, err := transformers.NewTransformInt64OptsFromConfig(config.GetTransformInt64Config())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME:
		opts, err := transformers.NewTransformLastNameOptsFromConfig(config.GetTransformLastNameConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER:
		opts, err := transformers.NewTransformStringPhoneNumberOptsFromConfig(config.GetTransformPhoneNumberConfig(), &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING:
		minLength := int64(3) // todo: we need to pull in this value from the database schema
		opts, err := transformers.NewTransformStringOptsFromConfig(config.GetTransformStringConfig(), &minLength, &maxLen)
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL:
		return shared.NullString, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT:
		return `"DEFAULT"`, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE:
		opts, err := transformers.NewTransformCharacterScrambleOptsFromConfig(config.GetTransformCharacterScrambleConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(formattedColPath), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_COUNTRY:
		opts, err := transformers.NewGenerateCountryOptsFromConfig(config.GetGenerateCountryConfig())
		if err != nil {
			return "", err
		}
		return opts.BuildBloblangString(), nil
	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}
