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

The runtime image sets `LD_LIBRARY_PATH=/opt/libmxl/lib:/opt/libfabric/lib`,
so the linked `libmxl.so` / `libmxl-fabrics.so` and the EFA-capable
`libfabric.so.1` shipped under `/opt/libfabric/lib` resolve without
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

## RDMA stack source

Both stages install libfabric and rdma-core from the upstream
GitHub release tarballs into `/opt/libfabric/`, not from Debian
apt. Debian trixie's `libfabric1` 2.1.0-1.1 fails on AWS EFA in
two independent ways:

1. **Missing 2.4 read-only-mmap fix.** `fi_mr_regattr` with
   `FI_WRITE` on a `PROT_READ`-only mapping returns `EFAULT` in
   `libfabric < 2.4`; libmxl-fabrics hits this path on every
   reader-side memory region (dmf-mxl/mxl#516).
2. **No EFA provider.** Debian's `debian/rules` builds libfabric
   with `--enable-efa` off, and `ibverbs-providers` 61.x in
   trixie ships drivers for hns/mlx4/mlx5/… but no EFA userspace
   driver. Inside a Debian-trixie container on an EFA node:
   `fi_getinfo: provider efa output empty list`.

A dedicated `libfabric-stack` build stage compiles `rdma-core` and
then `libfabric` from upstream release tarballs. The two versions
are pinned via `RDMA_CORE_VERSION` and `LIBFABRIC_VERSION` ARGs at
the top of `docker/Dockerfile`, each annotated with a
`# renovate: datasource=github-releases` comment so Renovate
proposes bumps directly against the Dockerfile — no separate pin
file. Both the builder and runtime stages `COPY --from=libfabric-stack`
the installed `/opt/libfabric/` tree, so the upstream stack caches
independently from the libmxl rebuild loop.

## Provider and hardware coverage

The build flags pin the libfabric provider matrix below. Providers
left at libfabric's `auto` default enable themselves when their
build-time deps are present; the rdma-core install satisfies the
verbs deps for every NIC driver the kernel exposes. Vendor and
experimental providers without an SDK in the image are disabled
explicitly so a future libfabric bump cannot surprise-enable them.

**Supported**

| libfabric provider | Hardware reached | Kernel-side prereq |
| --- | --- | --- |
| `efa` | AWS Elastic Fabric Adapter on EC2 HPC / GPU instance families. | Linux `efa` kernel module (ships with AWS-built AMIs). |
| `verbs` | NVIDIA / Mellanox ConnectX-3 onward (CX-3, CX-4, CX-4 Lx, CX-5, CX-6, CX-7) and BlueField-1 / -2 / -3 DPUs over InfiniBand or RoCE/v2; Intel E810 iWARP; Intel Omni-Path 100-series HFI1 (verbs fallback path); Broadcom NetXtreme-E RoCE; Chelsio T4 / T5 / T6 iWARP; HiSilicon Kunpeng HiP06 / 07 / 08; Marvell / QLogic FastLinQ; AMD Pensando DSC; Microsoft Azure MANA; Alibaba ERDMA; VMware paravirtual RDMA in ESXi guests; Soft-RoCE (`rxe`) over any Ethernet NIC; Soft-iWARP (`siw`) over any TCP-capable NIC. | Matching in-tree kernel driver: `mlx4_ib`, `mlx5_ib`, `irdma`, `hfi1`, `bnxt_re`, `iw_cxgb4`, `hns_roce`, `qedr`, `ionic_rdma`, `mana_ib`, `erdma`, `vmw_pvrdma`, `rdma_rxe`, `siw`, … (Linux >= 5.x ships them all). |
| `tcp` | Any IP-capable NIC. Fallback path; no RDMA required. | None. |
| `udp` | Same as `tcp`, datagram. | None. |
| `shm`, `sm2` | Intra-host, between processes on the same kernel. | `/dev/shm` writeable. |
| `rxm`, `rxd` | Utility providers; layer reliable-message or reliable-datagram semantics over an underlying provider (typically `tcp` or `verbs`). Not selected directly; picked up via `FI_PROVIDER=tcp;ofi_rxm` style chaining. | Inherited from the base provider. |
| Hook providers (`hook_debug`, `hook_hmem`, `dmabuf_peer_mem`, …) | Debug instrumentation + GPU dma-buf integration. Not transports. | None. |

**Not supported in this build**

| libfabric provider compiled out | Hardware this would have reached | Why off |
| --- | --- | --- |
| `psm2` | Intel Omni-Path 100 / 200 series via the PSM2 high-performance path. (HFI1 hardware remains reachable through the `verbs` provider — slower than PSM2 but functional.) | psm2 userspace SDK not present; Intel deprecated Omni-Path in 2019. |
| `psm3` | Intel Ethernet 800 series via PSM3. | psm3 userspace SDK not present. |
| `opx` | Intel Omni-Path Express (Cornelis CN5000). | OPX SDK not present; niche HPC hardware. |
| `usnic` | Cisco UCS VIC adapters via the native usNIC interface. (Basic verbs may still bind on some VICs.) | Cisco usnic userspace lib not present. |
| `cxi` | HPE Cray Slingshot-11 interconnect. | HPE Cassini SDK not present; restricted distribution. |
| `gni` | Cray Gemini / Aries (pre-Slingshot). | Hardware EOL; SDK not present. |
| `bgq` | IBM Blue Gene/Q. | Hardware EOL. |
| `ucx` | Higher-level wrapper over the same NVIDIA/Mellanox cards the `verbs` provider already covers. Useful for DC transport, GPUDirect, richer atomics; not used by libmxl-fabrics today. | The Go binding's `Provider` enum has no `ProviderUCX`; including it would only affect `ProviderAuto`'s registration order. One-line build flag away when libmxl exposes the enum. |
| `mlx` | Legacy Mellanox MXM (predecessor of UCX). | Deprecated by NVIDIA. |
| `mrail`, `lnx`, `lpp` | Experimental upstream composite / linked-provider work. | Not stable. |

So in short:

- **All major RDMA NIC vendors with mainline Linux drivers
  (NVIDIA/Mellanox CX-3 onward, Intel iWARP, Broadcom, Chelsio,
  HiSilicon, Marvell, AMD Pensando, Azure MANA, Alibaba) plus
  AWS EFA are supported via `verbs` or `efa`.**
- **Hardware needing vendor-specific HPC SDKs (Intel Omni-Path
  PSM2/PSM3/OPX, Cisco usNIC native, HPE Slingshot, Cray
  Gemini/Aries, IBM BG/Q) is not supported.**
- **TCP / UDP / SHM fallbacks are always available.**

## Why trixie-slim and not distroless

The runtime image is trixie-slim plus the libmxl shared objects
and the `/opt/libfabric/` tree from the `libfabric-stack` stage.
Trixie's apt resolves the handful of shared objects libfabric and
rdma-core depend on (`libnuma`, `libuv`, libc, libstdc++,
`libnl-3`, `libnl-route-3`, `libudev`, …) without manual
curation; replicating that set with explicit `COPY`s from a
distroless base is brittle and breaks the day a new provider
lands. The resulting image is still under 200 MB.
