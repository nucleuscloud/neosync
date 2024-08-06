package main

import (
	"log"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	// Corrected JSON string with snake_case field names
	jsonStr := `{
		"config": {
			"dynamodb": {
				"connection_id": "54836a9f-a147-4f03-8070-b811f1078d2b",
				"unmapped_transforms": {
					"b": {
						"source": 1,
						"config": {
							"config": {
								"passthrough_config": {}
							}
						}
					},
					"boolean": {
						"source": 6,
						"config": {
							"config": {
								"generate_bool_config": {}
							}
						}
					},
					"n": {
						"source": 36,
						"config": {
							"config": {
								"transform_int64_config": {
									"randomization_range_min": 20,
									"randomization_range_max": 50
								}
							}
						}
					},
					"s": {
						"source": 25,
						"config": {
							"config": {
								"generate_string_config": {
									"min": 1,
									"max": 100
								}
							}
						}
					}
				}
			}
		}
	}`

	// Define the JobSourceOptions message
	var jobSourceOptions mgmtv1alpha1.JobSourceOptions

	// Unmarshal the JSON string into the JobSourceOptions struct
	err := protojson.Unmarshal([]byte(jsonStr), &jobSourceOptions)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Print the unmarshalled JobSourceOptions
	// fmt.Printf("Unmarshalled JobSourceOptions: %+v\n", jobSourceOptions)
}
