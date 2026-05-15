# syntax=docker/dockerfile:1.7
#
# go-mxl builder + runtime base images.
#
# - builder: Debian trixie with the Go toolchain, clang-19, and libmxl
#   + libmxl-fabrics installed under /opt/libmxl. Downstream projects
#   FROM this as their build stage; the libmxl pkg-config files are
#   already on PKG_CONFIG_PATH so `#cgo pkg-config: libmxl
#   libmxl-fabrics` resolves out of the box.
# - runtime: Debian trixie-slim with the libmxl runtime shared objects
#   and libfabric. No toolchain, no headers, no vcpkg artefacts.
#
# The libmxl revision is taken from .github/libmxl.version (a GitHub
# tree URL); pass --build-arg LIBMXL_VERSION=<url> to override.

ARG GO_VERSION=1.26.0
ARG DEBIAN_TAG=trixie


FROM debian:${DEBIAN_TAG} AS builder

ARG GO_VERSION
ARG TARGETARCH=amd64

# libmxl-fabrics calls fi_fabric2, introduced in libfabric 2.x; trixie
# ships 2.1 via apt. The gstreamer packages cover libmxl's bundled
# tools — they're not used at runtime but the cmake build references
# them.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential cmake ninja-build clang-19 lld-19 \
      pkg-config bison flex curl zip unzip tar git ca-certificates \
      libfabric-dev libfabric-bin \
      libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
 && rm -rf /var/lib/apt/lists/* \
 && ln -sf /usr/bin/clang-19    /usr/local/bin/clang \
 && ln -sf /usr/bin/clang++-19  /usr/local/bin/clang++ \
 && ln -sf /usr/bin/ld.lld-19   /usr/local/bin/ld.lld

RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" \
      | tar -C /usr/local -xz
ENV PATH=/usr/local/go/bin:${PATH} \
    GOTOOLCHAIN=local

# libmxl's CMake preset hard-codes the toolchain file at
# `~/vcpkg/scripts/buildsystems/vcpkg.cmake`. CMake expands `~` to the
# user's $HOME, so vcpkg has to live there (not at /opt/vcpkg).
# --filter=tree:0 keeps the clone small but still lets `git show`
# resolve the builtin-baseline commit referenced in libmxl's vcpkg.json;
# plain --depth 1 breaks vcpkg's baseline checkout.
RUN git clone --filter=tree:0 https://github.com/microsoft/vcpkg.git /root/vcpkg \
 && /root/vcpkg/bootstrap-vcpkg.sh -disableMetrics
ENV VCPKG_ROOT=/root/vcpkg \
    PATH=/root/vcpkg:${PATH}

ARG LIBMXL_VERSION=
COPY .github/libmxl.version /tmp/libmxl.version
RUN set -eu; \
    url=${LIBMXL_VERSION:-$(tr -d '[:space:]' < /tmp/libmxl.version)}; \
    rest=${url#https://github.com/}; \
    repo=${rest%%/tree/*}; \
    ref=${rest#*/tree/}; \
    if [ "$rest" = "$ref" ] || [ -z "$repo" ] || [ -z "$ref" ]; then \
      echo "expected https://github.com/<owner>/<repo>/tree/<ref>, got: $url" >&2; \
      exit 1; \
    fi; \
    git clone "https://github.com/$repo.git" /tmp/mxl; \
    git -C /tmp/mxl checkout "$ref"; \
    cd /tmp/mxl; \
    cmake --preset Linux-Clang-Release \
          -DBUILD_DOCS=OFF \
          -DMXL_ENABLE_FABRICS_OFI=ON \
          -DCMAKE_INSTALL_PREFIX=/opt/libmxl; \
    cmake --build  build/Linux-Clang-Release; \
    cmake --install build/Linux-Clang-Release; \
    # libmxl.pc and libmxl-fabrics.pc declare `Requires.private: spdlog`
    # (and fmt). `cmake --install` only ships libmxl's own .pc files;
    # vcpkg's spdlog.pc/fmt.pc live in the build tree. Copy them into
    # the install prefix so the image is self-contained on a single
    # PKG_CONFIG_PATH entry. spdlog and fmt are statically linked into
    # libmxl{,-fabrics}.so, so the runtime image needs neither.
    cp build/Linux-Clang-Release/vcpkg_installed/x64-linux/lib/pkgconfig/*.pc \
       /opt/libmxl/lib/pkgconfig/; \
    rm -rf /tmp/mxl /tmp/libmxl.version

ENV PKG_CONFIG_PATH=/opt/libmxl/lib/pkgconfig \
    LD_LIBRARY_PATH=/opt/libmxl/lib

# `go build` runs `git status` for -buildvcs stamping. When the image
# is mounted on a checkout owned by an arbitrary uid, git refuses with
# the safe.directory error unless we whitelist it.
RUN git config --system --add safe.directory '*'

WORKDIR /src


FROM debian:${DEBIAN_TAG}-slim AS runtime

# libfabric1 pulls in libnuma1, libucx0, libuv1, libstdc++6 — i.e.
# everything libmxl-fabrics dlopen()s at runtime via libfabric's
# provider plugins.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      libfabric1 ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /opt/libmxl/lib/ /opt/libmxl/lib/
# Drop install-time artefacts (pkg-config, cmake exports) and libmxl's
# bundled gstreamer plugin — the latter pulls a full gstreamer runtime
# that this image deliberately omits.
RUN rm -rf \
      /opt/libmxl/lib/pkgconfig \
      /opt/libmxl/lib/cmake \
      /opt/libmxl/lib/gstreamer-1.0

ENV LD_LIBRARY_PATH=/opt/libmxl/lib
