# Contributor notes (Claude and humans)

Rules for working in this repository. Read them before opening a PR or
running an automated assistant against this tree.

## Documentation

- Keep `README.md` and code comments tight. State facts; don't speculate.
- Do not write about how the binding "should be used" or "is ideal for"
  any particular system. The binding wraps libmxl; downstream choice is
  not ours to characterize.
- Do not reference other projects (mediamtx, gstreamer, ffmpeg, …) as
  motivating consumers in docs, comments, or commit messages. They are
  not part of the public contract.
- Don't invent API behavior. If you can't verify it by reading libmxl's
  headers or testing against a real install, leave it out.
- Don't add SPDX headers or copyright lines to new files unless you're
  preserving existing ones from an external source.

## Commits

- Use Conventional Commits: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`,
  `test:`, `ci:`, `build:`, `perf:`, `style:`. Breaking changes get `!`
  (`feat!:`) or a `BREAKING CHANGE:` footer.
- Prefer small, focused commits. The release tooling derives version
  bumps and the changelog from commit subjects.
- Subject line ≤ 72 chars. Body wraps at 72.

## Versioning and tags

- The module path is `github.com/qvest-digital/go-mxl`. While the major
  version is 0 or 1, tags are `vMAJOR.MINOR.PATCH` at the repo root.
- For a future v2+, the module path must include the `/v2` suffix and
  tags become `v2.0.0` etc. — see <https://go.dev/ref/mod#major-version-suffixes>.
- Releases are automated by `release-please` (see `.github/workflows/`).
  Don't hand-tag or hand-edit `CHANGELOG.md` — let the workflow do it.

## Build

- The package is cgo. `libmxl` must be installed with headers and a
  pkg-config file before `go build` works. See `README.md`.
- Tests that exercise a writer↔reader round-trip live under build tag
  `mxl_integration` and need a tmpfs-mounted `/dev/shm`. CI runs them
  in the builder image alongside the unit tests; the build tag keeps
  them out of a plain `go test ./...` for callers without that mount.

## When in doubt

Ask the maintainer before changing the public Go API, the module path,
or the release/tagging strategy.
