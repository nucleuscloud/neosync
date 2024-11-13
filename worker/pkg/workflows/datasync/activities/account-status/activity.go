package accountstatus_activity

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
)

type Activity struct {
	userclient mgmtv1alpha1connect.UserAccountServiceClient
}

func New(
	userclient mgmtv1alpha1connect.UserAccountServiceClient,
) *Activity {
	return &Activity{
		userclient: userclient,
	}
}

type CheckAccountStatusRequest struct {
	AccountId            string
	RequestedRecordCount *uint64
}

type CheckAccountStatusResponse struct {
	IsValid    bool
	Reason     *string
	ShouldPoll bool
}

func (a *Activity) CheckAccountStatus(
	ctx context.Context,
	req *CheckAccountStatusRequest,
) (*CheckAccountStatusResponse, error) {
	activityInfo := activity.GetInfo(ctx)
	logger := log.With(
		activity.GetLogger(ctx),
		"accountId", req.AccountId,
		"WorkflowID", activityInfo.WorkflowExecution.ID,
		"RunID", activityInfo.WorkflowExecution.RunID,
	)

	go func() {
		for {
			select {
			case <-time.After(30 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	logger.Debug("checking account status")

	resp, err := a.userclient.IsAccountStatusValid(ctx, connect.NewRequest(&mgmtv1alpha1.IsAccountStatusValidRequest{
		AccountId:            req.AccountId,
		RequestedRecordCount: req.RequestedRecordCount,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve account status: %w", err)
	}

	logger.Debug(
		fmt.Sprintf("account status: %v", resp.Msg.GetIsValid()),
		"reason", withReasonOrDefault(resp.Msg.GetReason()),
	)

	return &CheckAccountStatusResponse{IsValid: resp.Msg.GetIsValid(), Reason: resp.Msg.Reason, ShouldPoll: resp.Msg.GetShouldPoll()}, nil
}

const defaultReason = "no reason provided"

func withReasonOrDefault(reason string) string {
	if reason == "" {
		return defaultReason
	}
	return reason
}
