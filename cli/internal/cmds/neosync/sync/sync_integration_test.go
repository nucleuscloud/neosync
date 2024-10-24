package sync_cmd

import (
	"context"
	"testing"

	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/cli/internal/output"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

const neosyncDbMigrationsPath = "../../../../../backend/sql/postgresql/schema"

// TODO fix types then this will pass
func Test_Sync_Postgres(t *testing.T) {
	t.Parallel()
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	ctx := context.Background()

	var neosyncApi *tcneosyncapi.NeosyncApiTestClient
	var postgres *tcpostgres.PostgresTestSyncContainer

	errgrp := errgroup.Group{}
	errgrp.Go(func() error {
		p, err := tcpostgres.NewPostgresTestSyncContainer(ctx, []tcpostgres.Option{}, []tcpostgres.Option{})
		if err != nil {
			return err
		}
		postgres = p
		return nil
	})

	errgrp.Go(func() error {
		api, err := tcneosyncapi.NewNeosyncApiTestClient(ctx, t, tcneosyncapi.WithMigrationsDirectory(neosyncDbMigrationsPath))
		if err != nil {
			return err
		}
		neosyncApi = api
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		panic(err)
	}

	testdataFolder := "../../../../../internal/testutil/testdata/postgres/humanresources"
	err = postgres.Source.RunSqlFiles(ctx, &testdataFolder, []string{"create-tables.sql"})
	if err != nil {
		panic(err)
	}
	err = postgres.Target.RunSqlFiles(ctx, &testdataFolder, []string{"create-schema.sql"})
	if err != nil {
		panic(err)
	}

	connclient := neosyncApi.UnauthdClients.Connections
	conndataclient := neosyncApi.UnauthdClients.ConnectionData

	sqlmanagerclient := tcneosyncapi.NewTestSqlManagerClient()

	discardLogger := testutil.GetTestCharmSlogger()

	accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, neosyncApi.UnauthdClients.Users)
	sourceConn := tcneosyncapi.CreatePostgresConnection(ctx, t, neosyncApi.UnauthdClients.Connections, accountId, "source", postgres.Source.URL)
	t.Run("sync_postgres", func(t *testing.T) {
		outputType := output.PlainOutput
		cmdConfig := &cmdConfig{
			Source: &sourceConfig{
				ConnectionId: sourceConn.Id,
			},
			Destination: &sqlDestinationConfig{
				ConnectionUrl:        postgres.Target.URL,
				Driver:               sqlmanager_shared.PostgresDriver,
				InitSchema:           true,
				TruncateBeforeInsert: true,
				TruncateCascade:      true,
			},
			OutputType: &outputType,
			AccountId:  &accountId,
		}
		sync := &clisync{
			connectiondataclient: conndataclient,
			connectionclient:     connclient,
			sqlmanagerclient:     sqlmanagerclient,
			ctx:                  ctx,
			logger:               discardLogger,
			cmd:                  cmdConfig,
		}
		err := sync.configureAndRunSync()
		require.NoError(t, err)
	})

	err = postgres.TearDown(ctx)
	if err != nil {
		panic(err)
	}
	err = neosyncApi.TearDown(ctx)
	if err != nil {
		panic(err)
	}
}
