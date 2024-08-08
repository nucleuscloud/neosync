package genbenthosconfigs_activity

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"google.golang.org/protobuf/encoding/protojson"
)

func buildSqlUpdateProcessorConfigs(
	config *tabledependency.RunConfig,
	redisConfig *shared.RedisConfig,
	jobId, runId string,
	transformedFktoPkMap map[string][]*referenceKey,
) ([]*neosync_benthos.ProcessorConfig, error) {
	processorConfigs := []*neosync_benthos.ProcessorConfig{}
	for fkCol, pks := range transformedFktoPkMap {
		for _, pk := range pks {
			if !slices.Contains(config.InsertColumns, fkCol) {
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
		for _, pk := range config.PrimaryKeys {
			// primary key
			hashedKey := neosync_benthos.HashBenthosCacheKey(jobId, runId, config.Table, pk)
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
	transformedFktoPkMap map[string][]*referenceKey,
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
		if slices.Contains(runconfig.InsertColumns, col.Column) {
			filteredCols = append(filteredCols, col)
		}
	}
	jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, filteredCols)
	if err != nil {
		return nil, err
	}

	mutations, err := buildMutationConfigs(ctx, transformerclient, filteredCols, tableColumnInfo, runconfig.SplitColumnPaths)
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
			mappings = append(mappings, fmt.Sprintf("meta neosync_%s_%s_%s = this.%q", col.Schema, col.Table, col.Column, col.Column))
		}
	}
	return strings.Join(mappings, "\n")
}

