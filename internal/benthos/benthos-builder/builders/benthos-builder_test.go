package benthosbuilder_builders

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	bb_internal "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/internal"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	rc "github.com/nucleuscloud/neosync/internal/runconfigs"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	neosync_benthos_transformers "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

const (
	mockJobId      = "b1767636-3992-4cb4-9bf2-4bb9bddbf43c"
	mockWorkflowId = "b1767636-3992-4cb4-9bf2-4bb9bddbf43c-workflowid"
	mockRunId      = "26444272-0bb0-4325-ae60-17dcd9744785"
)

var driver = sqlmanager_shared.PostgresDriver

func Test_ProcessorConfigEmpty(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := map[string]*tableMapping{
		"public.users": {Schema: "public",
			Table: "users",
			Mappings: []*shared.JobTransformationMapping{
				{
					DestinationSchema: "public",
					DestinationTable:  "users",
					JobMapping: &mgmtv1alpha1.JobMapping{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
							},
						},
					},
				},
				{
					DestinationSchema: "public",
					DestinationTable:  "users",
					JobMapping: &mgmtv1alpha1.JobMapping{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
							},
						},
					},
				},
			},
		},
	}

	groupedSchemas := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"public.users": {
			"id": &sqlmanager_shared.DatabaseSchemaRow{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             false,
				DataType:               "",
				CharacterMaximumLength: 0,
				NumericPrecision:       0,
				NumericScale:           0,
			},
			"name": &sqlmanager_shared.DatabaseSchemaRow{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             false,
				DataType:               "",
				CharacterMaximumLength: 0,
				NumericPrecision:       0,
				NumericScale:           0,
			},
		},
	}
	groupedTransformers := map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{
		"public.users": {
			"id":   &mgmtv1alpha1.JobMappingTransformer{},
			"name": &mgmtv1alpha1.JobMappingTransformer{},
		},
	}

	runconfigId := "public.users.insert"
	queryMap := map[string]*sqlmanager_shared.SelectQuery{
		"public.users.insert": {Query: ""},
	}
	runconfigs := []*rc.RunConfig{
		rc.NewRunConfig(runconfigId, sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"}, rc.RunTypeInsert, []string{"id"}, nil, []string{"id", "name"}, []string{"id", "name"}, []*rc.DependsOn{}, false),
	}
	logger := testutil.GetTestLogger(t)
	connectionId := uuid.NewString()

	res, err := buildBenthosSqlSourceConfigResponses(
		logger,
		context.Background(),
		mockTransformerClient,
		tableMappings,
		runconfigs,
		connectionId,
		queryMap,
		groupedSchemas,
		map[string][]*sqlmanager_shared.ForeignConstraint{},
		groupedTransformers,
		mockJobId,
		mockRunId,
		nil,
	)
	require.Nil(t, err)
	require.Empty(t, res[0].Config.StreamConfig.Pipeline.Processors)
}

