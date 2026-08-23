// icon/tray.go
package icon

import _ "embed"

// TrayPNG is the application icon in PNG form, used for the Wails v3
// system tray on all platforms (darwin/windows/linux). It is separate
// from the platform-specific Data variables (icondarwin.go, iconwin.go),
// which remain in their original .ico/.png forms for other uses (e.g.
// the Windows executable icon via go-winres).
//
//go:embed albiondata-client.png
var TrayPNG []byte
