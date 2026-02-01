//go:build darwin

package systray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ao-data/albiondata-client/icon"
	"github.com/ao-data/albiondata-client/log"
	"github.com/getlantern/systray"
)

var ConsoleHidden bool = false

const CanHideConsole = false

func HideConsole() {
	// Not supported on macOS
}

func ShowConsole() {
	// Not supported on macOS
}

func Run() {
	systray.Run(onReady, onExit)
}

func onExit() {
	// Cleanup if needed
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("") // Clear text since we have an icon now
	systray.SetTooltip("Albion Data Client")

	mOpenLog := systray.AddMenuItem("Open Log File", "Open the log file in default viewer")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Close the Albion Data Client")

	go func() {
		for {
			select {
			case <-mOpenLog.ClickedCh:
				openLogFile()

			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				os.Exit(0)
			}
		}
	}()
}

func openLogFile() {
	// Try to find and open the log file
	logFile := "albiondata-client.log"
	
	// Check current directory first
	if _, err := os.Stat(logFile); err == nil {
		absPath, _ := filepath.Abs(logFile)
		cmd := exec.Command("open", absPath)
		if err := cmd.Start(); err != nil {
			log.Errorf("Failed to open log file: %v", err)
		}
		return
	}
	
	// If no log file exists, show a message
	log.Info("No log file found yet.")
}
