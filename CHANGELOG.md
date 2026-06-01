# Changelog

## [1.0.0-rc.9](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.8...v1.0.0-rc.9) (2026-06-01)


### Features

* **fabrics:** add sample read/transfer methods ([#54](https://github.com/qvest-digital/go-mxl/issues/54)) ([191c94e](https://github.com/qvest-digital/go-mxl/commit/191c94e81fe6040fb9567436406d4004a029477e))


### Bug Fixes

* **libmxl:** pin upstream dmf-mxl revision ([#52](https://github.com/qvest-digital/go-mxl/issues/52)) ([464bbb6](https://github.com/qvest-digital/go-mxl/commit/464bbb65b1df73a8d7a402311d96f0fceebd6b95))

## [1.0.0-rc.8](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.7...v1.0.0-rc.8) (2026-05-27)


### Bug Fixes

* **docker:** install libibverbs.d under prefix ([#49](https://github.com/qvest-digital/go-mxl/issues/49)) ([f7dacfc](https://github.com/qvest-digital/go-mxl/commit/f7dacfc645a8504e77773a64234d16b6e4e72a55))

## [1.0.0-rc.7](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.6...v1.0.0-rc.7) (2026-05-27)


### Bug Fixes

* **fabrics:** match mxl.StatusErrNotReady under ErrNotReady ([#46](https://github.com/qvest-digital/go-mxl/issues/46)) ([29f89e4](https://github.com/qvest-digital/go-mxl/commit/29f89e43d50999f39faba3e70adfb1294ca5ec85))


### Build System

* **docker:** build libfabric+rdma-core upstream ([#47](https://github.com/qvest-digital/go-mxl/issues/47)) ([ba94687](https://github.com/qvest-digital/go-mxl/commit/ba94687d5d58368f56d3458fb064be2f77085691))

## [1.0.0-rc.6](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.5...v1.0.0-rc.6) (2026-05-27)


### Build System

* **docker:** ship EFA-capable libfabric from aws-efa-installer ([#43](https://github.com/qvest-digital/go-mxl/issues/43)) ([e9828d0](https://github.com/qvest-digital/go-mxl/commit/e9828d0b42e678bf9da43ad3c9fc026e34711d94))


### Miscellaneous

* **config:** migrate config renovate.json ([#39](https://github.com/qvest-digital/go-mxl/issues/39)) ([c9211a8](https://github.com/qvest-digital/go-mxl/commit/c9211a89e1dff772201e2a6c1a12389b4e121389))
* **deps:** update debian docker tag to trixie-20260518 ([#40](https://github.com/qvest-digital/go-mxl/issues/40)) ([e367bee](https://github.com/qvest-digital/go-mxl/commit/e367bee6443ea5c7e43e456b2275e18254ae8ac2))
* ignore .omc/ except skills/ ([#42](https://github.com/qvest-digital/go-mxl/issues/42)) ([4efe9c4](https://github.com/qvest-digital/go-mxl/commit/4efe9c457c359070939f56215bdca2abb2d2aaef))

## [1.0.0-rc.5](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.4...v1.0.0-rc.5) (2026-05-17)


### Build System

* **libmxl:** bump pin to qvest-digital/mxl-dmf-demo@5703b41 ([#36](https://github.com/qvest-digital/go-mxl/issues/36)) ([b1e5011](https://github.com/qvest-digital/go-mxl/commit/b1e50114bc976a6f37c478f75e6fcf04f73f617d))

## [1.0.0-rc.4](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.3...v1.0.0-rc.4) (2026-05-17)


### Continuous Integration

* **deps:** update github actions ([#33](https://github.com/qvest-digital/go-mxl/issues/33)) ([3e32a3d](https://github.com/qvest-digital/go-mxl/commit/3e32a3def4d2b509f21e34f549f7ce1acf74d446))
* **devcontainer:** two configs reusing docker/Dockerfile ([#34](https://github.com/qvest-digital/go-mxl/issues/34)) ([5e6db97](https://github.com/qvest-digital/go-mxl/commit/5e6db970ab61266697dfbf3f1fabfeae2f91dd39))
* **docker:** always tag main's HEAD with :dev ([#29](https://github.com/qvest-digital/go-mxl/issues/29)) ([472d095](https://github.com/qvest-digital/go-mxl/commit/472d095f67452c62719bd702ea64c15cb60aa66e))
* matrix build of linux/amd64 + linux/arm64 docker images ([#32](https://github.com/qvest-digital/go-mxl/issues/32)) ([9bd7375](https://github.com/qvest-digital/go-mxl/commit/9bd73757407f50f7e874a2ab03fd928ddccd4a54))
* **release-please:** include build, ci, chore in release changelog ([#30](https://github.com/qvest-digital/go-mxl/issues/30)) ([3324b90](https://github.com/qvest-digital/go-mxl/commit/3324b90065df1f0c319c499762b1f5bac62f5f9f))


### Miscellaneous

* **deps:** update docker/dockerfile docker tag to v1.24 ([#24](https://github.com/qvest-digital/go-mxl/issues/24)) ([e8c7c95](https://github.com/qvest-digital/go-mxl/commit/e8c7c954f03c1158ad3f065c0e91e94c7bc03186))

## [1.0.0-rc.3](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.2...v1.0.0-rc.3) (2026-05-16)


### Continuous Integration

* detect release tags via git tag --points-at, drop workflow_dispatch ([#26](https://github.com/qvest-digital/go-mxl/issues/26)) ([6ac58da](https://github.com/qvest-digital/go-mxl/commit/6ac58da462008102f057af6a5cd8beb10c51dae0))

## [1.0.0-rc.2](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.1...v1.0.0-rc.2) (2026-05-16)


### Continuous Integration

* dispatch ci.yml on the tag release-please creates ([#25](https://github.com/qvest-digital/go-mxl/issues/25)) ([6603900](https://github.com/qvest-digital/go-mxl/commit/660390052b9f262491b70f1959035455c86b475e))

## [1.0.0-rc.1](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.0...v1.0.0-rc.1) (2026-05-16)


### Continuous Integration

* **docker:** publish prerelease tags as :pre and :&lt;version&gt; ([#22](https://github.com/qvest-digital/go-mxl/issues/22)) ([e1c7704](https://github.com/qvest-digital/go-mxl/commit/e1c7704b2f7c1a0aefeb405e208d9d1a32f22e75))


### Miscellaneous

* **deps:** update debian docker tag to trixie-20260505 ([#8](https://github.com/qvest-digital/go-mxl/issues/8)) ([e330e97](https://github.com/qvest-digital/go-mxl/commit/e330e970babfa8ec61e31145e22775233f9d9f4b))
* **deps:** update docker/dockerfile docker tag to v1.23 ([#12](https://github.com/qvest-digital/go-mxl/issues/12)) ([4002a05](https://github.com/qvest-digital/go-mxl/commit/4002a05fe3c3a58f86c21d268c13da682d127805))

## 1.0.0-rc.0 (2026-05-16)


### Features

* add cgo preamble, status type and package docs ([8222fb2](https://github.com/qvest-digital/go-mxl/commit/8222fb28928fd358fe1bc9798563c3593cb64719))
* add examples for the public API ([1dc6ec3](https://github.com/qvest-digital/go-mxl/commit/1dc6ec33c36a5bc57e8e8b21c9ab177316f53125))
* add Instance, time helpers, FlowInfo ([64fb1b1](https://github.com/qvest-digital/go-mxl/commit/64fb1b1fa5e43dd624f83214057814a6e74d9b9c))
* add Reader, Grain and SamplesReader ([ade830b](https://github.com/qvest-digital/go-mxl/commit/ade830baaa6e5c2cfd1f80eb2e739712612b72e7))
* add Writer and SyncGroup ([ae8f41c](https://github.com/qvest-digital/go-mxl/commit/ae8f41cb0eaec9daa260a554ef22462154810e3b))
* **docker:** builder + runtime base images ([#7](https://github.com/qvest-digital/go-mxl/issues/7)) ([8dab02b](https://github.com/qvest-digital/go-mxl/commit/8dab02b6d92646fd2d8349008693ec346cb3a873))
* **fabrics:** wrap libmxl-fabrics C API ([#6](https://github.com/qvest-digital/go-mxl/issues/6)) ([0291363](https://github.com/qvest-digital/go-mxl/commit/0291363bedc833e6863fc2f644a3570a9949b4f1))
* tests for mxl, fabrics, and a fabrics example pair ([#17](https://github.com/qvest-digital/go-mxl/issues/17)) ([432b5b1](https://github.com/qvest-digital/go-mxl/commit/432b5b1178317a650536d70ed11fb7810a87739d))


### Code Refactoring

* move package under mxl/ subdir ([0492b38](https://github.com/qvest-digital/go-mxl/commit/0492b38c495b297d1f8d18c6b4659ab426cb1082))
