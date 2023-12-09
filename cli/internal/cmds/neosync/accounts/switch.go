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
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type listKeyMap struct {
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
}

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
			return switchAccount(cmd.Context(), &apiKey, &id, &name)
		},
	}
	cmd.Flags().String("id", "", "Account id to switch to")
	cmd.Flags().String("name", "", "Account name to switch to")
	cmd.MarkFlagsOneRequired("id", "name")
	return cmd
}

func switchAccount(
	ctx context.Context,
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
	accounts := accountsResp.Msg.Accounts
	if len(accounts) == 0 {
		return errors.New("unable to find accounts for user")
	}

	if _, err := tea.NewProgram(newModel(accounts)).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
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

	fmt.Println("Switched accounts")                            // nolint
	fmt.Printf("Name: %s  Id: %s \n", account.Name, account.Id) // nolint

	return nil
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

type model struct {
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
}

func newModel(accounts []*mgmtv1alpha1.UserAccount) model {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	items := make([]list.Item, len(accounts))
	for i := 0; i < len(accounts); i++ {
		a := accounts[i]
		items[i] = item{
			title:       a.Name,
			description: a.Id,
		}
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	accountList := list.New(items, delegate, 0, 0)
	accountList.Title = "Accounts"
	accountList.Styles.Title = titleStyle
	accountList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
		}
	}

	return model{
		list:         accountList,
		keys:         listKeys,
		delegateKeys: delegateKeys,
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.togglePagination):
			m.list.SetShowPagination(!m.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.keys.toggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string

		if i, ok := m.SelectedItem().(item); ok {
			title = i.Title()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.choose):
				return m.NewStatusMessage(statusMessageStyle("You chose " + title))

			}
		}

		return nil
	}

	help := []key.Binding{keys.choose}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	choose key.Binding
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
	}
}