func buildBranchCacheConfigs(
	cols []*mgmtv1alpha1.JobMapping,
	transformedFktoPkMap map[string][]*referenceKey,
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
function setNestedProperty(obj, path, value) {
	path.split('.').reduce((current, part, index, parts) => {
		if (index === parts.length - 1) {
			current[part] = value;
		} else {
			current[part] = current[part] ?? {};
		}
		return current[part];
	}, obj);
}
%s
const input = benthos.v0_msg_as_structured();
const output = { ...input };
%s
benthos.v0_msg_set_structured(output);
})();`, jsFunctionStrings, benthosOutputString)
	return jsCode
}

func constructBenthosJavascriptObject(col string, source mgmtv1alpha1.TransformerSource) string {
	switch source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT:
		return fmt.Sprintf(
			`setNestedProperty(output, %q, fn_%s(%s, input));`,
			col,
			sanitizeJsFunctionName(col),
			convertJsObjPathToOptionalChain(fmt.Sprintf("input.%s", col)),
		)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return fmt.Sprintf(
			`setNestedProperty(output, %q, fn_%s());`,
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

	switch col.Transformer.Source {
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL:
		categories := col.Transformer.Config.GetGenerateCategoricalConfig().Categories
		return fmt.Sprintf(`generate_categorical(categories: %q)`, categories), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL:
		emailType := col.GetTransformer().GetConfig().GetGenerateEmailConfig().GetEmailType()
		if emailType == mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UNSPECIFIED {
			emailType = mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4
		}
		return fmt.Sprintf(`generate_email(max_length:%d,email_type:%q)`, maxLen, dtoEmailTypeToBenthosEmailType(emailType)), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL:
		pd := col.Transformer.Config.GetTransformEmailConfig().PreserveDomain
		pl := col.Transformer.Config.GetTransformEmailConfig().PreserveLength
		excludedDomains := col.Transformer.Config.GetTransformEmailConfig().ExcludedDomains

		excludedDomainsStr, err := convertStringSliceToString(excludedDomains)
		if err != nil {
			return "", err
		}
		emailType := col.GetTransformer().GetConfig().GetTransformEmailConfig().GetEmailType()
		if emailType == mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UNSPECIFIED {
			emailType = mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4
		}

		invalidEmailAction := col.GetTransformer().GetConfig().GetTransformEmailConfig().GetInvalidEmailAction()
		if invalidEmailAction == mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_UNSPECIFIED {
			invalidEmailAction = mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_REJECT
		}

		return fmt.Sprintf(
			"transform_email(value:this.%s,preserve_domain:%t,preserve_length:%t,excluded_domains:%v,max_length:%d,email_type:%q,invalid_email_action:%q)",
			formattedColPath, pd, pl, excludedDomainsStr, maxLen, dtoEmailTypeToBenthosEmailType(emailType), dtoInvalidEmailActionToBenthosInvalidEmailAction(invalidEmailAction),
		), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL:
		return "generate_bool()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER:
		luhn := col.Transformer.Config.GetGenerateCardNumberConfig().ValidLuhn
		return fmt.Sprintf(`generate_card_number(valid_luhn:%t)`, luhn), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY:
		return fmt.Sprintf(`generate_city(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER:
		minValue := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Min
		maxValue := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Max
		return fmt.Sprintf(`generate_e164_phone_number(min:%d,max:%d)`, minValue, maxValue), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME:
		return fmt.Sprintf(`generate_first_name(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64:
		randomSign := col.Transformer.Config.GetGenerateFloat64Config().RandomizeSign
		minValue := col.Transformer.Config.GetGenerateFloat64Config().Min
		maxValue := col.Transformer.Config.GetGenerateFloat64Config().Max

		var precision *int64
		if col.GetTransformer().GetConfig().GetGenerateFloat64Config().GetPrecision() > 0 {
			userDefinedPrecision := col.GetTransformer().GetConfig().GetGenerateFloat64Config().GetPrecision()
			precision = &userDefinedPrecision
		}
		if colInfo != nil && colInfo.NumericPrecision != nil && *colInfo.NumericPrecision > 0 {
			newPrecision := transformer_utils.Ceil(*precision, int64(*colInfo.NumericPrecision))
			precision = &newPrecision
		}

		var scale *int64
		if colInfo != nil && colInfo.NumericScale != nil && *colInfo.NumericScale >= 0 {
			newScale := int64(*colInfo.NumericScale)
			scale = &newScale
		}

		fnStr := []string{"randomize_sign:%t", "min:%f", "max:%f"}
		params := []any{randomSign, minValue, maxValue}

		if precision != nil {
			fnStr = append(fnStr, "precision: %d")
			params = append(params, *precision)
		}
		if scale != nil {
			fnStr = append(fnStr, "scale: %d")
			params = append(params, *scale)
		}
		template := fmt.Sprintf("generate_float64(%s)", strings.Join(fnStr, ", "))
		return fmt.Sprintf(template, params...), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS:
		return fmt.Sprintf(`generate_full_address(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME:
		return fmt.Sprintf(`generate_full_name(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_GENDER:
		ab := col.Transformer.Config.GetGenerateGenderConfig().Abbreviate
		return fmt.Sprintf(`generate_gender(abbreviate:%t,max_length:%d)`, ab, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER:
		return "generate_int64_phone_number()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64:
		sign := col.Transformer.Config.GetGenerateInt64Config().RandomizeSign
		minValue := col.Transformer.Config.GetGenerateInt64Config().Min
		maxValue := col.Transformer.Config.GetGenerateInt64Config().Max
		return fmt.Sprintf(`generate_int64(randomize_sign:%t,min:%d, max:%d)`, sign, minValue, maxValue), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME:
		return fmt.Sprintf(`generate_last_name(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH:
		return `generate_sha256hash()`, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN:
		return "generate_ssn()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE:
		generateFullName := col.Transformer.Config.GetGenerateStateConfig().GenerateFullName
		return fmt.Sprintf(`generate_state(generate_full_name:%t)`, generateFullName), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS:
		return fmt.Sprintf(`generate_street_address(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER:
		minValue := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Min
		maxValue := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Max
		minValue = transformer_utils.MinInt(minValue, maxLen)
		maxValue = transformer_utils.Ceil(maxValue, maxLen)
		return fmt.Sprintf("generate_string_phone_number(min:%d,max:%d)", minValue, maxValue), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING:
		minValue := col.Transformer.Config.GetGenerateStringConfig().Min
		maxValue := col.Transformer.Config.GetGenerateStringConfig().Max
		minValue = transformer_utils.MinInt(minValue, maxLen) // ensure the min is not larger than the max allowed length
		maxValue = transformer_utils.Ceil(maxValue, maxLen)
		// todo: we need to pull in the min from the database schema
		return fmt.Sprintf(`generate_string(min:%d,max:%d)`, minValue, maxValue), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP:
		return "generate_unixtimestamp()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_USERNAME:
		return fmt.Sprintf(`generate_username(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP:
		return "generate_utctimestamp()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID:
		ih := col.Transformer.Config.GetGenerateUuidConfig().IncludeHyphens
		return fmt.Sprintf("generate_uuid(include_hyphens:%t)", ih), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_ZIPCODE:
		return "generate_zipcode()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER:
		pl := col.Transformer.Config.GetTransformE164PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_e164_phone_number(value:this.%s,preserve_length:%t,max_length:%d)", formattedColPath, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME:
		pl := col.Transformer.Config.GetTransformFirstNameConfig().PreserveLength
		return fmt.Sprintf("transform_first_name(value:this.%s,preserve_length:%t,max_length:%d)", formattedColPath, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64:
		rMin := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMax

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

		fnStr := []string{"value:this.%s", "randomization_range_min:%f", "randomization_range_max:%f"}
		params := []any{formattedColPath, rMin, rMax}

		if precision != nil {
			fnStr = append(fnStr, "precision:%d")
			params = append(params, *precision)
		}
		if scale != nil {
			fnStr = append(fnStr, "scale:%d")
			params = append(params, *scale)
		}
		template := fmt.Sprintf(`transform_float64(%s)`, strings.Join(fnStr, ", "))
		return fmt.Sprintf(template, params...), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME:
		pl := col.Transformer.Config.GetTransformFullNameConfig().PreserveLength
		return fmt.Sprintf("transform_full_name(value:this.%s,preserve_length:%t,max_length:%d)", formattedColPath, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER:
		pl := col.Transformer.Config.GetTransformInt64PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_int64_phone_number(value:this.%s,preserve_length:%t)", formattedColPath, pl), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64:
		rMin := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_int64(value:this.%s,randomization_range_min:%d,randomization_range_max:%d)`, formattedColPath, rMin, rMax), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME:
		pl := col.Transformer.Config.GetTransformLastNameConfig().PreserveLength
		return fmt.Sprintf("transform_last_name(value:this.%s,preserve_length:%t,max_length:%d)", formattedColPath, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER:
		pl := col.Transformer.Config.GetTransformPhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_phone_number(value:this.%s,preserve_length:%t,max_length:%d)", formattedColPath, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING:
		pl := col.Transformer.Config.GetTransformStringConfig().PreserveLength
		minLength := int64(3) // todo: we need to pull in this value from the database schema
		return fmt.Sprintf(`transform_string(value:this.%s,preserve_length:%t,min_length:%d,max_length:%d)`, formattedColPath, pl, minLength, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL:
		return shared.NullString, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT:
		return `"DEFAULT"`, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE:
		regex := col.Transformer.Config.GetTransformCharacterScrambleConfig().UserProvidedRegex

		if regex != nil {
			regexValue := *regex
			return fmt.Sprintf(`transform_character_scramble(value:this.%s,user_provided_regex:%q)`, formattedColPath, regexValue), nil
		} else {
			return fmt.Sprintf(`transform_character_scramble(value:this.%s)`, formattedColPath), nil
		}

	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}

func dtoEmailTypeToBenthosEmailType(dto mgmtv1alpha1.GenerateEmailType) transformers.GenerateEmailType {
	switch dto {
	case mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_FULLNAME:
		return transformers.GenerateEmailType_FullName
	default:
		return transformers.GenerateEmailType_UuidV4
	}
}

func dtoInvalidEmailActionToBenthosInvalidEmailAction(dto mgmtv1alpha1.InvalidEmailAction) transformers.InvalidEmailAction {
	switch dto {
	case mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_GENERATE:
		return transformers.InvalidEmailAction_Generate
	case mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_NULL:
		return transformers.InvalidEmailAction_Null
	case mgmtv1alpha1.InvalidEmailAction_INVALID_EMAIL_ACTION_PASSTHROUGH:
		return transformers.InvalidEmailAction_Passthrough
	default:
		return transformers.InvalidEmailAction_Reject
	}
}
