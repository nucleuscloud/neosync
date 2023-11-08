package up_cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run all database migrations",
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
			return Up(
				cmd.Context(),
				dbUrl,
				schemaDir,
				slog.New(slog.NewJSONHandler(os.Stdout, nil)),
			)
		},
	}
	cmd.Flags().StringP("database", "db", "", "optionally set the database url, otherwise it will pull from the environment")
	cmd.Flags().StringP("source", "source", "", "optionally set the migrations dir, otherwise pull from DB_SCHEMA_DIR env")
	return cmd
}

func Up(
	ctx context.Context,
	connStr string,
	schemaDir string,
	logger *slog.Logger,
) error {

	var absSchemaDir string
	if filepath.IsAbs(schemaDir) {
		absSchemaDir = schemaDir
	} else {
		a, err := filepath.Abs(schemaDir)
		if err != nil {
			return err
		}
		absSchemaDir = a
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", strings.TrimPrefix(absSchemaDir, "file://")),
		connStr,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		return err
	}

	return nil
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

	return nucleusdb.GetDbUrl(&nucleusdb.ConnectConfig{
		Host:     dbHost,
		Port:     dbPort,
		Database: dbName,
		User:     dbUser,
		Pass:     dbPass,
		SslMode:  &sslMode,
	}), nil
}
