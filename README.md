# go-mxl

Go bindings for [libmxl](https://github.com/dmf-mxl/mxl), the Media
eXchange Layer C SDK.

The package mirrors the public API exposed by libmxl's
`mxl/{mxl,flow,flowinfo,time,rational,dataformat}.h`: instance management,
discrete-grain and continuous-sample I/O on both the reader and writer
sides, synchronization groups, and the time/index helpers.

## Install

```sh
go get github.com/qvest-digital/go-mxl
```

## Build requirements

The package is a cgo wrapper around libmxl. To build:

- `libmxl` installed with headers and a working `libmxl.pc`
- `pkg-config` on `PATH`
- `cgo` enabled (`CGO_ENABLED=1`, the default for native builds)

Verify with:

```sh
pkg-config --cflags --libs libmxl
```

If that returns include and link flags, `go build` works without any
extra environment.

### Non-default install locations

If libmxl is installed somewhere `pkg-config` does not search by
default, point it at the directory containing the `libmxl.pc` file. A `.env` at the repo root
(gitignored) is one convenient place; pick whatever your editor or
shell tooling reads from.

```sh
# .env
PKG_CONFIG_PATH=/opt/libmxl/lib

# Some libmxl installs declare transitive libs as Requires.private,
# which pkg-config only emits with --static. Set CGO_LDFLAGS in
# addition to the path above if `go build` reports unresolved symbols
# against spdlog / fmt / libmxl-common.
#CGO_LDFLAGS=-L/opt/libmxl/lib -lmxl -lmxl-common -lspdlog -lfmt -lstdc++
```

## Usage

```go
package main

import (
    "errors"
    "log"
    "time"

    "github.com/qvest-digital/go-mxl/mxl"
)

func main() {
    inst, err := mxl.NewInstance("/dev/shm/mxl", "")
    if err != nil {
        log.Fatal(err)
    }
    defer inst.Close()

    r, err := inst.NewReader("<flow-uuid>")
    if err != nil {
        log.Fatal(err)
    }
    defer r.Close()

    info, _ := r.Info()
    rate := info.Config.Common.GrainRate
    idx := mxl.CurrentIndex(rate)

    for {
        g, err := r.GetGrain(idx, 200*time.Millisecond)
        switch {
        case errors.Is(err, mxl.ErrTimeout):
            idx = mxl.CurrentIndex(rate)
        case err != nil:
            log.Fatal(err)
        default:
            log.Printf("grain %d: %d bytes", g.Index, len(g.Payload))
            idx++
        }
    }
}
```

## Examples

Small `main` programs exercising each side of the API live under
[`examples/`](examples/) — local `write-*`/`read-*` pipelines, a
synchronization-group demo, and a libfabric `fabrics-target` /
`fabrics-initiator` pair. See [`examples/README.md`](examples/README.md)
for the canonical command lines.

## Container images

Prebuilt builder and runtime images are published to GHCR. See
[`docs/docker.md`](docs/docker.md).

## Releases

Releases are cut by [release-please](https://github.com/googleapis/release-please)
from Conventional Commits in two automated stages:

1. Every merge to `main` opens (or updates) a **prerelease PR**
   proposing the next `vX.Y.Z-rc.N`. Merge it to tag the candidate.
2. Tagging a prerelease automatically opens a **release PR** for the
   matching final `vX.Y.Z`. Merge that to graduate.

Both PRs are optional — merge only when you want the corresponding
tag. Downstream pins either flavour explicitly:

```sh
go get github.com/qvest-digital/go-mxl@v1.0.0-rc.1
go get github.com/qvest-digital/go-mxl@v1.0.0
```

The two stages are driven by `.github/prerelease-config.json` and
`.github/release-config.json`; the release config keeps the prerelease
manifest in sync via `extra-files`.

## Memory safety

Byte slices returned by reads (`Grain.Payload`, `SamplesView` fragments)
and by writer-side `OpenGrain` / `OpenSamples` alias libmxl's shared
memory directly. They are only valid until the next read, the matching
`Commit` / `Cancel`, or the `Reader` / `Writer` being closed. Use
`Grain.Copy()` or `SamplesView.CopyChannel()` to retain data.

## Graphify

This repo carries a [Graphify](https://github.com/safishamsi/graphify)
knowledge graph under `graphify-out/`. The graph is committed so a
fresh clone already has it; `.graphifyignore` controls what gets
indexed.

Graphify is optional. To rebuild the graph locally, query it, or
have it auto-rebuild after each commit, install the `graphifyy`
PyPI package (CLI: `graphify`) and run `graphify hook install`.
Manage the hook with `graphify hook status` and
`graphify hook uninstall`.

See the upstream
[common commands](https://github.com/safishamsi/graphify#common-commands)
for usage.

## License

[Apache-2.0](LICENSE)
