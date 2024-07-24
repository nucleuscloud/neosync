package neosync_benthos_dynamodb

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/warpstreamlabs/bento/public/service"
)

func dynamoInputConfigSpec() *service.ConfigSpec {
	spec := service.NewConfigSpec().
		Categories("Services").
		Summary("Scans an entire dynamodb table and creates a message for each document received").
		Field(service.NewStringField("table").
			Description("The table to retrieve items from."))

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
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func newDynamoDbBatchInput(conf *service.ParsedConfig, logger *service.Logger) (service.BatchInput, error) {
	table, err := conf.FieldString("table")
	if err != nil {
		return nil, err
	}

	sess, err := getAwsSession(context.Background(), conf)
	if err != nil {
		return nil, err
	}

	return &dynamodbInput{
		awsConfig: sess,
		logger:    logger,

		table: table,
	}, nil
}

type dynamodbInput struct {
	client    dynamoDBAPIV2 // lazy
	awsConfig aws.Config
	logger    *service.Logger
	readMu    sync.Mutex

	table            string
	lastEvaluatedKey map[string]types.AttributeValue
	done             bool
}

var _ service.BatchInput = &dynamodbInput{}

func (d *dynamodbInput) Connect(ctx context.Context) error {
	d.readMu.Lock()
	defer d.readMu.Unlock()

	if d.client == nil {
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
	result, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:         &d.table,
		ExclusiveStartKey: d.lastEvaluatedKey,
		ConsistentRead:    aws.Bool(true),
	})
	if err != nil {
		return nil, nil, err
	}
	batch := service.MessageBatch{}
	for _, item := range result.Items {
		if item == nil {
			continue
		}

		resMap := attributeValueMapToStandardJSON(item)
		msg := service.NewMessage(nil)
		msg.SetStructuredMut(resMap)
		batch = append(batch, msg)
	}
	d.lastEvaluatedKey = result.LastEvaluatedKey
	d.done = result.LastEvaluatedKey == nil

	return batch, emptyAck, nil
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

func getAwsSession(ctx context.Context, parsedConf *service.ParsedConfig, opts ...func(*config.LoadOptions) error) (aws.Config, error) {
	if region, _ := parsedConf.FieldString("region"); region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	credsConf := parsedConf.Namespace("credentials")
	if profile, _ := credsConf.FieldString("profile"); profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	} else if id, _ := credsConf.FieldString("id"); id != "" {
		secret, _ := credsConf.FieldString("secret")
		token, _ := credsConf.FieldString("token")
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			id, secret, token,
		)))
	}

	conf, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return conf, err
	}

	if endpoint, _ := parsedConf.FieldString("endpoint"); endpoint != "" {
		conf.BaseEndpoint = &endpoint
	}

	if role, _ := credsConf.FieldString("role"); role != "" {
		stsSvc := sts.NewFromConfig(conf)

		var stsOpts []func(*stscreds.AssumeRoleOptions)
		if externalID, _ := credsConf.FieldString("role_external_id"); externalID != "" {
			stsOpts = append(stsOpts, func(aro *stscreds.AssumeRoleOptions) {
				aro.ExternalID = &externalID
			})
		}

		creds := stscreds.NewAssumeRoleProvider(stsSvc, role, stsOpts...)
		conf.Credentials = aws.NewCredentialsCache(creds)
	}

	if useEC2, _ := credsConf.FieldBool("from_ec2_role"); useEC2 {
		conf.Credentials = aws.NewCredentialsCache(ec2rolecreds.New())
	}
	return conf, nil
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

func attributeValueMapToStandardJSON(item map[string]types.AttributeValue) map[string]any {
	standardJSON := make(map[string]any)
	for k, v := range item {
		standardJSON[k] = attributeValueToStandardValue(v)
	}
	return standardJSON
}

// attributeValueToStandardValue converts a DynamoDB AttributeValue to a standard value
func attributeValueToStandardValue(v types.AttributeValue) any {
	switch t := v.(type) {
	case *types.AttributeValueMemberB:
		return t.Value
	case *types.AttributeValueMemberBOOL:
		return t.Value
	case *types.AttributeValueMemberBS:
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			lAny[i] = v
		}
		return lAny
	case *types.AttributeValueMemberL:
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			lAny[i] = attributeValueToStandardValue(v)
		}
		return lAny
	case *types.AttributeValueMemberM:
		mAny := make(map[string]any, len(t.Value))
		for k, v := range t.Value {
			mAny[k] = attributeValueToStandardValue(v)
		}
		return mAny
	case *types.AttributeValueMemberN:
		return t.Value
	case *types.AttributeValueMemberNS:
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
		lAny := make([]any, len(t.Value))
		for i, v := range t.Value {
			lAny[i] = v
		}
		return lAny
	}
	return nil
}
