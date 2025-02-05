package neosync_benthos_dynamodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jeffail/gabs/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cenkalti/backoff/v4"

	neosync_types "github.com/nucleuscloud/neosync/internal/types"
	neosync_benthos_metadata "github.com/nucleuscloud/neosync/worker/pkg/benthos/metadata"
	"github.com/redpanda-data/benthos/v4/public/service"
)

const (
	// DynamoDB Output Fields
	ddboField               = "namespace"
	ddboFieldTable          = "table"
	ddboFieldStringColumns  = "string_columns"
	ddboFieldJSONMapColumns = "json_map_columns"
	ddboFieldTTL            = "ttl"
	ddboFieldTTLKey         = "ttl_key"
	ddboFieldBatching       = "batching"

	crboFieldMaxRetries     = "max_retries"
	crboFieldBackOff        = "backoff"
	crboFieldInitInterval   = "initial_interval"
	crboFieldMaxInterval    = "max_interval"
	crboFieldMaxElapsedTime = "max_elapsed_time"
)

type ddboConfig struct {
	Table          string
	StringColumns  map[string]*service.InterpolatedString
	JSONMapColumns map[string]string
	TTL            string
	TTLKey         string

	backoffCtor func() backoff.BackOff
	awsConfig   aws.Config
}

func ddboConfigFromParsed(pConf *service.ParsedConfig) (conf *ddboConfig, err error) {
	c := &ddboConfig{}
	if c.Table, err = pConf.FieldString(ddboFieldTable); err != nil {
		return
	}
	if c.StringColumns, err = pConf.FieldInterpolatedStringMap(ddboFieldStringColumns); err != nil {
		return
	}
	if c.JSONMapColumns, err = pConf.FieldStringMap(ddboFieldJSONMapColumns); err != nil {
		return
	}
	if c.TTL, err = pConf.FieldString(ddboFieldTTL); err != nil {
		return
	}
	if c.TTLKey, err = pConf.FieldString(ddboFieldTTLKey); err != nil {
		return
	}
	if c.backoffCtor, err = commonRetryBackOffCtorFromParsed(pConf); err != nil {
		return
	}
	sess, err := getAwsSession(context.Background(), pConf)
	if err != nil {
		return
	}
	c.awsConfig = *sess
	return c, nil
}

func dynamoOutputConfigSpec() *service.ConfigSpec {
	spec := service.NewConfigSpec().
		Stable().
		Version("1.0.0").
		Categories("Services", "AWS").
		Summary(`Inserts items into a DynamoDB table.`).
		Fields(
			service.NewStringField(ddboFieldTable).
				Description("The table to store messages in."),
			service.NewInterpolatedStringMapField(ddboFieldStringColumns).
				Description("A map of column keys to string values to store.").
				Default(map[string]any{}).
				Example(map[string]any{
					"id":           "${!json(\"id\")}",
					"title":        "${!json(\"body.title\")}",
					"topic":        "${!meta(\"kafka_topic\")}",
					"full_content": "${!content()}",
				}),
			service.NewStringMapField(ddboFieldJSONMapColumns).
				Description("A map of column keys to [field paths](/docs/configuration/field_paths) pointing to value data within messages.").
				Default(map[string]any{}).
				Example(map[string]any{
					"user":           "path.to.user",
					"whole_document": ".",
				}).
				Example(map[string]string{
					"": ".",
				}),
			service.NewStringField(ddboFieldTTL).
				Description("An optional TTL to set for items, calculated from the moment the message is sent.").
				Default("").
				Advanced(),
			service.NewStringField(ddboFieldTTLKey).
				Description("The column key to place the TTL value within.").
				Default("").
				Advanced(),
			service.NewOutputMaxInFlightField(),
			service.NewBatchPolicyField(ddboFieldBatching),
		).
		Fields(commonRetryBackOffFields(3, "1s", "5s", "30s")...)
	for _, f := range awsSessionFields() {
		spec = spec.Field(f)
	}
	return spec
}

