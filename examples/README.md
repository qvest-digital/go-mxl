# examples

Each subdirectory is a small `main` program exercising one side of the
public API. Build them with `go build ./examples/...` from the repo
root (or run inline with `go run ./examples/<name>`). The local
pipeline pairs a `write-*` producer with the matching `read-*` consumer
on the same MXL domain; the fabric pipeline bridges two domains across
libfabric.

`flow.json` below is a libmxl NMOS flow definition and `<flow-id>` is
its `id` field — supply your own.

## Local grain (video/data) flow

```sh
write-grain -domain /dev/shm/mxl -flow-def flow.json
read-grain  -domain /dev/shm/mxl -flow <flow-id>
```

## Local samples (audio) flow

```sh
write-samples -domain /dev/shm/mxl -flow-def flow.json
read-samples  -domain /dev/shm/mxl -flow <flow-id>
```

## Multiple flows in lock-step

`sync-group` ticks every time data is available at the requested
timestamp across every reader.

```sh
sync-group -domain /dev/shm/mxl -rate 30000/1001 \
    -flow <flow-id-1> -flow <flow-id-2>
```

## Fabric transport

`fabrics-target` binds a libfabric endpoint and accepts incoming grains
into a local flow. `fabrics-initiator` reads from a local flow and
pushes grains to one or more targets. The same `flow.json` must be used
at both ends.

```sh
# 1. produce into the source domain
write-grain -domain /dev/shm/mxl-src -flow-def flow.json

# 2. bind the target on the destination domain; writes target.info
fabrics-target -domain /dev/shm/mxl-tgt -flow-def flow.json \
    -node 127.0.0.1 -service 23456 -target-info target.info

# 3. push grains from source to target
fabrics-initiator -domain /dev/shm/mxl-src -flow <flow-id> \
    -node 127.0.0.1 -service 23457 -target-info target.info

# 4. consume the replicated flow
read-grain -domain /dev/shm/mxl-tgt -flow <flow-id>
```

`-provider` defaults to `tcp`, so the pipeline runs on a single host
without RDMA hardware. Switch to `verbs`, `efa`, or `shm` when the
deployment supports it. The published runtime image
(`ghcr.io/qvest-digital/go-mxl-runtime`) ships libfabric from
[`aws-efa-installer`](../.github/aws-efa-installer.version) so the
`efa` provider works on EFA-enabled EC2 nodes out of the box; a
local builder that uses Debian apt's libfabric instead will not.
