//go:build windows
// +build windows

package systray

import (
	"fmt"
	"os"

	"github.com/ao-data/albiondata-client/client"

	"time"

	"github.com/ao-data/albiondata-client/icon"
	"github.com/ao-data/albiondata-client/log"
	"github.com/getlantern/systray"
	"github.com/gonutz/w32"
	"github.com/shirou/gopsutil/process"
)

var consoleHidden bool

func hideConsole() {
	console := w32.GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_HIDE)
		}
	}

	consoleHidden = true
}

func showConsole() {
	console := w32.GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_SHOW)
		}
	}

	consoleHidden = false
}

func GetActionTitle() string {
	if consoleHidden {
		return "Show Console"
	} else {
		return "Hide Console"
	}
}

func Run() {
	systray.Run(onReady, onExit)
}

func onExit() {

}

const targetProcessName string = "Albion-Online.exe"
const albionProcessTimeBetweenChecks int = 5

// Gets all processes periodically and stops when albion is found running.
func findAlbionProcess(found chan<- bool) {

	log.Info("Waiting for Albion to start...")

	// Endless loop to check periodically if {targetProcessName} process is running
	for {

		// Get all processes
		processes, err := process.Processes()
		if err != nil {
			log.Errorf("Error getting processes:", err)
			return
		}

		// Check if {targetProcessName} is running
		for _, p := range processes {
			name, err := p.Name()

			if err == nil && name == targetProcessName {
				log.Info("Albion is running.")
				found <- true
				return
			}
		}

		// Check again every n seconds
		time.Sleep(time.Duration(albionProcessTimeBetweenChecks) * time.Second)
	}

}

func onReady() {

	found := make(chan bool)
	go findAlbionProcess(found)
	<-found

	// Don't hide the console automatically
	// Unless started from the scheduled task or with the parameter
	// People think it is crashing
	if client.ConfigGlobal.Minimize {
		hideConsole()
	}
	systray.SetIcon(icon.Data)
	systray.SetTitle("Albion Data Client")
	systray.SetTooltip("Albion Data Client")
	mConHideShow := systray.AddMenuItem(GetActionTitle(), "Show/Hide Console")
	mQuit := systray.AddMenuItem("Quit", "Close the Albion Data Client")

	func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				os.Exit(0)
				fmt.Println("Finished quitting")

			case <-mConHideShow.ClickedCh:
				if consoleHidden == true {
					showConsole()
					mConHideShow.SetTitle(GetActionTitle())
				} else {
					hideConsole()
					mConHideShow.SetTitle(GetActionTitle())
				}
			}
		}
	}()
}
