package integrationtests_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	integrationtests_test "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v81"
)

func (s *IntegrationTestSuite) Test_AnonymizeService_AnonymizeMany() {
	t := s.T()

	t.Run("OSS-fail", func(t *testing.T) {
		userclient := s.OSSUnauthenticatedLicensedClients.Users()
		s.setUser(s.ctx, userclient)
		accountId := s.createPersonalAccount(s.ctx, userclient)
		resp, err := s.OSSUnauthenticatedLicensedClients.Anonymize().AnonymizeMany(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.AnonymizeManyRequest{
				AccountId:           accountId,
				InputData:           []string{},
				HaltOnFailure:       false,
				DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{},
				TransformerMappings: []*mgmtv1alpha1.TransformerMapping{},
			}),
		)
		requireErrResp(t, resp, err)
		requireConnectError(t, err, connect.CodeUnimplemented)
	})

	t.Run("cloud-personal-fail", func(t *testing.T) {
		userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		anonclient := s.NeosyncCloudAuthenticatedLicensedClients.Anonymize(integrationtests_test.WithUserId(testAuthUserId))
		s.setUser(s.ctx, userclient)
		accountId := s.createPersonalAccount(s.ctx, userclient)
		resp, err := anonclient.AnonymizeMany(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.AnonymizeManyRequest{
				AccountId:           accountId,
				InputData:           []string{},
				HaltOnFailure:       false,
				DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{},
				TransformerMappings: []*mgmtv1alpha1.TransformerMapping{},
			}),
		)
		requireErrResp(t, resp, err)
		requireConnectError(t, err, connect.CodePermissionDenied)
	})

	t.Run("cloud-team-ok", func(t *testing.T) {
		jsonStrs := []string{
			`{
  "user": {
      "name": "John Doe",
      "age": 300,
      "email": "john@example.com"
  },
  "details": {
      "address": "123 Main St",
      "phone": "555-1234",
      "favorites": ["dog", "cat", "bird"],
      "name": "jake"
  },
  "active": true,
  "sports": ["soccer", "golf", "tennis"]
}`,
			`{
  "user": {
      "name": "Jane Doe",
      "age": 420,
      "email": "jane@example.com"
  },
  "details": {
      "address": "123 Other St",
      "phone": "555-1234",
      "favorites": ["lizard", "cat", "bird"],
      "name": "jan"
  },
  "active": false,
  "sports": ["basketball", "golf", "swimming"]
}`,
		}

		userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		anonclient := s.NeosyncCloudAuthenticatedLicensedClients.Anonymize(integrationtests_test.WithUserId(testAuthUserId))

		s.setUser(s.ctx, userclient)
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "team1", "foo")
		s.Mocks.Billingclient.On("GetSubscriptions", "foo").Once().Return(&testSubscriptionIter{subscriptions: []*stripe.Subscription{
			{Status: stripe.SubscriptionStatusIncompleteExpired},
			{Status: stripe.SubscriptionStatusActive},
		}}, nil)
		resp, err := anonclient.AnonymizeMany(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.AnonymizeManyRequest{
				AccountId:     accountId,
				InputData:     jsonStrs,
				HaltOnFailure: false,
				DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
					N: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
							GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
								Min: gotypeutil.ToPtr(int64(18)),
								Max: gotypeutil.ToPtr(int64(30)),
							},
						},
					},
				},
				TransformerMappings: []*mgmtv1alpha1.TransformerMapping{
					{
						Expression: ".details.name",
						Transformer: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
								TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
							},
						},
					},
					{
						Expression: ".sports[]",
						Transformer: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
								GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
							},
						},
					},
				},
			}),
		)
		requireNoErrResp(s.T(), resp, err)
		require.NotEmpty(s.T(), resp.Msg.OutputData)

		var inputObjects []map[string]any
		err = json.Unmarshal([]byte(fmt.Sprintf("[%s]", strings.Join(jsonStrs, ","))), &inputObjects)
		require.NoError(s.T(), err)

		for i, output := range resp.Msg.OutputData {
			var result map[string]any
			err = json.Unmarshal([]byte(output), &result)
			require.NoError(s.T(), err)
			require.NotEqual(s.T(), inputObjects[i]["details"].(map[string]any)["name"], result["details"].(map[string]any)["name"])
			require.NotEqual(s.T(), inputObjects[i]["user"].(map[string]any)["age"], result["user"].(map[string]any)["age"])
			for j, sport := range result["sports"].([]any) {
				require.NotEqual(s.T(), inputObjects[i]["sports"].([]any)[j], sport)
			}
		}
	})
}

