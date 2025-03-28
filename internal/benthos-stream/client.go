package benthosstream

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redpanda-data/benthos/v4/public/service"
)

type BenthosStreamClient interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	StopWithin(d time.Duration) error
}

type BenthosStreamManagerClient interface {
	NewBenthosStreamFromBuilder(streambldr *service.StreamBuilder) (BenthosStreamClient, error)
}

type BenthosStreamManager struct{}

func NewBenthosStreamManager() *BenthosStreamManager {
	return &BenthosStreamManager{}
}

func (b *BenthosStreamManager) NewBenthosStreamFromBuilder(
	streambldr *service.StreamBuilder,
) (BenthosStreamClient, error) {
	stream, err := streambldr.Build()
	if err != nil {
		return nil, err
	}
	return NewBenthosStreamAdapter(stream), nil
}

type BenthosStreamAdapter struct {
	mu     *sync.RWMutex
	Stream *service.Stream
}

func NewBenthosStreamAdapter(stream *service.Stream) *BenthosStreamAdapter {
	return &BenthosStreamAdapter{
		Stream: stream,
		mu:     &sync.RWMutex{},
	}
}

func (b *BenthosStreamAdapter) Run(ctx context.Context) error {
	b.mu.RLock()
	stream := b.Stream
	b.mu.RUnlock()

	if stream == nil {
		return fmt.Errorf("benthos stream is nil during Run")
	}
	return stream.Run(ctx)
}

func (b *BenthosStreamAdapter) Stop(ctx context.Context) error {
	b.mu.Lock()
	stream := b.Stream
	b.Stream = nil
	b.mu.Unlock()

	if stream == nil {
		return nil
	}
	return stream.Stop(ctx)
}

func (b *BenthosStreamAdapter) StopWithin(d time.Duration) error {
	b.mu.Lock()
	stream := b.Stream
	b.Stream = nil
	b.mu.Unlock()

	if stream == nil {
		return nil
	}
	return stream.StopWithin(d)
}
