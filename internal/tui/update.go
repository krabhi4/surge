package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressMsg:
		m.Downloaded = msg.Downloaded
		m.Speed = msg.Speed
		m.Elapsed = time.Since(m.StartTime)

		// Update progress bar
		if m.Total > 0 {
			percentage := float64(m.Downloaded) / float64(m.Total)
			cmd := m.progress.SetPercent(percentage)
			return m, tea.Batch(cmd, m.listenProgress())
		}
		return m, m.listenProgress()

	case DownloadCompleteMsg:
		m.Downloaded = msg.Total
		m.Elapsed = msg.Elapsed
		m.done = true
		return m, tea.Quit

	case DownloadErrorMsg:
		m.err = msg.Err
		m.done = true
		return m, tea.Quit

	case TickMsg:
		return m, tea.Batch(m.listenProgress(), tickCmd())

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}
