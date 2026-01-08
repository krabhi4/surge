package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"surge/internal/downloader"
	"surge/internal/messages"
	"surge/internal/tui"
	"surge/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Shared channel for download events (start/complete/error)
var eventCh chan tea.Msg
var program *tea.Program

// initTUI sets up the shared event channel and BubbleTea program
func initTUI() {
	eventCh = make(chan tea.Msg, DefaultProgressChannelBuffer)
	program = tea.NewProgram(tui.InitialRootModel(), tea.WithAltScreen())

	// Pump events to TUI
	go func() {
		for msg := range eventCh {
			program.Send(msg)
		}
	}()
}

// runTUI starts the TUI and blocks until quit
func runTUI() error {
	_, err := program.Run()
	return err
}

// runHeadless runs a download without TUI, printing progress to stdout
func runHeadless(ctx context.Context, url, outPath string, verbose bool, md5sum, sha256sum string) error {
	eventCh := make(chan tea.Msg, DefaultProgressChannelBuffer)

	startTime := time.Now()
	var totalSize int64
	var lastProgress int64

	// Start download in background
	errCh := make(chan error, 1)
	go func() {
		err := downloader.Download(ctx, url, outPath, verbose, md5sum, sha256sum, eventCh, 1)
		errCh <- err
		close(eventCh)
	}()

	// Process events
	for msg := range eventCh {
		switch m := msg.(type) {
		case messages.DownloadStartedMsg:
			totalSize = m.Total
			fmt.Fprintf(os.Stderr, "Downloading: %s (%s)\n", m.Filename, utils.ConvertBytesToHumanReadable(totalSize))
		case messages.ProgressMsg:
			// Only print progress every 10%
			if totalSize > 0 {
				percent := m.Downloaded * 100 / totalSize
				lastPercent := lastProgress * 100 / totalSize
				if percent/10 > lastPercent/10 {
					speed := float64(m.Downloaded) / time.Since(startTime).Seconds() / (1024 * 1024)
					fmt.Fprintf(os.Stderr, "  %d%% (%s) - %.2f MB/s\n", percent,
						utils.ConvertBytesToHumanReadable(m.Downloaded), speed)
				}
				lastProgress = m.Downloaded
			}
		case messages.DownloadCompleteMsg:
			elapsed := time.Since(startTime)
			speed := float64(totalSize) / elapsed.Seconds() / (1024 * 1024)
			fmt.Fprintf(os.Stderr, "Complete: %s in %s (%.2f MB/s)\n",
				utils.ConvertBytesToHumanReadable(totalSize),
				elapsed.Round(time.Millisecond), speed)
		case messages.DownloadErrorMsg:
			return m.Err
		}
	}

	return <-errCh
}

// sendToServer sends a download request to a running surge server
func sendToServer(url, outPath string, port int) error {
	reqBody := DownloadRequest{
		URL:  url,
		Path: outPath,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	serverURL := fmt.Sprintf("http://127.0.0.1:%d/download", port)
	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request to server: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s - %s", resp.Status, string(body))
	}

	fmt.Printf("Download queued on server: %s\n", string(body))
	return nil
}

var getCmd = &cobra.Command{
	Use:   "get [url]",
	Short: "get downloads a file from a URL",
	Long: `get downloads a file from a URL and saves it to the local filesystem.
If --port is specified, the download request is sent to a running surge server instead.
Use --headless for CLI-only mode (no TUI), useful for scripting and benchmarks.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		outPath, _ := cmd.Flags().GetString("path")
		verbose, _ := cmd.Flags().GetBool("verbose")
		headless, _ := cmd.Flags().GetBool("headless")
		// concurrent, _ := cmd.Flags().GetInt("concurrent") Have to implement this later
		md5sum, _ := cmd.Flags().GetString("md5")
		sha256sum, _ := cmd.Flags().GetString("sha256")
		port, _ := cmd.Flags().GetInt("port")

		if outPath == "" {
			outPath = "."
		}

		// If port is specified, send to server instead of downloading locally
		if port > 0 {
			if err := sendToServer(url, outPath, port); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		ctx := context.Background()

		// Headless mode: CLI-only, no TUI
		if headless {
			if err := runHeadless(ctx, url, outPath, verbose, md5sum, sha256sum); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Normal TUI mode
		initTUI()
		go func() {
			defer close(eventCh)
			if err := downloader.Download(ctx, url, outPath, verbose, md5sum, sha256sum, eventCh, 1); err != nil {
				program.Send(messages.DownloadErrorMsg{DownloadID: 1, Err: err})
			}
		}()

		if err := runTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	getCmd.Flags().StringP("path", "o", "", "the path to the download folder")
	getCmd.Flags().IntP("concurrent", "c", DefaultConcurrentConnections, "number of concurrent connections (1 = single thread)")
	getCmd.Flags().BoolP("verbose", "v", false, "enable verbose output")
	getCmd.Flags().Bool("headless", false, "run without TUI (CLI-only mode for scripting/benchmarks)")
	getCmd.Flags().String("md5", "", "MD5 checksum for verification")
	getCmd.Flags().String("sha256", "", "SHA256 checksum for verification")
	getCmd.Flags().IntP("port", "p", 0, "port of running surge server to send download to (0 = run locally)")
}
