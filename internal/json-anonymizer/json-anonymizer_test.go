package jsonanonymizer

import (
	"encoding/json"
	"os"
	"testing"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/testutil"
	"github.com/stretchr/testify/require"
)

func Test_NewAnonymizer(t *testing.T) {
	t.Run("Initialize with no options", func(t *testing.T) {
		anonymizer, err := NewAnonymizer()
		require.Error(t, err)
		require.Nil(t, anonymizer)
	})

	t.Run("Initialize with transformer mappings", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.TransformerMapping{
			{
				Expression: ".city",
				Transformer: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
						GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
					},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings))
		require.NoError(t, err)
		require.NotNil(t, anonymizer)
		require.Equal(t, mappings, anonymizer.transformerMappings)
		require.NotEmpty(t, anonymizer.transformerExecutors)
	})

	t.Run("Initialize with default transformers", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
			Boolean: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
					GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithDefaultTransformers(defaults))
		require.NoError(t, err)
		require.NotNil(t, anonymizer)
		require.Equal(t, defaults, anonymizer.defaultTransformers)
		require.NotNil(t, anonymizer.defaultTransformerExecutor)
	})

	t.Run("Initialize with halt on failure", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithDefaultTransformers(defaults), WithHaltOnFailure(true))
		require.NoError(t, err)
		require.NotNil(t, anonymizer)
		require.True(t, anonymizer.haltOnFailure)
	})
}

func Test_AnonymizeJSONObjects(t *testing.T) {
	t.Run("Anonymize with transformer mappings", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.TransformerMapping{
			{
				Expression: ".name",
				Transformer: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
						GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
					},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings))
		require.NoError(t, err)

		input := `{"name": "John Doe", "age": 30}`
		output, anonErrors := anonymizer.AnonymizeJSONObjects([]string{input})
		require.Empty(t, anonErrors)

		var result map[string]any
		err = json.Unmarshal([]byte(output[0]), &result)
		require.NoError(t, err)
		require.NotEqual(t, "John Doe", result["name"])
		require.Equal(t, float64(30), result["age"])
	})

	t.Run("Anonymize with default transformers", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithDefaultTransformers(defaults))
		require.NoError(t, err)

		input := `{"name": "John Doe", "city": "New York"}`
		output, anonErrors := anonymizer.AnonymizeJSONObjects([]string{input})
		require.Empty(t, anonErrors)

		var result map[string]any
		err = json.Unmarshal([]byte(output[0]), &result)
		require.NoError(t, err)
		require.NotEqual(t, "John Doe", result["name"])
		require.NotEqual(t, "New York", result["city"])
	})

	t.Run("Anonymize with invalid JSON", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithDefaultTransformers(defaults))
		require.NoError(t, err)

		input := `invalid json`
		_, anonErrors := anonymizer.AnonymizeJSONObjects([]string{input})
		require.NotEmpty(t, anonErrors)
	})

	t.Run("Anonymize error should halt", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.TransformerMapping{
			{
				Expression: ".name",
				Transformer: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
						TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
					},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings), WithHaltOnFailure(true))
		require.NoError(t, err)

		inputs := []string{`{"id": 1, "name": "John Doe", "city": "New York"}`, `{"id": 2, "name": 1, "city": "New York"}`, `{"id": 3, "name": "John Doe", "city": "New York"}`}
		outputs, anonErrors := anonymizer.AnonymizeJSONObjects(inputs)
		require.Len(t, anonErrors, 1)
		require.Equal(t, int64(1), anonErrors[0].InputIndex)

		require.Len(t, outputs, 1)
		var result map[string]any
		err = json.Unmarshal([]byte(outputs[0]), &result)
		require.NoError(t, err)
		require.NotEqual(t, "John Doe", result["name"])
		require.Equal(t, "New York", result["city"])
	})
	t.Run("Anonymize error should not halt", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.TransformerMapping{
			{
				Expression: ".name",
				Transformer: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
						TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
					},
				},
			},
		}
		anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings), WithHaltOnFailure(false))
		require.NoError(t, err)

		inputs := []string{`{"id": 0, "name": "John Doe", "city": "New York"}`, `{"id": 1, "name": 1, "city": "New York"}`, `{"id": 2, "name": "John Doe", "city": "New York"}`}
		outputs, anonErrors := anonymizer.AnonymizeJSONObjects(inputs)
		require.Len(t, anonErrors, 1)
		require.Equal(t, int64(1), anonErrors[0].InputIndex)

		for idx, o := range outputs {
			if idx == 1 {
				require.Empty(t, o)
				continue
			}
			var result map[string]any
			err = json.Unmarshal([]byte(o), &result)
			require.NoError(t, err)
			require.NotEqual(t, "John Doe", result["name"])
			require.Equal(t, "New York", result["city"])
			require.Equal(t, float64(idx), result["id"])
		}
	})
}

