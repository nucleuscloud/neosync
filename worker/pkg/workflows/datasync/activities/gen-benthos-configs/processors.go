package genbenthosconfigs_activity

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

func buildSqlUpdateProcessorConfigs(
	config *tabledependency.RunConfig,
	redisConfig *shared.RedisConfig,
	jobId, runId string,
	mappings []*mgmtv1alpha1.JobMapping,
	fkMap map[string]*dbschemas_utils.ForeignKey,
) ([]*neosync_benthos.ProcessorConfig, error) {
	processorConfigs := []*neosync_benthos.ProcessorConfig{}
	colSourceMap := map[string]mgmtv1alpha1.TransformerSource{}
	for _, col := range mappings {
		colSourceMap[col.Column] = col.GetTransformer().Source
	}
	for pkCol, fk := range fkMap {
		// only need redis processors if the primary key has a transformer
		if !hasTransformer(colSourceMap[pkCol]) || !slices.Contains(config.Columns, fk.Column) {
			continue
		}

		// circular dependent foreign key
		hashedKey := neosync_benthos.HashBenthosCacheKey(jobId, runId, fk.Table, pkCol)
		requestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, fk.Column)
		argsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, fk.Column)
		resultMap := fmt.Sprintf("root.%q = this", fk.Column)
		fkBranch, err := buildRedisGetBranchConfig(resultMap, argsMapping, &requestMap, redisConfig)
		if err != nil {
			return nil, err
		}
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Branch: fkBranch})

		// primary key
		pkRequestMap := fmt.Sprintf(`root = if this.%q == null { deleted() } else { this }`, pkCol)
		pkArgsMapping := fmt.Sprintf(`root = [%q, json(%q)]`, hashedKey, pkCol)
		pkResultMap := fmt.Sprintf("root.%q = this", pkCol)
		pkBranch, err := buildRedisGetBranchConfig(pkResultMap, pkArgsMapping, &pkRequestMap, redisConfig)
		if err != nil {
			return nil, err
		}
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Branch: pkBranch})
	}
	if len(processorConfigs) > 0 {
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
	tableColumnInfo map[string]*dbschemas_utils.ColumnInfo,
	columnConstraints map[string]*dbschemas_utils.ForeignKey,
	primaryKeys []string,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
) ([]*neosync_benthos.ProcessorConfig, error) {
	jsCode, err := extractJsFunctionsAndOutputs(ctx, transformerclient, cols)
	if err != nil {
		return nil, err
	}

	mutations, err := buildMutationConfigs(ctx, transformerclient, cols, tableColumnInfo)
	if err != nil {
		return nil, err
	}

	cacheBranches, err := buildBranchCacheConfigs(cols, columnConstraints, jobId, runId, redisConfig)
	if err != nil {
		return nil, err
	}

	pkMapping := buildPrimaryKeyMappingConfigs(cols, primaryKeys)

	var processorConfigs []*neosync_benthos.ProcessorConfig
	if pkMapping != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Mapping: &pkMapping})
	}
	if mutations != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Mutation: &mutations})
	}
	if jsCode != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Javascript: &neosync_benthos.JavascriptConfig{Code: jsCode}})
	}
	if len(cacheBranches) > 0 {
		for _, config := range cacheBranches {
			processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Branch: config})
		}
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
	tableColumnInfo map[string]*dbschemas_utils.ColumnInfo,
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
				mutation, err := computeMutationFunction(col, colInfo)
				if err != nil {
					return "", fmt.Errorf("%s is not a supported transformer: %w", col.Transformer, err)
				}
				mutations = append(mutations, fmt.Sprintf("root.%q = %s", col.Column, mutation))
			}
		}
	}

	return strings.Join(mutations, "\n"), nil
}

func buildPrimaryKeyMappingConfigs(cols []*mgmtv1alpha1.JobMapping, primaryKeys []string) string {
	mappings := []string{}
	for _, col := range cols {
		if shouldProcessColumn(col.Transformer) && slices.Contains(primaryKeys, col.Column) {
			mappings = append(mappings, fmt.Sprintf("meta neosync_%s = this.%q", col.Column, col.Column))
		}
	}
	return strings.Join(mappings, "\n")
}

