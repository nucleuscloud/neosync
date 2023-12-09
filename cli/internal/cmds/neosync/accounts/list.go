package accounts_cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/spf13/cobra"

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
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return listAccounts(cmd.Context(), &apiKey)
		},
	}
}

func listAccounts(
	ctx context.Context,
	apiKey *string,
) error {
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}
	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)

	accountsResp, err := userclient.GetUserAccounts(
		ctx,
		connect.NewRequest[mgmtv1alpha1.GetUserAccountsRequest](&mgmtv1alpha1.GetUserAccountsRequest{}),
	)
	if err != nil {
		return err
	}
	accounts := accountsResp.Msg.Accounts
	if len(accounts) == 0 {
		return errors.New("unable to find accounts for user")
	}

	printAccountTable(accounts)
	return nil
}

func printAccountTable(accounts []*mgmtv1alpha1.UserAccount) {
	columns := []table.Column{
		{Title: "Name", Width: 20},
		{Title: "Type", Width: 20},
		{Title: "Id", Width: 50},
	}

	rows := []table.Row{}
	for _, a := range accounts {
		var accountType = "Personal"
		if a.Type.String() == "USER_ACCOUNT_TYPE_TEAM" {
			accountType = "Team"
		}
		rows = append(rows, table.Row{
			a.Name, accountType, a.Id,
		})
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
