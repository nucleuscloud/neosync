package accounts_cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
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
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

func newSwitchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch",
		Short: "switch accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}

			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			cmd.SilenceUsage = true
			flagCount := cmd.Flags().NFlag()
			return switchAccount(cmd.Context(), flagCount, &apiKey, &id, &name)
		},
	}
	cmd.Flags().String("id", "", "Account id to switch to")
	cmd.Flags().String("name", "", "Account name to switch to")
	return cmd
}

func switchAccount(
	ctx context.Context,
	flagCount int,
	apiKey, id, name *string,
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

	currentAccountId, _ := userconfig.GetAccountId()

	accounts := accountsResp.Msg.Accounts
	if len(accounts) == 0 {
		return errors.New("unable to find accounts for user")
	}

	if flagCount == 0 {
		items := []list.Item{}
		for _, a := range accounts {
			isCurrent := a.Id == currentAccountId
			items = append(items, item{
				title:       a.Name,
				description: a.Id,
				isCurrent:   isCurrent,
			})
		}
		items = append(items, item{title: "Cancel"})

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
			fmt.Println("Error running program:", err) // nolint
			os.Exit(1)
		}
		return nil

	}

	var account *mgmtv1alpha1.UserAccount
	if id != nil && *id != "" {
		for _, a := range accounts {
			if a.Id == *id {
				account = a

			}
		}
	} else if name != nil && *name != "" {
		for _, a := range accounts {
			if a.Name == *name {
				account = a
			}
		}
	}

	if account == nil {
		return errors.New("unable to find account for user")
	}

	err = userconfig.SetAccountId(account.Id)
	if err != nil {
		fmt.Println("unable to switch accounts") // nolint
		return err
	}

	fmt.Println(selectedItemStyle.Render(fmt.Sprintf("\n Switched account to %s (%s) \n", account.Name, account.Id))) // nolint

	return nil
}

type item struct {
	title       string
	description string
	isCurrent   bool
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) { // nolint
	i, ok := listItem.(item)
	if !ok {
		return
	}

	var str = i.title
	if i.description != "" {
		str = fmt.Sprintf("%s (%s)", str, i.description)
	}
	if i.isCurrent {
		str = fmt.Sprintf("%s (current)", str)
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
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
			return quitTextStyle.Render(fmt.Sprintf("Failed to switch accounts. Error %s", err.Error()))
		}
		return quitTextStyle.Render(fmt.Sprintf("Switched account to %s", m.choice.title))
	}
	if m.quitting || m.choice.title == "Cancel" {
		return quitTextStyle.Render("Canceling...")
	}
	return "\n" + m.list.View()
}
