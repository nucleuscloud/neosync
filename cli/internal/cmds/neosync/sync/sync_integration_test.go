package sync_cmd

import (
	"context"
	"os"
	"testing"

	charmlog "github.com/charmbracelet/log"
	tcneosyncapi "github.com/nucleuscloud/neosync/backend/pkg/integration-test"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcpostgres "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/postgres"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

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
		api := tcneosyncapi.NewNeosyncApiTestClient(ctx, t)
		neosyncApi = api
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		panic(err)
	}

	testdataFolder := "../../../../../internal/testutil/testdata/postgres/alltypes"
	err = postgres.Source.RunSqlFiles(ctx, &testdataFolder, []string{"create-tables.sql"})
	if err != nil {
		panic(err)
	}
	err = postgres.Target.RunSqlFiles(ctx, &testdataFolder, []string{"create-schema.sql"})
	if err != nil {
		panic(err)
	}

	connclient := neosyncApi.UnathdClients.Connections
	conndataclient := neosyncApi.UnathdClients.ConnectionData

	sqlmanagerclient := tcneosyncapi.NewTestSqlManagerClient()

	// discardLogger := testutil.GetTestCharmLogger()
	discardLogger := charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		ReportTimestamp: true,
		Level:           charmlog.DebugLevel,
	})

	accountId := tcneosyncapi.CreatePersonalAccount(ctx, t, neosyncApi.UnathdClients.Users)
	sourceConn := tcneosyncapi.CreatePostgresConnection(ctx, t, neosyncApi.UnathdClients.Connections, accountId, "source", postgres.Source.URL)
	t.Run("sync_postgres", func(t *testing.T) {
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
		}
		sync := &clisync{
			connectiondataclient: conndataclient,
			connectionclient:     connclient,
			sqlmanagerclient:     sqlmanagerclient,
		}
		err := sync.configureAndRunSync(ctx, discardLogger, "plain", &accountId, cmdConfig)
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
