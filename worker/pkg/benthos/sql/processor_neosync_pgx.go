package neosync_benthos_sql

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	"github.com/warpstreamlabs/bento/public/service"
)

func neosyncToPgxProcessorConfig() *service.ConfigSpec {
	return service.NewConfigSpec().Field(service.NewStringMapField("column_data_types"))
}

func RegisterNeosyncToPgxProcessor(env *service.Environment) error {
	return env.RegisterBatchProcessor(
		"neosync_to_pgx",
		neosyncToPgxProcessorConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchProcessor, error) {
			proc := newNeosyncToPgxProcessor(conf, mgr)
			return proc, nil
		})
}

type neosyncToPgxProcessor struct {
	logger *service.Logger
}

func newNeosyncToPgxProcessor(_ *service.ParsedConfig, mgr *service.Resources) *neosyncToPgxProcessor {
	return &neosyncToPgxProcessor{
		logger: mgr.Logger(),
	}
}

func (p *neosyncToPgxProcessor) ProcessBatch(ctx context.Context, batch service.MessageBatch) ([]service.MessageBatch, error) {
	newBatch := make(service.MessageBatch, 0, len(batch))
	for _, msg := range batch {
		root, err := msg.AsStructuredMut()
		if err != nil {
			return nil, err
		}
		newRoot := p.transform(root)
		newMsg := msg.Copy()
		newMsg.SetStructured(newRoot)
		newBatch = append(newBatch, newMsg)
	}

	if len(newBatch) == 0 {
		return nil, nil
	}
	return []service.MessageBatch{newBatch}, nil
}

func (m *neosyncToPgxProcessor) Close(context.Context) error {
	return nil
}

func (p *neosyncToPgxProcessor) transform(root any) any {
	switch v := root.(type) {
	case map[string]any:
		newMap := make(map[string]any)
		for k, v2 := range v {
			newValue := p.transform(v2)
			newMap[k] = newValue
		}
		return newMap
	case nil:
		return v
	default:
		// Check if the type implements Value() method
		if valuer, ok := v.(neosynctypes.NeosyncPgxValuer); ok {
			value, err := valuer.ValuePgx()
			if err != nil {
				p.logger.Warn(fmt.Sprintf("unable to get PGX value: %v", err))
				return v
			}
			if gotypeutil.IsSlice(value) {
				return pq.Array(value)
			}
			return value
		}

		return v
	}
}
