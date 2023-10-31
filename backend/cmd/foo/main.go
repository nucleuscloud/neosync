package main

import (
	"fmt"

	namegenerator "github.com/nucleuscloud/neosync/backend/internal/name-generator"
)

func main() {
	// nsclient, err := client.NewNamespaceClient(client.Options{})
	// if err != nil {
	// 	panic(err)
	// }
	// _ = nsclient
	// ctx := context.Background()
	// nsres, err := nsclient.Describe(ctx, "foobar")
	// if err != nil {
	// 	fmt.Println(err.Error() == serviceerror.NewNamespaceNotFound("foobar").Error(), errors.Is(err, serviceerror.NewNamespaceNotFound("foobar")))
	// 	panic(err)
	// }
	// fmt.Println(nsres.NamespaceInfo.Name)
	ng := namegenerator.New()
	fmt.Println(ng.Generate())
	fmt.Println(ng.Generate())
	fmt.Println(ng.Generate())
	fmt.Println(ng.Generate())
	fmt.Println(ng.Generate())
}
