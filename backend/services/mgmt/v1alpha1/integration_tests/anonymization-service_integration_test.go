package integrationtests_test

import (
	"fmt"

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
				// },
				S: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
						TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
					},
				},
			},
			TransformerMappings: map[string]*mgmtv1alpha1.TransformerConfig{
				`.details.name`: {
					Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
						TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
					},
				},
				// ".user.age": {
				// 	Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
				// 		GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
				// 			Min: 18,
				// 			Max: 30,
				// 		},
				// 	},
				// },
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
	for _, s := range resp.Msg.OutputData {
		fmt.Println()
		fmt.Println(s)
		fmt.Println()
	}
	require.Empty(s.T(), resp.Msg.OutputData)
}
