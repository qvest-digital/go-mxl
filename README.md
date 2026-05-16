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

See [examples/](examples/) for read/write programs covering both
discrete grain and continuous sample flows, plus a synchronization
group example.

## Container images

Prebuilt builder and runtime images are published to GHCR. See
[`docs/docker.md`](docs/docker.md).

## Releases

Releases are cut by [release-please](https://github.com/googleapis/release-please)
from Conventional Commits. Until the API stabilises, the release PR
proposes `vX.Y.Z-rc.N` prereleases; merge one to publish that tag.
Downstream pins explicitly:

```sh
go get github.com/qvest-digital/go-mxl@v1.0.0-rc.1
```

To cut a non-prerelease tag, land a commit on `main` with a
`Release-As: <version>` footer, or drop `versioning-strategy`,
`prerelease-type`, and `prerelease` from `.release-please-config.json`.

## Memory safety

Byte slices returned by reads (`Grain.Payload`, `SamplesView` fragments)
and by writer-side `OpenGrain` / `OpenSamples` alias libmxl's shared
memory directly. They are only valid until the next read, the matching
`Commit` / `Cancel`, or the `Reader` / `Writer` being closed. Use
`Grain.Copy()` or `SamplesView.CopyChannel()` to retain data.

## License

[Apache-2.0](LICENSE)