func (s *IntegrationTestSuite) Test_AnonymizeService_AnonymizeSingle() {
	t := s.T()

	t.Run("user-defined-transformer", func(t *testing.T) {
		jsonStr := `{
  "sports": ["basketball", "golf", "swimming"]
}`

		accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())
		categories := "A"
		createTransformerResp, err := s.OSSUnauthenticatedLicensedClients.Transformers().CreateUserDefinedTransformer(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.CreateUserDefinedTransformerRequest{
				AccountId: accountId,
				Source:    mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_CATEGORICAL,
				TransformerConfig: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateCategoricalConfig{
						GenerateCategoricalConfig: &mgmtv1alpha1.GenerateCategorical{
							Categories: &categories,
						},
					},
				},
			}),
		)
		requireNoErrResp(t, createTransformerResp, err)
		transformerId := createTransformerResp.Msg.GetTransformer().GetId()
		resp, err := s.OSSUnauthenticatedLicensedClients.Anonymize().AnonymizeSingle(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
				AccountId: accountId,
				InputData: jsonStr,
				TransformerMappings: []*mgmtv1alpha1.TransformerMapping{
					{
						Expression: ".sports[]",
						Transformer: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig{
								UserDefinedTransformerConfig: &mgmtv1alpha1.UserDefinedTransformerConfig{
									Id: transformerId,
								},
							},
						},
					},
				},
			}),
		)
		requireNoErrResp(t, resp, err)
		require.NotEmpty(t, resp.Msg.OutputData)

		var inputObject map[string]any
		err = json.Unmarshal([]byte(jsonStr), &inputObject)
		require.NoError(t, err)

		output := resp.Msg.OutputData
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)
		for _, sport := range result["sports"].([]any) {
			require.Equal(t, "A", sport)
		}
	})

	t.Run("javascript-transformers", func(t *testing.T) {
		jsonStr := `{
  "sports": ["basketball", "golf", "swimming"],
	"name": "bill"
}`

		accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

		resp, err := s.OSSUnauthenticatedLicensedClients.Anonymize().AnonymizeSingle(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
				AccountId: accountId,
				InputData: jsonStr,
				TransformerMappings: []*mgmtv1alpha1.TransformerMapping{
					{
						Expression: ".sports",
						Transformer: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
								TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{
									Code: "return value.map(v => v + ' updated');",
								},
							},
						},
					},
					{
						Expression: ".name",
						Transformer: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_GenerateJavascriptConfig{
								GenerateJavascriptConfig: &mgmtv1alpha1.GenerateJavascript{
									Code: "return 'jim';",
								},
							},
						},
					},
				},
			}),
		)
		requireNoErrResp(t, resp, err)
		require.NotEmpty(t, resp.Msg.OutputData)

		var inputObject map[string]any
		err = json.Unmarshal([]byte(jsonStr), &inputObject)
		require.NoError(t, err)

		output := resp.Msg.OutputData
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)
		require.Equal(t, "jim", result["name"])
		require.Equal(t, []any{"basketball updated", "golf updated", "swimming updated"}, result["sports"])
	})

	t.Run("ok", func(t *testing.T) {
		jsonStr :=
			`{
  "user": {
      "name": "Jane Doe",
      "age": 42,
      "email": "jane@example.com"
  },
  "details": {
      "address": "123 Other St",
      "phone": "555-1234",
      "favorites": ["lizard", "cat", "bird"],
      "name": "jan"
  },
  "active": false,
  "sports": ["basketball", "golf", "swimming"]
}`

		accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())
		resp, err := s.OSSUnauthenticatedLicensedClients.Anonymize().AnonymizeSingle(
			s.ctx,
			connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
				AccountId: accountId,
				InputData: jsonStr,
				DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
					N: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
							GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
								Min: gotypeutil.ToPtr(int64(18)),
								Max: gotypeutil.ToPtr(int64(30)),
							},
						},
					},
				},
				TransformerMappings: []*mgmtv1alpha1.TransformerMapping{
					{
						Expression: ".details.name",
						Transformer: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
								TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
							},
						},
					},
					{
						Expression: ".sports[]",
						Transformer: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
								GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
							},
						},
					},
				},
			}),
		)
		requireNoErrResp(s.T(), resp, err)
		require.NotEmpty(s.T(), resp.Msg.OutputData)

		var inputObject map[string]any
		err = json.Unmarshal([]byte(jsonStr), &inputObject)
		require.NoError(s.T(), err)

		output := resp.Msg.OutputData
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(s.T(), err)
		require.NotEqual(s.T(), inputObject["details"].(map[string]any)["name"], result["details"].(map[string]any)["name"])
		require.NotEqual(s.T(), inputObject["user"].(map[string]any)["age"], result["user"].(map[string]any)["age"])
		for j, sport := range result["sports"].([]any) {
			require.NotEqual(s.T(), inputObject["sports"].([]any)[j], sport)
		}
	})
}