func Test_ProcessorConfigEmptyJavascript(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	tableMappings := map[string]*tableMapping{
		"public.users": {Schema: "public",
			Table: "users",
			Mappings: []*shared.JobTransformationMapping{
				{
					DestinationSchema: "public",
					DestinationTable:  "users",
					JobMapping: &mgmtv1alpha1.JobMapping{
						Schema: "public",
						Table:  "users",
						Column: "id",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
									PassthroughConfig: &mgmtv1alpha1.Passthrough{},
								},
							},
						},
					},
				},
				{
					DestinationSchema: "public",
					DestinationTable:  "users",
					JobMapping: &mgmtv1alpha1.JobMapping{
						Schema: "public",
						Table:  "users",
						Column: "name",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
									TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: ""},
								},
							},
						},
					},
				},
			},
		},
	}

	groupedSchemas := map[string]map[string]*sqlmanager_shared.DatabaseSchemaRow{
		"public.users": {
			"id": &sqlmanager_shared.DatabaseSchemaRow{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             false,
				DataType:               "",
				CharacterMaximumLength: 0,
				NumericPrecision:       0,
				NumericScale:           0,
			},
			"name": &sqlmanager_shared.DatabaseSchemaRow{
				OrdinalPosition:        1,
				ColumnDefault:          "324",
				IsNullable:             false,
				DataType:               "",
				CharacterMaximumLength: 0,
				NumericPrecision:       0,
				NumericScale:           0,
			},
		},
	}

	groupedTransformers := map[string]map[string]*mgmtv1alpha1.JobMappingTransformer{
		"public.users": {
			"id":   &mgmtv1alpha1.JobMappingTransformer{},
			"name": &mgmtv1alpha1.JobMappingTransformer{},
		},
	}

	schemaTable := sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"}
	runconfigs := []*rc.RunConfig{
		rc.NewRunConfig(schemaTable.String(), schemaTable, rc.RunTypeInsert, []string{"id"}, nil, []string{"id", "name"}, []string{"id", "name"}, []*rc.DependsOn{}, false),
	}

	queryMap := map[string]*sqlmanager_shared.SelectQuery{
		schemaTable.String(): {Query: ""},
	}
	logger := testutil.GetTestLogger(t)
	connectionId := uuid.NewString()
	res, err := buildBenthosSqlSourceConfigResponses(
		logger,
		context.Background(),
		mockTransformerClient,
		tableMappings,
		runconfigs,
		connectionId,
		queryMap,
		groupedSchemas,
		map[string][]*sqlmanager_shared.ForeignConstraint{},
		groupedTransformers,
		mockJobId,
		mockRunId,
		nil,
	)
	require.NoError(t, err)
	require.Empty(t, res[0].Config.StreamConfig.Pipeline.Processors)
}

func Test_buildProcessorConfigsMutation(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	schemaTable := sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"}
	runconfig := rc.NewRunConfig(schemaTable.String(), schemaTable, rc.RunTypeInsert, []string{}, nil, []string{}, []string{}, []*rc.DependsOn{}, false)
	output, err := buildProcessorConfigs(ctx, mockTransformerClient, []*shared.JobTransformationMapping{}, map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, runconfig, nil, []string{})
	require.Nil(t, err)
	require.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*shared.JobTransformationMapping{}, map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, runconfig, nil, []string{})
	require.Nil(t, err)
	require.Empty(t, output)

	runconfig = rc.NewRunConfig(schemaTable.String(), schemaTable, rc.RunTypeInsert, []string{}, nil, []string{}, []string{"id"}, []*rc.DependsOn{}, false)
	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*shared.JobTransformationMapping{
		{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "id"}},
	}, map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, runconfig, nil, []string{})
	require.Nil(t, err)
	require.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*shared.JobTransformationMapping{
		{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{}}},
	}, map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, runconfig, nil, []string{})
	require.Nil(t, err)
	require.Empty(t, output)

	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*shared.JobTransformationMapping{
		{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{},
		}}}},
	}, map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, runconfig, nil, []string{})
	require.Nil(t, err)
	require.Empty(t, output)

	runconfig = rc.NewRunConfig(schemaTable.String(), schemaTable, rc.RunTypeInsert, []string{}, nil, []string{}, []string{"id", "name"}, []*rc.DependsOn{}, false)
	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*shared.JobTransformationMapping{
		{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}}},
		{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "name", Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
				Nullconfig: &mgmtv1alpha1.Null{},
			},
		}}}},
	}, map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, runconfig, nil, []string{})

	require.Nil(t, err)

	require.Equal(t, *output[0].Mutation, "root.\"id\" = null\nroot.\"name\" = null")

	jsT := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain:  gotypeutil.ToPtr(true),
					PreserveLength:  gotypeutil.ToPtr(false),
					ExcludedDomains: []string{},
				},
			},
		},
	}

	emailLength := 40

	groupedSchemas := map[string]*sqlmanager_shared.DatabaseSchemaRow{

		"email": {
			OrdinalPosition:        2,
			ColumnDefault:          "",
			IsNullable:             true,
			DataType:               "timestamptz",
			CharacterMaximumLength: emailLength,
			NumericPrecision:       0,
			NumericScale:           0,
		},
	}

	runconfig = rc.NewRunConfig(schemaTable.String(), schemaTable, rc.RunTypeInsert, []string{"id"}, nil, []string{"email"}, []string{"email"}, []*rc.DependsOn{}, false)
	output, err = buildProcessorConfigs(ctx, mockTransformerClient, []*shared.JobTransformationMapping{
		{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "email", Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT.Config}}},
	}, groupedSchemas, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, runconfig, nil, []string{})

	require.Nil(t, err)
	require.Equal(t, `root."email" = transform_email(value:this."email",preserve_length:false,preserve_domain:true,excluded_domains:[],max_length:40,email_type:"uuidv4",invalid_email_action:"reject")`, *output[0].Mutation)
}

