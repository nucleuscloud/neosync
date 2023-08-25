package main

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect"
	greetv1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/greet/v1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/greet/v1/greetv1connect"
)

func main() {
	client := greetv1connect.NewGreetServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)
	res, err := client.Greet(
		context.Background(),
		connect.NewRequest(&greetv1.GreetRequest{Name: "Nick"}),
	)
	if err != nil {
		panic(err)
	}
	log.Println(res.Msg.Greeting)
}
