package sync_cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/cli/internal/output"
	benthosbuilder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Builds the sync command config by pulling from the cobra command flags
func newCobraCmdConfig(
	cmd *cobra.Command,
	getAccountId func(accountIdFlag string) (string, error),
) (*cmdConfig, error) {
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
	accountId, err := getAccountId(accountIdFlag)
	if err != nil {
		return nil, err
	}
	config.AccountId = &accountId

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
	if cmd.Flags().Changed("destination-open-limit") {
		openLimit, err := cmd.Flags().GetInt32("destination-open-limit")
		if err != nil {
			return nil, err
		}
		config.Destination.ConnectionOpts.OpenLimit = &openLimit
	}

	if cmd.Flags().Changed("destination-idle-limit") {
		idleLimit, err := cmd.Flags().GetInt32("destination-idle-limit")
		if err != nil {
			return nil, err
		}
		config.Destination.ConnectionOpts.IdleLimit = &idleLimit
	}
	if cmd.Flags().Changed("destination-open-duration") {
		openDuration, err := cmd.Flags().GetString("destination-open-duration")
		if err != nil {
			return nil, err
		}
		if _, err := time.ParseDuration(openDuration); err != nil {
			return nil, fmt.Errorf("unable to parse destination-open-duration as a valid duration string: %w", err)
		}
		config.Destination.ConnectionOpts.OpenDuration = &openDuration
	}
	if cmd.Flags().Changed("destination-idle-duration") {
		idleDuration, err := cmd.Flags().GetString("destination-idle-duration")
		if err != nil {
			return nil, err
		}
		if _, err := time.ParseDuration(idleDuration); err != nil {
			return nil, fmt.Errorf("unable to parse destination-idle-duration as valid duration string: %w", err)
		}
		config.Destination.ConnectionOpts.IdleDuration = &idleDuration
	}
	if cmd.Flags().Changed("destination-max-in-flight") {
		mif, err := cmd.Flags().GetUint32("destination-max-in-flight")
		if err != nil {
			return nil, err
		}
		config.Destination.MaxInFlight = &mif
	}
	if cmd.Flags().Changed("destination-batch-count") {
		batchcount, err := cmd.Flags().GetUint32("destination-batch-count")
		if err != nil {
			return nil, err
		}
		if config.Destination.Batch == nil {
			config.Destination.Batch = &batchConfig{}
		}
		config.Destination.Batch.Count = &batchcount
	}
	if cmd.Flags().Changed("destination-batch-period") {
		batchperiod, err := cmd.Flags().GetString("destination-batch-period")
		if err != nil {
			return nil, err
		}
		if config.Destination.Batch == nil {
			config.Destination.Batch = &batchConfig{}
		}
		if _, err := time.ParseDuration(batchperiod); err != nil {
			return nil, fmt.Errorf("unable to parse destination-batch-period as valid duration string: %w", err)
		}
		config.Destination.Batch.Period = &batchperiod
	}
	return config, nil
}

func isConfigValid(cmd *cmdConfig, logger *slog.Logger, sourceConnection *mgmtv1alpha1.Connection, sourceConnectionType benthosbuilder_shared.ConnectionType) error {
	if sourceConnectionType == benthosbuilder_shared.ConnectionTypeAwsS3 && (cmd.Source.ConnectionOpts.JobId == nil || *cmd.Source.ConnectionOpts.JobId == "") && (cmd.Source.ConnectionOpts.JobRunId == nil || *cmd.Source.ConnectionOpts.JobRunId == "") {
		return errors.New("S3 source connection type requires job-id or job-run-id.")
	}
	if sourceConnectionType == benthosbuilder_shared.ConnectionTypeGCP && (cmd.Source.ConnectionOpts.JobId == nil || *cmd.Source.ConnectionOpts.JobId == "") && (cmd.Source.ConnectionOpts.JobRunId == nil || *cmd.Source.ConnectionOpts.JobRunId == "") {
		return errors.New("GCP Cloud Storage source connection type requires job-id or job-run-id")
	}

	if (sourceConnectionType == benthosbuilder_shared.ConnectionTypeAwsS3 || sourceConnectionType == benthosbuilder_shared.ConnectionTypeGCP) && cmd.Destination.InitSchema {
		return errors.New("init schema is only supported when source is a SQL Database")
	}

	if cmd.Destination != nil && cmd.Destination.TruncateCascade && cmd.Destination.Driver == mysqlDriver {
		return fmt.Errorf("truncate cascade is only supported in postgres")
	}

	if sourceConnectionType == benthosbuilder_shared.ConnectionTypeMysql || sourceConnectionType == benthosbuilder_shared.ConnectionTypePostgres {
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

	if sourceConnectionType == benthosbuilder_shared.ConnectionTypeDynamodb {
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

	var destinationDriver *DriverType
	if cmd.Destination != nil {
		destinationDriver = &cmd.Destination.Driver
	}
	logger.Debug("Checking if source and destination are compatible")
	err := areSourceAndDestCompatible(sourceConnection, destinationDriver)
	if err != nil {
		return err
	}
	return nil
}
