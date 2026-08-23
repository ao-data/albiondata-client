package main

import (
	"embed"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ao-data/albiondata-client/client"
	"github.com/ao-data/albiondata-client/icon"
	"github.com/ao-data/albiondata-client/internal/console"
	"github.com/ao-data/albiondata-client/internal/dashboard"
	"github.com/ao-data/albiondata-client/internal/dockicon"
	"github.com/ao-data/albiondata-client/internal/pcapdriver"
	"github.com/ao-data/albiondata-client/internal/winstate"
	"github.com/ao-data/albiondata-client/log"

	"github.com/ao-data/go-githubupdate/updater"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

var version string

// dashboardWindowMu guards dashboardWindowRef, which is set once
// runDashboardApp() creates the window. runClient() runs in its own
// goroutine and may race the window's creation, so access to the
// reference must be synchronized.
var (
	dashboardWindowMu  sync.Mutex
	dashboardWindowRef *application.WebviewWindow
)

// showDashboardWindow shows the dashboard window if it has been created
// yet. It is a no-op if runDashboardApp() hasn't gotten there yet, which
// can happen if capture fails very early in startup.
func showDashboardWindow() {
	dashboardWindowMu.Lock()
	w := dashboardWindowRef
	dashboardWindowMu.Unlock()

	if w != nil {
		// Show() alone only calls makeKeyAndOrderFront on macOS, which
		// doesn't activate the app - since this is an accessory
		// (menu-bar-only) app, the window can end up ordered-front within
		// its own layer but still hidden behind whatever app is currently
		// frontmost. Focus() also activates the app, which actually
		// brings it above everything else.
		w.Show()
		w.Focus()
	}
}

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	client.ConfigGlobal.SetupFlags()
	application.RegisterEvent[dashboard.Status]("status:changed")
	application.RegisterEvent[map[string]int64]("counters:snapshot")
	application.RegisterEvent[dashboard.LogLine]("log:line")
}

func main() {
	// Hide the console immediately, before any other startup work, so
	// nothing flashes on screen before it disappears. Only when launched
	// with no arguments — any flag (e.g. -version, -h) means the user is
	// running this from a terminal and expects to see output, so leave
	// the console visible in that case.
	if runtime.GOOS == "windows" && len(os.Args) == 1 && console.Owned() {
		console.Hide()
	}

	if client.ConfigGlobal.PrintVersion {
		log.Infof("Albion Data Client, version: %s", version)
		return
	}

	log.AddHook(dashboard.NewLogHook())

	// Delayed rather than called inline here: this early in startup it'd
	// print before Wails' own boot noise (Build Info/AssetServer Info/
	// Platform Info/WebView2 environment lines, all logged during
	// runDashboardApp() below), landing at the very top of the console
	// where it's easy to miss. Waiting lets it land after that settles,
	// closer to the bottom where a user's eye actually lands.
	go func() {
		time.Sleep(3 * time.Second)
		checkCaptureDriver()
	}()
	startUpdater()

	// Wails owns the main thread on every platform, so packet capture
	// always runs in its own goroutine now (previously this was
	// conditional on darwin because only macOS's systray required the
	// main thread).
	go runClient()

	runDashboardApp()
}

// checkCaptureDriver warns the user on startup if the system's
// packet-capture driver looks like it'll prevent capture from working -
// e.g. only the abandoned WinPcap is installed instead of Npcap, or
// Npcap is installed admin-only but this process isn't elevated. A no-op
// on non-Windows platforms. Logged and surfaced on the dashboard rather
// than just logged, since a log line alone is easy to miss - that's
// exactly what turned this into a multi-session investigation before
// the actual cause (missing/outdated driver, not a code bug) was found.
func checkCaptureDriver() {
	w := pcapdriver.Check()
	if w.Message == "" {
		return
	}
	if w.HelpURL != "" {
		// The dashboard renders HelpURL as a clickable link, but the
		// console/log-file output only ever gets Message - spell out the
		// URL there too so it's visible without the GUI.
		log.Warnf("%s %s", w.Message, w.HelpURL)
	} else {
		log.Warn(w.Message)
	}
	dashboard.SetDriverWarning(w.Message, w.HelpURL)
}

