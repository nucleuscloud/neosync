package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobHookDto(
	input *db_queries.NeosyncApiJobHook,
) (*mgmtv1alpha1.JobHook, error) {
	if input == nil {
		input = &db_queries.NeosyncApiJobHook{}
	}
	priority := uint32(0)
	if input.Priority > 0 {
		priority = uint32(input.Priority)
	}

	config := &mgmtv1alpha1.JobHookConfig{}
	err := config.UnmarshalJSON(input.Config)
	if err != nil {
		return nil, err
	}

	output := &mgmtv1alpha1.JobHook{
		Id:              neosyncdb.UUIDString(input.ID),
		Name:            input.Name,
		Description:     input.Description,
		JobId:           neosyncdb.UUIDString(input.JobID),
		CreatedByUserId: neosyncdb.UUIDString(input.CreatedByUserID),
		CreatedAt:       timestamppb.New(input.CreatedAt.Time),
		UpdatedByUserId: neosyncdb.UUIDString(input.UpdatedByUserID),
		UpdatedAt:       timestamppb.New(input.UpdatedAt.Time),
		Enabled:         input.Enabled,
		Priority:        priority,
		Config:          config,
	}

	return output, nil
}
