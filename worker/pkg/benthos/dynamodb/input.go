package neosync_benthos_dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	awsmanager "github.com/nucleuscloud/neosync/internal/aws"
	"github.com/warpstreamlabs/bento/public/service"
)

const (
	metaTypeMapStr = "neosync_key_type_map"
)

type KeyType int

const (
	StringSet KeyType = iota
	NumberSet
)

func dynamoInputConfigSpec() *service.ConfigSpec {
	spec := service.NewConfigSpec().
		Categories("Services").
		Summary("Scans an entire dynamodb table and creates a message for each document received").
		Field(service.NewStringField("table").
			Description("The table to retrieve items from.")).
		Field(service.NewStringField("where").
			Description("Optional PartiQL where clause that gets tacked on to the end of the select query").
			Optional())

	for _, f := range awsSessionFields() {
		spec = spec.Field(f)
	}

	return spec
}

func RegisterDynamoDbInput(env *service.Environment) error {
	return env.RegisterBatchInput(
		"aws_dynamodb", dynamoInputConfigSpec(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return newDynamoDbBatchInput(conf, mgr.Logger())
		},
	)
}

type dynamoDBAPIV2 interface {
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	ExecuteStatement(ctx context.Context, params *dynamodb.ExecuteStatementInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ExecuteStatementOutput, error)
}

func newDynamoDbBatchInput(conf *service.ParsedConfig, logger *service.Logger) (service.BatchInput, error) {
	table, err := conf.FieldString("table")
	if err != nil {
		return nil, err
	}

	var whereClause *string
	if conf.Contains("where") {
		where, err := conf.FieldString("where")
		if err != nil {
			return nil, err
		}
		whereClause = &where
	}

	sess, err := getAwsSession(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	return &dynamodbInput{
		awsConfig: *sess,
		logger:    logger,

		table: table,
		where: whereClause,
	}, nil
}

type dynamodbInput struct {
	client    dynamoDBAPIV2 // lazy
	awsConfig aws.Config
	logger    *service.Logger
	readMu    sync.Mutex

	table string
	where *string

	nextToken *string
	done      bool
}

var _ service.BatchInput = &dynamodbInput{}

func (d *dynamodbInput) Connect(ctx context.Context) error {
	d.readMu.Lock()
	defer d.readMu.Unlock()

	if d.client != nil {
		return nil
	}

	client := dynamodb.NewFromConfig(d.awsConfig)

	tableOutput, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &d.table,
	})
	if err != nil {
		return fmt.Errorf("unable to describe dynamodb table when connecting to read: %w", err)
	}
	if !isTableActive(tableOutput) {
		return fmt.Errorf("dynamodb table %q must be active to read", d.table)
	}

	d.client = client
	return nil
}

func isTableActive(output *dynamodb.DescribeTableOutput) bool {
	return output != nil && output.Table != nil && output.Table.TableStatus == types.TableStatusActive
}

func (d *dynamodbInput) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	d.readMu.Lock()
	defer d.readMu.Unlock()
	if d.client == nil {
		return nil, nil, service.ErrNotConnected
	}
	if d.done {
		return nil, nil, service.ErrEndOfInput
	}

	// todo: allow specifying batch size
	result, err := d.client.ExecuteStatement(ctx, &dynamodb.ExecuteStatementInput{
		Statement:      aws.String(buildExecStatement(d.table, d.where)),
		NextToken:      d.nextToken,
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return nil, nil, err
	}
	batch := service.MessageBatch{}
	for _, item := range result.Items {
		if item == nil {
			continue
		}

		resMap, keyTypeMap := attributeValueMapToStandardJSON(item)
		msg := service.NewMessage(nil)
		msg.MetaSetMut(metaTypeMapStr, keyTypeMap)
		msg.SetStructuredMut(resMap)
		batch = append(batch, msg)
	}
	d.nextToken = result.NextToken
	d.done = result.NextToken == nil

	return batch, emptyAck, nil
}

func buildExecStatement(table string, where *string) string {
	stmt := fmt.Sprintf("SELECT * FROM %q", table)
	if where != nil && *where != "" {
		return fmt.Sprintf("%s WHERE %s", stmt, *where)
	}
	return stmt
}

func emptyAck(ctx context.Context, err error) error {
	return nil
}

func (d *dynamodbInput) Close(ctx context.Context) error {
	d.readMu.Lock()
	defer d.readMu.Unlock()
	if d.client == nil {
		return nil
	}
	d.client = nil
	return nil
}

func getAwsSession(ctx context.Context, parsedConf *service.ParsedConfig, opts ...func(*config.LoadOptions) error) (*aws.Config, error) {
	awsCfg, err := awsmanager.GetAwsConfig(ctx, getAwsCredentialsConfigFromParsedConf(parsedConf), opts...)
	if err != nil {
		return aws.NewConfig(), err
	}
	return awsCfg, nil
}

