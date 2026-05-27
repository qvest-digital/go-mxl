# syntax=docker/dockerfile:1.24
#
# go-mxl builder + runtime base images.
#
# - libfabric-stack: Debian trixie with rdma-core and libfabric built
#   from upstream release tarballs, installed under /opt/libfabric.
#   Cached independently from builder and runtime — only invalidates
#   when the RDMA_CORE_VERSION / LIBFABRIC_VERSION ARGs or the base
#   image change.
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
#
# libfabric does not come from Debian: trixie ships 2.1, but EFA on
# AWS needs >=2.4 to register PROT_READ-only memory (dmf-mxl/mxl#516),
# and Debian's build has --enable-efa off / no EFA ibverbs provider.
# The libfabric-stack stage below builds libfabric and rdma-core from
# upstream releases instead, so EFA, verbs (all rdma-core providers),
# tcp, udp, and shm are all compiled in.

ARG GO_VERSION=1.26.0
ARG DEBIAN_TAG=trixie-20260518

# renovate: datasource=github-releases depName=linux-rdma/rdma-core
ARG RDMA_CORE_VERSION=v63.0
# renovate: datasource=github-releases depName=ofiwg/libfabric
ARG LIBFABRIC_VERSION=v2.5.1


# Stage 1: build the full rdma-core + libfabric userspace stack from
# upstream releases, installed under /opt/libfabric. Cached
# independently from builder and runtime — only invalidates when the
# two ARGs above (or the base image) change.
FROM debian:${DEBIAN_TAG} AS libfabric-stack

ARG RDMA_CORE_VERSION
ARG LIBFABRIC_VERSION

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential cmake ninja-build pkg-config \
      autoconf automake libtool m4 \
      libnl-3-dev libnl-route-3-dev libudev-dev libsystemd-dev \
      python3 \
      curl ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# rdma-core: ships libibverbs, librdmacm, and the full set of
