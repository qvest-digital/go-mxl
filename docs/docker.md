# Container images

Two images are published to GHCR on every push to `main` and on every
`v*.*.*` tag:

| Image | Purpose |
| --- | --- |
| `ghcr.io/qvest-digital/go-mxl-builder` | Build stage. Debian trixie with Go, clang-19, and libmxl + libmxl-fabrics installed under `/opt/libmxl`. `PKG_CONFIG_PATH` is preset so `#cgo pkg-config: libmxl libmxl-fabrics` resolves with no extra environment. |
| `ghcr.io/qvest-digital/go-mxl-runtime` | Runtime stage. Debian trixie-slim with the libmxl shared objects and libfabric. No toolchain, no headers. |

## Tags

- `:latest` — the most recently tagged release.
- `:<version>` — e.g. `:0.2.0`, for the matching Git tag.
- `:dev` — head of `main`.
- `:libmxl-<shortsha>` — content-addressed tag derived from the libmxl
  revision pinned in [`.github/libmxl.version`](../.github/libmxl.version).
  Use this when you need to pin to a specific libmxl build across
  go-mxl versions.

## Usage

```dockerfile
FROM ghcr.io/qvest-digital/go-mxl-builder:latest AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o /out/app ./cmd/app

FROM ghcr.io/qvest-digital/go-mxl-runtime:latest
COPY --from=builder /out/app /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
```

The runtime image already sets `LD_LIBRARY_PATH=/opt/libmxl/lib`, so the
linked `libmxl.so` / `libmxl-fabrics.so` resolve without extra wiring.

## Building locally

```sh
docker build -f docker/Dockerfile --target builder -t go-mxl-builder .
docker build -f docker/Dockerfile --target runtime -t go-mxl-runtime .
```

The libmxl revision is read from `.github/libmxl.version` inside the
build context. Pass `--build-arg LIBMXL_VERSION=<github-tree-url>` to
override.

## Why trixie-slim and not distroless

`libmxl-fabrics` links against libfabric 2.x. Debian trixie is the
oldest stable distribution that ships libfabric 2 (`libfabric1` v2.1).
At runtime libfabric `dlopen()`s its provider plugins, which pull a
handful of transitive shared objects (`libnuma`, `libucx`, `libuv`,
…). Reaching for trixie-slim and `apt install libfabric1` lets apt's
dependency solver pick those up; replicating the same set with manual
`COPY`s from a distroless base is brittle and silently breaks the day
a new provider lands. The resulting image is still under 200 MB.
