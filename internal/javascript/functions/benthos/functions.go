package benthos_functions

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dop251/goja"
	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
)

const (
	namespace = "benthos"
)

func Get() []*javascript_functions.FunctionDefinition {
	return []*javascript_functions.FunctionDefinition{
		getV0Fetch(namespace),
		getV0MsgSetString(namespace),
		getV0MsgAsString(namespace),
		getV0MsgSetStructured(namespace),
		getV0MsgAsStructured(namespace),
		getV0MsgSetMeta(namespace),
		getV0MsgGetMeta(namespace),
		getV0MsgMetaExists(namespace),
	}
}

func getV0Fetch(namespace string) *javascript_functions.FunctionDefinition {
	return javascript_functions.NewFunctionDefinition(namespace, "v0_fetch", func(r javascript_functions.Runner) javascript_functions.Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			var (
				url         string
				httpHeaders map[string]any
				method      = "GET"
				payload     = ""
			)
			if err := javascript_functions.ParseFunctionArguments(call, &url, &httpHeaders, &method, &payload); err != nil {
				return nil, err
			}

			var payloadReader io.Reader
			if payload != "" {
				payloadReader = strings.NewReader(payload)
			}

			req, err := http.NewRequestWithContext(ctx, method, url, payloadReader)
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
}

func getV0MsgSetString(namespace string) *javascript_functions.FunctionDefinition {
	return javascript_functions.NewFunctionDefinition(namespace, "v0_msg_set_string", func(r javascript_functions.Runner) javascript_functions.Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			var value string
			if err := javascript_functions.ParseFunctionArguments(call, &value); err != nil {
				return nil, err
			}

			r.ValueApi().SetBytes([]byte(value))
			return nil, nil
		}
	})
}

func getV0MsgAsString(namespace string) *javascript_functions.FunctionDefinition {
	return javascript_functions.NewFunctionDefinition(namespace, "v0_msg_as_string", func(r javascript_functions.Runner) javascript_functions.Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			b, err := r.ValueApi().AsBytes()
			if err != nil {
				return nil, err
			}
			return string(b), nil
		}
	})
}

func getV0MsgSetStructured(namespace string) *javascript_functions.FunctionDefinition {
	return javascript_functions.NewFunctionDefinition(namespace, "v0_msg_set_structured", func(r javascript_functions.Runner) javascript_functions.Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			var value any
			if err := javascript_functions.ParseFunctionArguments(call, &value); err != nil {
				return nil, err
			}

			r.ValueApi().SetStructured(value)
			return nil, nil
		}
	})
}

func getV0MsgAsStructured(namespace string) *javascript_functions.FunctionDefinition {
	return javascript_functions.NewFunctionDefinition(namespace, "v0_msg_as_structured", func(r javascript_functions.Runner) javascript_functions.Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			return r.ValueApi().AsStructured()
		}
	})
}

func getV0MsgSetMeta(namespace string) *javascript_functions.FunctionDefinition {
	return javascript_functions.NewFunctionDefinition(namespace, "v0_msg_set_meta", func(r javascript_functions.Runner) javascript_functions.Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			var key string
			var value any
			if err := javascript_functions.ParseFunctionArguments(call, &key, &value); err != nil {
				return nil, err
			}
			r.ValueApi().MetaSetMut(key, value)
			return nil, nil
		}
	})
}

func getV0MsgGetMeta(namespace string) *javascript_functions.FunctionDefinition {
	return javascript_functions.NewFunctionDefinition(namespace, "v0_msg_get_meta", func(r javascript_functions.Runner) javascript_functions.Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			var key string
			if err := javascript_functions.ParseFunctionArguments(call, &key); err != nil {
				return nil, err
			}
			result, ok := r.ValueApi().MetaGet(key)
			if !ok {
				return nil, fmt.Errorf("key %s not found", key)
			}
			return result, nil
		}
	})
}

func getV0MsgMetaExists(namespace string) *javascript_functions.FunctionDefinition {
	return javascript_functions.NewFunctionDefinition(namespace, "v0_msg_exists_meta", func(r javascript_functions.Runner) javascript_functions.Function {
		return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (any, error) {
			var key string
			if err := javascript_functions.ParseFunctionArguments(call, &key); err != nil {
				return nil, err
			}
			_, ok := r.ValueApi().MetaGet(key)
			return ok, nil
		}
	})
}
