package dtomaps

import (
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobDto(
	inputJob *db_queries.NeosyncApiJob,
	inputDestConnections []pgtype.UUID,
) *mgmtv1alpha1.Job {
	mappings := []*mgmtv1alpha1.JobMapping{}
	for _, mapping := range inputJob.Mappings {
		mappings = append(mappings, mapping.ToDto())
	}
	return &mgmtv1alpha1.Job{
		Id:                       nucleusdb.UUIDString(inputJob.ID),
		Name:                     inputJob.Name,
		CreatedAt:                timestamppb.New(inputJob.CreatedAt.Time),
		UpdatedAt:                timestamppb.New(inputJob.UpdatedAt.Time),
		CreatedByUserId:          nucleusdb.UUIDString(inputJob.CreatedByID),
		UpdatedByUserId:          nucleusdb.UUIDString(inputJob.UpdatedByID),
		Status:                   mgmtv1alpha1.JobStatus(inputJob.Status),
		ConnectionSourceId:       nucleusdb.UUIDString(inputJob.ConnectionSourceID),
		CronSchedule:             nucleusdb.ToNullableString(inputJob.CronSchedule),
		HaltOnNewColumnAddition:  nucleusdb.Int16ToBool(inputJob.HaltOnNewColumnAddition),
		ConnectionDestinationIds: nucleusdb.UUIDStrings(inputDestConnections),
		Mappings:                 mappings,
	}
}
