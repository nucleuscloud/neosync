package integrationtests_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *IntegrationTestSuite) Test_AnonymizeService_AnonymizeMany() {
	jsonStrs := []string{
		`{
  "user": {
      "name": "John Doe",
      "age": 30,
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
}`,
	}

	resp, err := s.unauthdClients.anonymize.AnonymizeMany(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.AnonymizeManyRequest{
			InputData:     jsonStrs,
			HaltOnFailure: false,
			DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
				N: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
						GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
							Min: 18,
							Max: 30,
						},
					},
				},
			},
			TransformerMappings: map[string]*mgmtv1alpha1.TransformerConfig{
				`.details.name`: {
					Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
						TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
					},
				},
				".sports[]": {
					Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
						GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
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
}

func (s *IntegrationTestSuite) Test_AnonymizeService_AnonymizeSingle() {
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

	resp, err := s.unauthdClients.anonymize.AnonymizeSingle(
		s.ctx,
		connect.NewRequest(&mgmtv1alpha1.AnonymizeSingleRequest{
			InputData: jsonStr,
			DefaultTransformers: &mgmtv1alpha1.DefaultTransformersConfig{
				N: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
						GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
							Min: 18,
							Max: 30,
						},
					},
				},
			},
			TransformerMappings: map[string]*mgmtv1alpha1.TransformerConfig{
				`.details.name`: {
					Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
						TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
					},
				},
				".sports[]": {
					Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
						GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
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
}