func Test_DerefPointer(t *testing.T) {
	t.Run("Deref string pointer", func(t *testing.T) {
		str := "test"
		ptr := &str
		result := derefPointer(ptr)
		require.Equal(t, str, result)
	})

	t.Run("Deref nil pointer", func(t *testing.T) {
		var ptr *string
		result := derefPointer(ptr)
		require.Nil(t, result)
	})

	t.Run("Deref non-pointer", func(t *testing.T) {
		value := 42
		result := derefPointer(value)
		require.Equal(t, value, result)
	})
}

func Test_InitTransformerExecutors(t *testing.T) {
	t.Run("Initialize valid transformer", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.TransformerMapping{
			{
				Expression: ".field1",
				Transformer: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateFirstNameConfig{
						GenerateFirstNameConfig: &mgmtv1alpha1.GenerateFirstName{},
					},
				},
			},
		}
		executors, err := initTransformerExecutors(mappings, nil, testutil.GetTestLogger(t))
		require.NoError(t, err)
		require.Len(t, executors, 1)
	})

	t.Run("Initialize invalid transformer", func(t *testing.T) {
		mappings := []*mgmtv1alpha1.TransformerMapping{
			{
				Expression:  ".field1",
				Transformer: nil,
			},
		}
		_, err := initTransformerExecutors(mappings, nil, testutil.GetTestLogger(t))
		require.Error(t, err)
	})
}

func Test_InitDefaultTransformerExecutors(t *testing.T) {
	t.Run("Initialize all default transformers", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
				},
			},
			N: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
					GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{},
				},
			},
			Boolean: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{},
			},
		}
		executors, err := initDefaultTransformerExecutors(defaults, nil, testutil.GetTestLogger(t))
		require.NoError(t, err)
		require.NotNil(t, executors.S)
		require.NotNil(t, executors.N)
		require.NotNil(t, executors.Boolean)
	})

	t.Run("Initialize partial default transformers", func(t *testing.T) {
		defaults := &mgmtv1alpha1.DefaultTransformersConfig{
			S: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
					TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
				},
			},
		}
		executors, err := initDefaultTransformerExecutors(defaults, nil, testutil.GetTestLogger(t))
		require.NoError(t, err)
		require.NotNil(t, executors.S)
		require.Nil(t, executors.N)
		require.Nil(t, executors.Boolean)
	})
}

func Test_AnonymizeJSON_Largedata(t *testing.T) {
	inputStrings, inputObjects, err := getTestData("./testdata/company.json")
	require.NoError(t, err)

	preserveLength := true
	// Define transformer mappings
	mappings := []*mgmtv1alpha1.TransformerMapping{
		{
			Expression: ".companyName",
			Transformer: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformStringConfig{
					TransformStringConfig: &mgmtv1alpha1.TransformString{PreserveLength: &preserveLength},
				},
			},
		},
		{
			Expression: ".leadership.CEO.name",
			Transformer: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
					TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{},
				},
			},
		},
		{
			Expression: ".departments[].projects[]?.teamMembers[]?.name",
			Transformer: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_GenerateFullNameConfig{
					GenerateFullNameConfig: &mgmtv1alpha1.GenerateFullName{},
				},
			},
		},
	}

	anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings))
	require.NoError(t, err)

	outputs, anonErrors := anonymizer.AnonymizeJSONObjects(inputStrings)
	require.Empty(t, anonErrors)

	for i, output := range outputs {
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Check if company name was anonymized
		require.NotEqual(t, inputObjects[i]["companyName"], result["companyName"])

		// Check if CEO name was anonymized
		originalCEO := inputObjects[i]["leadership"].(map[string]any)["CEO"].(map[string]any)["name"]
		resultCEO := result["leadership"].(map[string]any)["CEO"].(map[string]any)["name"]
		require.NotEqual(t, originalCEO, resultCEO)

		// Check if team member names were anonymized
		for j, dept := range result["departments"].([]any) {
			projects, ok := dept.(map[string]any)["projects"].([]any)
			if !ok {
				continue
			}
			for k, project := range projects {
				for l, member := range project.(map[string]any)["teamMembers"].([]any) {
					originalName := inputObjects[i]["departments"].([]any)[j].(map[string]any)["projects"].([]any)[k].(map[string]any)["teamMembers"].([]any)[l].(map[string]any)["name"]
					resultName := member.(map[string]any)["name"]
					require.NotEqual(t, originalName, resultName)
				}
			}
		}

		// Check if non-anonymized fields remain unchanged
		require.Equal(t, inputObjects[i]["foundedYear"], result["foundedYear"])
		require.Equal(t, inputObjects[i]["headquarters"].(map[string]any)["address"].(map[string]any)["city"], result["headquarters"].(map[string]any)["address"].(map[string]any)["city"])
	}
}

