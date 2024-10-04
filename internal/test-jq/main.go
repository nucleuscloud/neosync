package main

import (
	"encoding/json"
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	jsonanonymizer "github.com/nucleuscloud/neosync/internal/json-anonymizer"
)

func main() {
	transformerMappings := map[string]*mgmtv1alpha1.TransformerConfig{
		// `(.. | objects | select(has("name")) | .name)`: {
		// 	Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
		// 		TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
		// 	},
		// },
		// `.details.name`: {
		// 	Config: &mgmtv1alpha1.TransformerConfig_TransformFirstNameConfig{
		// 		TransformFirstNameConfig: &mgmtv1alpha1.TransformFirstName{},
		// 	},
		// },
		// ".user.age": {
		// 	Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
		// 		GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
		// 			Min: 18,
		// 			Max: 30,
		// 		},
		// 	},
		// },
		// ".sports[0]": {
		// 	Config: &mgmtv1alpha1.TransformerConfig_GenerateCityConfig{
		// 		GenerateCityConfig: &mgmtv1alpha1.GenerateCity{},
		// 	},
		// },
	}

	// Define default transformers
	defaultTransformers := &mgmtv1alpha1.DefaultTransformersConfig{
		N: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateInt64Config{
				GenerateInt64Config: &mgmtv1alpha1.GenerateInt64{
					Min: 18,
					Max: 30,
				},
			},
		},
		S: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformCharacterScrambleConfig{
				TransformCharacterScrambleConfig: &mgmtv1alpha1.TransformCharacterScramble{},
			},
		},
	}

	// Sample JSON input
	// 	jsonStr := `{
	//     "user": {
	//         "name": "John Doe",
	//         "age": 30,
	//         "email": "john@example.com"
	//     },
	//     "details": {
	//         "address": "123 Main St",
	//         "phone": "555-1234",
	//         "favorites": ["dog", "cat", "bird"]
	//     },
	//     "active": true,
	//     "sports": ["soccer", "golf", "tennis"]
	// }`

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
      "name": 1
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
      "name": "1"
  },
  "active": false,
  "sports": ["basketball", "golf", "swimming"]
}`,
	}

	fmt.Println()
	fmt.Println("## INPUT")
	fmt.Println(jsonStrs)
	fmt.Println()

	anon, err := jsonanonymizer.NewAnonymizer(
		jsonanonymizer.WithTransformerMappings(transformerMappings),
		jsonanonymizer.WithDefaultTransformers(defaultTransformers),
		jsonanonymizer.WithHaltOnFailure(false),
	)
	if err != nil {
		panic(err)
	}

	// Anonymize the JSON input
	result, anonErrors := anon.AnonymizeJSONObjects(jsonStrs)
	fmt.Println()
	fmt.Println("## OUTPUT")
	for _, r := range result {
		var data any
		_ = json.Unmarshal([]byte(r), &data)
		jsonF, _ := json.MarshalIndent(data, "", " ")
		fmt.Printf("%s \n", string(jsonF))

	}
	fmt.Println()

	fmt.Println("## ERRORS")
	jsonF, _ := json.MarshalIndent(anonErrors, "", " ")
	fmt.Printf("%s \n", string(jsonF))
	fmt.Println()
}

// test javascript transformer