func Test_ShouldProcessColumnTrue(t *testing.T) {
	val := &mgmtv1alpha1.JobMappingTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
				GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{},
			},
		},
	}

	res := shouldProcessColumn(val)
	require.Equal(t, true, res)
}

func Test_ShouldProcessColumnFalse(t *testing.T) {
	val := &mgmtv1alpha1.JobMappingTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
				PassthroughConfig: &mgmtv1alpha1.Passthrough{},
			},
		},
	}

	res := shouldProcessColumn(val)
	require.Equal(t, false, res)
}

func Test_buildProcessorConfigsJavascriptEmpty(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)
	ctx := context.Background()

	jsT := mgmtv1alpha1.SystemTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
					Code: ``,
				},
			},
		},
	}

	schemaTable := sqlmanager_shared.SchemaTable{Schema: "public", Table: "users"}
	runconfig := rc.NewRunConfig(schemaTable.String(), schemaTable, rc.RunTypeInsert, []string{"id"}, nil, []string{"id"}, []string{"id"}, []*rc.DependsOn{}, false)
	resp, err := buildProcessorConfigs(ctx, mockTransformerClient, []*shared.JobTransformationMapping{
		{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "id", Transformer: &mgmtv1alpha1.JobMappingTransformer{Config: jsT.Config}}},
	}, map[string]*sqlmanager_shared.DatabaseSchemaRow{}, map[string][]*bb_internal.ReferenceKey{}, []string{}, mockJobId, mockRunId, runconfig, nil, []string{})

	require.NoError(t, err)
	require.Empty(t, resp)
}

func Test_convertUserDefinedFunctionConfig(t *testing.T) {
	mockTransformerClient := mgmtv1alpha1connect.NewMockTransformersServiceClient(t)

	ctx := context.Background()

	mockTransformerClient.On(
		"GetUserDefinedTransformerById",
		mock.Anything,
		connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
			TransformerId: "123",
		}),
	).Return(connect.NewResponse(&mgmtv1alpha1.GetUserDefinedTransformerByIdResponse{
		Transformer: &mgmtv1alpha1.UserDefinedTransformer{
			Id:          "123",
			Name:        "stage",
			Description: "description",
			DataType:    mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			Source:      mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain:  gotypeutil.ToPtr(true),
						PreserveLength:  gotypeutil.ToPtr(false),
						ExcludedDomains: []string{},
					},
				},
			},
		},
	}), nil)

	jmt := &mgmtv1alpha1.JobMappingTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig{
				UserDefinedTransformerConfig: &mgmtv1alpha1.UserDefinedTransformerConfig{
					Id: "123",
				},
			},
		},
	}

	expected := &mgmtv1alpha1.JobMappingTransformer{
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
				TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
					PreserveDomain:  gotypeutil.ToPtr(true),
					PreserveLength:  gotypeutil.ToPtr(false),
					ExcludedDomains: []string{},
				},
			},
		},
	}

	resp, err := convertUserDefinedFunctionConfig(ctx, mockTransformerClient, jmt)
	require.NoError(t, err)
	require.Equal(t, resp, expected)
}

func Test_buildPlainColumns(t *testing.T) {
	require.Empty(t, buildPlainColumns(nil))
	require.Empty(t, buildPlainColumns([]*shared.JobTransformationMapping{}))
	require.Equal(
		t,
		buildPlainColumns([]*shared.JobTransformationMapping{
			{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "foo"}},
			{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "bar"}},
			{DestinationSchema: "public", DestinationTable: "users", JobMapping: &mgmtv1alpha1.JobMapping{Schema: "public", Table: "users", Column: "baz"}},
		}),
		[]string{"foo", "bar", "baz"},
	)
}