# userspace verbs providers (efa, mlx5, mlx4, hns, irdma, siw, rxe,
# …). All shared, no pyverbs, no man pages.
#
# CMAKE_INSTALL_SYSCONFDIR is pinned to /opt/libfabric/etc so that
# rdma-core's CONFIG_DIR (${SYSCONFDIR}/libibverbs.d) and the matching
# compile-time macro inside libibverbs.so.1 both land inside the
# /opt/libfabric prefix. Without the override, GNUInstallDirs picks
# /etc/opt/libfabric for a /opt prefix per FHS, the .driver files
# install outside the prefix, the single `COPY --from=libfabric-stack
# /opt/libfabric/ /opt/libfabric/` below misses them, and the runtime
# verbs provider returns -FI_ENODATA because libibverbs sees no
# drivers.
RUN set -eu; \
    v=${RDMA_CORE_VERSION#v}; \
    curl -fsSL \
      "https://github.com/linux-rdma/rdma-core/archive/refs/tags/${RDMA_CORE_VERSION}.tar.gz" \
      | tar -xz -C /tmp; \
    cd /tmp/rdma-core-${v}; \
    cmake -B build -G Ninja \
        -DCMAKE_INSTALL_PREFIX=/opt/libfabric \
        -DCMAKE_INSTALL_SYSCONFDIR=/opt/libfabric/etc \
        -DCMAKE_BUILD_TYPE=Release \
        -DNO_PYVERBS=1 -DNO_MAN_PAGES=1 -DENABLE_STATIC=0; \
    ninja -C build; \
    ninja -C build install; \
    rm -rf /tmp/rdma-core-${v}

# libfabric: stable providers stay at "auto" so they enable themselves
# when their deps are present (efa + verbs find the rdma-core install
# above; tcp / udp / shm / sm2 / rxm / rxd have no external deps). The
# --disable list explicitly turns off vendor providers we have no SDK
# for and experimental providers we do not want surprise-enabling on a
# future libfabric bump.
RUN set -eu; \
    v=${LIBFABRIC_VERSION#v}; \
    curl -fsSL \
      "https://github.com/ofiwg/libfabric/archive/refs/tags/${LIBFABRIC_VERSION}.tar.gz" \
      | tar -xz -C /tmp; \
    cd /tmp/libfabric-${v}; \
    PKG_CONFIG_PATH=/opt/libfabric/lib/pkgconfig \
    CPPFLAGS=-I/opt/libfabric/include \
    LDFLAGS="-L/opt/libfabric/lib -Wl,-rpath-link,/opt/libfabric/lib" \
    ./autogen.sh; \
    PKG_CONFIG_PATH=/opt/libfabric/lib/pkgconfig \
    CPPFLAGS=-I/opt/libfabric/include \
    LDFLAGS="-L/opt/libfabric/lib -Wl,-rpath-link,/opt/libfabric/lib" \
    ./configure --prefix=/opt/libfabric \
        --disable-debug --disable-static \
        --disable-psm2 --disable-psm3 --disable-opx \
        --disable-usnic --disable-cxi \
        --disable-mlx --disable-ucx \
        --disable-mrail --disable-lpp --disable-lnx; \
    make -j"$(nproc)"; \
    make install; \
    rm -rf /tmp/libfabric-${v}


FROM debian:${DEBIAN_TAG} AS builder

ARG GO_VERSION
# buildx auto-populates TARGETARCH (amd64, arm64, …) when a value is
# not supplied. An explicit default would shadow that, so leave it
# blank.
ARG TARGETARCH

# gstreamer-dev satisfies libmxl's bundled gstreamer plugin build; the
# plugin itself is pruned out of the runtime image. libnl/libudev
# runtime libs are pulled in so the rdma-core .so's COPY'd from
# libfabric-stack resolve at link time (libmxl-fabrics links against
# libfabric, which links against librdmacm/libibverbs, which link
# against libnl).
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential cmake ninja-build clang-19 lld-19 \
      pkg-config bison flex curl zip unzip tar git ca-certificates \
      libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev \
      libnl-3-200 libnl-route-3-200 libudev1 \
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

# Pull in the rdma-core + libfabric userspace stack from the dedicated
# build stage. ld.so.conf.d/libfabric.conf wires /opt/libfabric/lib
# into the dynamic linker search path so libfabric.so.1 resolves
# without LD_LIBRARY_PATH at link time.
COPY --from=libfabric-stack /opt/libfabric/ /opt/libfabric/
RUN echo /opt/libfabric/lib > /etc/ld.so.conf.d/libfabric.conf \
 && ldconfig

# PKG_CONFIG_PATH must point at libfabric.pc before the libmxl build
# below runs cmake (the .pc file is not on the default search path
# because it lives in /opt/libfabric/lib/pkgconfig). The libmxl install
# populates /opt/libmxl/lib/pkgconfig at the end of its RUN, so
# prepending that entry now is harmless — pkg-config silently skips
# non-existent dirs.
ENV PKG_CONFIG_PATH=/opt/libfabric/lib/pkgconfig:/opt/libmxl/lib/pkgconfig \
    LD_LIBRARY_PATH=/opt/libmxl/lib:/opt/libfabric/lib \
    PATH=/opt/libfabric/bin:${PATH}

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
    # vcpkg installs into a single per-host directory under
    # vcpkg_installed/ (e.g. x64-linux/ on amd64, arm64-linux/ on
    # arm64); the *-linux/ glob matches it without arch detection.
    cp build/Linux-Clang-Release/vcpkg_installed/*-linux/lib/pkgconfig/*.pc \
       /opt/libmxl/lib/pkgconfig/; \
    rm -rf /tmp/mxl /tmp/libmxl.version

# `go build` runs `git status` for -buildvcs stamping. When the image
# is mounted on a checkout owned by an arbitrary uid, git refuses with
# the safe.directory error unless we whitelist it.
RUN git config --system --add safe.directory '*'

WORKDIR /src


FROM debian:${DEBIAN_TAG}-slim AS runtime

# Debian's libfabric1 used to drag libnuma1/libuv1/libstdc++6 in as
# transitive deps; the upstream-built libfabric does not, so they are
# listed explicitly. libnl-3-200 / libnl-route-3-200 / libudev1 are
# the runtime link deps of the rdma-core shared objects COPY'd from
# libfabric-stack.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates libnuma1 libuv1 libstdc++6 \
      libnl-3-200 libnl-route-3-200 libudev1 \
 && rm -rf /var/lib/apt/lists/*

# Pull in the rdma-core + libfabric userspace stack from the dedicated
# build stage. ld.so.conf.d/libfabric.conf matches what is written in
# the builder stage so /opt/libfabric/lib/libfabric.so.1 resolves
# without LD_LIBRARY_PATH.
COPY --from=libfabric-stack /opt/libfabric/ /opt/libfabric/
RUN echo /opt/libfabric/lib > /etc/ld.so.conf.d/libfabric.conf \
 && ldconfig

COPY --from=builder /opt/libmxl/lib/ /opt/libmxl/lib/
# Drop install-time artefacts (pkg-config, cmake exports) and libmxl's
# bundled gstreamer plugin — the latter pulls a full gstreamer runtime
# that this image deliberately omits.
RUN rm -rf \
      /opt/libmxl/lib/pkgconfig \
      /opt/libmxl/lib/cmake \
      /opt/libmxl/lib/gstreamer-1.0

ENV LD_LIBRARY_PATH=/opt/libmxl/lib:/opt/libfabric/lib
