package javascript

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dop251/goja"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

type jsFunction func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error)

type jsFunctionParam struct {
	name    string
	typeStr string
	what    string
}

type jsFunctionDefinition struct {
	namespace   string
	name        string
	description string
	params      []jsFunctionParam
	examples    []string
	ctor        func(r *vmRunner) jsFunction
}

func (j *jsFunctionDefinition) Param(name, typeStr, what string) *jsFunctionDefinition {
	j.params = append(j.params, jsFunctionParam{
		name:    name,
		typeStr: typeStr,
		what:    what,
	})
	return j
}

func (j *jsFunctionDefinition) Example(example string) *jsFunctionDefinition {
	j.examples = append(j.examples, example)
	return j
}

func (j *jsFunctionDefinition) FnCtor(ctor func(r *vmRunner) jsFunction) *jsFunctionDefinition {
	j.ctor = ctor
	return j
}

func (j *jsFunctionDefinition) Namespace(namespace string) *jsFunctionDefinition {
	j.namespace = namespace
	return j
}

func (j *jsFunctionDefinition) String() string {
	var description strings.Builder

	_, _ = fmt.Fprintf(&description, "### `benthos.%v`\n\n", j.name)
	_, _ = description.WriteString(j.description + "\n\n")
	if len(j.params) > 0 {
		_, _ = description.WriteString("#### Parameters\n\n")
		for _, p := range j.params {
			_, _ = fmt.Fprintf(&description, "**`%v`** &lt;%v&gt; %v  \n", p.name, p.typeStr, p.what)
		}
		_, _ = description.WriteString("\n")
	}

	if len(j.examples) > 0 {
		_, _ = description.WriteString("#### Examples\n\n")
		for _, e := range j.examples {
			_, _ = description.WriteString("```javascript\n")
			_, _ = description.WriteString(strings.Trim(e, "\n"))
			_, _ = description.WriteString("\n```\n")
		}
	}

	return description.String()
}

var vmRunnerFunctionCtors = map[string]*jsFunctionDefinition{}

func registerVMRunnerFunction(name, description string) *jsFunctionDefinition {
	fn := &jsFunctionDefinition{
		name:        name,
		description: description,
	}
	vmRunnerFunctionCtors[name] = fn
	return fn
}

// ------------------------------------------------------------------------------
// type Param struct {
// 	Name    string
// 	TypeStr string
// 	What    string
// }
// type TemplateData struct {
// 	Name        string
// 	Description string
// 	Params      []*Param
// }

// var f NeosyncTransformer = &TransformFloat{}

// type TransformFloatOpts struct {
// 	randomizer rng.Rand
// }
// type TransformFloat struct {
// 	maxnumgetter any
// }

// func NewTransformFloat() *TransformFloat {
// 	return &TransformFloat{
// 		maxnumgetter: 1,
// 	}
// }

// func (t *TransformFloat) GetTemplateData() (*TemplateData, error) {
// 	return nil, nil
// }
// func (t *TransformFloat) ParseOptions(opts map[string]any) (any, error) {
// 	// seed comes from the user opts
// 	// var udSeed = 1

// 	return &TransformFloatOpts{
// 		randomizer: rng.New(1),
// 	}, nil
// }

// func (t *TransformFloat) Transform(value any, opts any) (any, error) {
// 	parsedOpts, ok := opts.(*TransformFloatOpts)
// 	if !ok {
// 		return nil, errors.New("invalid parse opts")
// 	}
// 	_ = parsedOpts

// 	return 1, nil
// }

// type NeosyncTransformer interface {
// 	GetTemplateData() (*TemplateData, error)
// 	ParseOptions(opts map[string]any) (any, error)

// 	GetJsTemplateData() (*TemplateData, error)
// 	GetBenthosTemplateData() (any, error)

// 	// Get() func(value any, opts TOpts) (any, error)
// 	Transform(value any, opts any) (any, error)
// }

// type NeosyncGenerator[TOpts any] interface {
// 	GetTemplateData() (*TemplateData, error)
// 	ParseOptions(opts map[string]any) (TOpts, error)
// 	Generate(opts TOpts) (any, error)
// }

