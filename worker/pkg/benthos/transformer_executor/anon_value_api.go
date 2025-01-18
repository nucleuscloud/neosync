package transformer_executor

import (
	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
	"github.com/warpstreamlabs/bento/public/service"
)

type anonValueApi struct {
	message *service.Message
}

func newAnonValueApi() *anonValueApi {
	return &anonValueApi{}
}

func (b *anonValueApi) SetMessage(message *service.Message) {
	b.message = message
}

func (b *anonValueApi) Message() *service.Message {
	return b.message
}

var _ javascript_functions.ValueApi = (*anonValueApi)(nil)

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
