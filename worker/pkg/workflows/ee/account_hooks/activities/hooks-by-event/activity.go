package hooks_by_event_activity

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	temporallogger "github.com/nucleuscloud/neosync/worker/internal/temporal-logger"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type Activity struct {
	accounthookclient mgmtv1alpha1connect.AccountHookServiceClient
}

func New(
	accounthookclient mgmtv1alpha1connect.AccountHookServiceClient,
) *Activity {
	return &Activity{accounthookclient: accounthookclient}
}

type RunHooksByEventRequest struct {
	AccountId string
	EventName mgmtv1alpha1.AccountHookEvent
}

type RunHooksByEventResponse struct {
	HookIds []string
}

func (a *Activity) GetAccountHooksByEvent(
	ctx context.Context,
	req *RunHooksByEventRequest,
) (*RunHooksByEventResponse, error) {
	activityInfo := activity.GetInfo(ctx)
	loggerKeyVals := []any{
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
		"AccountId", req.AccountId,
		"EventName", req.EventName,
	}

	slogger := temporallogger.NewSlogger(log.With(
		activity.GetLogger(ctx),
		loggerKeyVals...,
	))

	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	slogger.Debug("retrieving hooks by event")

	resp, err := a.accounthookclient.GetActiveAccountHooksByEvent(ctx, connect.NewRequest(&mgmtv1alpha1.GetActiveAccountHooksByEventRequest{
		AccountId: req.AccountId,
		Event:     req.EventName,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve active hooks by event: %w", err)
	}
	hooks := resp.Msg.GetHooks()
	slogger.Debug(fmt.Sprintf("found %d active hooks", len(hooks)))

	hookIds := make([]string, len(hooks))
	for i, hook := range hooks {
		hookIds[i] = hook.GetId()
	}

	return &RunHooksByEventResponse{HookIds: hookIds}, nil
}
