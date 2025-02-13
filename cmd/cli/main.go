package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
)

const (
	apiURL = "http://localhost:8080/mealplan"
	// apiURL = "https://httpbin.org/post"
)

type model struct {
	plan           string         // The meal plan we are viewing
	loading        bool           // Whether the request is loading
	loadingSpinner spinner.Model  // Loading spinner
	err            error          // Error message
	width          int            // Width of the terminal
	height         int            // Height of the terminal
	viewport       viewport.Model // Viewport for the meal plan
	keys           keyMap         // The key bindings shown in the viewport
	help           help.Model     // The help model in the viewport
}

type keyMap struct {
	Refresh key.Binding
	Quit    key.Binding
	Up      key.Binding
	Down    key.Binding
	Help    key.Binding
}

var keys = keyMap{
	Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Quit:    key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑", "scroll up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓", "scroll down")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Refresh, k.Quit, k.Help}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Refresh, k.Quit},
		{k.Help},
	}
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		loading:        true,
		viewport:       viewport.New(0, 0),
		loadingSpinner: s,
		keys:           keys,
		help:           help.New(),
	}
}

type gotPlanMsg string
type errMsg error
type tickMsg struct{}

func (m model) Init() tea.Cmd {
	return m.requestPlan()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		m.help.Width = msg.Width

		return m, nil

	case gotPlanMsg:
		m.plan = string(msg)
		m.loading = false
		wrapped := wordwrap.String(m.plan, m.width)
		indented := indent.String(wrapped, 2)

		out, err := glamour.Render(indented, "dark")
		if err != nil {
			m.err = err
			m.loading = false
			return m, nil
		}

		m.viewport.SetContent(out)
		return m, nil

	case tickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
			return m, cmd
		}

		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.Refresh):
			m.loading = true
			return m, m.requestPlan()

		case key.Matches(msg, m.keys.Up):
			m.viewport.LineUp(1)

		case key.Matches(msg, m.keys.Down):
			m.viewport.LineDown(1)
		}
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)

	var cmd tea.Cmd
	if m.loading {
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
		return m, tea.Batch(cmd, vpCmd)
	}
	return m, vpCmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.loading {
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			AlignVertical(lipgloss.Center).
			Align(lipgloss.Center).
			Render(
				lipgloss.JoinHorizontal(lipgloss.Center,
					m.loadingSpinner.View(),
					"Generating meal plan",
				),
			)
	}

	helpView := lipgloss.NewStyle().PaddingLeft(2).MarginTop(1).Render(m.help.View(m.keys))
	contentHeight := m.height - lipgloss.Height(helpView)
	if contentHeight < 0 {
		contentHeight = 0
	}
	m.viewport.Height = contentHeight

	return lipgloss.JoinVertical(lipgloss.Left, m.viewport.View(), helpView)
}

func (m model) requestPlan() tea.Cmd {
	return tea.Batch(
		m.loadingSpinner.Tick,
		getPlanCmd(make(chan bool)),
		tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg{}
		}),
	)
}

func getPlanCmd(done chan bool) tea.Cmd {
	return func() tea.Msg {
		defer close(done)

		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer([]byte{}))
		if err != nil {
			return errMsg(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errMsg(fmt.Errorf("HTTP error: %s", resp.Status))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errMsg(err)
		}

		var data map[string]string
		err = json.Unmarshal(body, &data)
		if err != nil {
			return errMsg(err)
		}

		plan, ok := data["plan"]
		if !ok {
			return errMsg(fmt.Errorf("missing 'plan' key in JSON response"))
		}

		return gotPlanMsg(plan)
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