func buildBranchCacheConfigs(
	cols []*mgmtv1alpha1.JobMapping,
	columnConstraints map[string]*dbschemas_utils.ForeignKey,
	jobId, runId string,
	redisConfig *shared.RedisConfig,
) ([]*neosync_benthos.BranchConfig, error) {
	branchConfigs := []*neosync_benthos.BranchConfig{}
	for _, col := range cols {
		fk, ok := columnConstraints[col.Column]
		if ok {
			// skip self referencing cols
			if fk.Table == fmt.Sprintf("%s.%s", col.Schema, col.Table) {
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
`, col, jsCode)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return fmt.Sprintf(`
function fn_%s(){
  %s
};
`, col, jsCode)
	default:
		return ""
	}
}

func constructBenthosJsProcessor(jsFunctions, benthosOutputs []string) string {
	jsFunctionStrings := strings.Join(jsFunctions, "\n")

	benthosOutputString := strings.Join(benthosOutputs, "\n")

	jsCode := fmt.Sprintf(`
(() => {
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
		return fmt.Sprintf(`output["%[1]s"] = fn_%[1]s(input["%[1]s"], input);`, col)
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT:
		return fmt.Sprintf(`output["%[1]s"] = fn_%[1]s();`, col)
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

func computeMutationFunction(col *mgmtv1alpha1.JobMapping, colInfo *dbschemas_utils.ColumnInfo) (string, error) {
	var maxLen int64 = 10000
	if colInfo != nil && colInfo.CharacterMaximumLength != nil && *colInfo.CharacterMaximumLength > 0 {
		maxLen = int64(*colInfo.CharacterMaximumLength)
	}

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

		return fmt.Sprintf(
			"transform_email(email:this.%q,preserve_domain:%t,preserve_length:%t,excluded_domains:%v,max_length:%d,email_type:%q)",
			col.Column, pd, pl, excludedDomainsStr, maxLen, dtoEmailTypeToBenthosEmailType(emailType),
		), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL:
		return "generate_bool()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER:
		luhn := col.Transformer.Config.GetGenerateCardNumberConfig().ValidLuhn
		return fmt.Sprintf(`generate_card_number(valid_luhn:%t)`, luhn), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY:
		return fmt.Sprintf(`generate_city(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER:
		min := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Min
		max := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Max
		return fmt.Sprintf(`generate_e164_phone_number(min:%d,max:%d)`, min, max), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME:
		return fmt.Sprintf(`generate_first_name(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64:
		randomSign := col.Transformer.Config.GetGenerateFloat64Config().RandomizeSign
		min := col.Transformer.Config.GetGenerateFloat64Config().Min
		max := col.Transformer.Config.GetGenerateFloat64Config().Max
		precision := col.Transformer.Config.GetGenerateFloat64Config().Precision
		return fmt.Sprintf(`generate_float64(randomize_sign:%t, min:%f, max:%f, precision:%d)`, randomSign, min, max, precision), nil
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
		min := col.Transformer.Config.GetGenerateInt64Config().Min
		max := col.Transformer.Config.GetGenerateInt64Config().Max
		return fmt.Sprintf(`generate_int64(randomize_sign:%t,min:%d, max:%d)`, sign, min, max), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME:
		return fmt.Sprintf(`generate_last_name(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH:
		return `generate_sha256hash()`, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN:
		return "generate_ssn()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE:
		return "generate_state()", nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS:
		return fmt.Sprintf(`generate_street_address(max_length:%d)`, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER:
		min := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Min
		max := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Max
		min = transformer_utils.MinInt(min, maxLen)
		max = transformer_utils.Ceil(max, maxLen)
		return fmt.Sprintf("generate_string_phone_number(min:%d,max:%d)", min, max), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING:
		min := col.Transformer.Config.GetGenerateStringConfig().Min
		max := col.Transformer.Config.GetGenerateStringConfig().Max
		min = transformer_utils.MinInt(min, maxLen) // ensure the min is not larger than the max allowed length
		max = transformer_utils.Ceil(max, maxLen)
		// todo: we need to pull in the min from the database schema
		return fmt.Sprintf(`generate_string(min:%d,max:%d)`, min, max), nil
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
		return fmt.Sprintf("transform_e164_phone_number(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME:
		pl := col.Transformer.Config.GetTransformFirstNameConfig().PreserveLength
		return fmt.Sprintf("transform_first_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64:
		rMin := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_float64(value:this.%q,randomization_range_min:%f,randomization_range_max:%f)`, col.Column, rMin, rMax), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME:
		pl := col.Transformer.Config.GetTransformFullNameConfig().PreserveLength
		return fmt.Sprintf("transform_full_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER:
		pl := col.Transformer.Config.GetTransformInt64PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_int64_phone_number(value:this.%q,preserve_length:%t)", col.Column, pl), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64:
		rMin := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_int64(value:this.%q,randomization_range_min:%d,randomization_range_max:%d)`, col.Column, rMin, rMax), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME:
		pl := col.Transformer.Config.GetTransformLastNameConfig().PreserveLength
		return fmt.Sprintf("transform_last_name(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER:
		pl := col.Transformer.Config.GetTransformPhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_phone_number(value:this.%q,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING:
		pl := col.Transformer.Config.GetTransformStringConfig().PreserveLength
		minLength := int64(3) // todo: we need to pull in this value from the database schema
		return fmt.Sprintf(`transform_string(value:this.%q,preserve_length:%t,min_length:%d,max_length:%d)`, col.Column, pl, minLength, maxLen), nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL:
		return shared.NullString, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT:
		return `"DEFAULT"`, nil
	case mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE:
		regex := col.Transformer.Config.GetTransformCharacterScrambleConfig().UserProvidedRegex

		if regex != nil {
			regexValue := *regex
			return fmt.Sprintf(`transform_character_scramble(value:this.%q,user_provided_regex:%q)`, col.Column, regexValue), nil
		} else {
			return fmt.Sprintf(`transform_character_scramble(value:this.%q)`, col.Column), nil
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
