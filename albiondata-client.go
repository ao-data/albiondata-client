package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ao-data/albiondata-client/client"
	"github.com/ao-data/albiondata-client/log"
	"github.com/ao-data/albiondata-client/systray"

	"github.com/ao-data/go-githubupdate/updater"
)

var version string

func init() {
	client.ConfigGlobal.SetupFlags()
}

func main() {
	if client.ConfigGlobal.PrintVersion {
		log.Infof("Albion Data Client, version: %s", version)
		return
	}

	startUpdater()

	// On macOS, the systray requires the Cocoa event loop to run on the main thread.
	// So we run the client in a goroutine and systray on the main thread.
	// On other platforms, we do the opposite for backward compatibility.
	if runtime.GOOS == "darwin" {
		go runClient()
		systray.Run() // This blocks on the main thread (required for macOS)
	} else {
		go systray.Run()
		runClient()
	}
}

func runClient() {
	c := client.NewClient(version)
	err := c.Run()
	if err != nil {
		log.Error(err)
		log.Error("The program encountered an error. Press any key to close this window.")
		var b = make([]byte, 1)
		_, _ = os.Stdin.Read(b)
	}
}

func startUpdater() {
	if version != "" && !strings.Contains(version, "dev") {
		u := updater.NewUpdater(
			version,
			client.ConfigGlobal.UpdateGithubOwner,
			client.ConfigGlobal.UpdateGithubRepo,
			"update-",
		)

		go func() {
			for {
				if tryUpdate(u) {
					restartProcess()
					return // This line won't be reached if restart succeeds, but included for clarity
				}
				// Wait 1 hour before checking again
				time.Sleep(time.Hour)
			}
		}()
	}
}

// tryUpdate attempts to check and apply an update with retry logic.
// Returns true if an update was successfully applied.
func tryUpdate(u *updater.Updater) bool {
	maxTries := 2
	for i := 0; i < maxTries; i++ {
		updated, err := u.BackgroundUpdater()
		if err != nil {
			log.Error(err.Error())
			if i < maxTries-1 {
				log.Info("Will try again in 60 seconds. You may need to run the client as Administrator.")
				time.Sleep(time.Second * 60)
			}
			continue
		}
		if updated {
			return true
		}
		// No update available, no need to retry
		return false
	}
	return false
}

// restartProcess replaces the current process with the updated version.
// On Unix systems (macOS/Linux), it uses syscall.Exec to seamlessly take over the terminal.
// On Windows, it starts a new process and exits since exec-style replacement isn't supported.
func restartProcess() {
	execPath, err := os.Executable()
	if err != nil {
		log.Errorf("Failed to get executable path for restart: %v", err)
		return
	}

	log.Info("Restarting with updated version...")

	if runtime.GOOS == "windows" {
		// Windows doesn't support exec-style process replacement
		// Start a new process and exit
		cmd := exec.Command(execPath, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err = cmd.Start()
		if err != nil {
			log.Errorf("Failed to start new process: %v", err)
			return
		}

		log.Info("New process started, exiting current process.")
		os.Exit(0)
	} else {
		// On Unix systems (macOS/Linux), use syscall.Exec to replace the current process
		// This seamlessly takes over the terminal - same PID, same terminal session
		err = syscall.Exec(execPath, os.Args, os.Environ())
		if err != nil {
			log.Errorf("Failed to exec new process: %v", err)
			// Fall back to starting a new process
			cmd := exec.Command(execPath, os.Args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			_ = cmd.Start()
			cmd.Wait()
			os.Exit(0)
		}
		// If syscall.Exec succeeds, this line is never reached
		// because the current process is replaced
	}
}
