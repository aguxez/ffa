package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
)

const (
	apiURL = "http://localhost:8080/mealplan"
)

type model struct {
	plan     string
	loading  bool
	err      error
	quitting bool
	width    int
	height   int
	viewport viewport.Model
}

func initialModel() model {
	return model{
		loading:  true,
		width:    80,
		height:   24,
		viewport: viewport.New(80, 24),
	}
}

type gotPlanMsg string
type errMsg error

func (m model) Init() tea.Cmd {
	return getPlanCmd()
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
		m.viewport.SetContent(indented)
		return m, nil

	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			m.viewport.LineUp(1)
		case "down", "j":
			m.viewport.LineDown(1)
		}
	}
	m.viewport, _ = m.viewport.Update(msg)
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
		return "Loading...\n"
	}

	return m.viewport.View()
}

func getPlanCmd() tea.Cmd {
	return func() tea.Msg {
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
