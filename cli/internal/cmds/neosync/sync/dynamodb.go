package sync_cmd

import (
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
