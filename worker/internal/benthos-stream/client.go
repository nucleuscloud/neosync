package benthosstream

import (
	"context"
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

func (b *BenthosStreamManager) NewBenthosStreamFromBuilder(streambldr *service.StreamBuilder) (BenthosStreamClient, error) {
	stream, err := streambldr.Build()
	if err != nil {
		return nil, err
	}
	return NewBenthosStreamAdapter(stream), nil
}

type BenthosStreamAdapter struct {
	Stream *service.Stream
}

func NewBenthosStreamAdapter(stream *service.Stream) *BenthosStreamAdapter {
	return &BenthosStreamAdapter{
		Stream: stream,
	}
}

func (b *BenthosStreamAdapter) Run(ctx context.Context) error {
	return b.Stream.Run(ctx)
}

func (b *BenthosStreamAdapter) Stop(ctx context.Context) error {
	return b.Stream.Stop(ctx)
}

func (b *BenthosStreamAdapter) StopWithin(d time.Duration) error {
	return b.Stream.StopWithin(d)
}
