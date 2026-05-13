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

## Usage

```go
package main

import (
    "errors"
    "log"
    "time"

    mxl "github.com/qvest-digital/go-mxl"
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

## Memory safety

Byte slices returned by reads (`Grain.Payload`, `SamplesView` fragments)
and by writer-side `OpenGrain` / `OpenSamples` alias libmxl's shared
memory directly. They are only valid until the next read, the matching
`Commit` / `Cancel`, or the `Reader` / `Writer` being closed. Use
`Grain.Copy()` or `SamplesView.CopyChannel()` to retain data.

## License

[Apache-2.0](LICENSE)
