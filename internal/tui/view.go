package tui

import (
	"fmt"
	"time"

	"surge/internal/utils"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI
func (m Model) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(
			fmt.Sprintf("Error: %v\n", m.err),
		)
	}

	if m.done && m.err == nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(
			fmt.Sprintf("Download complete!\nFile: %s\nTime: %s\n",
				m.Filename,
				m.Elapsed.Round(time.Second),
			),
		)
	}

	// Calculate percentage
	percentage := 0.0
	if m.Total > 0 {
		percentage = float64(m.Downloaded) / float64(m.Total)
	}

	// Calculate ETA
	eta := "N/A"
	if m.Speed > 0 && m.Total > 0 {
		remainingBytes := m.Total - m.Downloaded
		remainingSeconds := float64(remainingBytes) / m.Speed
		eta = time.Duration(remainingSeconds * float64(time.Second)).Round(time.Second).String()
	}

	// Build the display
	title := lipgloss.NewStyle().Bold(true).Render("⬇ Downloading")
	filename := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("File: %s", m.Filename))
	progressBar := m.progress.View()
	stats := fmt.Sprintf(
		"%.1f%%  |  %s / %s  |  %.1f MB/s  |  ETA: %s",
		percentage*100,
		utils.ConvertBytesToHumanReadable(m.Downloaded),
		utils.ConvertBytesToHumanReadable(m.Total),
		m.Speed/1024.0/1024.0,
		eta,
	)
	connections := fmt.Sprintf("Connections: %d  |  Elapsed: %s",
		m.Connections,
		m.Elapsed.Round(time.Second),
	)

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n",
		title,
		filename,
		progressBar,
		stats,
		connections,
	)
}
