package run_stripe_usage_cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosynclogger "github.com/nucleuscloud/neosync/backend/pkg/logger"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stripe-usage",
		Short: "Sends daily usage metrics to Stripe",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return run(cmd.Context())
		},
	}
}

func run(ctx context.Context) error {
	slogger, _ := neosynclogger.NewLoggers()

	neoenv := viper.GetString("NUCLEUS_ENV")
	if neoenv != "" {
		slogger = slogger.With(
			"nucleusEnv", neoenv,
			"env", neoenv,
			"neosyncEnv", neoenv,
		)
	}

	slog.SetDefault(slogger)

	ingestDateStr := viper.GetString("INGEST_DATE")
	ingestDateOffset := viper.GetString("INGEST_DURATION_OFFSET")
	accountIds := viper.GetStringSlice("ACCOUNT_IDS")
	meterName := viper.GetString("METER_NAME")
	if meterName == "" {
		return errors.New("must provide valid meter name")
	}

	ingestDate, err := getIngestDate(ingestDateStr, ingestDateOffset)
	if err != nil {
		return fmt.Errorf("unable to calculate ingest date: %w", err)
	}

	neosyncurl := shared.GetNeosyncUrl()
	httpclient := shared.GetNeosyncHttpClient()

	usersclient := mgmtv1alpha1connect.NewUserAccountServiceClient(httpclient, neosyncurl)
	metricsclient := mgmtv1alpha1connect.NewMetricsServiceClient(httpclient, neosyncurl)

	if len(accountIds) > 0 {
		slogger.DebugContext(ctx, fmt.Sprintf("%d accounts provided as input", len(accountIds)))
	}

	accountsResp, err := usersclient.GetBillingAccounts(ctx, connect.NewRequest(&mgmtv1alpha1.GetBillingAccountsRequest{
		AccountIds: accountIds,
	}))
	if err != nil {
		return err
	}
	accounts := accountsResp.Msg.GetAccounts()

	slogger.InfoContext(ctx, fmt.Sprintf("processing %d accounts", len(accounts)))

	for _, account := range accounts {
		err := processAccount(
			ctx,
			metricsclient, usersclient,
			account,
			ingestDate, meterName,
			slogger.With("accountId", account.GetId()),
		)
		if err != nil {
			return err
		}
	}

	slogger.InfoContext(ctx, "processing completed successfully")
	return nil
}

func processAccount(
	ctx context.Context,
	metricsclient mgmtv1alpha1connect.MetricsServiceClient,
	userclient mgmtv1alpha1connect.UserAccountServiceClient,
	account *mgmtv1alpha1.UserAccount,
	ingestdate *mgmtv1alpha1.Date,
	meterName string,
	logger *slog.Logger,
) error {
	logger.DebugContext(ctx, "retrieving daily metric count")
	resp, err := metricsclient.GetDailyMetricCount(ctx, connect.NewRequest(&mgmtv1alpha1.GetDailyMetricCountRequest{
		Metric: mgmtv1alpha1.RangedMetricName_RANGED_METRIC_NAME_INPUT_RECEIVED,
		Start:  ingestdate,
		End:    ingestdate,
		Identifier: &mgmtv1alpha1.GetDailyMetricCountRequest_AccountId{
			AccountId: account.GetId(),
		},
	}))
	if err != nil {
		return err
	}
	results := resp.Msg.GetResults()
	for _, result := range results {
		if compareDates(result.Date, ingestdate) == 0 {
			recordCount := result.GetCount()
			logger.DebugContext(ctx, "found record count", "count", recordCount)
			if recordCount > 0 {
				ts := getEventTimestamp(ingestdate)
				logger.DebugContext(ctx, "record count was greater than 0, creating meter event")
				_, err := userclient.SetBillingMeterEvent(ctx, connect.NewRequest(&mgmtv1alpha1.SetBillingMeterEventRequest{
					AccountId: account.GetId(),
					EventName: meterName,
					Value:     strconv.FormatUint(recordCount, 10),
					EventId:   getEventId(account.GetId(), ingestdate),
					Timestamp: &ts,
				}))
				if err != nil {
					return err
				}
			}
		} else {
			logger.Warn("response returned date outside of range")
		}
	}
	return nil
}

func getEventTimestamp(date *mgmtv1alpha1.Date) uint64 {
	t := time.Date(int(date.GetYear()), time.Month(date.GetMonth()), int(date.GetDay()), 23, 59, 59, 0, time.UTC)
	return uint64(t.Unix())
}

func getEventId(accountId string, ingestDate *mgmtv1alpha1.Date) string {
	return fmt.Sprintf("%s-%s", accountId, formatDate(ingestDate))
}

func formatDate(d *mgmtv1alpha1.Date) string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

const (
	ingestDateFormat = "2006-01-02"
)

func getIngestDate(
	ingestDate string,
	durationOffset string,
) (*mgmtv1alpha1.Date, error) {
	var date time.Time
	var err error

	if ingestDate == "" {
		// If INGEST_DATE is not provided, use today's date in UTC
		date = time.Now().UTC()
	} else {
		date, err = time.Parse(ingestDateFormat, ingestDate)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	// Apply offset if provided
	if durationOffset != "" {
		offset, err := time.ParseDuration(durationOffset)
		if err != nil {
			return nil, fmt.Errorf("invalid offset format: %w", err)
		}
		date = date.Add(offset)
	}

	return &mgmtv1alpha1.Date{
		Year:  uint32(date.Year()),  //nolint:gosec
		Month: uint32(date.Month()), //nolint:gosec
		Day:   uint32(date.Day()),   //nolint:gosec
	}, nil
}

func compareDates(date1, date2 *mgmtv1alpha1.Date) int {
	// Compare years
	if date1.Year < date2.Year {
		return -1
	}
	if date1.Year > date2.Year {
		return 1
	}

	// Years are equal, compare months
	if date1.Month < date2.Month {
		return -1
	}
	if date1.Month > date2.Month {
		return 1
	}

	// Months are equal, compare days
	if date1.Day < date2.Day {
		return -1
	}
	if date1.Day > date2.Day {
		return 1
	}

	// All fields are equal
	return 0
}
