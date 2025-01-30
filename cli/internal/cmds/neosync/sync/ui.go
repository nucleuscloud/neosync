package sync_cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	syncmap "sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/nucleuscloud/neosync/cli/internal/output"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	_ "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	_ "github.com/redpanda-data/benthos/v4/public/components/io"
	_ "github.com/redpanda-data/benthos/v4/public/components/pure"
	_ "github.com/redpanda-data/benthos/v4/public/components/pure/extended"
	"github.com/redpanda-data/benthos/v4/public/service"
	_ "github.com/redpanda-data/connect/v4/public/components/aws"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	ctx              context.Context
	logger           *slog.Logger
	benv             *service.Environment
	groupedConfigs   [][]*benthosbuilder.BenthosConfigResponse
	tableSynced      int
	index            int
	width            int
	height           int
	spinner          spinner.Model
	done             bool
	totalConfigCount int
	outputType       output.OutputType
}

var (
	bold                = lipgloss.NewStyle().PaddingLeft(2).Bold(true)
	printlog            = lipgloss.NewStyle().PaddingLeft(2)
	currentPkgNameStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("42")).SetString("✓")
	helpStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle            = helpStyle.UnsetMargins()
	durationStyle       = dotStyle
)

func newModel(ctx context.Context, benv *service.Environment, groupedConfigs [][]*benthosbuilder.BenthosConfigResponse, logger *slog.Logger, outputType output.OutputType) *model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return &model{
		ctx:              ctx,
		groupedConfigs:   groupedConfigs,
		tableSynced:      0,
		spinner:          s,
		totalConfigCount: getConfigCount(groupedConfigs),
		logger:           logger,
		outputType:       outputType,
		benv:             benv,
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.syncConfigs(m.ctx, m.groupedConfigs[m.index]), m.spinner.Tick)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case syncedDataMsg:
		successStrs := []string{}
		for _, msgStr := range msg {
			successStrs = append(successStrs, msgStr)
			m.tableSynced++
		}
		if m.totalConfigCount == m.tableSynced {
			m.done = true
			m.logger.Info(fmt.Sprintf("Done! Completed %d tables.", m.tableSynced))
			return m, tea.Sequence(
				tea.Println(strings.Join(successStrs, " \n")),
				tea.Quit,
			)
		}

		m.index++
		return m, tea.Batch(
			tea.Println(strings.Join(successStrs, " \n")),
			m.syncConfigs(m.ctx, m.groupedConfigs[m.index]),
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) View() string {
	configCount := getConfigCount(m.groupedConfigs)
	w := lipgloss.Width(fmt.Sprintf("%d", configCount))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("Done! Completed %d tables.\n", configCount))
	}

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.tableSynced, w, configCount)

	spin := m.spinner.View() + " "
	cellsAvail := maxInt(0, m.width-lipgloss.Width(spin+pkgCount))

	processingTables := []string{}
	for _, config := range m.groupedConfigs[m.index] {
		processingTables = append(processingTables, config.Name)
	}

	var pkgName string
	if len(processingTables) > 5 {
		pkgName = currentPkgNameStyle.Render(fmt.Sprintf("%s \n + %d others...", strings.Join(processingTables[:5], "\n"), len(processingTables)))
	} else {
		pkgName = currentPkgNameStyle.Render(strings.Join(processingTables, "\n"))
	}
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Syncing " + pkgCount + " \n" + pkgName)
	return printlog.Render("\n") + spin + info
}

type syncedDataMsg map[string]string

func (m *model) syncConfigs(ctx context.Context, configs []*benthosbuilder.BenthosConfigResponse) tea.Cmd {
	return func() tea.Msg {
		messageMap := syncmap.Map{}
		errgrp, errctx := errgroup.WithContext(ctx)
		errgrp.SetLimit(5)
		for _, cfg := range configs {
			cfg := cfg
			errgrp.Go(func() error {
				start := time.Now()
				m.logger.Info(fmt.Sprintf("Syncing table %s", cfg.Name))
				err := syncData(errctx, m.benv, cfg, m.logger, m.outputType)
				if err != nil {
					fmt.Printf("Error syncing table: %s", err.Error()) //nolint:forbidigo
					return err
				}
				duration := time.Since(start)
				messageMap.Store(cfg.Name, duration)
				m.logger.Info(fmt.Sprintf("Finished syncing table %s %s", cfg.Name, duration.String()))
				return nil
			})
		}

		if err := errgrp.Wait(); err != nil {
			tea.Printf("Error syncing data: %s \n", err.Error())
			return tea.Quit
		}

		results := map[string]string{}
		//nolint:gofmt
		messageMap.Range(func(key, value interface{}) bool {
			d := value.(time.Duration)
			results[key.(string)] = fmt.Sprintf("%s %s %s", checkMark, key,
				durationStyle.Render(d.String()))
			return true
		})
		message := ""
		for _, config := range configs {
			message = fmt.Sprintf("%s, %s", message, config.Name)
		}
		return syncedDataMsg(results)
	}
}

func getConfigCount(groupedConfigs [][]*benthosbuilder.BenthosConfigResponse) int {
	count := 0
	for _, group := range groupedConfigs {
		for _, config := range group {
			if config != nil {
				count++
			}
		}
	}
	return count
}

func runSync(ctx context.Context, outputType output.OutputType, benv *service.Environment, groupedConfigs [][]*benthosbuilder.BenthosConfigResponse, logger *slog.Logger) error {
	var opts []tea.ProgramOption
	var synclogger = logger
	if outputType == output.PlainOutput {
		// Plain mode don't render the TUI
		opts = []tea.ProgramOption{tea.WithoutRenderer(), tea.WithInput(nil)}
	} else {
		fmt.Println(bold.Render(" \n Completed Tables")) //nolint:forbidigo
		// TUI mode, discard log output
		synclogger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	if _, err := tea.NewProgram(newModel(ctx, benv, groupedConfigs, synclogger, outputType), opts...).Run(); err != nil {
		logger.Error(fmt.Sprintf("Error syncing data: %v", err))
		return fmt.Errorf("unable to finish syncing data: %w", err)
	}
	return nil
}
