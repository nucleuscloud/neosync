package transformers

type Param struct {
	Name    string
	TypeStr string
	What    string
}
type TemplateData struct {
	Name        string
	Description string
	Params      []*Param
}

type NeosyncTransformer interface {
	// GetTemplateData() (*TemplateData, error)
	ParseOptions(opts map[string]any) (any, error)

	GetJsTemplateData() (*TemplateData, error)
	// GetBenthosTemplateData() (any, error)

	Transform(value any, opts any) (any, error)
}

type NeosyncGenerator interface {
	GetTemplateData() (*TemplateData, error)
	ParseOptions(opts map[string]any) (any, error)

	GetJsTemplateData() (*TemplateData, error)
	// GetBenthosTemplateData() (any, error)

	Generate(opts any) (any, error)
}
