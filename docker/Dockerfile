# syntax=docker/dockerfile:1.24
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
#
# libfabric does not come from Debian: trixie ships 2.1, but EFA on
# AWS needs >=2.4 to register PROT_READ-only memory (dmf-mxl/mxl#516),
# and Debian's build has --enable-efa off / no EFA ibverbs provider.
# Both stages install libfabric + rdma-core from aws-efa-installer
# under /opt/amazon/efa/ instead. The installer URL + sha256 are
# pinned in .github/aws-efa-installer.version; override with
# --build-arg AWS_EFA_INSTALLER_URL=<url> (sha256 check is skipped
# when the URL is overridden).

ARG GO_VERSION=1.26.0
ARG DEBIAN_TAG=trixie-20260505


FROM debian:${DEBIAN_TAG} AS builder

ARG GO_VERSION
# buildx auto-populates TARGETARCH (amd64, arm64, …) when a value is
# not supplied. An explicit default would shadow that, so leave it
# blank.
ARG TARGETARCH

# gstreamer-dev satisfies libmxl's bundled gstreamer plugin build;
# the plugin itself is pruned out of the runtime image. libfabric is
# installed from aws-efa-installer further down, not from apt.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential cmake ninja-build clang-19 lld-19 \
      pkg-config bison flex curl zip unzip tar git ca-certificates \
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

# aws-efa-installer ships native DEBIAN13 .debs for x86_64 and aarch64.
# Install libfabric1-aws + libfabric-aws-dev (headers for the
# libmxl-fabrics cmake build) + libfabric-aws-bin (fi_info for the
# CI smoke check), plus the matching AWS rdma-core set — the only
# place the EFA userspace ibverbs provider ships on Debian.
# Runtime-only .debs are staged under /opt/aws-efa-debs so the
# runtime stage can COPY+install them without re-downloading the
# 600+ MB tarball.
ARG AWS_EFA_INSTALLER_URL=
COPY .github/aws-efa-installer.version /tmp/aws-efa.version
RUN set -eu; \
    if [ -n "$AWS_EFA_INSTALLER_URL" ]; then \
      url=$AWS_EFA_INSTALLER_URL; sha=; \
    else \
      url=$(awk -F= '$1=="url"{print $2}'    /tmp/aws-efa.version); \
      sha=$(awk -F= '$1=="sha256"{print $2}' /tmp/aws-efa.version); \
    fi; \
    case "$TARGETARCH" in \
      amd64) efa_arch=x86_64 ;; \
      arm64) efa_arch=aarch64 ;; \
      *) echo "unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    curl -fsSL "$url" -o /tmp/aws-efa.tgz; \
    if [ -n "$sha" ]; then echo "$sha  /tmp/aws-efa.tgz" | sha256sum -c -; fi; \
    tar -xzf /tmp/aws-efa.tgz -C /tmp/; \
    debs=/tmp/aws-efa-installer/DEBS/DEBIAN13/${efa_arch}; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
        "$debs"/libfabric1-aws_*.deb \
        "$debs"/libfabric-aws-dev_*.deb \
        "$debs"/libfabric-aws-bin_*.deb \
        "$debs"/rdma-core/*.deb; \
    # Stage runtime-only .debs (drop -dev / -utils variants) for the
    # runtime stage's COPY --from=builder. Open MPI, NCCL plugin,
    # prrte, pmix, the efa kmod and the libfabric debug build stay
    # out of both stages.
    mkdir -p /opt/aws-efa-debs; \
    cp "$debs"/libfabric1-aws_*.deb /opt/aws-efa-debs/; \
    cp "$debs"/rdma-core/*.deb      /opt/aws-efa-debs/; \
    rm -f /opt/aws-efa-debs/*-dev_*.deb \
          /opt/aws-efa-debs/*-utils_*.deb; \
    echo /opt/amazon/efa/lib > /etc/ld.so.conf.d/efa.conf; \
    ldconfig; \
    rm -rf /var/lib/apt/lists/* \
           /tmp/aws-efa.tgz /tmp/aws-efa-installer /tmp/aws-efa.version

# PKG_CONFIG_PATH must point at AWS's libfabric.pc before the libmxl
# build below runs cmake (the .pc file is not on the default search
# path because it lives in /opt/amazon/efa/lib/pkgconfig). The
# libmxl install populates /opt/libmxl/lib/pkgconfig at the end of
# its RUN, so prepending that entry now is harmless — pkg-config
# silently skips non-existent dirs.
ENV PKG_CONFIG_PATH=/opt/amazon/efa/lib/pkgconfig:/opt/libmxl/lib/pkgconfig \
    LD_LIBRARY_PATH=/opt/libmxl/lib:/opt/amazon/efa/lib \
    PATH=/opt/amazon/efa/bin:${PATH}

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

# Debian's libfabric1 used to drag libnuma1/libucx0/libuv1/libstdc++6
# in as transitive deps; AWS's libfabric1-aws does not, so they are
# listed explicitly. libucx0 is intentionally omitted — AWS's
# libfabric build does not link UCX, and the ucx-cma provider is not
# in scope for EFA.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates libnuma1 libuv1 libstdc++6 \
 && rm -rf /var/lib/apt/lists/*

# Install AWS libfabric + rdma-core from the .debs staged in the
# builder. Using `apt-get install -y <path>.deb` (apt 1.1+) so apt
# resolves intra-tarball deps in one shot. ld.so.conf.d/efa.conf
# matches what the AWS installer itself writes in the non-container
# case so /opt/amazon/efa/lib/libfabric.so.1 resolves without
# LD_LIBRARY_PATH.
COPY --from=builder /opt/aws-efa-debs/ /opt/aws-efa-debs/
RUN apt-get update \
 && apt-get install -y --no-install-recommends /opt/aws-efa-debs/*.deb \
 && echo /opt/amazon/efa/lib > /etc/ld.so.conf.d/efa.conf \
 && ldconfig \
 && rm -rf /opt/aws-efa-debs /var/lib/apt/lists/*

COPY --from=builder /opt/libmxl/lib/ /opt/libmxl/lib/
# Drop install-time artefacts (pkg-config, cmake exports) and libmxl's
# bundled gstreamer plugin — the latter pulls a full gstreamer runtime
# that this image deliberately omits.
RUN rm -rf \
      /opt/libmxl/lib/pkgconfig \
      /opt/libmxl/lib/cmake \
      /opt/libmxl/lib/gstreamer-1.0

ENV LD_LIBRARY_PATH=/opt/libmxl/lib:/opt/amazon/efa/lib
