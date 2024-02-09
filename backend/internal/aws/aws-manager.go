package awsmanager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
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
}

func New() *NeosyncAwsManager {
	return &NeosyncAwsManager{}
}

func (n *NeosyncAwsManager) NewS3Client(ctx context.Context, config *mgmtv1alpha1.AwsS3ConnectionConfig) (*s3.Client, error) {
	cfg, err := n.getAwsConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(*cfg), nil
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

func (n *NeosyncAwsManager) getAwsConfig(ctx context.Context, config *mgmtv1alpha1.AwsS3ConnectionConfig) (*aws.Config, error) {
	awsCfg := aws.NewConfig()

	configCreds := config.GetCredentials()
	if profile := configCreds.GetProfile(); profile != "" {
		cfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithSharedConfigProfile(profile),
		)
		if err != nil {
			return nil, err
		}
		awsCfg = &cfg
	} else if accessKeyId := configCreds.GetAccessKeyId(); accessKeyId != "" {
		secretAccessKey := configCreds.GetSecretAccessKey()
		token := configCreds.GetSessionToken()
		staticCredsProvider := credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, token)
		awsCfg.Credentials = aws.NewCredentialsCache(staticCredsProvider)
	}

	if region := config.GetRegion(); region != "" {
		awsCfg.Region = region
	}
	if endpoint := config.GetEndpoint(); endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpoint,
				SigningRegion: awsCfg.Region,
			}, nil
		})

		cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithEndpointResolverWithOptions(customResolver))
		if err != nil {
			return nil, err
		}
		awsCfg = &cfg
	}

	if role := configCreds.GetRoleArn(); role != "" {
		if externalId := configCreds.GetRoleExternalId(); externalId != "" {
			awsCfg.Credentials = aws.NewCredentialsCache(
				stscreds.NewAssumeRoleProvider(sts.NewFromConfig(*awsCfg), role, func(aro *stscreds.AssumeRoleOptions) {
					aro.ExternalID = aws.String(externalId)
					aro.RoleSessionName = "neosync-mgmt-api"
				}),
			)
		}
	}

	if useEC2 := configCreds.GetFromEc2Role(); useEC2 {
		awsCfg.Credentials = ec2rolecreds.New()
	}

	return awsCfg, nil
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

		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return true
		}
	}
	return false
}
