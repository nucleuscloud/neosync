package neosync_functions

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/dop251/goja"
	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

const (
	namespace = "neosync"
)

func Get(
	transformPiiTextApi transformers.TransformPiiTextApi,
) ([]*javascript_functions.FunctionDefinition, error) {
	generatorFns, err := getNeosyncGenerators()
	if err != nil {
		return nil, err
	}
	transformerFns, err := getNeosyncTransformers(transformPiiTextApi)
	if err != nil {
		return nil, err
	}
	patchStructuredMessage := getPatchStructuredMessage(namespace)

	output := make(
		[]*javascript_functions.FunctionDefinition,
		0,
		len(generatorFns)+len(transformerFns)+1,
	)
	output = append(output, generatorFns...)
	output = append(output, transformerFns...)
	output = append(output, patchStructuredMessage)
	return output, nil
}

func getPatchStructuredMessage(namespace string) *javascript_functions.FunctionDefinition {
	fnName := "patchStructuredMessage"
	return javascript_functions.NewFunctionDefinition(
		namespace,
		fnName,
		func(r javascript_functions.Runner) javascript_functions.Function {
			return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (result any, err error) {
				defer func() {
					if r := recover(); r != nil {
						// we set the named "err" argument to the error so that it can be returned
						err = fmt.Errorf("panic recovered: %s.%s: %v", namespace, fnName, r)
						l.Error(
							"recovered from panic in custom neosync function",
							"error", err,
							"function", fmt.Sprintf("%s.%s", namespace, fnName),
							"stack", string(debug.Stack()),
						)
					}
				}()
				var updates map[string]any
				if err := javascript_functions.ParseFunctionArguments(call, &updates); err != nil {
					return nil, err
				}

				originalData, err := r.ValueApi().AsStructured()
				if err != nil {
					return nil, fmt.Errorf("failed to get structured data: %w", err)
				}

				originalMap, ok := originalData.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("structured data is not a map")
				}

				for key, value := range updates {
					setNestedProperty(originalMap, key, value)
				}

				r.ValueApi().SetStructured(originalMap)

				return nil, nil
			}
		},
	)
}

func setNestedProperty(obj map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := obj

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
		} else {
			if _, ok := current[part]; !ok {
				current[part] = make(map[string]any)
			}
			current = current[part].(map[string]any)
		}
	}
}

func getNeosyncGenerators() ([]*javascript_functions.FunctionDefinition, error) {
	generators := transformers.GetNeosyncGenerators()
	fns := make([]*javascript_functions.FunctionDefinition, 0, len(generators))
	for _, f := range generators {
		templateData, err := f.GetJsTemplateData()
		if err != nil {
			return nil, err
		}

		fn := javascript_functions.NewFunctionDefinition(
			namespace,
			templateData.Name,
			func(r javascript_functions.Runner) javascript_functions.Function {
				return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (result any, err error) {
					defer func() {
						if r := recover(); r != nil {
							// we set the named "err" argument to the error so that it can be returned
							err = fmt.Errorf(
								"panic recovered: %s.%s: %v",
								namespace,
								templateData.Name,
								r,
							)
							l.Error(
								"recovered from panic in custom neosync function",
								"error", err,
								"function", fmt.Sprintf("%s.%s", namespace, templateData.Name),
								"stack", string(debug.Stack()),
							)
						}
					}()
					var (
						opts map[string]any
					)

					if err := javascript_functions.ParseFunctionArguments(call, &opts); err != nil {
						return nil, err
					}
					goOpts, err := f.ParseOptions(opts)
					if err != nil {
						return nil, err
					}
					return f.Generate(goOpts)
				}
			},
		)
		fns = append(fns, fn)
	}
	return fns, nil
}

func getNeosyncTransformers(
	transformPiiTextApi transformers.TransformPiiTextApi,
) ([]*javascript_functions.FunctionDefinition, error) {
	neosyncTransformers := transformers.GetNeosyncTransformers()
	if transformPiiTextApi != nil {
		neosyncTransformers = append(neosyncTransformers, transformers.NewTransformPiiText(transformPiiTextApi))
	}
	fns := make([]*javascript_functions.FunctionDefinition, 0, len(neosyncTransformers))
	for _, f := range neosyncTransformers {
		templateData, err := f.GetJsTemplateData()
		if err != nil {
			return nil, err
		}

		fn := javascript_functions.NewFunctionDefinition(
			namespace,
			templateData.Name,
			func(r javascript_functions.Runner) javascript_functions.Function {
				return func(ctx context.Context, call goja.FunctionCall, rt *goja.Runtime, l *slog.Logger) (result any, err error) {
					defer func() {
						if r := recover(); r != nil {
							// we set the named "err" argument to the error so that it can be returned
							err = fmt.Errorf(
								"panic recovered: %s.%s: %v",
								namespace,
								templateData.Name,
								r,
							)
							l.Error(
								"recovered from panic in custom neosync function",
								"error", err,
								"function", fmt.Sprintf("%s.%s", namespace, templateData.Name),
								"stack", string(debug.Stack()),
							)
						}
					}()
					var (
						value any
						opts  map[string]any
					)

					if err := javascript_functions.ParseFunctionArguments(call, &value, &opts); err != nil {
						return nil, err
					}
					goOpts, err := f.ParseOptions(opts)
					if err != nil {
						return nil, err
					}
					return f.Transform(value, goOpts)
				}
			},
		)
		fns = append(fns, fn)
	}
	return fns, nil
}
