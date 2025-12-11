package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	URL         string
	Filename    string
	Total       int64
	Downloaded  int64
	Speed       float64
	Connections int

	StartTime time.Time
	Elapsed   time.Duration

	progress progress.Model

	progressCh chan ProgressMsg
	done       bool
	err        error
}

// NewModel creates a new TUI model
func NewModel(url string, filename string, total int64, connections int) Model {
	return Model{
		URL:         url,
		Filename:    filename,
		Total:       total,
		Downloaded:  0,
		Speed:       0,
		Connections: connections,
		StartTime:   time.Now(),
		progress:    progress.New(progress.WithDefaultGradient()),
		done:        false,
		err:         nil,
		progressCh:  make(chan ProgressMsg, 10),
	}
}

// GetProgressChannel returns the channel for receiving progress updates
func (m *Model) GetProgressChannel() chan ProgressMsg {
	return m.progressCh
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// tickCmd sends a tick message periodically
func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
		return TickMsg{}
	})
}

// listenProgress waits for a progress message from the channel
func (m Model) listenProgress() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.progressCh
		if !ok {
			return nil
		}
		return msg
	}
}
