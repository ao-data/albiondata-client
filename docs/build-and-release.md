# Build, CI, and release

## The single-file build convention (read this before adding a file to the repo root)

Every build script (`scripts/build-{windows,linux,darwin}.sh`,
`scripts/run.sh`) and `buildall_main.sh` builds with:

```
go build ... albiondata-client.go
```

Naming the single file, not `.` or the package. Go only compiles that
one file plus whatever it imports - any other `.go` file sitting
directly in `package main` in the repo root is silently invisible to
these scripts and will fail to link at the very last step with an
`undefined:` error for whatever it defined. **Anything new that `main`
needs has to live in an `internal/` package and be imported normally**
(imports resolve fine regardless of the single-file invocation) -
never add a second `.go` file directly in the repo root expecting the
build to pick it up.

## Local dev commands

```
make run              # go run the client directly (scripts/run.sh)
make frontend         # npm ci && npm run build in frontend/
make fmt              # goimports -w across the repo (vendor/-safe, see below)
make validate-fmt     # goimports -l, non-zero exit if anything's unformatted
make build-windows    # cross-compiles from any OS; needs nsis, go-winres
make build-linux      # native on Linux; needs libpcap-dev, patchelf, gtk4/webkitgtk-6.0 dev headers
make build-darwin     # native on macOS; forces amd64 (see below), needs libpcap headers (Xcode CLT)

go build ./...        # compile-check everything (needs frontend/dist to exist first, see below)
go vet ./...
go test ./...          # no known-failing tests as of this writing
go test ./client/... -run TestName -v   # single test
```

`go build`/`go vet`/`go test` on the root package need `frontend/dist`
to exist first (`make frontend`) - `albiondata-client.go` has
`//go:embed all:frontend/dist`, and embed directives are validated at
compile time even if nothing in the current build actually touches the
embedded FS (e.g. `go vet` on a package with no test files still fails
on a missing embed target).

`fmt.sh`/`validate-fmt.sh` use `find . -name '*.go' -not -path
'./vendor/*'` rather than filtering `go list`'s directory output - the
latter looks correct but isn't: `go list -f '{{.Dir}}' ./...` includes
the repo-root package dir (`.`, since it contains
`albiondata-client.go`), and `goimports`/`gofmt` recurse into
subdirectories when given a bare directory argument, so a `grep -v
/vendor/` on the directory list doesn't stop the walk once it's inside
a non-vendor directory that contains `vendor/` as a child. Don't
"simplify" this back to the `go list` form.

## macOS: always amd64, even natively

`scripts/build-darwin.sh` and the darwin leg of CI force `GOARCH=amd64`
with `CC="clang -arch x86_64"`/`CGO_LDFLAGS="-arch x86_64"`, even
though GitHub's `macos-latest` runners (and most current dev machines)
are arm64 hardware. This is deliberate, not a leftover - there is no
native arm64 build target, on purpose (see self-update.md for why this
matters beyond just "smaller binary"). If you build a local test binary
with a plain `go build` on Apple Silicon without these flags, it'll be
a native arm64 binary that behaves differently from what's actually
shipped - see self-update.md for a concrete way this bites.

## CI: one reusable workflow, three callers

`.github/workflows/build.yml` is a `workflow_call` reusable workflow: a
matrix over `{linux, windows, darwin}`, each leg just running `make
build-<os>` and uploading whatever that produces as a workflow
artifact. `test-build.yml` (push/PR) and `release.yml` (on a GitHub
release being created) both call it - `release.yml` adds a `publish`
job that downloads the artifacts and attaches them to the release.

This exists because the three build paths used to be duplicated almost
verbatim across two workflow files, and drifted: a fix to
`scripts/build-darwin.sh` (adding the frontend build step) was never
applied to the *inlined* darwin steps in the workflow YAML, because
those steps didn't call the script at all - they ran a raw `go build`
directly. That meant darwin CI was very likely failing outright (no
`frontend/dist` -> the embed directive fails to compile) for some
unknown period before anyone caught it. If you need to change how an
OS builds, change the `scripts/build-<os>.sh` script - never add
build steps directly to a workflow YAML again, that's exactly how this
happened.

The linux leg also runs `go vet ./...` and `go test ./...` (it already
has libpcap-dev, and the frontend built, from the build step above -
no separate job needed). `test-build.yml` used to be build-only despite
the name.

## Windows-specific build notes

- `go-winres` embeds the `.exe`'s icon and Windows resource metadata
  (`winres/winres.json`) at build time, regenerating
  `rsrc_windows_*.syso` which `go build` then picks up automatically.
  Those `.syso` files are gitignored, not committed.
- The NSIS installer (`pkg/nsis/`) derives its version from
  `$GITHUB_REF_NAME`, but only when it looks like a real tag
  (`^v?[0-9]+\.[0-9]+\.[0-9]+$`) - falls back to `"0.0.1"` otherwise.
  This matters because `GITHUB_REF_NAME` on a `pull_request`-triggered
  build is the merge ref (e.g. `5/merge`), not a version, and NSIS's
  `VIProductVersion "${PACKAGE_VERSION}.0"` directive hard-rejects
  anything that isn't a clean `X.X.X.X` - a bare ref there crashes the
  whole Windows build, not just mis-versions the installer. Don't go
  back to using the ref unconditionally.
- `winres/winres.json`'s version fields (`file_version`,
  `product_version`, etc.) are static zeros/empty strings - the built
  `.exe`'s Windows resource metadata is never actually stamped with the
  real version. Known, not yet fixed.

## No vendor/ directory

Dependencies resolve from the module cache (`go.sum`), not a committed
`vendor/` tree - nothing in any build script or CI job passes
`-mod=vendor`, so there's nothing to keep in sync. If `vendor/`
reappears (e.g. from running `go mod vendor` locally), it's gitignored
- don't commit it back.

## What no longer exists, and why

`Dockerfile`, `Dockerfile.build.darwin`, and `docker-compose.yml` were
deleted - they were a CircleCI/osxcross-era Linux-hosted
cross-compilation setup (Go 1.14/1.16 base images) for building macOS
binaries from Linux, superseded once macOS builds moved to running
natively on a real `macos-latest` runner. Nothing referenced them by
the time they were removed.