func RegisterDynamoDbOutput(env *service.Environment) error {
	return env.RegisterBatchOutput("aws_dynamodb", dynamoOutputConfigSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (out service.BatchOutput, batchPolicy service.BatchPolicy, maxInFlight int, err error) {
			if maxInFlight, err = conf.FieldMaxInFlight(); err != nil {
				return
			}
			if batchPolicy, err = conf.FieldBatchPolicy(ddboFieldBatching); err != nil {
				return
			}
			var wConf *ddboConfig
			if wConf, err = ddboConfigFromParsed(conf); err != nil {
				return
			}
			out, err = newDynamoDBWriter(wConf, mgr)
			return
		})
}

type dynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	BatchExecuteStatement(ctx context.Context, params *dynamodb.BatchExecuteStatementInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchExecuteStatementOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

type dynamoDBWriter struct {
	client dynamoDBAPI

	conf ddboConfig
	log  *service.Logger

	boffPool sync.Pool

	table *string
	ttl   time.Duration
}

func newDynamoDBWriter(conf *ddboConfig, mgr *service.Resources) (*dynamoDBWriter, error) {
	db := &dynamoDBWriter{
		conf:  *conf,
		log:   mgr.Logger(),
		table: aws.String(conf.Table),
	}
	if len(conf.StringColumns) == 0 && len(conf.JSONMapColumns) == 0 {
		return nil, errors.New("you must provide at least one column")
	}
	for k, v := range conf.JSONMapColumns {
		if v == "." {
			conf.JSONMapColumns[k] = ""
		}
	}
	if conf.TTL != "" {
		ttl, err := time.ParseDuration(conf.TTL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TTL: %v", err)
		}
		db.ttl = ttl
	}
	db.boffPool = sync.Pool{
		New: func() any {
			return db.conf.backoffCtor()
		},
	}
	return db, nil
}

func (d *dynamoDBWriter) Connect(ctx context.Context) error {
	if d.client != nil {
		return nil
	}

	client := dynamodb.NewFromConfig(d.conf.awsConfig)
	out, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: d.table,
	})
	if err != nil {
		return err
	} else if out == nil || out.Table == nil || out.Table.TableStatus != types.TableStatusActive {
		return fmt.Errorf("dynamodb table '%s' must be active", d.conf.Table)
	}

	d.client = client
	return nil
}

func (d *dynamoDBWriter) WriteBatch(ctx context.Context, b service.MessageBatch) error {
	if d.client == nil {
		return service.ErrNotConnected
	}

	boff := d.boffPool.Get().(backoff.BackOff)
	defer func() {
		boff.Reset()
		d.boffPool.Put(boff)
	}()

	writeReqs := []types.WriteRequest{}
	if err := b.WalkWithBatchedErrors(func(i int, p *service.Message) error {
		keyTypeMap, err := getKeyTypMap(p)
		if err != nil {
			return err
		}
		items := map[string]types.AttributeValue{}
		if d.ttl != 0 && d.conf.TTLKey != "" {
			items[d.conf.TTLKey] = &types.AttributeValueMemberN{
				Value: strconv.FormatInt(time.Now().Add(d.ttl).Unix(), 10),
			}
		}
		for k, v := range d.conf.StringColumns {
			s, err := b.TryInterpolatedString(i, v)
			if err != nil {
				return fmt.Errorf("string column %v interpolation error: %w", k, err)
			}
			items[k] = &types.AttributeValueMemberS{
				Value: s,
			}
		}
		if len(d.conf.JSONMapColumns) > 0 {
			jRoot, err := p.AsStructured()
			if err != nil {
				d.log.Errorf("Failed to extract JSON maps from document: %v", err)
				return err
			}
			for k, v := range d.conf.JSONMapColumns {
				attr := marshalJSONToDynamoDBAttribute(k, v, jRoot, keyTypeMap)
				if k == "" {
					if mv, ok := attr.(*types.AttributeValueMemberM); ok {
						for ak, av := range mv.Value {
							items[ak] = av
						}
					} else {
						items[k] = attr
					}
				} else {
					items[k] = attr
				}
			}
		}
		writeReqs = append(writeReqs, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: items,
			},
		})
		return nil
	}); err != nil {
		return err
	}

	batchResult, err := d.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			*d.table: writeReqs,
		},
	})
	if err != nil {
		headlineErr := err

		// None of the messages were successful, attempt to send individually
	individualRequestsLoop:
		for err != nil {
			batchErr := service.NewBatchError(b, headlineErr)
			for i, req := range writeReqs {
				if req.PutRequest == nil {
					continue
				}
				if _, iErr := d.client.PutItem(ctx, &dynamodb.PutItemInput{
					TableName: d.table,
					Item:      req.PutRequest.Item,
				}); iErr != nil {
					d.log.Errorf("Put error: %v\n", iErr)
					wait := boff.NextBackOff()
					if wait == backoff.Stop {
						break individualRequestsLoop
					}
					select {
					case <-time.After(wait):
					case <-ctx.Done():
						break individualRequestsLoop
					}
					err = batchErr.Failed(i, iErr)
				} else {
					writeReqs[i].PutRequest = nil
				}
			}
			if batchErr.IndexedErrors() == 0 {
				err = nil
			} else {
				err = batchErr
			}
		}
		return err
	}

	unproc := batchResult.UnprocessedItems[*d.table]
