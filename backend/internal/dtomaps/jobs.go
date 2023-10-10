package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	temporalclient "go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToJobDto(
	inputJob *db_queries.NeosyncApiJob,
	inputDestConnections []db_queries.NeosyncApiJobDestinationConnectionAssociation,
	inputSchedule *temporalclient.ScheduleDescription,
) *mgmtv1alpha1.Job {
	mappings := []*mgmtv1alpha1.JobMapping{}
	for _, mapping := range inputJob.Mappings {
		mappings = append(mappings, mapping.ToDto())
	}
	destinations := []*mgmtv1alpha1.JobDestination{}
	for i := range inputDestConnections {
		dest := inputDestConnections[i]
		destinations = append(destinations, toDestinationDto(&dest))
	}

	return &mgmtv1alpha1.Job{
		Id:              nucleusdb.UUIDString(inputJob.ID),
		Name:            inputJob.Name,
		CreatedAt:       timestamppb.New(inputJob.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(inputJob.UpdatedAt.Time),
		CreatedByUserId: nucleusdb.UUIDString(inputJob.CreatedByID),
		UpdatedByUserId: nucleusdb.UUIDString(inputJob.UpdatedByID),
		Status:          toJobStatus(inputJob, inputSchedule),
		CronSchedule:    nucleusdb.ToNullableString(inputJob.CronSchedule),
		Mappings:        mappings,
		Source: &mgmtv1alpha1.JobSource{
			ConnectionId: nucleusdb.UUIDString(inputJob.ConnectionSourceID),
			Options:      inputJob.ConnectionOptions.ToDto(),
		},
		Destinations: destinations,
		AccountId:    nucleusdb.UUIDString(inputJob.AccountID),
		RecentRuns:   toRecentRunsDto(inputSchedule),
		NextRuns:     toNextRunsDto(inputSchedule),
	}

}

func toDestinationDto(input *db_queries.NeosyncApiJobDestinationConnectionAssociation) *mgmtv1alpha1.JobDestination {
	return &mgmtv1alpha1.JobDestination{
		ConnectionId: nucleusdb.UUIDString(input.ConnectionID),
		Options:      input.Options.ToDto(),
		Id:           nucleusdb.UUIDString(input.ID),
	}
}

func toJobStatus(inputJob *db_queries.NeosyncApiJob, inputSchedule *temporalclient.ScheduleDescription) mgmtv1alpha1.JobStatus {
	if inputSchedule.Schedule.State.Paused {
		return mgmtv1alpha1.JobStatus_JOB_STATUS_PAUSED
	}
	return mgmtv1alpha1.JobStatus(inputJob.Status)
}

func toRecentRunsDto(inputSchedule *temporalclient.ScheduleDescription) *mgmtv1alpha1.JobRecentRuns {
	recentRuns := []*mgmtv1alpha1.JobRecentRun{}
	if inputSchedule == nil {
		return nil
	}
	for _, run := range inputSchedule.Info.RecentActions {
		recentRuns = append(recentRuns, &mgmtv1alpha1.JobRecentRun{
			StartTime: timestamppb.New(run.ActualTime),
			JobRunId:  run.StartWorkflowResult.WorkflowID,
		})
	}
	return &mgmtv1alpha1.JobRecentRuns{
		Runs: recentRuns,
	}
}

func toNextRunsDto(inputSchedule *temporalclient.ScheduleDescription) *mgmtv1alpha1.JobNextRuns {
	nextRunTimes := []*timestamppb.Timestamp{}
	if inputSchedule == nil {
		return nil
	}
	for _, t := range inputSchedule.Info.NextActionTimes {
		nextRunTimes = append(nextRunTimes, timestamppb.New(t))
	}
	return &mgmtv1alpha1.JobNextRuns{
		NextRunTimes: nextRunTimes,
	}
}
