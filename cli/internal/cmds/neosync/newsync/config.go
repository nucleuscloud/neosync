package newsync_cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/cli/internal/output"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func buildCmdConfig(cmd *cobra.Command) (*cmdConfig, error) {
	config := &cmdConfig{
		Source: &sourceConfig{
			ConnectionOpts: &connectionOpts{},
		},
		Destination:            &sqlDestinationConfig{},
		AwsDynamoDbDestination: &dynamoDbDestinationConfig{},
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	if configPath != "" {
		fileBytes, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		err = yaml.Unmarshal(fileBytes, &config)
		if err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
	}

	connectionId, err := cmd.Flags().GetString("connection-id")
	if err != nil {
		return nil, err
	}
	if connectionId != "" {
		config.Source.ConnectionId = connectionId
	}

	destConnUrl, err := cmd.Flags().GetString("destination-connection-url")
	if err != nil {
		return nil, err
	}
	if destConnUrl != "" {
		config.Destination.ConnectionUrl = destConnUrl
	}

	driver, err := cmd.Flags().GetString("destination-driver")
	if err != nil {
		return nil, err
	}
	pDriver, ok := parseDriverString(driver)
	if ok {
		config.Destination.Driver = pDriver
	}

	initSchema, err := cmd.Flags().GetBool("init-schema")
	if err != nil {
		return nil, err
	}
	if initSchema {
		config.Destination.InitSchema = initSchema
	}

	truncateBeforeInsert, err := cmd.Flags().GetBool("truncate-before-insert")
	if err != nil {
		return nil, err
	}
	if truncateBeforeInsert {
		config.Destination.TruncateBeforeInsert = truncateBeforeInsert
	}

	truncateCascade, err := cmd.Flags().GetBool("truncate-cascade")
	if err != nil {
		return nil, err
	}
	if truncateCascade {
		config.Destination.TruncateCascade = truncateCascade
	}

	onConflictDoNothing, err := cmd.Flags().GetBool("on-conflict-do-nothing")
	if err != nil {
		return nil, err
	}
	if onConflictDoNothing {
		config.Destination.OnConflict.DoNothing = onConflictDoNothing
	}

	jobId, err := cmd.Flags().GetString("job-id")
	if err != nil {
		return nil, err
	}
	if jobId != "" {
		config.Source.ConnectionOpts.JobId = &jobId
	}

	jobRunId, err := cmd.Flags().GetString("job-run-id")
	if err != nil {
		return nil, err
	}
	if jobRunId != "" {
		config.Source.ConnectionOpts.JobRunId = &jobRunId
	}

	config, err = buildAwsCredConfig(cmd, config)
	if err != nil {
		return nil, err
	}

	if config.Source.ConnectionId == "" {
		return nil, fmt.Errorf("must provide connection-id")
	}

	accountIdFlag, err := cmd.Flags().GetString("account-id")
	if err != nil {
		return nil, err
	}
	accountId := accountIdFlag
	if accountId == "" {
		aId, err := userconfig.GetAccountId()
		if err != nil {
			return nil, errors.New("Unable to retrieve account id. Please use account switch command to set account.")
		}
		accountId = aId
	}
	config.AccountId = &accountId

	if accountId == "" {
		return nil, errors.New("Account Id not found. Please use account switch command to set account.")
	}

	outputType, err := output.ValidateAndRetrieveOutputFlag(cmd)
	if err != nil {
		return nil, err
	}
	config.OutputType = &outputType

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return nil, err
	}
	config.Debug = debug
	return config, nil
}

func isConfigValid(cmd *cmdConfig, logger *slog.Logger, sourceConnection *mgmtv1alpha1.Connection, sourceConnectionType ConnectionType) error {
	if sourceConnectionType == awsS3Connection && (cmd.Source.ConnectionOpts.JobId == nil || *cmd.Source.ConnectionOpts.JobId == "") && (cmd.Source.ConnectionOpts.JobRunId == nil || *cmd.Source.ConnectionOpts.JobRunId == "") {
		return errors.New("S3 source connection type requires job-id or job-run-id.")
	}
	if sourceConnectionType == gcpCloudStorageConnection && (cmd.Source.ConnectionOpts.JobId == nil || *cmd.Source.ConnectionOpts.JobId == "") && (cmd.Source.ConnectionOpts.JobRunId == nil || *cmd.Source.ConnectionOpts.JobRunId == "") {
		return errors.New("GCP Cloud Storage source connection type requires job-id or job-run-id")
	}

	if (sourceConnectionType == awsS3Connection || sourceConnectionType == gcpCloudStorageConnection) && cmd.Destination.InitSchema {
		return errors.New("init schema is only supported when source is a SQL Database")
	}

	if cmd.Destination.TruncateCascade && cmd.Destination.Driver == mysqlDriver {
		return fmt.Errorf("truncate cascade is only supported in postgres")
	}

	if sourceConnectionType == mysqlConnection || sourceConnectionType == postgresConnection {
		if cmd.Destination.Driver == "" {
			return fmt.Errorf("must provide destination-driver")
		}
		if cmd.Destination.ConnectionUrl == "" {
			return fmt.Errorf("must provide destination-connection-url")
		}

		if cmd.Destination.Driver != mysqlDriver && cmd.Destination.Driver != postgresDriver {
			return errors.New("unsupported destination driver. only pgx (postgres) and mysql are currently supported")
		}
	}

	if sourceConnectionType == awsDynamoDBConnection {
		if cmd.AwsDynamoDbDestination == nil {
			return fmt.Errorf("must provide destination aws credentials")
		}

		if cmd.AwsDynamoDbDestination.AwsCredConfig.Region == "" {
			return fmt.Errorf("must provide destination aws region")
		}
	}

	if sourceConnection.AccountId != *cmd.AccountId {
		return fmt.Errorf("Connection not found. AccountId: %s", *cmd.AccountId)
	}

	logger.Debug("Checking if source and destination are compatible")
	err := areSourceAndDestCompatible(sourceConnection, cmd.Destination.Driver)
	if err != nil {
		return err
	}
	return nil
}
