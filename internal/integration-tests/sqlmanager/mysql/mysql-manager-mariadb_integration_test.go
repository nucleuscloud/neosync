package sqlmanager_mysql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/testutil"
	tcmariadb "github.com/nucleuscloud/neosync/internal/testutil/testcontainers/mariadb"
	"golang.org/x/sync/errgroup"
)

func Test_MysqlManager_MariaDb(t *testing.T) {
	ok := testutil.ShouldRunIntegrationTest()
	if !ok {
		return
	}
	t.Log("Running integration tests for Mysql Manager for MariaDB")
	t.Parallel()

	ctx := context.Background()
	containers, err := tcmariadb.NewMariaDBTestSyncContainer(ctx, []tcmariadb.Option{}, []tcmariadb.Option{})
	if err != nil {
		t.Fatal(err)
	}
	source := containers.Source
	target := containers.Target
	t.Log("Successfully created source and target mysql test containers")

	err = setupMariaDb(ctx, containers)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Successfully setup source and target databases")

	sourceDB, err := sql.Open(sqlmanager_shared.MysqlDriver, source.URL)
	if err != nil {
		t.Fatal(err)
	}

	testMysqlManager(t, ctx, sourceDB, target.DB)
}

func setupMariaDb(ctx context.Context, containers *tcmariadb.MariaDBTestSyncContainer) error {
	baseDir := "testdata"

	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		err := containers.Source.RunSqlFiles(errctx, &baseDir, []string{"mariadb/setup.sql"})
		if err != nil {
			return fmt.Errorf("encountered error when executing source setup statement: %w", err)
		}
		return nil
	})
	errgrp.Go(func() error {
		err := containers.Target.RunSqlFiles(errctx, &baseDir, []string{"mariadb/init.sql"})
		if err != nil {
			return fmt.Errorf("encountered error when executing dest setup statement: %w", err)
		}
		return nil
	})

	err := errgrp.Wait()
	if err != nil {
		return err
	}

	return nil
}
