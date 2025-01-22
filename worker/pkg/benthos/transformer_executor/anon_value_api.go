package transformer_executor

import (
	"encoding/json"
	"fmt"

	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
	"github.com/warpstreamlabs/bento/public/service"
)

type anonValueApi struct {
	message *service.Message
}

var _ javascript_functions.ValueApi = (*anonValueApi)(nil)

func newAnonValueApi() *anonValueApi {
	return &anonValueApi{}
}

func (b *anonValueApi) SetMessage(message *service.Message) {
	b.message = message
}

func (b *anonValueApi) Message() *service.Message {
	return b.message
}

func (b *anonValueApi) SetBytes(bytes []byte) {
	b.message.SetBytes(bytes)
}

func (b *anonValueApi) AsBytes() ([]byte, error) {
	return b.message.AsBytes()
}

func (b *anonValueApi) SetStructured(value any) {
	b.message.SetStructured(value)
}

func (b *anonValueApi) AsStructured() (any, error) {
	return b.message.AsStructured()
}

func (b *anonValueApi) MetaGet(key string) (any, bool) {
	return b.message.MetaGet(key)
}

func (b *anonValueApi) MetaSetMut(key string, value any) {
	b.message.MetaSetMut(key, value)
}

func (b *anonValueApi) GetPropertyPathValue(propertyPath string) (any, error) {
	if b.message == nil {
		return nil, fmt.Errorf("message is nil")
	}
	structuredValue, err := b.message.AsStructured()
	if err != nil {
		return nil, fmt.Errorf("failed to get structured message: %w", err)
	}
	structuredValueMap, ok := structuredValue.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("structured value is not a map[string]any")
	}
	return structuredValueMap[propertyPath], nil
}

func NewMessage(input map[string]any) (*service.Message, error) {
	bits, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input map: %w", err)
	}
	return service.NewMessage(bits), nil
}
