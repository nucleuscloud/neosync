package sync_cmd

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	neosync_benthos "github.com/nucleuscloud/neosync/cli/internal/benthos"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "One off sync job to local resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return sync(cmd.Context(), &apiKey)
		},
	}

	return cmd
}

func sync(ctx context.Context, apiKey *string) error {
	fmt.Println("CLI")
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}

	connectionclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)
	fmt.Println("connection client")

	stream, err := connectionclient.GetConnectionDataStream(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionDataStreamRequest{
		SourceConnectionId: "3b4db2af-ef33-4e26-b0b9-f6df7518e78b",
		Schema:             "public",
		Table:              "locations",
	}))
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("get data stream")

	for {
		response := stream.Receive()
		if response {
			fmt.Println(stream.Msg().Data)
		} else {
			return nil
		}

	}

	// create benthos config

}

type benthosConfigResponse struct {
	Name      string
	DependsOn []string
	Config    *neosync_benthos.BenthosConfig
}

func generateBenthosConfig()

/*

{
  "Name": "public.regions",
  "DependsOn": [],
  "Config": {
   "input": {
    "label": "",
    "sql_select": {
     "driver": "postgres",
     "dsn": "postgres://postgres:foofar@10.244.0.50:5432/nucleus?sslmode=disable",
     "table": "public.regions",
     "columns": [
      "region_id",
      "region_name"
     ]
    }
   },
   "buffer": null,
   "pipeline": {
    "threads": -1,
    "processors": []
   },
   "output": {
    "label": "",
    "broker": {
     "pattern": "fan_out",
     "outputs": [
      {
       "sql_insert": {
        "driver": "postgres",
        "dsn": "postgres://postgres:foofar@10.244.0.107:5432/nucleus?sslmode=disable",
        "table": "public.regions",
        "columns": [
         "region_id",
         "region_name"
        ],
        "args_mapping": "root = [this.region_id, this.region_name]",
        "init_statement": "CREATE TABLE IF NOT EXISTS public.regions (region_id integer NOT NULL DEFAULT nextval('regions_region_id_seq'::regclass), region_name character varying NULL, CONSTRAINT regions_pkey PRIMARY KEY (region_id));\nTRUNCATE TABLE public.regions CASCADE;",
        "conn_max_idle": 2,
        "conn_max_open": 2,
        "batching": {
         "count": 32767,
         "byte_size": 0,
         "period": "1s",
         "check": "",
         "processors": null
        }
       }
      },
      {
       "sql_insert": {
        "driver": "postgres",
        "dsn": "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
        "table": "public.regions",
        "columns": [
         "region_id",
         "region_name"
        ],
        "args_mapping": "root = [this.region_id, this.region_name]",
        "init_statement": "CREATE TABLE IF NOT EXISTS public.regions (region_id integer NOT NULL DEFAULT nextval('regions_region_id_seq'::regclass), region_name character varying NULL, CONSTRAINT regions_pkey PRIMARY KEY (region_id));",
        "conn_max_idle": 2,
        "conn_max_open": 2,
        "batching": {
         "count": 32767,
         "byte_size": 0,
         "period": "1s",
         "check": "",
         "processors": null
        }
       }
      }
     ]
    }
   }
  }
 },


*/

/*
				input:
			  label: ""
			  http_client:
			    url: "" # No default (required)
			    verb: GET
			    headers: {}
			    rate_limit: "" # No default (optional)
			    timeout: 5s
			    payload: "" # No default (optional)
			    stream:
			      enabled: false
			      reconnect: true
			      codec: lines

				pipeline:
					processors: []

				output:
					output:
						sql_insert:
							driver: mysql
							dsn: foouser:foopassword@tcp(localhost:3306)/foodb
							table: footable
							columns: [ id, name, topic ]
							args_mapping: |
								root = [
									this.user.id,
									this.user.name,
									meta("kafka_topic"),
	      ]
*/
