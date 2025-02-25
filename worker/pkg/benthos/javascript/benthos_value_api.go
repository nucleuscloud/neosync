package javascript_processor

import (
	"fmt"
	"sync"

	javascript_functions "github.com/nucleuscloud/neosync/internal/javascript/functions"
	"github.com/redpanda-data/benthos/v4/public/service"
)

// benthosValueApi is thread safe
type benthosValueApi struct {
	message *service.Message
	mu      sync.RWMutex // Mutex to protect access to message
}

func newBatchBenthosValueApi() *benthosValueApi {
	return &benthosValueApi{}
}

// used by batch processor to update the target message while being able to reuse the same VM
func (b *benthosValueApi) SetMessage(message *service.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.message = message
}

func (b *benthosValueApi) Message() *service.Message {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.message
}

var _ javascript_functions.ValueApi = (*benthosValueApi)(nil)

func (b *benthosValueApi) SetBytes(bytes []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.message == nil {
		return
	}
	b.message.SetBytes(bytes)
}

func (b *benthosValueApi) AsBytes() ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.message == nil {
		return nil, fmt.Errorf("message is nil")
	}
	return b.message.AsBytes()
}

func (b *benthosValueApi) SetStructured(value any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.message == nil {
		return
	}
	b.message.SetStructured(value)
}

func (b *benthosValueApi) AsStructured() (any, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.message == nil {
		return nil, fmt.Errorf("message is nil")
	}
	return b.message.AsStructured()
}

func (b *benthosValueApi) MetaGet(key string) (any, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.message == nil {
		return nil, false
	}
	return b.message.MetaGet(key)
}

func (b *benthosValueApi) MetaSetMut(key string, value any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.message == nil {
		return
	}
	b.message.MetaSetMut(key, value)
}
