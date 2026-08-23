// Package winstate persists the dashboard window's position and size
// across restarts.
package winstate

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// userConfigDir is a seam for tests; production code always uses
// os.UserConfigDir.
var userConfigDir = os.UserConfigDir

// Bounds is a persisted window position and size.
type Bounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

func filePath() (string, error) {
	dir, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "albiondata-client", "window.json"), nil
}

// Load returns the previously saved window bounds. The second return
// value is false if nothing has been saved yet, or the saved state can't
// be read - callers should fall back to their own defaults in that case.
func Load() (Bounds, bool) {
	path, err := filePath()
	if err != nil {
		return Bounds{}, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Bounds{}, false
	}

	var b Bounds
	if err := json.Unmarshal(data, &b); err != nil {
		return Bounds{}, false
	}
	if b.Width <= 0 || b.Height <= 0 {
		return Bounds{}, false
	}
	return b, true
}

// Save persists the given window bounds, overwriting whatever was saved
// before.
func Save(b Bounds) error {
	path, err := filePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(b)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