func init() {
	neosyncFns := transformers.GetNeosyncTransformers()
	for _, f := range neosyncFns {
		templateData, err := f.GetJsTemplateData()
		if err != nil {
			panic(err)
		}

		def := registerVMRunnerFunction(templateData.Name, templateData.Description)
		def.Param("value", "any", "The value to be transformed.")
		def.Param("opts", "object", "Transformer options config")
		def.Namespace(neosyncFnCtxName)
		def.FnCtor(func(r *vmRunner) jsFunction {
			return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (any, error) {
				var (
					value any
					opts  map[string]any
				)
				if err := parseArgs(call, &value, &opts); err != nil {
					return nil, err
				}
				goOpts, err := f.ParseOptions(opts)
				if err != nil {
					return nil, err
				}
				return f.Transform(value, goOpts)
			}
		})
	}
}

var _ = registerVMRunnerFunction(
	"v0_fetch",
	`Executes an HTTP request synchronously and returns the result as an object of the form `+"`"+`{"status":200,"body":"foo"}`+"`"+`.`,
).
	Namespace(benthosFnCtxName).
	Param("url", "string", "The URL to fetch").
	Param("headers", "object(string,string)", "An object of string/string key/value pairs to add the request as headers.").
	Param("method", "string", "The method of the request.").
	Param("body", "(optional) string", "A body to send.").
	Example(`
let result = benthos.v0_fetch("http://example.com", {}, "GET", "")
benthos.v0_msg_set_structured(result);
`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			var (
				url         string
				httpHeaders map[string]any
				method      = "GET"
				payload     = ""
			)
			if err := parseArgs(call, &url, &httpHeaders, &method, &payload); err != nil {
				return nil, err
			}

			var payloadReader io.Reader
			if payload != "" {
				payloadReader = strings.NewReader(payload)
			}

			req, err := http.NewRequest(method, url, payloadReader)
			if err != nil {
				return nil, err
			}

			// Parse HTTP headers
			for k, v := range httpHeaders {
				vStr, _ := v.(string)
				req.Header.Add(k, vStr)
			}

			// Do request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"status": resp.StatusCode,
				"body":   string(respBody),
			}, nil
		}
	})

var _ = registerVMRunnerFunction("v0_msg_set_string", `Set the contents of the processed message to a given string.`).
	Namespace(benthosFnCtxName).
	Param("value", "string", "The value to set it to.").
	Example(`benthos.v0_msg_set_string("hello world");`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			var value string
			if err := parseArgs(call, &value); err != nil {
				return nil, err
			}

			r.targetMessage.SetBytes([]byte(value))
			return nil, nil
		}
	})

var _ = registerVMRunnerFunction("v0_msg_as_string", `Obtain the raw contents of the processed message as a string.`).
	Namespace(benthosFnCtxName).
	Example(`let contents = benthos.v0_msg_as_string();`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			b, err := r.targetMessage.AsBytes()
			if err != nil {
				return nil, err
			}
			return string(b), nil
		}
	})

var _ = registerVMRunnerFunction("v0_msg_set_structured", `Set the root of the processed message to a given value of any type.`).
	Namespace(benthosFnCtxName).
	Param("value", "anything", "The value to set it to.").
	Example(`
benthos.v0_msg_set_structured({
  "foo": "a thing",
  "bar": "something else",
  "baz": 1234
});
`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			var value any
			if err := parseArgs(call, &value); err != nil {
				return nil, err
			}

			r.targetMessage.SetStructured(value)
			return nil, nil
		}
	})

var _ = registerVMRunnerFunction("v0_msg_as_structured", `Obtain the root of the processed message as a structured value. If the message is not valid JSON or has not already been expanded into a structured form this function will throw an error.`).
	Namespace(benthosFnCtxName).
	Example(`let foo = benthos.v0_msg_as_structured().foo;`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			return r.targetMessage.AsStructured()
		}
	})

var _ = registerVMRunnerFunction("v0_msg_exists_meta", `Check that a metadata key exists.`).
	Namespace(benthosFnCtxName).
	Param("name", "string", "The metadata key to search for.").
	Example(`if (benthos.v0_msg_exists_meta("kafka_key")) {}`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			var name string
			if err := parseArgs(call, &name); err != nil {
				return nil, err
			}

			_, ok := r.targetMessage.MetaGet(name)
			if !ok {
				return false, nil
			}
			return true, nil
		}
	})

