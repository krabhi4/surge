package tui

import "time"

// ProgressMsg represents a progress update from the downloader
type ProgressMsg struct {
	Downloaded int64
	Total      int64
	Speed      float64 // bytes per second
}

// DownloadCompleteMsg signals that the download finished successfully
type DownloadCompleteMsg struct {
	Filename string
	Elapsed  time.Duration
	Total    int64
}

// DownloadErrorMsg signals that an error occurred
type DownloadErrorMsg struct {
	Err error
}

// TickMsg is sent periodically to update the UI
type TickMsg struct{}