func Test_buildBenthosS3Credentials(t *testing.T) {
	require.Nil(t, buildBenthosS3Credentials(nil))

	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{}),
		&neosync_benthos.AwsCredentials{},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{Profile: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Profile: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{AccessKeyId: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Id: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{SecretAccessKey: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Secret: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{SessionToken: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Token: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{FromEc2Role: shared.Ptr(true)}),
		&neosync_benthos.AwsCredentials{FromEc2Role: true},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{RoleArn: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{Role: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{RoleExternalId: shared.Ptr("foo")}),
		&neosync_benthos.AwsCredentials{RoleExternalId: "foo"},
	)
	require.Equal(
		t,
		buildBenthosS3Credentials(&mgmtv1alpha1.AwsS3Credentials{
			Profile:         shared.Ptr("profile"),
			AccessKeyId:     shared.Ptr("access-key"),
			SecretAccessKey: shared.Ptr("secret"),
			SessionToken:    shared.Ptr("session"),
			FromEc2Role:     shared.Ptr(false),
			RoleArn:         shared.Ptr("role"),
			RoleExternalId:  shared.Ptr("foo"),
		}),
		&neosync_benthos.AwsCredentials{
			Profile:        "profile",
			Id:             "access-key",
			Secret:         "secret",
			Token:          "session",
			FromEc2Role:    false,
			Role:           "role",
			RoleExternalId: "foo",
		},
	)
}

func Test_computeMutationFunction_null(t *testing.T) {
	val, err := computeMutationFunction(
		&shared.JobTransformationMapping{
			DestinationSchema: "public",
			DestinationTable:  "users",
			JobMapping: &mgmtv1alpha1.JobMapping{
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{}},
				},
			},
		},
		&sqlmanager_shared.DatabaseSchemaRow{},
		false,
	)
	require.NoError(t, err)
	require.Equal(t, val, "null")
}

