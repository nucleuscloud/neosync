package javascript_processor

import (
	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
	"github.com/redpanda-data/benthos/v4/public/service"
)

// this is not thread safe
type benthosValueApi struct {
	message *service.Message
}

func newBatchBenthosValueApi() *benthosValueApi {
	return &benthosValueApi{}
}

// used by batch processor to update the target message while being able to reuse the same VM
func (b *benthosValueApi) SetMessage(message *service.Message) {
	b.message = message
}

func (b *benthosValueApi) Message() *service.Message {
	return b.message
}

var _ javascript_functions.ValueApi = (*benthosValueApi)(nil)

func (b *benthosValueApi) SetBytes(bytes []byte) {
	b.message.SetBytes(bytes)
}

func (b *benthosValueApi) AsBytes() ([]byte, error) {
	return b.message.AsBytes()
}

func (b *benthosValueApi) SetStructured(value any) {
	b.message.SetStructured(value)
}

func (b *benthosValueApi) AsStructured() (any, error) {
	return b.message.AsStructured()
}

func (b *benthosValueApi) MetaGet(key string) (any, bool) {
	return b.message.MetaGet(key)
}

func (b *benthosValueApi) MetaSetMut(key string, value any) {
	b.message.MetaSetMut(key, value)
}
