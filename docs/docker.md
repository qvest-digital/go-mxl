# Container images

Two images are published to GHCR on every push to `main` and on every
`v*.*.*` tag. Each tag is a multi-arch manifest covering `linux/amd64`
and `linux/arm64`, so `docker pull` selects the right variant
automatically (Apple Silicon machines under Docker Desktop included):

| Image | Purpose |
| --- | --- |
| `ghcr.io/qvest-digital/go-mxl-builder` | Build stage. Debian trixie with Go, clang-19, and libmxl + libmxl-fabrics installed under `/opt/libmxl`. `PKG_CONFIG_PATH` is preset so `#cgo pkg-config: libmxl libmxl-fabrics` resolves with no extra environment. |
| `ghcr.io/qvest-digital/go-mxl-runtime` | Runtime stage. Debian trixie-slim with the libmxl shared objects and libfabric. No toolchain, no headers. |

## Tags

- `:latest` — the most recently tagged **stable** release.
- `:pre` — the most recently tagged **prerelease** (e.g. `vX.Y.Z-rc.N`).
  Never overlaps with `:latest`.
- `:<version>` — exact version for every Git tag, prerelease or
  stable (e.g. `:1.0.0`, `:1.0.0-rc.0`).
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

The runtime image sets `LD_LIBRARY_PATH=/opt/libmxl/lib:/opt/amazon/efa/lib`,
so the linked `libmxl.so` / `libmxl-fabrics.so` and the EFA-capable
`libfabric.so.1` shipped under `/opt/amazon/efa/lib` resolve without
extra wiring.

## Building locally

```sh
docker build -f docker/Dockerfile --target builder -t go-mxl-builder .
docker build -f docker/Dockerfile --target runtime -t go-mxl-runtime .
```

The libmxl revision is read from `.github/libmxl.version` inside the
build context. Pass `--build-arg LIBMXL_VERSION=<github-tree-url>` to
override.

## Dev containers

`.devcontainer/` ships two configurations that reuse the same
`docker/Dockerfile` CI publishes. VS Code's "Dev Containers: Reopen
in Container" shows a picker:

| Config | What it does | When to pick it |
| --- | --- | --- |
| `image` | Pulls `ghcr.io/qvest-digital/go-mxl-builder:dev` | Default. Fast start, matches CI exactly. |
| `local` | Builds the same Dockerfile locally (`target: builder`) | Iterating on the Dockerfile or the libmxl pin without going through CI. |

Both mount the repo at `/src`, pass `--tmpfs=/dev/shm` so the
`mxl_integration` suite has a real tmpfs, and preinstall the `golang.go`
extension with `go.buildTags=mxl_integration` so gopls sees the
integration-tagged sources.

The `image` variant doesn't auto-refresh when CI publishes a new
`:dev`; run "Dev Containers: Rebuild Container" to pull it.

## Libfabric source

Both stages install libfabric and rdma-core from
[`aws-efa-installer`](../.github/aws-efa-installer.version) into
`/opt/amazon/efa/`, not from Debian apt. Debian trixie's
`libfabric1` 2.1.0-1.1 fails on AWS EFA in two independent ways:

1. **Missing 2.4 read-only-mmap fix.** `fi_mr_regattr` with
   `FI_WRITE` on a `PROT_READ`-only mapping returns `EFAULT` in
   `libfabric < 2.4`; libmxl-fabrics hits this path on every
   reader-side memory region (dmf-mxl/mxl#516).
2. **No EFA provider.** Debian's `debian/rules` builds libfabric
   with `--enable-efa` off, and `ibverbs-providers` 61.x in
   trixie ships drivers for hns/mlx4/mlx5/… but no EFA userspace
   driver. Inside a Debian-trixie container on an EFA node:
   `fi_getinfo: provider efa output empty list`.

The `aws-efa-installer` tarball ships native Debian 13 `.deb`s for
both `x86_64` and `aarch64` with EFA enabled (`libfabric1-aws`) and
a matching AWS `rdma-core` that includes the EFA userspace
provider. The installer URL and sha256 are pinned in
[`.github/aws-efa-installer.version`](../.github/aws-efa-installer.version);
that file documents the manual bump procedure (AWS publishes no
release feed Renovate or any other tool can track).

## Why trixie-slim and not distroless

The runtime image is trixie-slim plus the libmxl and
`aws-efa-installer` `.deb`s. Trixie's apt resolves the handful of
shared objects the AWS libfabric depends on (`libnuma`, `libuv`,
libc, libstdc++, …) without manual curation; replicating that set
with explicit `COPY`s from a distroless base is brittle and breaks
the day a new provider lands. The resulting image is still under
200 MB.