func Test_AnonymizeJSON_Largedata_WithDefaults(t *testing.T) {
	inputStrings, inputObjects, err := getTestData("./testdata/company.json")
	require.NoError(t, err)

	// Define transformer mappings
	mappings := []*mgmtv1alpha1.TransformerMapping{
		{
			Expression: ".companyName",
			Transformer: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
					PassthroughConfig: &mgmtv1alpha1.Passthrough{},
				},
			},
		},
		{
			Expression: ".leadership.CEO.name",
			Transformer: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
					PassthroughConfig: &mgmtv1alpha1.Passthrough{},
				},
			},
		},
		{
			Expression: ".departments[].projects[]?.teamMembers[]?.name",
			Transformer: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
					PassthroughConfig: &mgmtv1alpha1.Passthrough{},
				},
			},
		},
	}

	// Define transformer defaults
	defaults := &mgmtv1alpha1.DefaultTransformersConfig{
		S: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
				TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
			},
		},
		N: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
				GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{},
			},
		},
		Boolean: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateBoolConfig{
				GenerateBoolConfig: &mgmtv1alpha1.GenerateBool{},
			},
		},
	}

	anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings), WithDefaultTransformers(defaults))
	require.NoError(t, err)

	outputs, anonErrors := anonymizer.AnonymizeJSONObjects(inputStrings)
	require.Empty(t, anonErrors)

	for i, output := range outputs {
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Check if company name was passed through
		require.Equal(t, inputObjects[i]["companyName"], result["companyName"])

		// Check if CEO name was passed through
		originalCEO := inputObjects[i]["leadership"].(map[string]any)["CEO"].(map[string]any)["name"]
		resultCEO := result["leadership"].(map[string]any)["CEO"].(map[string]any)["name"]
		require.Equal(t, originalCEO, resultCEO)

		// Check if team member names were passed through
		for j, dept := range result["departments"].([]any) {
			projects, ok := dept.(map[string]any)["projects"].([]any)
			if !ok {
				continue
			}
			for k, project := range projects {
				for l, member := range project.(map[string]any)["teamMembers"].([]any) {
					originalName := inputObjects[i]["departments"].([]any)[j].(map[string]any)["projects"].([]any)[k].(map[string]any)["teamMembers"].([]any)[l].(map[string]any)["name"]
					resultName := member.(map[string]any)["name"]
					require.Equal(t, originalName, resultName)
				}
			}
		}

		// Check if other fields where anonymized
		require.NotEqual(t, inputObjects[i]["foundedYear"], result["foundedYear"])
		require.NotEqual(t, inputObjects[i]["headquarters"].(map[string]any)["address"].(map[string]any)["city"], result["headquarters"].(map[string]any)["address"].(map[string]any)["city"])
	}
}

func Test_AnonymizeJSON_Largedata_Advanced(t *testing.T) {
	inputStrings, inputObjects, err := getTestData("./testdata/company.json")
	require.NoError(t, err)

	// Transform all name fields in objects
	mappings := []*mgmtv1alpha1.TransformerMapping{
		{
			Expression: `(.. | objects | select(has("name")) | .name)`,
			Transformer: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_TransformFullNameConfig{
					TransformFullNameConfig: &mgmtv1alpha1.TransformFullName{},
				},
			},
		},
	}

	defaults := &mgmtv1alpha1.DefaultTransformersConfig{}

	anonymizer, err := NewAnonymizer(WithTransformerMappings(mappings), WithDefaultTransformers(defaults))
	require.NoError(t, err)

	outputs, anonErrors := anonymizer.AnonymizeJSONObjects(inputStrings)
	require.Empty(t, anonErrors)

	for i, output := range outputs {
		var result map[string]any
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err)

		// Check if company name was not anonymized
		require.Equal(t, inputObjects[i]["companyName"], result["companyName"])

		// Check if CEO name was anonymized
		originalCEO := inputObjects[i]["leadership"].(map[string]any)["CEO"].(map[string]any)["name"]
		resultCEO := result["leadership"].(map[string]any)["CEO"].(map[string]any)["name"]
		require.NotEqual(t, originalCEO, resultCEO)

		// Check if team member names were anonymized
		for j, dept := range result["departments"].([]any) {
			projects, ok := dept.(map[string]any)["projects"].([]any)
			if !ok {
				continue
			}
			for k, project := range projects {
				for l, member := range project.(map[string]any)["teamMembers"].([]any) {
					originalName := inputObjects[i]["departments"].([]any)[j].(map[string]any)["projects"].([]any)[k].(map[string]any)["teamMembers"].([]any)[l].(map[string]any)["name"]
					resultName := member.(map[string]any)["name"]
					require.NotEqual(t, originalName, resultName)
				}
			}
		}

		// Check if non-anonymized fields remain unchanged
		require.Equal(t, inputObjects[i]["foundedYear"], result["foundedYear"])
		require.Equal(t, inputObjects[i]["headquarters"].(map[string]any)["address"].(map[string]any)["city"], result["headquarters"].(map[string]any)["address"].(map[string]any)["city"])
	}
}

func getTestData(filePath string) (jsonStrings []string, jsonObjects []map[string]any, err error) {
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	var inputObjects []map[string]any
	err = json.Unmarshal(jsonData, &inputObjects)
	if err != nil {
		return nil, nil, err
	}

	var inputStrings []string
	for _, obj := range inputObjects {
		jsonStr, err := json.Marshal(obj)
		if err != nil {
			return nil, nil, err
		}
		inputStrings = append(inputStrings, string(jsonStr))
	}
	return inputStrings, inputObjects, nil
}
