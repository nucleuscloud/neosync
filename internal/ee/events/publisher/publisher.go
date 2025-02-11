package events_publisher

import (
	"context"

	accounthook_events "github.com/nucleuscloud/neosync/internal/ee/events"
)

type Interface interface {
	Publish(ctx context.Context, event *accounthook_events.Event) error
}

type WorkflowPublisher struct{}

func (p *WorkflowPublisher) Publish(ctx context.Context, event *accounthook_events.Event) error {
	return nil
}