unprocessedLoop:
	for len(unproc) > 0 {
		wait := boff.NextBackOff()
		if wait == backoff.Stop {
			break unprocessedLoop
		}

		select {
		case <-time.After(wait):
		case <-ctx.Done():
			break unprocessedLoop
		}
		if batchResult, err = d.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				*d.table: unproc,
			},
		}); err != nil {
			d.log.Errorf("Write multi error: %v\n", err)
		} else if unproc = batchResult.UnprocessedItems[*d.table]; len(unproc) > 0 {
			err = fmt.Errorf("failed to set %v items", len(unproc))
		} else {
			unproc = nil
		}
	}

	if len(unproc) > 0 {
		if err == nil {
			err = errors.New("ran out of request retries")
		}
	}
	return err
}

func (d *dynamoDBWriter) Close(context.Context) error {
	return nil
}

func getKeyTypMap(p *service.Message) (map[string]neosync_types.KeyType, error) {
	keyTypeMap := map[string]neosync_types.KeyType{}
	meta, ok := p.MetaGetMut(neosync_benthos_metadata.MetaTypeMapStr)
	if ok {
		kt, err := convertToMapStringKeyType(meta)
		if err != nil {
			return nil, err
		}
		keyTypeMap = kt
	}
	return keyTypeMap, nil
}

func convertToMapStringKeyType(i any) (map[string]neosync_types.KeyType, error) {
	if m, ok := i.(map[string]neosync_types.KeyType); ok {
		return m, nil
	}

	return nil, errors.New("input is not of type map[string]KeyType")
}

func commonRetryBackOffFields(
	defaultMaxRetries int,
	defaultInitInterval string,
	defaultMaxInterval string,
	defaultMaxElapsed string,
) []*service.ConfigField {
	return []*service.ConfigField{
		service.NewIntField(crboFieldMaxRetries).
			Description("The maximum number of retries before giving up on the request. If set to zero there is no discrete limit.").
			Default(defaultMaxRetries).
			Advanced(),
		service.NewObjectField(crboFieldBackOff,
			service.NewDurationField(crboFieldInitInterval).
				Description("The initial period to wait between retry attempts.").
				Default(defaultInitInterval),
			service.NewDurationField(crboFieldMaxInterval).
				Description("The maximum period to wait between retry attempts.").
				Default(defaultMaxInterval),
			service.NewDurationField(crboFieldMaxElapsedTime).
				Description("The maximum period to wait before retry attempts are abandoned. If zero then no limit is used.").
				Default(defaultMaxElapsed),
		).
			Description("Control time intervals between retry attempts.").
			Advanced(),
	}
}

