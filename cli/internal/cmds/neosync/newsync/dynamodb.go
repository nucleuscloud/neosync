package newsync_cmd

import (
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/spf13/cobra"
)

func buildAwsCredConfig(cmd *cobra.Command, config *cmdConfig) (*cmdConfig, error) {
	region, err := cmd.Flags().GetString("aws-region")
	if err != nil {
		return nil, err
	}
	if region != "" {
		config.AwsDynamoDbDestination.AwsCredConfig.Region = region
	}

	dynamoDBAccessKeyID, err := cmd.Flags().GetString("aws-access-key-id")
	if err != nil {
		return nil, err
	}
	if dynamoDBAccessKeyID != "" {
		config.AwsDynamoDbDestination.AwsCredConfig.AccessKeyID = &dynamoDBAccessKeyID
	}
	dynamoDBSecretAccessKey, err := cmd.Flags().GetString("aws-secret-access-key")
	if err != nil {
		return nil, err
	}
	if dynamoDBSecretAccessKey != "" {
		config.AwsDynamoDbDestination.AwsCredConfig.SecretAccessKey = &dynamoDBSecretAccessKey
	}
	dynamoDBSessionToken, err := cmd.Flags().GetString("aws-session-token")
	if err != nil {
		return nil, err
	}
	if dynamoDBSessionToken != "" {
		config.AwsDynamoDbDestination.AwsCredConfig.SessionToken = &dynamoDBSessionToken
	}
	dynamoDBRoleARN, err := cmd.Flags().GetString("aws-role-arn")
	if err != nil {
		return nil, err
	}
	if dynamoDBRoleARN != "" {
		config.AwsDynamoDbDestination.AwsCredConfig.RoleARN = &dynamoDBRoleARN
	}
	dynamoDBRoleExternalID, err := cmd.Flags().GetString("aws-role-external-id")
	if err != nil {
		return nil, err
	}
	if dynamoDBRoleExternalID != "" {
		config.AwsDynamoDbDestination.AwsCredConfig.RoleExternalID = &dynamoDBRoleExternalID
	}
	dynamoDBEndpoint, err := cmd.Flags().GetString("aws-endpoint")
	if err != nil {
		return nil, err
	}
	if dynamoDBEndpoint != "" {
		config.AwsDynamoDbDestination.AwsCredConfig.Endpoint = &dynamoDBEndpoint
	}
	return config, nil
}

func generateDynamoDbBenthosConfig(
	cmd *cmdConfig,
	table string,
) *benthosConfigResponse {
	bc := &neosync_benthos.BenthosConfig{
		StreamConfig: neosync_benthos.StreamConfig{
			Logger: &neosync_benthos.LoggerConfig{
				Level:        "ERROR",
				AddTimestamp: true,
			},
			Input: &neosync_benthos.InputConfig{
				Inputs: neosync_benthos.Inputs{
					NeosyncConnectionData: &neosync_benthos.NeosyncConnectionData{
						// ApiKey:         authToken,
						// ApiUrl:         apiUrl,
						ConnectionId:   cmd.Source.ConnectionId,
						ConnectionType: string(awsDynamoDBConnection),
						Schema:         "dynamodb",
						Table:          table,
					},
				},
			},
			Pipeline: &neosync_benthos.PipelineConfig{},
			Output: &neosync_benthos.OutputConfig{
				Outputs: neosync_benthos.Outputs{
					AwsDynamoDB: &neosync_benthos.OutputAwsDynamoDB{
						Table: table,
						JsonMapColumns: map[string]string{
							"": ".",
						},

						Batching: &neosync_benthos.Batching{
							// https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_BatchWriteItem.html
							// A single call to BatchWriteItem can transmit up to 16MB of data over the network, consisting of up to 25 item put or delete operations
							// Specifying the count here may not be enough if the overall data is above 16MB.
							// Benthos will fall back on error to single writes however
							Period: "5s",
							Count:  25,
						},

						Region:      cmd.AwsDynamoDbDestination.AwsCredConfig.Region,
						Endpoint:    *cmd.AwsDynamoDbDestination.AwsCredConfig.Endpoint,
						Credentials: buildBenthosAwsCredentials(cmd),
					},
				},
			},
		},
	}
	return &benthosConfigResponse{
		Name:      table,
		Config:    bc,
		DependsOn: []*tabledependency.DependsOn{},
		Table:     table,
		Columns:   []string{},
	}
}

func buildBenthosAwsCredentials(cmd *cmdConfig) *neosync_benthos.AwsCredentials {
	if cmd.AwsDynamoDbDestination == nil || cmd.AwsDynamoDbDestination.AwsCredConfig == nil {
		return nil
	}
	cc := cmd.AwsDynamoDbDestination.AwsCredConfig
	creds := &neosync_benthos.AwsCredentials{}
	if cc.Profile != nil {
		creds.Profile = *cc.Profile
	}
	if cc.AccessKeyID != nil {
		creds.Id = *cc.AccessKeyID
	}
	if cc.SecretAccessKey != nil {
		creds.Secret = *cc.SecretAccessKey
	}
	if cc.SessionToken != nil {
		creds.Token = *cc.SessionToken
	}
	if cc.RoleARN != nil {
		creds.Role = *cc.RoleARN
	}
	if cc.RoleExternalID != nil {
		creds.RoleExternalId = *cc.RoleExternalID
	}
	return creds
}
