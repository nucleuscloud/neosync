package pg_models

import mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

type JobHookConfig struct {
}

func (j *JobHookConfig) ToDto() (*mgmtv1alpha1.JobHookConfig, error) {
	return &mgmtv1alpha1.JobHookConfig{}, nil
}

func (j *JobHookConfig) FromDto(dto *mgmtv1alpha1.JobHookConfig) error {
	return nil
}
