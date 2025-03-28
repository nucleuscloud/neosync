package awsmanager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type NeosyncAwsManager struct {
}

type NeosyncAwsManagerClient interface {
	NewS3Client(ctx context.Context, config *mgmtv1alpha1.AwsS3ConnectionConfig) (*s3.Client, error)
	ListObjectsV2(
		ctx context.Context,
		s3Client *s3.Client,
		region *string,
		params *s3.ListObjectsV2Input,
	) (*s3.ListObjectsV2Output, error)
	GetObject(
		ctx context.Context,
		s3Client *s3.Client,
		region *string,
		params *s3.GetObjectInput,
	) (*s3.GetObjectOutput, error)

	NewDynamoDbClient(
		ctx context.Context,
		connCfg *mgmtv1alpha1.DynamoDBConnectionConfig,
	) (*DynamoDbClient, error)
}

func New() *NeosyncAwsManager {
	return &NeosyncAwsManager{}
}

// Returns a wrapper dynamodb client
func (n *NeosyncAwsManager) NewDynamoDbClient(
	ctx context.Context,
	connCfg *mgmtv1alpha1.DynamoDBConnectionConfig,
) (*DynamoDbClient, error) {
	client, err := n.newDynamoDbClient(ctx, connCfg)
	if err != nil {
		return nil, err
	}
	return NewDynamoDbClient(client), nil
}

// returns the raw, underlying aws client
func (n *NeosyncAwsManager) newDynamoDbClient(
	ctx context.Context,
	connCfg *mgmtv1alpha1.DynamoDBConnectionConfig,
) (*dynamodb.Client, error) {
	cfg, err := getDynamoAwsConfig(ctx, connCfg)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(*cfg, func(o *dynamodb.Options) {
		if connCfg.GetEndpoint() != "" {
			o.BaseEndpoint = aws.String(connCfg.GetEndpoint())
		}
	}), nil
}

func (n *NeosyncAwsManager) NewS3Client(
	ctx context.Context,
	connCfg *mgmtv1alpha1.AwsS3ConnectionConfig,
) (*s3.Client, error) {
	cfg, err := getS3AwsConfig(ctx, connCfg)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(*cfg, func(o *s3.Options) {
		if connCfg.GetEndpoint() != "" {
			o.BaseEndpoint = aws.String(connCfg.GetEndpoint())
		}
	}), nil
}

func (n *NeosyncAwsManager) ListObjectsV2(
	ctx context.Context,
	s3Client *s3.Client,
	region *string,
	params *s3.ListObjectsV2Input,
) (*s3.ListObjectsV2Output, error) {
	output, err := s3Client.ListObjectsV2(ctx, params, withS3Region(region))
	if err != nil && !IsNotFound(err) {
		return nil, fmt.Errorf("error getting object list from S3: %w", err)
	}
	if (err != nil && IsNotFound(err)) || *output.KeyCount == 0 {
		return nil, nil
	}
	return output, nil
}

func (n *NeosyncAwsManager) GetObject(
	ctx context.Context,
	s3Client *s3.Client,
	region *string,
	params *s3.GetObjectInput,
) (*s3.GetObjectOutput, error) {
	output, err := s3Client.GetObject(ctx, params, withS3Region(region))
	if err != nil {
		return nil, fmt.Errorf("error getting object from S3: %w", err)
	}
	return output, nil
}

func withS3Region(region *string) func(o *s3.Options) {
	return func(o *s3.Options) {
		if region != nil && *region != "" {
			o.Region = *region
		}
	}
}

func getS3AwsConfig(
	ctx context.Context,
	s3ConnConfig *mgmtv1alpha1.AwsS3ConnectionConfig,
) (*aws.Config, error) {
	return GetAwsConfig(ctx, &AwsCredentialsConfig{
		Region:          s3ConnConfig.GetRegion(),
		Endpoint:        s3ConnConfig.GetEndpoint(),
		Profile:         s3ConnConfig.GetCredentials().GetProfile(),
		Id:              s3ConnConfig.GetCredentials().GetAccessKeyId(),
		Secret:          s3ConnConfig.GetCredentials().GetSecretAccessKey(),
		Token:           s3ConnConfig.GetCredentials().GetSessionToken(),
		Role:            s3ConnConfig.GetCredentials().GetRoleArn(),
		RoleExternalId:  s3ConnConfig.GetCredentials().GetRoleExternalId(),
		RoleSessionName: "neosync",
		UseEc2:          s3ConnConfig.GetCredentials().GetFromEc2Role(),
	})
}

func getDynamoAwsConfig(
	ctx context.Context,
	dynConnConfig *mgmtv1alpha1.DynamoDBConnectionConfig,
) (*aws.Config, error) {
	return GetAwsConfig(ctx, &AwsCredentialsConfig{
		Region:          dynConnConfig.GetRegion(),
		Endpoint:        dynConnConfig.GetEndpoint(),
		Profile:         dynConnConfig.GetCredentials().GetProfile(),
		Id:              dynConnConfig.GetCredentials().GetAccessKeyId(),
		Secret:          dynConnConfig.GetCredentials().GetSecretAccessKey(),
		Token:           dynConnConfig.GetCredentials().GetSessionToken(),
		Role:            dynConnConfig.GetCredentials().GetRoleArn(),
		RoleExternalId:  dynConnConfig.GetCredentials().GetRoleExternalId(),
		RoleSessionName: "neosync",
		UseEc2:          dynConnConfig.GetCredentials().GetFromEc2Role(),
	})
}

func IsNotFound(err error) bool {
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" {
				return true
			}
		}

		var notFound *types.NoSuchKey
		if ok := errors.As(err, &notFound); ok {
			return true
		}
		var dynResourceNotFound *dynamodbtypes.ResourceNotFoundException
		if ok := errors.As(err, &dynResourceNotFound); ok {
			return true
		}

		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return true
		}
	}
	return false
}

type AwsCredentialsConfig struct {
	Region   string
	Endpoint string

	Profile string
	Id      string
	Secret  string
	Token   string

	Role            string
	RoleExternalId  string
	RoleSessionName string

	UseEc2 bool
}

func GetAwsConfig(
	ctx context.Context,
	cfg *AwsCredentialsConfig,
	opts ...func(*config.LoadOptions) error,
) (*aws.Config, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cfg input was nil, expected *AwsCredentialsConfig")
	}

	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}
	if cfg.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(cfg.Profile))
	} else if cfg.Id != "" {
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.Id, cfg.Secret, cfg.Token)))
	}

	conf, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if cfg.Endpoint != "" {
		conf.BaseEndpoint = &cfg.Endpoint
	}
	if cfg.Role != "" {
		stsSvc := sts.NewFromConfig(conf)

		var stsOpts []func(*stscreds.AssumeRoleOptions)
		if cfg.RoleExternalId != "" {
			stsOpts = append(stsOpts, func(aro *stscreds.AssumeRoleOptions) {
				aro.ExternalID = &cfg.RoleExternalId
				aro.RoleSessionName = cfg.RoleSessionName
			})
		}

		creds := stscreds.NewAssumeRoleProvider(stsSvc, cfg.Role, stsOpts...)
		conf.Credentials = aws.NewCredentialsCache(creds)
	}
	if cfg.UseEc2 {
		conf.Credentials = aws.NewCredentialsCache(ec2rolecreds.New())
	}
	return &conf, nil
}
