package accounthook_events

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type Event struct {
	Name      mgmtv1alpha1.AccountHookEvent `json:"name"`
	AccountId string                        `json:"accountId"`

	JobRunCreated   *Event_JobRunCreated   `json:"jobRunCreated,omitempty"`
	JobRunSucceeded *Event_JobRunSucceeded `json:"jobRunSucceeded,omitempty"`
	JobRunFailed    *Event_JobRunFailed    `json:"jobRunFailed,omitempty"`
}
type Event_BaseJobRun struct {
	JobId    string `json:"jobId"`
	JobRunId string `json:"jobRunId"`
}

func newEvent_BaseJobRun(
	jobId string,
	jobRunId string,
) *Event_BaseJobRun {
	return &Event_BaseJobRun{
		JobId:    jobId,
		JobRunId: jobRunId,
	}
}

type Event_JobRunCreated struct {
	*Event_BaseJobRun
}

func NewEvent_JobRunCreated(
	accountId string,
	jobId string,
	jobRunId string,
) *Event {
	return &Event{
		Name:      mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_CREATED,
		AccountId: accountId,
		JobRunCreated: &Event_JobRunCreated{
			Event_BaseJobRun: newEvent_BaseJobRun(jobId, jobRunId),
		},
	}
}

type Event_JobRunSucceeded struct {
	*Event_BaseJobRun
}

func NewEvent_JobRunSucceeded(
	accountId string,
	jobId string,
	jobRunId string,
) *Event {
	return &Event{
		Name:      mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED,
		AccountId: accountId,
		JobRunSucceeded: &Event_JobRunSucceeded{
			Event_BaseJobRun: newEvent_BaseJobRun(jobId, jobRunId),
		},
	}
}

type Event_JobRunFailed struct {
	*Event_BaseJobRun
}

func NewEvent_JobRunFailed(
	accountId string,
	jobId string,
	jobRunId string,
) *Event {
	return &Event{
		Name:      mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED,
		AccountId: accountId,
		JobRunFailed: &Event_JobRunFailed{
			Event_BaseJobRun: newEvent_BaseJobRun(jobId, jobRunId),
		},
	}
}
