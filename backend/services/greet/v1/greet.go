package v1

import (
	"context"
	"fmt"
	"log"

	"connectrpc.com/connect"
	greetv1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/greet/v1"
)

type GreetServer struct{}

func (s *GreetServer) Greet(
	ctx context.Context,
	req *connect.Request[greetv1.GreetRequest],
) (*connect.Response[greetv1.GreetResponse], error) {
	log.Println("Request headers: ", req.Header())

	res := connect.NewResponse(&greetv1.GreetResponse{
		Greeting: fmt.Sprintf("Hello, %s! - method: %s", req.Msg.Name, req.HTTPMethod()),
	})
	res.Header().Set("Greet-Version", "v1")

	return res, nil
}