func runClient() {
	c := client.NewClient(version)
	err := c.Run()
	if err != nil {
		log.Error(err)
		log.Error("Packet capture has stopped due to an error. Check the dashboard's log tail for details.")
		dashboard.SetCaptureRunning(false)
		dashboard.SetCaptureError(true)
		showDashboardWindow()
	}
}

func runDashboardApp() {
	app := application.New(application.Options{
		Name:        "Albion Data Client",
		Description: "Live status dashboard for the Albion Data Client",
		Services: []application.Service{
			application.NewService(&dashboard.DashboardService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	winOpts := application.WebviewWindowOptions{
		Title:  "Albion Data Client",
		Name:   "dashboard",
		Width:  900,
		Height: 600,
		Hidden: true,
		URL:    "/",
	}
	if saved, ok := winstate.Load(); ok {
		winOpts.Width = saved.Width
		winOpts.Height = saved.Height
		winOpts.X = saved.X
		winOpts.Y = saved.Y
		winOpts.InitialPosition = application.WindowXY
	}

	dashboardWindow := app.Window.NewWithOptions(winOpts)

	dashboardWindowMu.Lock()
	dashboardWindowRef = dashboardWindow
	dashboardWindowMu.Unlock()

	// Closing the window just hides it - the app (and packet capture)
	// keeps running, matching the old tray app's behavior.
	dashboardWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		dashboardWindow.Hide()
		e.Cancel()
	})

	// The app runs as a menu-bar-only accessory (see Mac.ActivationPolicy
	// below) so it doesn't clutter the Dock while just running in the
	// background, but gets a normal Dock icon once its window is first
	// shown. Deliberately one-directional: flipping the activation policy
	// back to Accessory on hide was tried and dropped the tray's status
	// item from the menu bar (a known-fragile interaction with menu-bar
	// managers like Bartender that track NSStatusItem across activation
	// policy changes) - going Accessory -> Regular once and staying there
	// avoids the transition that caused that.
	dashboardWindow.RegisterHook(events.Common.WindowShow, func(e *application.WindowEvent) {
		application.InvokeSync(func() {
			dockicon.SetVisible(true)
		})
	})

	// Persist the window's position and size shortly after the user
	// finishes moving or resizing it, so the next launch reopens where
	// they left it. Debounced since both events fire continuously
	// throughout a drag.
	var saveBoundsMu sync.Mutex
	var saveBoundsTimer *time.Timer
	scheduleSaveBounds := func() {
		saveBoundsMu.Lock()
		defer saveBoundsMu.Unlock()
		if saveBoundsTimer != nil {
			saveBoundsTimer.Stop()
		}
		saveBoundsTimer = time.AfterFunc(500*time.Millisecond, func() {
			b := dashboardWindow.Bounds()
			if err := winstate.Save(winstate.Bounds{X: b.X, Y: b.Y, Width: b.Width, Height: b.Height}); err != nil {
				log.Error(err)
			}
		})
	}
	dashboardWindow.RegisterHook(events.Common.WindowDidMove, func(e *application.WindowEvent) {
		scheduleSaveBounds()
	})
	dashboardWindow.RegisterHook(events.Common.WindowDidResize, func(e *application.WindowEvent) {
		scheduleSaveBounds()
	})

	dashboard.OnStatusChange(func(s dashboard.Status) {
		app.Event.Emit("status:changed", s)
	})
	dashboard.OnCountersChange(func(c map[string]int64) {
		app.Event.Emit("counters:snapshot", c)
	})
	dashboard.OnLogLine(func(l dashboard.LogLine) {
		app.Event.Emit("log:line", l)
	})

	setupTray(app, dashboardWindow)

	// Open and focus the dashboard automatically at launch. Registered as
	// an ApplicationStarted hook rather than a bare goroutine: a plain
	// `go showDashboardWindow()` here would race app.Run()'s internal
	// setup (globalApplication.impl isn't set until partway through
	// Run(), and Show()/Focus() silently no-op if it's still nil) -
	// ApplicationStarted only fires once that setup has completed.
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		// Use the same image as the tray icon for the Dock icon, rather
		// than leaving it as whatever generic default the OS falls back
		// to for an unbundled binary (there's no .icns/Info.plist - see
		// scripts/build-darwin.sh, which ships a bare executable, not a
		// .app bundle). Must happen here, not right after
		// application.New(): SetIcon silently no-ops until app.impl is
		// set partway through Run(), same as showDashboardWindow() below.
		app.SetIcon(icon.TrayPNG)

		showDashboardWindow()
		if !dashboardWindow.IsVisible() {
			// showDashboardWindow()'s Show() call can, in a narrow
			// goroutine-scheduling race, realize the window while it's
			// still marked Hidden in its options - which shows nothing
			// and leaves Focus() a no-op. IsVisible() is nil-safe and
			// false in that case, so retrying here (now that the window
			// is definitely realized) takes the normal Show()/Focus()
			// path.
			showDashboardWindow()
		}
	})

	if err := app.Run(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func setupTray(app *application.App, dashboardWindow *application.WebviewWindow) {
	tray := app.SystemTray.New()
	tray.SetTooltip("Albion Data Client")

	// icon.TrayPNG is a full-color logo with an opaque background, not a
	// monochrome silhouette, so it must be set as a regular icon rather
	// than a macOS "template" icon: template rendering discards all color
	// and uses only the alpha channel as a mask, and this icon has no
	// transparent margin - it would render as a solid black rectangle
	// that's nearly invisible against the menu bar.
	tray.SetIcon(icon.TrayPNG)

	// Double-clicking the tray icon opens the dashboard. This is done by
	// hand rather than via tray.OnDoubleClick: Wails v3's macOS systray
	// backend never wires that handler up at all (only Windows does), and
	// on macOS a tray icon with a menu attached shows that menu on the
	// very first mouse-down whenever no click handler is registered -
	// which consumes the click before a second one can ever be seen as a
	// double-click. Registering our own OnClick suppresses that
	// auto-show-on-first-click behavior, so we do the single/double-click
	// disambiguation ourselves: a click starts a short timer that opens
	// the menu if nothing follows, or is cancelled and treated as a
	// double-click if a second one arrives first.
	const trayDoubleClickWindow = 400 * time.Millisecond
	var (
		trayClickMu      sync.Mutex
		trayPendingClick *time.Timer
	)
	tray.OnClick(func() {
		trayClickMu.Lock()
		defer trayClickMu.Unlock()

		if trayPendingClick != nil {
			trayPendingClick.Stop()
			trayPendingClick = nil
			dashboardWindow.Show()
			dashboardWindow.Focus()
			return
		}

		trayPendingClick = time.AfterFunc(trayDoubleClickWindow, func() {
			trayClickMu.Lock()
			trayPendingClick = nil
			trayClickMu.Unlock()
			tray.ShowMenu()
		})
	})

	menu := app.NewMenu()
	menu.Add("Open Dashboard").OnClick(func(ctx *application.Context) {
		// Show() alone doesn't activate the app on macOS (see
		// showDashboardWindow's comment); Focus() does, so the window
		// actually comes to the front instead of staying hidden behind
		// whatever app currently has focus.
		dashboardWindow.Show()
		dashboardWindow.Focus()
	})
	menu.Add("Open Log File").OnClick(func(ctx *application.Context) {
		openLogFile()
	})

	if console.Supported() {
		consoleLabel := "Show Console"
		if !console.Hidden() {
			consoleLabel = "Hide Console"
		}
		menu.Add(consoleLabel).OnClick(func(ctx *application.Context) {
			item := ctx.ClickedMenuItem()
			if console.Hidden() {
				console.Show()
				item.SetLabel("Hide Console")
			} else {
				console.Hide()
				item.SetLabel("Show Console")
			}
		})
	}

	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	tray.SetMenu(menu)
}

func openLogFile() {
	path := client.GetLogFilePath()
	if _, err := os.Stat(path); err != nil {
		log.Info("No log file found yet.")
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		log.Errorf("Failed to open log file: %v", err)
	}
}

func startUpdater() {
	dashboard.SetVersionInfo(version, "")
	dashboard.SetCustomPublicIngest(client.ConfigGlobal.PublicIngestBaseUrls != "https+pow://albion-online-data.com")

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
		if available, checkErr := u.CheckUpdateAvailable(); checkErr == nil {
			dashboard.SetVersionInfo(version, available)
		}

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
		err = syscall.Exec(execPath, os.Args, os.Environ())
		if err != nil {
			log.Errorf("Failed to exec new process: %v", err)
			cmd := exec.Command(execPath, os.Args[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if startErr := cmd.Start(); startErr != nil {
				log.Errorf("Fallback process start also failed: %v", startErr)
				return
			}
			os.Exit(0)
		}
	}
}
