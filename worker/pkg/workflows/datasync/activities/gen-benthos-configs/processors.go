package genbenthosconfigs_activity

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

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
				ErrorMsg: `${! meta("fallback_error")}`,
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