func commonRetryBackOffCtorFromParsed(pConf *service.ParsedConfig) (ctor func() backoff.BackOff, err error) {
	var maxRetries int
	if maxRetries, err = pConf.FieldInt(crboFieldMaxRetries); err != nil {
		return nil, err
	}

	var initInterval, maxInterval, maxElapsed time.Duration
	if pConf.Contains(crboFieldBackOff) {
		bConf := pConf.Namespace(crboFieldBackOff)
		if initInterval, err = fieldDurationOrEmptyStr(bConf, crboFieldInitInterval); err != nil {
			return nil, err
		}
		if maxInterval, err = fieldDurationOrEmptyStr(bConf, crboFieldMaxInterval); err != nil {
			return nil, err
		}
		if maxElapsed, err = fieldDurationOrEmptyStr(bConf, crboFieldMaxElapsedTime); err != nil {
			return nil, err
		}
	}

	return func() backoff.BackOff {
		boff := backoff.NewExponentialBackOff()

		boff.InitialInterval = initInterval
		boff.MaxInterval = maxInterval
		boff.MaxElapsedTime = maxElapsed

		if maxRetries > 0 {
			return backoff.WithMaxRetries(boff, uint64(maxRetries))
		}
		return boff
	}, nil
}
func fieldDurationOrEmptyStr(pConf *service.ParsedConfig, path ...string) (time.Duration, error) {
	if dStr, err := pConf.FieldString(path...); err == nil && dStr == "" {
		return 0, nil
	}
	return pConf.FieldDuration(path...)
}

func marshalToAttributeValue(key string, root any, keyTypeMap map[string]neosync_types.KeyType) types.AttributeValue {
	if typeStr, ok := keyTypeMap[key]; ok {
		switch typeStr {
		case neosync_types.StringSet:
			s, err := convertToStringSlice(root)
			if err == nil {
				return &types.AttributeValueMemberSS{
					Value: s,
				}
			}
		case neosync_types.NumberSet:
			s, err := convertToStringSlice(root)
			if err == nil {
				return &types.AttributeValueMemberNS{
					Value: s,
				}
			}
		}
	}
	switch v := root.(type) {
	case map[string]any:
		m := make(map[string]types.AttributeValue, len(v))
		for k, v2 := range v {
			path := k
			if key != "" {
				path = fmt.Sprintf("%s.%s", key, k)
			}
			m[k] = marshalToAttributeValue(path, v2, keyTypeMap)
		}
		return &types.AttributeValueMemberM{
			Value: m,
		}
	case []byte:
		return &types.AttributeValueMemberB{
			Value: v,
		}
	case [][]byte:
		return &types.AttributeValueMemberBS{
			Value: v,
		}
	case []any:
		l := make([]types.AttributeValue, len(v))
		for i, v2 := range v {
			l[i] = marshalToAttributeValue(fmt.Sprintf("%s[%d]", key, i), v2, keyTypeMap)
		}
		return &types.AttributeValueMemberL{
			Value: l,
		}
	case string:
		return &types.AttributeValueMemberS{
			Value: v,
		}
	case json.Number:
		return &types.AttributeValueMemberS{
			Value: v.String(),
		}
	case float64:
		return &types.AttributeValueMemberN{
			Value: formatFloat(v),
		}
	case int:
		return &types.AttributeValueMemberN{
			Value: strconv.Itoa(v),
		}
	case int64:
		return &types.AttributeValueMemberN{
			Value: strconv.Itoa(int(v)),
		}
	case bool:
		return &types.AttributeValueMemberBOOL{
			Value: v,
		}
	case nil:
		return &types.AttributeValueMemberNULL{
			Value: true,
		}
	}
	return &types.AttributeValueMemberS{
		Value: fmt.Sprintf("%v", root),
	}
}

func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 4, 64)
	s = strings.TrimRight(s, "0")
	if strings.HasSuffix(s, ".") {
		s += "0"
	}
	return s
}

func marshalJSONToDynamoDBAttribute(key, path string, root any, keyTypeMap map[string]neosync_types.KeyType) types.AttributeValue {
	gObj := gabs.Wrap(root)
	if path != "" {
		gObj = gObj.Path(path)
	}
	return marshalToAttributeValue(key, gObj.Data(), keyTypeMap)
}

func convertToStringSlice(slice any) ([]string, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("input is not a slice")
	}

	result := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i).Interface()
		result[i] = anyToString(elem)
	}

	return result, nil
}

func anyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case float32, float64:
		return formatFloat(reflect.ValueOf(v).Float())
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}
