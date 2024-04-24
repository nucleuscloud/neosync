package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	"github.com/benthosdev/benthos/v4/public/service"
	bsql "github.com/nucleuscloud/neosync/worker/internal/benthos/sql"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"
)

type pprovider struct {
}

func (p *pprovider) GetDb(driver, dsn string) (bsql.SqlDbtx, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	env := service.NewEnvironment()

	dbprovider := &pprovider{}

	err := bsql.RegisterPooledSqlInsertOutput(env, dbprovider, false)
	if err != nil {
		panic(err)
	}
	err = bsql.RegisterPooledSqlUpdateOutput(env, dbprovider)
	if err != nil {
		panic(err)
	}
	stop := make(chan error)
	err = bsql.RegisterPooledSqlRawInput(env, dbprovider, stop)
	if err != nil {
		panic(err)
	}
	go func() {
		stoperr := <-stop
		if stoperr != nil {
			panic(stoperr)
		}
	}()

	streambldr := env.NewStreamBuilder()
	ctx := withInterrupt(context.Background())

	config := "./benthos.yml"
	if len(os.Args) == 2 {
		config = os.Args[1]
	}

	fmt.Println("reading benthos config from ", config)
	benthosyamlbits, err := os.ReadFile(config)
	if err != nil {
		panic(err)
	}

	err = streambldr.SetYAML(string(benthosyamlbits))
	if err != nil {
		panic(err)
	}

	stream, err := streambldr.Build()
	if err != nil {
		panic(err)
	}

	ctx, _ = context.WithDeadline(ctx, time.Now().Add(3*time.Minute))
	err = stream.Run(ctx)
	if err != nil {
		panic(err)
	}
}

func withInterrupt(ctx context.Context) context.Context {
	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		oscall := <-signalChan
		slog.Info(fmt.Sprintf("system call:%+v\n", oscall))
		cancel()
	}()
	return ctx
}
