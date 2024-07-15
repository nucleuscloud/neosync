package transformers

type TemplateData struct {
	Name        string
	Description string
	Example     string
}

type NeosyncTransformer interface {
	ParseOptions(opts map[string]any) (any, error)
	GetJsTemplateData() (*TemplateData, error)
	Transform(value any, opts any) (any, error)
}

type NeosyncGenerator interface {
	ParseOptions(opts map[string]any) (any, error)
	GetJsTemplateData() (*TemplateData, error)
	Generate(opts any) (any, error)
}
