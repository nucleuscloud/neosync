package benthos_redis

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"
)

const (
	hoFieldKey           = "key"
	hoFieldWalkMetadata  = "walk_metadata"
	hoFieldWalkJSON      = "walk_json_object"
	hoFieldFieldsMapping = "fields_mapping"
)

func redisHashOutputConfig() *service.ConfigSpec {
	return service.NewConfigSpec().
		Stable().
		Summary(`Sets Redis hash objects using the HMSET command.`).
		Categories("Services").
		Fields(clientFields()...).
		Fields(
			service.NewInterpolatedStringField(hoFieldKey).
				Description("The key for each message, function interpolations should be used to create a unique key per message.").
				Examples("${! @.kafka_key )}", "${! this.doc.id }", "${! count(\"msgs\") }"),
			service.NewBoolField(hoFieldWalkMetadata).
				Description("Whether all metadata fields of messages should be walked and added to the list of hash fields to set.").
				Default(false),
			service.NewBoolField(hoFieldWalkJSON).
				Description("Whether to walk each message as a JSON object and add each key/value pair to the list of hash fields to set.").
				Default(false),
			service.NewBloblangField(hoFieldFieldsMapping),
			service.NewOutputMaxInFlightField(),
		)
}

func init() {
	err := service.RegisterOutput(
		"redis_hash_output", redisHashOutputConfig(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (out service.Output, maxInFlight int, err error) {
			if maxInFlight, err = conf.FieldMaxInFlight(); err != nil {
				return
			}
			out, err = newRedisHashWriter(conf, mgr)
			return
		})
	if err != nil {
		panic(err)
	}
}

type redisHashWriter struct {
	log *service.Logger

	key           *service.InterpolatedString
	walkMetadata  bool
	walkJSON      bool
	fieldsMapping *bloblang.Executor

	clientCtor func() (redis.UniversalClient, error)
	client     redis.UniversalClient
	connMut    sync.RWMutex
}

func newRedisHashWriter(conf *service.ParsedConfig, mgr *service.Resources) (r *redisHashWriter, err error) {
	r = &redisHashWriter{
		clientCtor: func() (redis.UniversalClient, error) {
			return getClient(conf)
		},
		log: mgr.Logger(),
	}
	if _, err = getClient(conf); err != nil {
		return
	}

	if r.key, err = conf.FieldInterpolatedString(hoFieldKey); err != nil {
		return
	}
	if r.walkMetadata, err = conf.FieldBool(hoFieldWalkMetadata); err != nil {
		return
	}
	if r.walkJSON, err = conf.FieldBool(hoFieldWalkJSON); err != nil {
		return
	}
	if r.fieldsMapping, err = conf.FieldBloblang(hoFieldFieldsMapping); err != nil {
		return nil, err
	}

	if !r.walkMetadata && !r.walkJSON && r.fieldsMapping == nil {
		return nil, errors.New("at least one mechanism for setting fields must be enabled")
	}
	return
}

func (r *redisHashWriter) Connect(ctx context.Context) error {
	r.connMut.Lock()
	defer r.connMut.Unlock()

	client, err := r.clientCtor()
	if err != nil {
		return err
	}
	if _, err = client.Ping(ctx).Result(); err != nil {
		return err
	}

	r.log.Info("Setting messages as hash objects to Redis")
	r.client = client
	return nil
}

//------------------------------------------------------------------------------

func walkForHashFields(msg *service.Message, fields map[string]any) error {
	jVal, err := msg.AsStructured()
	if err != nil {
		return err
	}
	jObj, ok := jVal.(map[string]any)
	if !ok {
		return fmt.Errorf("expected JSON object, found '%T'", jVal)
	}
	for k, v := range jObj {
		fields[k] = v
	}
	return nil
}

func (r *redisHashWriter) Write(ctx context.Context, msg *service.Message) error {
	r.connMut.RLock()
	client := r.client
	r.connMut.RUnlock()

	if client == nil {
		return service.ErrNotConnected
	}

	key, err := r.key.TryString(msg)
	if err != nil {
		return fmt.Errorf("key interpolation error: %w", err)
	}
	fields := map[string]any{}
	if r.walkMetadata {
		_ = msg.MetaWalkMut(func(k string, v any) error {
			fields[k] = v
			return nil
		})
	}
	if r.walkJSON {
		if err := walkForHashFields(msg, fields); err != nil {
			err = fmt.Errorf("failed to walk JSON object: %v", err)
			r.log.Errorf("HMSET error: %v\n", err)
			return err
		}
	}

	if r.fieldsMapping != nil {
		mapMsg, err := msg.BloblangQuery(r.fieldsMapping)
		if err != nil {
			return err
		}

		var mapVal any
		if mapMsg != nil {
			if mapVal, err = mapMsg.AsStructured(); err != nil {
				return err
			}
		}

		if mapVal != nil {
			fieldMappings, ok := mapVal.(map[string]interface{}) //nolint:gofmt
			if !ok {
				return fmt.Errorf("fieldMappings resulted in a non-object mapping: %T", mapVal)
			}
			for k, v := range fieldMappings {
				fields[k] = v
			}
		}
	}

	if err := client.HMSet(ctx, key, fields).Err(); err != nil {
		_ = r.disconnect()
		r.log.Errorf("Error from redis: %v\n", err)
		return service.ErrNotConnected
	}
	return nil
}

func (r *redisHashWriter) disconnect() error {
	r.connMut.Lock()
	defer r.connMut.Unlock()
	if r.client != nil {
		err := r.client.Close()
		r.client = nil
		return err
	}
	return nil
}

func (r *redisHashWriter) Close(context.Context) error {
	return r.disconnect()
}
