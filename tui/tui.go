package tui

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ContentReadyMsg struct {
	Response string
}

type ErrorMsg struct {
	Err error
}

func (e ErrorMsg) Error() string { return e.Err.Error() }

type ResponseModel struct {
	ApiKey   string
	Model    string
	Prompt   string
	Viewport viewport.Model
	Loading  bool
	Spinner  spinner.Model
}

func (r ResponseModel) Init() tea.Cmd {
	return fetchGAIResponse(r.ApiKey, r.Model, r.Prompt)
}

func (r ResponseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case ContentReadyMsg:
		r.Loading = false
		r.Viewport.SetContent(msg.Response)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return r, tea.Quit
		}

	case ErrorMsg:
		return r, func() tea.Msg { return ErrorMsg{msg.Err} }
	}
	r.Viewport, cmd = r.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return r, tea.Batch(cmds...)
}

func (r ResponseModel) View() string {
	if r.Loading {
		str := fmt.Sprintf("\n\n   %s Crunching numbers ...\n\n", r.Spinner.View())
		r.Viewport.SetContent(str)
	}
	return r.Viewport.View()

}

func fetchGAIResponse(apiKey, model, prompt string) tea.Cmd {
	return func() tea.Msg {
		resp, err := generateResponse(apiKey, model, prompt)
		if err != nil {
			return ErrorMsg{Err: err}
		}

		return ContentReadyMsg{
			Response: resp,
		}
	}
}

func generateResponse(apiKey, model, prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	m := client.GenerativeModel(model)

	resp, err := m.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		slog.Error("Error generating a response from Gemini AI", err)
		return "", err
	}

	content := ""
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			content = fmt.Sprintf("%s", part)
			break
		}
		break
	}

	return content, nil
}

func InvokeResponseTUI(apiKey, model, prompt string) {
	s := spinner.New()
	s.Spinner = spinner.Ellipsis
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(100, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	r := ResponseModel{
		ApiKey:   apiKey,
		Model:    model,
		Prompt:   prompt,
		Loading:  true,
		Viewport: vp,
		Spinner:  s,
	}
	p := tea.NewProgram(r, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
