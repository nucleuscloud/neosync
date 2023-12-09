package jobs_cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type tableModel struct {
	table table.Model
}

func (m tableModel) Init() tea.Cmd { return nil }

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tableModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}

			accountId, err := cmd.Flags().GetString("account-id")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return listJobs(cmd.Context(), &apiKey, &accountId)
		},
	}
	cmd.Flags().String("account-id", "", "Account to list jobs for. Defaults to account id in cli context")
	return cmd
}

func listJobs(
	ctx context.Context,
	apiKey, accountIdFlag *string,
) error {
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}

	var accountId = accountIdFlag
	if accountId == nil || *accountId == "" {
		aId, err := userconfig.GetAccountId()
		if err != nil {
			fmt.Println("Unable to retrieve account id. Please use account switch command to set account.") // nolint
			return err
		}
		accountId = &aId
	}

	if accountId == nil || *accountId == "" {
		return errors.New("Account Id not found. Please use account switch command to set account.")
	}

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)
	res, err := jobclient.GetJobs(ctx, connect.NewRequest[mgmtv1alpha1.GetJobsRequest](&mgmtv1alpha1.GetJobsRequest{
		AccountId: *accountId,
	}))
	if err != nil {
		return err
	}

	jobstatuses := make([]*mgmtv1alpha1.JobStatus, len(res.Msg.Jobs))
	errgrp, errctx := errgroup.WithContext(ctx)
	for idx := range res.Msg.Jobs {
		idx := idx
		errgrp.Go(func() error {
			jsres, err := jobclient.GetJobStatus(errctx, connect.NewRequest[mgmtv1alpha1.GetJobStatusRequest](&mgmtv1alpha1.GetJobStatusRequest{
				JobId: res.Msg.Jobs[idx].Id,
			}))
			if err != nil {
				return err
			}
			jobstatuses[idx] = &jsres.Msg.Status
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return err
	}

	printJobTable(res.Msg.Jobs, jobstatuses)
	return nil
}

func printJobTable(
	jobs []*mgmtv1alpha1.Job,
	jobstatuses []*mgmtv1alpha1.JobStatus,
) {
	columns := []table.Column{
		{Title: "Id", Width: 40},
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 20},
		{Title: "Created At", Width: 30},
		{Title: "Updated At", Width: 30},
	}

	rows := make([]table.Row, len(jobs))
	for idx := range jobs {
		job := jobs[idx]
		js := jobstatuses[idx]
		rows[idx] = table.Row{
			job.Id,
			job.Name,
			js.String(),
			job.CreatedAt.AsTime().Local().Format(time.RFC3339),
			job.UpdatedAt.AsTime().Local().Format(time.RFC3339),
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := tableModel{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err) // nolint
		os.Exit(1)
	}
}

// func printJobTable(
// 	jobs []*mgmtv1alpha1.Job,
// 	jobstatuses []*mgmtv1alpha1.JobStatus,
// ) {
// 	tbl := table.
// 		New("Id", "Name", "Status", "Created At", "Updated At").
// 		WithHeaderFormatter(
// 			color.New(color.FgGreen, color.Underline).SprintfFunc(),
// 		).
// 		WithFirstColumnFormatter(
// 			color.New(color.FgYellow).SprintfFunc(),
// 		)

// 	for idx := range jobs {
// 		job := jobs[idx]
// 		js := jobstatuses[idx]
// 		tbl.AddRow(
// 			job.Id,
// 			job.Name,
// 			js.String(),
// 			job.CreatedAt.AsTime().Local().Format(time.RFC3339),
// 			job.UpdatedAt.AsTime().Local().Format(time.RFC3339),
// 		)
// 	}
// 	tbl.Print()
// }
