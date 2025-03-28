package up_cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	neomigrate "github.com/nucleuscloud/neosync/internal/migrate"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Run all database migrations down",
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaDir, err := cmd.Flags().GetString("source")
			if err != nil {
				return err
			}
			if schemaDir == "" {
				schemaDir = viper.GetString("DB_SCHEMA_DIR")
				if schemaDir == "" {
					return errors.New("must provide schema dir as flag or env var")
				}
			}
			dbUrl, err := cmd.Flags().GetString("database")
			if err != nil {
				return err
			}
			if dbUrl == "" {
				dbUrl, err = getDbUrl()
				if err != nil {
					return err
				}
			}

			cmd.SilenceUsage = true
			return neomigrate.Down(
				cmd.Context(),
				dbUrl,
				schemaDir,
				slog.New(slog.NewJSONHandler(os.Stdout, nil)),
			)
		},
	}
	cmd.Flags().
		StringP("database", "d", "", "optionally set the database url, otherwise it will pull from the environment")
	cmd.Flags().
		StringP("source", "s", "", "optionally set the migrations dir, otherwise pull from DB_SCHEMA_DIR env")
	return cmd
}

func getDbUrl() (string, error) {
	dburl := viper.GetString("DB_URL")
	if dburl != "" {
		return dburl, nil
	}

	dbHost := viper.GetString("DB_HOST")
	if dbHost == "" {
		return "", fmt.Errorf("must provide DB_HOST in environment")
	}

	dbPort := viper.GetInt("DB_PORT")
	if dbPort == 0 {
		return "", fmt.Errorf("must provide DB_PORT in environment")
	}

	dbName := viper.GetString("DB_NAME")
	if dbName == "" {
		return "", fmt.Errorf("must provide DB_NAME in environment")
	}

	dbUser := viper.GetString("DB_USER")
	if dbUser == "" {
		return "", fmt.Errorf("must provide DB_USER in environment")
	}

	dbPass := viper.GetString("DB_PASS")
	if dbPass == "" {
		return "", fmt.Errorf("must provide DB_PASS in environment")
	}

	sslMode := "require"
	if viper.IsSet("DB_SSL_DISABLE") && viper.GetBool("DB_SSL_DISABLE") {
		sslMode = "disable"
	}

	var migrationsTable *string
	if viper.IsSet("DB_MIGRATIONS_TABLE") {
		table := viper.GetString("DB_MIGRATIONS_TABLE")
		migrationsTable = &table
	}

	var tableQuoted *bool
	if viper.IsSet("DB_MIGRATIONS_TABLE_QUOTED") {
		isQuoted := viper.GetBool("DB_MIGRATIONS_TABLE_QUOTED")
		tableQuoted = &isQuoted
	}

	var dbOptions *string
	if viper.IsSet("DB_MIGRATIONS_OPTIONS") {
		val := viper.GetString("DB_MIGRATIONS_OPTIONS")
		dbOptions = &val
	}

	return neosyncdb.GetDbUrl(&neosyncdb.ConnectConfig{
		Host:                  dbHost,
		Port:                  dbPort,
		Database:              dbName,
		User:                  dbUser,
		Pass:                  dbPass,
		SslMode:               &sslMode,
		MigrationsTableName:   migrationsTable,
		MigrationsTableQuoted: tableQuoted,
		Options:               dbOptions,
	}), nil
}
