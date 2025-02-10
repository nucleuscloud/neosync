package accounthook_events

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type Event struct {
	Name mgmtv1alpha1.AccountHookEvent `json:"name"`

	JobRunCreated   *Event_JobRunCreated   `json:"jobRunCreated,omitempty"`
	JobRunSucceeded *Event_JobRunSucceeded `json:"jobRunSucceeded,omitempty"`
	JobRunFailed    *Event_JobRunFailed    `json:"jobRunFailed,omitempty"`
}
type Event_BaseJobRun struct {
	AccountId string `json:"accountId"`
	JobId     string `json:"jobId"`
	JobRunId  string `json:"jobRunId"`
}

func newEvent_BaseJobRun(
	accountId string,
	jobId string,
	jobRunId string,
) *Event_BaseJobRun {
	return &Event_BaseJobRun{
		AccountId: accountId,
		JobId:     jobId,
		JobRunId:  jobRunId,
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
		Name: mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_CREATED,
		JobRunCreated: &Event_JobRunCreated{
			Event_BaseJobRun: newEvent_BaseJobRun(accountId, jobId, jobRunId),
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
		Name: mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED,
		JobRunSucceeded: &Event_JobRunSucceeded{
			Event_BaseJobRun: newEvent_BaseJobRun(accountId, jobId, jobRunId),
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
		Name: mgmtv1alpha1.AccountHookEvent_ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED,
		JobRunFailed: &Event_JobRunFailed{
			Event_BaseJobRun: newEvent_BaseJobRun(accountId, jobId, jobRunId),
		},
	}
}