func getAwsCredentialsConfigFromParsedConf(parsedConf *service.ParsedConfig) *awsmanager.AwsCredentialsConfig {
	output := &awsmanager.AwsCredentialsConfig{}
	if parsedConf == nil {
		return output
	}
	region, _ := parsedConf.FieldString("region")
	output.Region = region

	endpoint, _ := parsedConf.FieldString("endpoint")
	output.Endpoint = endpoint

	useEc2, _ := parsedConf.FieldBool("from_ec2_role")
	output.UseEc2 = useEc2

	role, _ := parsedConf.FieldString("role")
	output.Role = role

	roleExternalId, _ := parsedConf.FieldString("role_external_id")
	output.RoleExternalId = roleExternalId

	credsConf := parsedConf.Namespace("credentials")
	profile, _ := credsConf.FieldString("profile")
	output.Profile = profile

	id, _ := credsConf.FieldString("id")
	output.Id = id

	secret, _ := credsConf.FieldString("secret")
	output.Secret = secret

	token, _ := credsConf.FieldString("token")
	output.Token = token

	return output
}

// SessionFields defines a re-usable set of config fields for an AWS session
// that is compatible with the public service APIs and avoids importing the full
// AWS dependencies.
func awsSessionFields() []*service.ConfigField {
	return []*service.ConfigField{
		service.NewStringField("region").
			Description("The AWS region to target.").
			Default("").
			Advanced(),
		service.NewStringField("endpoint").
			Description("Allows you to specify a custom endpoint for the AWS API.").
			Default("").
			Advanced(),
		service.NewObjectField("credentials",
			service.NewStringField("profile").
				Description("A profile from `~/.aws/credentials` to use.").
				Default(""),
			service.NewStringField("id").
				Description("The ID of credentials to use.").
				Default("").Advanced(),
			service.NewStringField("secret").
				Description("The secret for the credentials being used.").
				Default("").Advanced().Secret(),
			service.NewStringField("token").
				Description("The token for the credentials being used, required when using short term credentials.").
				Default("").Advanced(),
			service.NewBoolField("from_ec2_role").
				Description("Use the credentials of a host EC2 machine configured to assume [an IAM role associated with the instance](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2.html).").
				Default(false).Version("4.2.0"),
			service.NewStringField("role").
				Description("A role ARN to assume.").
				Default("").Advanced(),
			service.NewStringField("role_external_id").
				Description("An external ID to provide when assuming a role.").
				Default("").Advanced()).
			Advanced().
			Description("Optional manual configuration of AWS credentials to use. More information can be found [in this document](/docs/guides/cloud/aws)."),
	}
}

func attributeValueMapToStandardJSON(item map[string]types.AttributeValue) (standardMap map[string]any, keyTypeMap map[string]KeyType) {
	standardJSON := make(map[string]any)
	ktm := make(map[string]KeyType)
	for k, v := range item {
		val := attributeValueToStandardValue(k, v, ktm)
		standardJSON[k] = val
	}
	return standardJSON, ktm
}

// attributeValueToStandardValue converts a DynamoDB AttributeValue to a standard value
func attributeValueToStandardValue(key string, v types.AttributeValue, keyTypeMap map[string]KeyType) any {
	switch t := v.(type) {
	case *types.AttributeValueMemberB:
		return t.Value
	case *types.AttributeValueMemberBOOL:
		return t.Value
	case *types.AttributeValueMemberBS:
		return t.Value
	case *types.AttributeValueMemberL:
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			val := attributeValueToStandardValue("", v, keyTypeMap)
			lAny[i] = val
		}
		return lAny
	case *types.AttributeValueMemberM:
		mAny := make(map[string]any, len(t.Value))
		for k, v := range t.Value {
			val := attributeValueToStandardValue(k, v, keyTypeMap)
			mAny[k] = val
		}
		return mAny
	case *types.AttributeValueMemberN:
		n, err := convertStringToNumber(t.Value)
		if err != nil {
			return t.Value
		}
		return n
	case *types.AttributeValueMemberNS:
		keyTypeMap[key] = NumberSet
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			lAny[i] = v
		}
		return lAny
	case *types.AttributeValueMemberNULL:
		return nil
	case *types.AttributeValueMemberS:
		return t.Value
	case *types.AttributeValueMemberSS:
		keyTypeMap[key] = StringSet
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			lAny[i] = v
		}
		return lAny
	}
	return nil
}

func convertStringToNumber(s string) (any, error) {
	if i, err := strconv.Atoi(s); err == nil {
		return i, nil
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, nil
	}

	return nil, errors.New("input string is neither a valid int nor a float")
}
