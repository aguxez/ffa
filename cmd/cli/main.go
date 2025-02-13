package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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
	quitting       bool           // Whether we are quitting the CLI
	width          int            // Width of the terminal
	height         int            // Height of the terminal
	viewport       viewport.Model // Viewport for the meal plan
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		loading:        true,
		width:          80,
		height:         24,
		viewport:       viewport.New(80, 24),
		loadingSpinner: s,
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
		switch msg.String() {

		case "q":
			m.quitting = true
			return m, tea.Quit

		case "r":
			m.loading = true
			return m, m.requestPlan()

		case "up", "k":
			m.viewport.LineUp(1)

		case "down", "j":
			m.viewport.LineDown(1)
		}
	}

	m.viewport, _ = m.viewport.Update(msg)

	var cmd tea.Cmd
	if m.loading {
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if m.quitting {
		return "Quitting.\n"
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

	return m.viewport.View()
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
