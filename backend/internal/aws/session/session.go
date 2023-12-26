package aws_session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func NewSession(config *mgmtv1alpha1.AwsS3ConnectionConfig) (*session.Session, error) {
	awsCfg := aws.NewConfig()

	if region := config.GetRegion(); region != "" {
		awsCfg = awsCfg.WithRegion(region)
	}
	if endpoint := config.GetEndpoint(); endpoint != "" {
		awsCfg = awsCfg.WithEndpoint(endpoint)
	}
	configCreds := config.GetCredentials()
	if profile := configCreds.GetProfile(); profile != "" {
		awsCfg = awsCfg.WithCredentials(credentials.NewSharedCredentials(
			"", profile,
		))
	} else if id := configCreds.GetAccessKeyId(); id != "" {
		secret := configCreds.GetSecretAccessKey()
		token := configCreds.GetSessionToken()
		awsCfg = awsCfg.WithCredentials(credentials.NewStaticCredentials(
			id, secret, token,
		))
	}

	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, err
	}

	if role := configCreds.GetRoleArn(); role != "" {
		var opts []func(*stscreds.AssumeRoleProvider)
		if externalID := configCreds.GetRoleExternalId(); externalID != "" {
			opts = []func(*stscreds.AssumeRoleProvider){
				func(p *stscreds.AssumeRoleProvider) {
					p.ExternalID = &externalID
				},
			}
		}
		sess.Config = sess.Config.WithCredentials(
			stscreds.NewCredentials(sess, role, opts...),
		)
	}

	if useEC2 := configCreds.GetFromEc2Role(); useEC2 {
		sess.Config = sess.Config.WithCredentials(ec2rolecreds.NewCredentials(sess))
	}

	return sess, nil
}
