package main

// import (
// 	"encoding/json"
// 	"fmt"

// 	"github.com/benthosdev/benthos/v4/internal/bloblang/query"
// 	"github.com/benthosdev/benthos/v4/public/bloblang"
// )

// func main() {
// 	spec := bloblang.NewPluginSpec().
// 		Param(bloblang.NewInt64Param("max_length").Default(10000)).
// 		Param(bloblang.NewAnyParam("value").Optional()).
// 		Param(bloblang.NewBoolParam("preserve_length").Default(false)).
// 		Param(bloblang.NewInt64Param("seed").Optional())
// 	specString := fmt.Sprintf("%v", spec)
// 	newSpec := bloblang.NewPluginSpec()
// 	err := newSpec.EncodeJSON([]byte(specString))
// 	if err != nil {
// 		panic(err)
// 	}
// 	def := struct {
// 		Description string       `json:"description"`
// 		Params      query.Params `json:"params"`
// 	}{}
// 	if err := json.Unmarshal([]byte(specString), &def); err != nil {
// 		panic(err)
// 	}
// 	jsonF, _ := json.MarshalIndent(newSpec, "", " ")
// 	fmt.Printf("def: %s \n", string(jsonF))

// }