var _ = registerVMRunnerFunction("v0_msg_get_meta", `Get the value of a metadata key from the processed message.`).
	Namespace(benthosFnCtxName).
	Param("name", "string", "The metadata key to search for.").
	Example(`let key = benthos.v0_msg_get_meta("kafka_key");`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			var name string
			if err := parseArgs(call, &name); err != nil {
				return nil, err
			}

			result, ok := r.targetMessage.MetaGet(name)
			if !ok {
				return nil, errors.New("key not found")
			}
			return result, nil
		}
	})

var _ = registerVMRunnerFunction("v0_msg_set_meta", `Set a metadata key on the processed message to a value.`).
	Namespace(benthosFnCtxName).
	Param("name", "string", "The metadata key to set.").
	Param("value", "anything", "The value to set it to.").
	Example(`benthos.v0_msg_set_meta("thing", "hello world");`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			var (
				name  string
				value any
			)
			if err := parseArgs(call, &name, &value); err != nil {
				return "", err
			}
			r.targetMessage.MetaSetMut(name, value)
			return nil, nil
		}
	})

var _ = registerVMRunnerFunction("hello", `Prefixes hello to string.`).
	Namespace(neosyncFnCtxName).
	Param("name", "string", "The metadata key to set.").
	Example(`neosync.hello("kevin");`).
	FnCtor(func(r *vmRunner) jsFunction {
		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
			var (
				name string
			)
			if err := parseArgs(call, &name); err != nil {
				return "", err
			}
			return fmt.Sprintf("hello %s", name), nil
		}
	})

	// var _ = registerVMRunnerFunction("transformFirstName", `Transforms first name`).
	// 	Namespace(neosyncFnCtxName).
	// 	Param("value", "any", "The metadata key to set.").
	// 	Param("opts", "object", "options config").
	// 	Example(`neosync.transformFirstName("kevin");`).
	// 	FnCtor(func(r *vmRunner) jsFunction {
	// 		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
	// 			var (
	// 				value any
	// 				opts  map[string]interface{}
	// 			)
	// 			if err := parseArgs(call, &value, &opts); err != nil {
	// 				return "", err
	// 			}
	// 			var seed int64
	// 			if opts != nil && opts["seed"] != nil {
	// 				seed = opts["seed"].(int64)
	// 			} else {
	// 				var err error
	// 				seed, err = transformer_utils.GenerateCryptoSeed()
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 			}

	// 			funcOpts := &transformer.TransformFirstNameOpts{}
	// 			funcOpts.PreserveLength = false
	// 			if opts != nil && opts["preserveLength"] != nil {
	// 				funcOpts.PreserveLength = opts["preserveLength"].(bool)
	// 			}
	// 			funcOpts.MaxLength = int64(10000)
	// 			if opts != nil && opts["maxLength"] != nil {
	// 				funcOpts.MaxLength = opts["maxLength"].(int64)
	// 			}

	// 			randomizer := rng.New(seed)
	// 			return transformer.TransformFirstName(randomizer, value, funcOpts)
	// 		}
	// 	})

	// var _ = registerVMRunnerFunction("generateFirstName", `Generates first name`).
	// 	Namespace(neosyncFnCtxName).
	// 	Param("opts", "object", "options config").
	// 	Example(`neosync.transformFirstName("kevin");`).
	// 	FnCtor(func(r *vmRunner) jsFunction {
	// 		return func(call goja.FunctionCall, rt *goja.Runtime, l *service.Logger) (interface{}, error) {
	// 			var (
	// 				opts map[string]interface{}
	// 			)
	// 			if err := parseArgs(call, &opts); err != nil {
	// 				return "", err
	// 			}
	// 			var seed int64
	// 			if opts != nil && opts["seed"] != nil {
	// 				seed = opts["seed"].(int64)
	// 			} else {
	// 				var err error
	// 				seed, err = transformer_utils.GenerateCryptoSeed()
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 			}

	// 			maxLength := int64(10000)
	// 			if opts != nil && opts["maxLength"] != nil {
	// 				maxLength = opts["maxLength"].(int64)
	// 			}

	// 			randomizer := rng.New(seed)
	// 			return transformer.GenerateRandomFirstName(randomizer, nil, maxLength)
	// 		}
	// 	})
