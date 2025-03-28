package accounts_cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	cli_logger "github.com/nucleuscloud/neosync/cli/internal/logger"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)
)

const listHeight = 14

var (
	header            = lipgloss.NewStyle().Faint(true).PaddingLeft(4)
	bold              = lipgloss.NewStyle().Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2).Height(1)
	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Height(1).
				Foreground(lipgloss.Color("170"))
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle   = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

func newSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch [name | id]",
		Short: "switch accounts",
		Example: `
    $ neosync accounts switch [name | id]

    - If the name and id is omitted, you can choose interactively

    NOTE: When you switch, everything in the CLI will be scoped that account!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}

			var accountIdOrName *string
			if len(args) > 0 {
				accountIdOrName = &args[0]
			}

			debugMode, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			return switchAccount(cmd.Context(), &apiKey, accountIdOrName, debugMode)
		},
	}
	return cmd
}

func switchAccount(
	ctx context.Context,
	apiKey,
	accountIdOrName *string,
	debug bool,
) error {
	logger := cli_logger.NewSLogger(cli_logger.GetCharmLevelOrDefault(debug))
	httpclient, err := auth.GetNeosyncHttpClient(ctx, logger, auth.WithApiKey(apiKey))
	if err != nil {
		return err
	}

	userclient := mgmtv1alpha1connect.NewUserAccountServiceClient(
		httpclient,
		auth.GetNeosyncUrl(),
	)

	accountsResp, err := userclient.GetUserAccounts(
		ctx,
		connect.NewRequest(&mgmtv1alpha1.GetUserAccountsRequest{}),
	)
	if err != nil {
		return err
	}

	currentAccountId, _ := userconfig.GetAccountId()

	accounts := accountsResp.Msg.GetAccounts()
	if len(accounts) == 0 {
		return errors.New("unable to find accounts for user")
	}

	if accountIdOrName == nil || *accountIdOrName == "" {
		items := []list.Item{}
		personalAccounts := []*mgmtv1alpha1.UserAccount{}
		teamAccounts := []*mgmtv1alpha1.UserAccount{}
		for _, a := range accounts {
			switch a.Type {
			case mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_PERSONAL:
				personalAccounts = append(personalAccounts, a)
			case mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_TEAM:
				teamAccounts = append(teamAccounts, a)
			}
		}
		for i, a := range personalAccounts {
			isCurrent := a.Id == currentAccountId
			items = append(items, item{
				title:       a.Name,
				description: a.Id,
				isCurrent:   isCurrent,
				header:      getHeader(a.Type, i == 0),
			})
		}

		for i, a := range teamAccounts {
			isCurrent := a.Id == currentAccountId
			items = append(items, item{
				title:       a.Name,
				description: a.Id,
				isCurrent:   isCurrent,
				header:      getHeader(a.Type, i == 0),
			})
		}
		cancelHeader := "─────────────────────────────────────────────────────────"
		items = append(items, item{title: "Cancel", header: &cancelHeader})

		const defaultWidth = 20

		l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
		l.Title = "Select an account"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.Styles.Title = titleStyle
		l.Styles.PaginationStyle = paginationStyle
		l.Styles.HelpStyle = helpStyle

		m := &model{list: l}

		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err) //nolint:forbidigo
			os.Exit(1)
		}
		return nil
	}

	var account *mgmtv1alpha1.UserAccount
	for _, a := range accounts {
		if strings.EqualFold(a.Name, *accountIdOrName) ||
			strings.EqualFold(a.Id, *accountIdOrName) {
			account = a
		}
	}

	if account == nil {
		return errors.New("unable to find account for user")
	}

	err = userconfig.SetAccountId(account.Id)
	if err != nil {
		return fmt.Errorf("unable to set account context: %w", err)
	}

	fmt.Println(
		itemStyle.Render(
			fmt.Sprintf("\n Switched account to %s (%s) \n", account.Name, account.Id),
		),
	) //nolint:forbidigo

	return nil
}

func getHeader(accountType mgmtv1alpha1.UserAccountType, shouldShowHeader bool) *string {
	var header string
	if shouldShowHeader && accountType == mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_PERSONAL {
		header = "── Personal Account ─────────────────────────────────────"
	}
	if shouldShowHeader && accountType == mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_TEAM {
		header = "── Team Account ─────────────────────────────────────────"
	}
	return &header
}

type item struct {
	title       string
	description string
	isCurrent   bool
	header      *string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(
	w io.Writer,
	m list.Model,
	index int,
	listItem list.Item,
) { //nolint:gocritic
	i, ok := listItem.(item)
	if !ok {
		return
	}

	var str = i.title
	if i.description != "" {
		str = fmt.Sprintf("%s (%s)", str, i.description)
	}
	if i.isCurrent {
		str = bold.Render(fmt.Sprintf("%s (current)", str))
	}

	var itemHeader string
	if i.header != nil && *i.header != "" {
		itemHeader = fmt.Sprintf("%s \n", header.Render(*i.header))
	}
	fn := func(s ...string) string {
		return fmt.Sprintf("%s%s", itemHeader, itemStyle.Render("○ "+strings.Join(s, " ")))
	}
	if index == m.Index() {
		fn = func(s ...string) string {
			return fmt.Sprintf(
				"%s%s",
				itemHeader,
				selectedItemStyle.Render("● "+strings.Join(s, " ")),
			)
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     list.Model
	choice   item
	quitting bool
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.choice.description != "" {
		err := userconfig.SetAccountId(m.choice.description)
		if err != nil {
			return quitTextStyle.Render(
				fmt.Sprintf("Failed to switch accounts. Error %s", err.Error()),
			)
		}
		return quitTextStyle.Render(
			fmt.Sprintf("Switched account to %s (%s)", m.choice.title, m.choice.description),
		)
	}
	if m.quitting || m.choice.title == "Cancel" {
		return quitTextStyle.Render("No changes made")
	}
	return "\n" + m.list.View()
}
