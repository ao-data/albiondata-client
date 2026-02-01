// Copyright 2017 René Jochum. All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package updater

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v68/github"
	"github.com/minio/selfupdate"
)

const (
	platform = runtime.GOOS + "-" + runtime.GOARCH
)

var (
	ErrorNoBinary        = errors.New("No binary for the update found")
	defaultHTTPRequester = HTTPRequester{}
)

// Updater is the configuration and runtime data for doing an update.
//
type Updater struct {
	CurrentVersion     string    // Currently running version.
	GithubOwner        string    // The owner of the repo like "ao-data"
	GithubRepo         string    // The repository like "go-githubupdate"
	FilePrefix         string    // A prefix like "update-" for the binaries to indicate these are updater files.
	Requester          Requester //Optional parameter to override existing http request handler
	latestReleasesResp *github.RepositoryRelease
}

// NewUpdater creates a new updater.
func NewUpdater(currentVersion, githubOwner, githubRepo, filePrefix string) *Updater {
	return &Updater{
		CurrentVersion: currentVersion,
		GithubOwner:    githubOwner,
		GithubRepo:     githubRepo,
		FilePrefix:     filePrefix,
	}
}

// BackgroundUpdater is the all in one update solution for ya. :)
// Returns true if an update was successfully applied, false otherwise.
func (u *Updater) BackgroundUpdater() (bool, error) {
	available, err := u.CheckUpdateAvailable()
	if err != nil {
		return false, err
	}

	if available != "" {
		fmt.Printf("Version %s available, installing now.\n", available)
		err := u.Update()
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// CheckUpdateAvailable fetches the latest releases from github and
// returns string new version of empty string on no update.
func (u *Updater) CheckUpdateAvailable() (string, error) {
	client := github.NewClient(nil)

	ctx := context.Background()
	release, _, err := client.Repositories.GetLatestRelease(ctx, u.GithubOwner, u.GithubRepo)
	if err != nil {
		return "", err
	}

	u.latestReleasesResp = release

	current, err := semver.Make(u.CurrentVersion)
	update, err := semver.Make(*u.latestReleasesResp.TagName)
	if current.LT(update) {
		return *u.latestReleasesResp.TagName, nil
	}

	return "", nil
}

func (u *Updater) Update() error {
	reqFilename := u.FilePrefix + platform + ".gz"
	if runtime.GOOS == "windows" {
		reqFilename = u.FilePrefix + platform + ".exe.gz"
	}
	var foundAsset *github.ReleaseAsset
	for _, asset := range u.latestReleasesResp.Assets {
		if asset.GetName() == reqFilename {
			foundAsset = asset
			break
		}
	}

	// Not found
	if foundAsset == nil {
		return ErrorNoBinary
	}

	dlURL := foundAsset.GetBrowserDownloadURL()

	bin, err := u.fetchGZ(dlURL)
	if err != nil {
		return err
	}

	err = selfupdate.Apply(bytes.NewReader(bin), selfupdate.Options{})
	if err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("update failed and rollback also failed: %v", rerr)
		}
		return fmt.Errorf("update failed: %v", err)
	}
	fmt.Println("Update installed successfully.")
	return nil
}

func (u *Updater) fetch(url string) (io.ReadCloser, error) {
	if u.Requester == nil {
		return defaultHTTPRequester.Fetch(url)
	}

	readCloser, err := u.Requester.Fetch(url)
	if err != nil {
		return nil, err
	}

	if readCloser == nil {
		return nil, fmt.Errorf("Fetch was expected to return non-nil ReadCloser")
	}

	return readCloser, nil
}

func (u *Updater) fetchGZ(url string) ([]byte, error) {
	r, err := u.fetch(url)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buf := new(bytes.Buffer)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(buf, gz); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