func (s *IntegrationTestSuite) Test_AnonymizeService_AnonymizeSingle_InvalidTransformerConfig() {
	t := s.T()

	t.Run("no-nested-transformpiitext", func(t *testing.T) {
		userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		anonclient := s.NeosyncCloudAuthenticatedLicensedClients.Anonymize(integrationtests_test.WithUserId(testAuthUserId))

		s.setUser(s.ctx, userclient)
		accountId := s.createBilledTeamAccount(s.ctx, userclient, "team34", "foo34")

		t.Run("default-boolean", func(t *testing.T) {
			resp, err := anonclient.AnonymizeSingle(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
					AccountId: accountId,
					InputData: "foo",
					DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
						Boolean: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{
								TransformPiiTextConfig: &mgmtv1alpha1.TransformPiiText{
									DefaultAnonymizer: &mgmtv1alpha1.PiiAnonymizer{
										Config: &mgmtv1alpha1.PiiAnonymizer_Transform_{Transform: &mgmtv1alpha1.PiiAnonymizer_Transform{
											Config: &mgmtv1alpha1.TransformerConfig{
												Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
											},
										}},
									},
								},
							},
						},
					},
				}),
			)
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeInvalidArgument)
		})

		t.Run("transformer-mappings", func(t *testing.T) {
			resp, err := anonclient.AnonymizeSingle(
				s.ctx,
				connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
					AccountId: accountId,
					InputData: "foo",
					DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
						Boolean: &mgmtv1alpha1.TransformerConfig{
							Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{
								TransformPiiTextConfig: &mgmtv1alpha1.TransformPiiText{
									DefaultAnonymizer: &mgmtv1alpha1.PiiAnonymizer{
										Config: &mgmtv1alpha1.PiiAnonymizer_Transform_{
											Transform: &mgmtv1alpha1.PiiAnonymizer_Transform{
												Config: &mgmtv1alpha1.TransformerConfig{
													Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
												},
											},
										},
									},
								},
							},
						},
					},
					TransformerMappings: []*mgmtv1alpha1.TransformerMapping{
						{
							Expression: ".details.name",
							Transformer: &mgmtv1alpha1.TransformerConfig{
								Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{
									TransformPiiTextConfig: &mgmtv1alpha1.TransformPiiText{
										EntityAnonymizers: map[string]*mgmtv1alpha1.PiiAnonymizer{
											"PERSON": {
												Config: &mgmtv1alpha1.PiiAnonymizer_Transform_{
													Transform: &mgmtv1alpha1.PiiAnonymizer_Transform{
														Config: &mgmtv1alpha1.TransformerConfig{
															Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}),
			)
			requireErrResp(t, resp, err)
			requireConnectError(t, err, connect.CodeInvalidArgument)
		})
	})
}

func (s *IntegrationTestSuite) Test_AnonymizeService_AnonymizeSingle_ForbiddenTransformers() {
	t := s.T()

	t.Run("OSS", func(t *testing.T) {
		accountId := s.createPersonalAccount(s.ctx, s.OSSUnauthenticatedLicensedClients.Users())

		t.Run("transformpiitext", func(t *testing.T) {
			t.Run("mappings", func(t *testing.T) {
				resp, err := s.OSSUnauthenticatedLicensedClients.Anonymize().AnonymizeSingle(
					s.ctx,
					connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
						AccountId: accountId,
						InputData: "foo",
						TransformerMappings: []*mgmtv1alpha1.TransformerMapping{
							{
								Transformer: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
								},
							},
						},
					}),
				)
				requireErrResp(t, resp, err)
				requireConnectError(t, err, connect.CodePermissionDenied)
			})

			t.Run("defaults", func(t *testing.T) {
				t.Run("Bool", func(t *testing.T) {
					resp, err := s.OSSUnauthenticatedLicensedClients.Anonymize().AnonymizeSingle(
						s.ctx,
						connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
							AccountId: accountId,
							InputData: "foo",
							DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
								Boolean: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
								},
							},
						}),
					)
					requireErrResp(t, resp, err)
					requireConnectError(t, err, connect.CodePermissionDenied)
				})
				t.Run("S", func(t *testing.T) {
					resp, err := s.OSSUnauthenticatedLicensedClients.Anonymize().AnonymizeSingle(
						s.ctx,
						connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
							AccountId: accountId,
							InputData: "foo",
							DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
								S: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
								},
							},
						}),
					)
					requireErrResp(t, resp, err)
					requireConnectError(t, err, connect.CodePermissionDenied)
				})
				t.Run("N", func(t *testing.T) {
					resp, err := s.OSSUnauthenticatedLicensedClients.Anonymize().AnonymizeSingle(
						s.ctx,
						connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
							AccountId: accountId,
							InputData: "foo",
							DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
								N: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
								},
							},
						}),
					)
					requireErrResp(t, resp, err)
					requireConnectError(t, err, connect.CodePermissionDenied)
				})
			})
		})
	})

	t.Run("cloud-personal", func(t *testing.T) {
		userclient := s.NeosyncCloudAuthenticatedLicensedClients.Users(integrationtests_test.WithUserId(testAuthUserId))
		anonclient := s.NeosyncCloudAuthenticatedLicensedClients.Anonymize(integrationtests_test.WithUserId(testAuthUserId))

		s.setUser(s.ctx, userclient)
		accountId := s.createPersonalAccount(s.ctx, userclient)

		t.Run("transformpiitext", func(t *testing.T) {
			t.Run("mappings", func(t *testing.T) {
				resp, err := anonclient.AnonymizeSingle(
					s.ctx,
					connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
						AccountId: accountId,
						InputData: "foo",
						TransformerMappings: []*mgmtv1alpha1.TransformerMapping{
							{
								Transformer: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
								},
							},
						},
					}),
				)
				requireErrResp(t, resp, err)
				requireConnectError(t, err, connect.CodePermissionDenied)
			})

			t.Run("defaults", func(t *testing.T) {
				t.Run("Bool", func(t *testing.T) {
					resp, err := anonclient.AnonymizeSingle(
						s.ctx,
						connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
							AccountId: accountId,
							InputData: "foo",
							DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
								Boolean: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
								},
							},
						}),
					)
					requireErrResp(t, resp, err)
					requireConnectError(t, err, connect.CodePermissionDenied)
				})
				t.Run("S", func(t *testing.T) {
					resp, err := anonclient.AnonymizeSingle(
						s.ctx,
						connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
							AccountId: accountId,
							InputData: "foo",
							DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
								S: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
								},
							},
						}),
					)
					requireErrResp(t, resp, err)
					requireConnectError(t, err, connect.CodePermissionDenied)
				})
				t.Run("N", func(t *testing.T) {
					resp, err := anonclient.AnonymizeSingle(
						s.ctx,
						connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
							AccountId: accountId,
							InputData: "foo",
							DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
								N: &mgmtv1alpha1.TransformerConfig{
									Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{},
								},
							},
						}),
					)
					requireErrResp(t, resp, err)
					requireConnectError(t, err, connect.CodePermissionDenied)
				})
			})
		})
	})
}
