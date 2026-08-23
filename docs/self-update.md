# Self-update mechanics

`startUpdater()` in `albiondata-client.go` only runs at all when
`version != "" && !strings.Contains(version, "dev")` - a dev build
(the default `main.version=dev` ldflags used by `scripts/run.sh` and
`buildall_main.sh`) never checks for updates. It polls hourly via
`github.com/ao-data/go-githubupdate`, which wraps `minio/selfupdate` to
replace the running binary in place (see gui-dashboard.md for why
that's always a single-file replace, never a bundle).

## Exact filename matching - no fuzzy match, no fallback

The updater builds the expected release-asset filename as:

```
"update-" + runtime.GOOS + "-" + runtime.GOARCH + ".gz"          // .exe.gz on windows
```

and does a literal `asset.GetName() == reqFilename` string comparison
against the target repo's latest release assets. No fallback, no
partial match - if the exact filename isn't there, it's
`ErrorNoBinary` ("No binary for the update found"), regardless of
whether *some* compatible binary exists in the release.

This is why macOS is always built as amd64 (see build-and-release.md)
even on arm64 hardware: a real end user runs the shipped
`albiondata-client-amd64-mac.zip` under Rosetta 2, so their process's
`runtime.GOARCH` reports `amd64` at runtime and matches
`update-darwin-amd64.gz`, the only darwin asset any release actually
publishes. But a **locally built test binary** on Apple Silicon, built
with a plain `go build` (no `GOARCH=amd64` override), links natively
for arm64 - `runtime.GOARCH` reports `arm64`, the updater looks for
`update-darwin-arm64.gz`, and that asset doesn't exist in any release.
Confirmed exactly this way once already: `file
albiondata-client-pre-gui` showed `Mach-O 64-bit executable arm64`,
which is why its self-update failed even though the target release had
a perfectly good `update-darwin-amd64.gz`. If this happens again,
rebuild the local binary with the same amd64-forcing flags
`scripts/build-darwin.sh` uses, rather than adding an arm64 release
target - see build-and-release.md for why arm64 was deliberately not
added.

## Target repo is compile-time hardcoded

`client.ConfigGlobal.UpdateGithubOwner`/`UpdateGithubRepo` default to
`"ao-data"`/`"albiondata-client"` (`client/config.go`). There *is* a
viper-based override for these via `config.yaml`, but it's commented
out in `setupWebsocketFlags()` ("Keeping for local development, but
commenting out so it's not live") - so in a normal build there is no
config-file or flag way to point the updater at a fork. Testing against
a personal fork (e.g. `phendryx/albiondata-client`) currently means
building a local binary with `UpdateGithubOwner`/`UpdateGithubRepo`
edited directly in `client/config.go` before building - not committed,
reverted before any real commit.