func Test_computeMutationFunction_Validate_Bloblang_Output(t *testing.T) {
	uuidEmailType := mgmtv1alpha1.GenerateEmailType_GENERATE_EMAIL_TYPE_UUID_V4
	transformers := []*mgmtv1alpha1.SystemTransformer{
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_EMAIL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{
					GenerateEmailConfig: &mgmtv1alpha1.GenerateEmail{
						EmailType: &uuidEmailType,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_EMAIL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{
					TransformEmailConfig: &mgmtv1alpha1.TransformEmail{
						PreserveDomain:  gotypeutil.ToPtr(false),
						PreserveLength:  gotypeutil.ToPtr(false),
						ExcludedDomains: []string{"gmail", "yahoo"},
						EmailType:       &uuidEmailType,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_BOOL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{
					GenerateCardNumberConfig: &mgmtv1alpha1.GenerateCardNumber{
						ValidLuhn: gotypeutil.ToPtr(true),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CITY,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
					GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{
					GenerateE164PhoneNumberConfig: &mgmtv1alpha1.GenerateE164PhoneNumber{
						Min: gotypeutil.ToPtr(int64(9)),
						Max: gotypeutil.ToPtr(int64(15)),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FIRST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
					GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FLOAT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{
					GenerateFloat64Config: &mgmtv1alpha1.GenerateFloat64{
						RandomizeSign: gotypeutil.ToPtr(true),
						Min:           gotypeutil.ToPtr(1.00),
						Max:           gotypeutil.ToPtr(100.00),
						Precision:     gotypeutil.ToPtr(int64(6)),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{
					GenerateFullAddressConfig: &mgmtv1alpha1.GenerateFullAddress{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_FULL_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_GENDER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{
					GenerateGenderConfig: &mgmtv1alpha1.GenerateGender{
						Abbreviate: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{
					GenerateInt64PhoneNumberConfig: &mgmtv1alpha1.GenerateInt64PhoneNumber{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_INT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
						RandomizeSign: gotypeutil.ToPtr(true),
						Min:           gotypeutil.ToPtr(int64(1)),
						Max:           gotypeutil.ToPtr(int64(40)),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_LAST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{
					GenerateLastNameConfig: &mgmtv1alpha1.GenerateLastName{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SHA256HASH,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{
					GenerateSha256HashConfig: &mgmtv1alpha1.GenerateSha256Hash{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_SSN,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{
					GenerateSsnConfig: &mgmtv1alpha1.GenerateSSN{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STATE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{
					GenerateStateConfig: &mgmtv1alpha1.GenerateState{
						GenerateFullName: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{
					GenerateStreetAddressConfig: &mgmtv1alpha1.GenerateStreetAddress{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{
					GenerateStringPhoneNumberConfig: &mgmtv1alpha1.GenerateStringPhoneNumber{
						Min: gotypeutil.ToPtr(int64(9)),
						Max: gotypeutil.ToPtr(int64(14)),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
					GenerateStringConfig: &mgmtv1alpha1.GenerateString{
						Min: gotypeutil.ToPtr(int64(2)),
						Max: gotypeutil.ToPtr(int64(7)),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{
					GenerateUnixtimestampConfig: &mgmtv1alpha1.GenerateUnixTimestamp{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_USERNAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{
					GenerateUsernameConfig: &mgmtv1alpha1.GenerateUsername{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{
					GenerateUtctimestampConfig: &mgmtv1alpha1.GenerateUtcTimestamp{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_UUID,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
					GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
						IncludeHyphens: gotypeutil.ToPtr(true),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_ZIPCODE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{
					GenerateZipcodeConfig: &mgmtv1alpha1.GenerateZipcode{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{
					TransformE164PhoneNumberConfig: &mgmtv1alpha1.TransformE164PhoneNumber{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
					TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FLOAT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFloat64Config{
					TransformFloat64Config: &mgmtv1alpha1.TransformFloat64{
						RandomizationRangeMin: gotypeutil.ToPtr(20.00),
						RandomizationRangeMax: gotypeutil.ToPtr(50.00),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
					TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{
					TransformInt64PhoneNumberConfig: &mgmtv1alpha1.TransformInt64PhoneNumber{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_INT64,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{
					TransformInt64Config: &mgmtv1alpha1.TransformInt64{
						RandomizationRangeMin: gotypeutil.ToPtr(int64(20)),
						RandomizationRangeMax: gotypeutil.ToPtr(int64(50)),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{
					TransformLastNameConfig: &mgmtv1alpha1.TransformLastName{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{
					TransformPhoneNumberConfig: &mgmtv1alpha1.TransformPhoneNumber{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_STRING,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
					TransformStringConfig: &mgmtv1alpha1.TransformString{
						PreserveLength: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
					GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
						Categories: gotypeutil.ToPtr("value1,value2"),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{
						UserProvidedRegex: nil,
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
					GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{
					Nullconfig: &mgmtv1alpha1.Null{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_COUNTRY,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateCountryConfig{
					GenerateCountryConfig: &mgmtv1alpha1.GenerateCountry{
						GenerateFullName: gotypeutil.ToPtr(false),
					},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_SCRAMBLE_IDENTITY,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformScrambleIdentityConfig{
					TransformScrambleIdentityConfig: &mgmtv1alpha1.TransformScrambleIdentity{},
				},
			},
		},
		{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PII_TEXT,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{
					TransformPiiTextConfig: &mgmtv1alpha1.TransformPiiText{
						Language: gotypeutil.ToPtr("en"),
					},
				},
			},
		},
	}

	emailColInfo := &sqlmanager_shared.DatabaseSchemaRow{
		OrdinalPosition:        2,
		ColumnDefault:          "",
		IsNullable:             true,
		DataType:               "timestamptz",
		CharacterMaximumLength: 40,
		NumericPrecision:       0,
		NumericScale:           0,
	}

	blobenv := bloblang.NewEnvironment()
	neosync_benthos_transformers.RegisterTransformIdentityScramble(blobenv, nil)
	neosync_benthos_transformers.RegisterTransformPiiText(blobenv, nil)

	for _, transformer := range transformers {
		t.Run(fmt.Sprintf("%s_%T_lint", t.Name(), transformer.Config.Config), func(t *testing.T) {
			val, err := computeMutationFunction(
				&shared.JobTransformationMapping{
					DestinationSchema: "public",
					DestinationTable:  "users",
					JobMapping: &mgmtv1alpha1.JobMapping{
						Column: "email",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: transformer.Config,
						},
					},
				}, emailColInfo, false)
			require.NoError(t, err)
			ex, err := blobenv.Parse(val)
			require.NoError(t, err, fmt.Sprintf("transformer lint failed, check that the transformer string is being constructed correctly. Failing Config: %T", transformer.Config.Config))
			_, err = ex.Query(nil)
			require.NoError(t, err)
		})
	}
}

func Test_computeMutationFunction_Validate_Bloblang_Output_EmptyConfigs(t *testing.T) {
	transformers := []*mgmtv1alpha1.SystemTransformer{
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateEmailConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformEmailConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateCardNumberConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateE164PhoneNumberConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateFullAddressConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateGenderConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64PhoneNumberConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateLastNameConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateSha256HashConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateSsnConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateStateConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateStreetAddressConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateStringPhoneNumberConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateUnixtimestampConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateUsernameConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateUtctimestampConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateZipcodeConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformE164PhoneNumberConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateFloat64Config{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformInt64PhoneNumberConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformInt64Config{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformLastNameConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformPhoneNumberConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_Nullconfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_GenerateCountryConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformScrambleIdentityConfig{}},
		},
		{
			Config: &mgmtv1alpha1.TransformerConfig{Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{}},
		},
	}

	emailColInfo := &sqlmanager_shared.DatabaseSchemaRow{
		OrdinalPosition:        2,
		ColumnDefault:          "",
		IsNullable:             true,
		DataType:               "timestamptz",
		CharacterMaximumLength: 40,
		NumericPrecision:       0,
		NumericScale:           0,
	}

	blobenv := bloblang.NewEnvironment()
	neosync_benthos_transformers.RegisterTransformIdentityScramble(blobenv, nil)
	neosync_benthos_transformers.RegisterTransformPiiText(blobenv, nil)

	for _, transformer := range transformers {
		t.Run(fmt.Sprintf("%s_%T_lint", t.Name(), transformer.Config.Config), func(t *testing.T) {
			val, err := computeMutationFunction(
				&shared.JobTransformationMapping{
					DestinationSchema: "public",
					DestinationTable:  "users",
					JobMapping: &mgmtv1alpha1.JobMapping{
						Column: "email",
						Transformer: &mgmtv1alpha1.JobMappingTransformer{
							Config: transformer.Config,
						},
					},
				}, emailColInfo, false)
			require.NoError(t, err)
			ex, err := blobenv.Parse(val)
			require.NoError(t, err, fmt.Sprintf("transformer lint failed, check that the transformer string is being constructed correctly. Failing Config: %T", transformer.Config.Config))
			_, err = ex.Query(nil)
			require.NoError(t, err)
		})
	}
}

func Test_computeMutationFunction_handles_Db_Maxlen(t *testing.T) {
	type testcase struct {
		jm       *shared.JobTransformationMapping
		ci       *sqlmanager_shared.DatabaseSchemaRow
		expected string
	}
	jm := &shared.JobTransformationMapping{
		DestinationSchema: "public",
		DestinationTable:  "users",
		JobMapping: &mgmtv1alpha1.JobMapping{
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateStringConfig{
						GenerateStringConfig: &mgmtv1alpha1.GenerateString{
							Min: gotypeutil.ToPtr(int64(2)),
							Max: gotypeutil.ToPtr(int64(7)),
						},
					},
				},
			},
		},
	}
	testcases := []testcase{
		{
			jm:       jm,
			ci:       &sqlmanager_shared.DatabaseSchemaRow{},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sqlmanager_shared.DatabaseSchemaRow{
				CharacterMaximumLength: 0,
			},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sqlmanager_shared.DatabaseSchemaRow{
				CharacterMaximumLength: -1,
			},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sqlmanager_shared.DatabaseSchemaRow{
				CharacterMaximumLength: 0,
			},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sqlmanager_shared.DatabaseSchemaRow{
				CharacterMaximumLength: 10,
			},
			expected: "generate_string(min:2,max:7)",
		},
		{
			jm: jm,
			ci: &sqlmanager_shared.DatabaseSchemaRow{
				CharacterMaximumLength: 3,
			},
			expected: "generate_string(min:2,max:3)",
		},
		{
			jm: jm,
			ci: &sqlmanager_shared.DatabaseSchemaRow{
				CharacterMaximumLength: 1,
			},
			expected: "generate_string(min:1,max:1)",
		},
	}

	for _, tc := range testcases {
		t.Run(t.Name(), func(t *testing.T) {
			out, err := computeMutationFunction(tc.jm, tc.ci, false)
			require.NoError(t, err)
			require.NotNil(t, out)
			require.Equal(t, tc.expected, out, "computed bloblang string was not expected")
			ex, err := bloblang.Parse(out)
			require.NoError(t, err)
			_, err = ex.Query(nil)
			require.NoError(t, err)
		})
	}
}

func Test_buildBranchCacheConfigs_null(t *testing.T) {
	cols := []*shared.JobTransformationMapping{
		{
			DestinationSchema: "public",
			DestinationTable:  "users",
			JobMapping: &mgmtv1alpha1.JobMapping{
				Schema: "public",
				Table:  "users",
				Column: "user_id",
			},
		},
	}

	constraints := map[string][]*bb_internal.ReferenceKey{
		"name": {
			{
				Table:  "public.orders",
				Column: "buyer_id",
			},
		},
	}

	resp := buildBranchCacheConfigs(cols, constraints, mockJobId, mockRunId)
	require.Len(t, resp, 0)
}

func Test_buildBranchCacheConfigs_missing_redis(t *testing.T) {
	cols := []*shared.JobTransformationMapping{
		{
			DestinationSchema: "public",
			DestinationTable:  "users",
			JobMapping: &mgmtv1alpha1.JobMapping{
				Schema: "public",
				Table:  "users",
				Column: "user_id",
			},
		},
	}

	constraints := map[string][]*bb_internal.ReferenceKey{
		"user_id": {
			{
				Table:  "public.orders",
				Column: "buyer_id",
			},
		},
	}

	resp := buildBranchCacheConfigs(cols, constraints, mockJobId, mockRunId)
	require.Len(t, resp, 1)
}

func Test_buildBranchCacheConfigs_success(t *testing.T) {
	cols := []*shared.JobTransformationMapping{
		{
			DestinationSchema: "public",
			DestinationTable:  "users",
			JobMapping: &mgmtv1alpha1.JobMapping{
				Schema: "public",
				Table:  "users",
				Column: "user_id",
			},
		},
		{
			DestinationSchema: "public",
			DestinationTable:  "users",
			JobMapping: &mgmtv1alpha1.JobMapping{
				Schema: "public",
				Table:  "users",
				Column: "name",
			},
		},
	}

	constraints := map[string][]*bb_internal.ReferenceKey{
		"user_id": {
			{
				Table:  "public.orders",
				Column: "buyer_id",
			},
		},
	}
	resp := buildBranchCacheConfigs(cols, constraints, mockJobId, mockRunId)

	require.Len(t, resp, 1)
	require.Equal(t, *resp[0].RequestMap, `root = if this."user_id" == null { deleted() } else { this }`)
	require.Equal(t, *resp[0].ResultMap, `root."user_id" = this`)
}

func Test_buildBranchCacheConfigs_self_referencing(t *testing.T) {
	cols := []*shared.JobTransformationMapping{
		{
			DestinationSchema: "public",
			DestinationTable:  "users",
			JobMapping: &mgmtv1alpha1.JobMapping{
				Schema: "public",
				Table:  "users",
				Column: "user_id",
			},
		},
	}

	constraints := map[string][]*bb_internal.ReferenceKey{
		"user_id": {
			{
				Table:  "public.users",
				Column: "other_id",
			},
		},
	}

	resp := buildBranchCacheConfigs(cols, constraints, mockJobId, mockRunId)
	require.Len(t, resp, 0)
}

func Test_getPrimaryKeyDependencyMap(t *testing.T) {
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"hr.countries": {
			{
				Columns:     []string{"region_id"},
				NotNullable: []bool{true},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "hr.regions",
					Columns: []string{"region_id"},
				},
			},
		},
		"hr.departments": {
			{
				Columns:     []string{"location_id"},
				NotNullable: []bool{false},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "hr.locations",
					Columns: []string{"location_id"},
				},
			},
		},
		"hr.dependents": {
			{
				Columns:     []string{"employee_id"},
				NotNullable: []bool{true},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "hr.employees",
					Columns: []string{"employee_id"},
				},
			},
		},
		"hr.employees": {
			{
				Columns:     []string{"job_id"},
				NotNullable: []bool{true},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "hr.jobs",
					Columns: []string{"job_id"},
				},
			},
			{
				Columns:     []string{"department_id"},
				NotNullable: []bool{false},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "hr.departments",
					Columns: []string{"department_id"},
				},
			},
			{
				Columns:     []string{"manager_id"},
				NotNullable: []bool{false},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "hr.employees",
					Columns: []string{"employee_id"},
				},
			},
		},
		"hr.locations": {
			{
				Columns:     []string{"country_id"},
				NotNullable: []bool{true},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "hr.countries",
					Columns: []string{"country_id"},
				},
			},
		},
	}

	expected := map[string]map[string][]*bb_internal.ReferenceKey{
		"hr.regions": {
			"region_id": {
				{
					Table:  "hr.countries",
					Column: "region_id",
				},
			},
		},
		"hr.locations": {
			"location_id": {
				{
					Table:  "hr.departments",
					Column: "location_id",
				},
			},
		},
		"hr.employees": {
			"employee_id": {
				{
					Table:  "hr.dependents",
					Column: "employee_id",
				},
				{
					Table:  "hr.employees",
					Column: "manager_id",
				},
			},
		},
		"hr.jobs": {
			"job_id": {
				{
					Table:  "hr.employees",
					Column: "job_id",
				},
			},
		},
		"hr.departments": {
			"department_id": {
				{
					Table:  "hr.employees",
					Column: "department_id",
				},
			},
		},
		"hr.countries": {
			"country_id": {
				{
					Table:  "hr.locations",
					Column: "country_id",
				},
			},
		},
	}

	actual := getPrimaryKeyDependencyMap(tableDependencies)
	for table, depsMap := range expected {
		actualDepsMap := actual[table]
		require.NotNil(t, actualDepsMap)
		for col, deps := range depsMap {
			actualDeps := actualDepsMap[col]
			require.ElementsMatch(t, deps, actualDeps)
		}
	}
}

func Test_getPrimaryKeyDependencyMap_compositekeys(t *testing.T) {
	tableDependencies := map[string][]*sqlmanager_shared.ForeignConstraint{
		"employees": {
			{
				Columns:     []string{"department_id"},
				NotNullable: []bool{false},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "department",
					Columns: []string{"department_id"},
				},
			},
		},
		"projects": {
			{
				Columns:     []string{"responsible_employee_id", "responsible_department_id"},
				NotNullable: []bool{true},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "employees",
					Columns: []string{"employee_id", "department_id"},
				},
			},
		},
	}

	expected := map[string]map[string][]*bb_internal.ReferenceKey{
		"department": {
			"department_id": {
				{
					Table:  "employees",
					Column: "department_id",
				},
			},
		},
		"employees": {
			"employee_id": {{
				Table:  "projects",
				Column: "responsible_employee_id",
			}},
			"department_id": {{
				Table:  "projects",
				Column: "responsible_department_id",
			}},
		},
	}

	actual := getPrimaryKeyDependencyMap(tableDependencies)
	require.Equal(t, expected, actual)
}
