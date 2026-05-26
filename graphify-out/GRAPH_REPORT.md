# Graph Report - go-mxl  (2026-05-26)

## Corpus Check
- 66 files · ~29,886 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 583 nodes · 779 edges · 43 communities (38 shown, 5 thin omitted)
- Extraction: 76% EXTRACTED · 24% INFERRED · 0% AMBIGUOUS · INFERRED: 188 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]

## God Nodes (most connected - your core abstractions)
1. `statusErr()` - 37 edges
2. `newTestInstance()` - 29 edges
3. `fabricsStatusErr()` - 25 edges
4. `newTestWriter()` - 20 edges
5. `CurrentIndex()` - 17 edges
6. `What You Must Do When Invoked` - 16 edges
7. `/graphify` - 15 edges
8. `ErrClosed()` - 12 edges
9. `newTestReader()` - 11 edges
10. `ParseProvider()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `RegionsForFlowReader()`  [INFERRED]
  examples/fabrics-initiator/main.go → fabrics/regions.go
- `main()` --calls--> `ParseTargetInfo()`  [INFERRED]
  examples/fabrics-initiator/main.go → fabrics/target_info.go
- `main()` --calls--> `CurrentIndex()`  [INFERRED]
  examples/fabrics-initiator/main.go → mxl/time.go
- `main()` --calls--> `RegionsForFlowWriter()`  [INFERRED]
  examples/fabrics-target/main.go → fabrics/regions.go
- `main()` --calls--> `CurrentIndex()`  [INFERRED]
  examples/read-grain/main.go → mxl/time.go

## Communities (43 total, 5 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.14
Nodes (13): cwd, effort, level, exceeds_200k_tokens, fast_mode, output_style, name, session_id (+5 more)

### Community 1 - "Community 1"
Cohesion: 0.14
Nodes (32): newTestInstance(), TestInstanceCloseIdempotent(), TestInstanceFlowDefMissing(), TestInstanceGarbageCollect(), TestInstanceIsFlowActiveMissing(), TestInstanceMethodsAfterClose(), newTestReader(), TestNewReaderMissingFlow() (+24 more)

### Community 2 - "Community 2"
Cohesion: 0.08
Nodes (19): EndpointAddress, endpointCBuf, ParseTargetInfo(), newTestTargetInfo(), TestParseTargetInfoEmpty(), TestParseTargetInfoInvalid(), TestTargetInfoCloseIdempotent(), TestTargetInfoMarshalAfterClose() (+11 more)

### Community 3 - "Community 3"
Cohesion: 0.11
Nodes (13): Initiator, Instance, InitiatorConfig, Instance, NewInstance(), fabricsStatusErr(), Target, Instance (+5 more)

### Community 4 - "Community 4"
Cohesion: 0.11
Nodes (22): newDomain(), TestAudioSamplesRoundTrip(), TestSyncGroupGrain(), TestVideoGrainRoundTrip(), TestWriterCancelGrainNotPublished(), CurrentIndex(), IndexToTimestamp(), Now() (+14 more)

### Community 5 - "Community 5"
Cohesion: 0.12
Nodes (19): newTestFabrics(), TestInitiatorAddRemoveTargetNil(), TestInitiatorAddTargetClosedInfo(), TestInitiatorCloseIdempotent(), TestInitiatorMethodsAfterClose(), TestInitiatorSetupClosedRegions(), TestInitiatorSetupNilRegions(), newTestMxlInstance() (+11 more)

### Community 6 - "Community 6"
Cohesion: 0.10
Nodes (15): CommonFlowConfig, ContinuousFlowConfig, DataFormat, DiscreteFlowConfig, FlowConfig, FlowInfo, goFlowConfig(), goFlowInfo() (+7 more)

### Community 7 - "Community 7"
Cohesion: 0.12
Nodes (16): 1.0.0-rc.0 (2026-05-16), [1.0.0-rc.1](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.0...v1.0.0-rc.1) (2026-05-16), [1.0.0-rc.2](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.1...v1.0.0-rc.2) (2026-05-16), [1.0.0-rc.3](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.2...v1.0.0-rc.3) (2026-05-16), [1.0.0-rc.4](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.3...v1.0.0-rc.4) (2026-05-17), [1.0.0-rc.5](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.4...v1.0.0-rc.5) (2026-05-17), Build System, Changelog (+8 more)

### Community 8 - "Community 8"
Cohesion: 0.11
Nodes (19): Branches and PRs, Build, code:sh (git fetch origin), code:block2 (feat(docker): support linux/arm64 in published images), code:sh (git worktree remove .claude/worktrees/<topic>-<id>), code:sh (go test ./...), code:sh (go test -tags mxl_integration ./...), Commits (+11 more)

### Community 9 - "Community 9"
Cohesion: 0.12
Nodes (16): Build requirements, code:sh (go get github.com/qvest-digital/go-mxl), code:sh (pkg-config --cflags --libs libmxl), code:sh (# .env), code:go (package main), code:sh (go get github.com/qvest-digital/go-mxl@v1.0.0-rc.1), Container images, Examples (+8 more)

### Community 10 - "Community 10"
Cohesion: 0.05
Nodes (17): GrainWriteAccess, durationNs(), Instance, Reader, makeGrain(), Reader, makeSamplesView(), SamplesView (+9 more)

### Community 11 - "Community 11"
Cohesion: 0.19
Nodes (8): main(), Provider, ParseProvider(), TestParseProviderAuto(), TestParseProviderEmpty(), TestParseProviderUnknown(), TestProviderRoundTrip(), main()

### Community 12 - "Community 12"
Cohesion: 0.14
Nodes (13): build, context, dockerfile, target, customizations, vscode, name, runArgs (+5 more)

### Community 13 - "Community 13"
Cohesion: 0.05
Nodes (38): code:block1 (/graphify                                             # full), code:bash (if [ ! -f graphify-out/.graphify_python ]; then), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash (if [ ! -f graphify-out/.graphify_extract.json ]; then), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c ") (+30 more)

### Community 14 - "Community 14"
Cohesion: 0.17
Nodes (11): changelog-sections, include-component-in-tag, include-v-in-tag, initial-version, package-name, packages, prerelease, prerelease-type (+3 more)

### Community 15 - "Community 15"
Cohesion: 0.18
Nodes (10): customizations, vscode, image, name, runArgs, go.buildTags, extensions, settings (+2 more)

### Community 17 - "Community 17"
Cohesion: 0.20
Nodes (9): code:sh (write-grain -domain /dev/shm/mxl -flow-def flow.json), code:sh (write-samples -domain /dev/shm/mxl -flow-def flow.json), code:sh (sync-group -domain /dev/shm/mxl -rate 30000/1001 \), code:sh (# 1. produce into the source domain), examples, Fabric transport, Local grain (video/data) flow, Local samples (audio) flow (+1 more)

### Community 18 - "Community 18"
Cohesion: 0.20
Nodes (9): changelog-sections, extra-files, include-component-in-tag, include-v-in-tag, package-name, packages, prerelease, release-type (+1 more)

### Community 19 - "Community 19"
Cohesion: 0.06
Nodes (36): code:bash (mkdir -p graphify-out), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash (LOCAL_PATH=$(graphify clone <github-url> [--branch <branch>]), code:bash (graphify export obsidian), code:bash (graphify export html  # auto-aggregates to community view if), code:bash (graphify export wiki), code:bash (graphify export neo4j), code:bash (graphify export neo4j --push bolt://localhost:7687 --user ne) (+28 more)

### Community 20 - "Community 20"
Cohesion: 0.22
Nodes (8): Building locally, code:dockerfile (FROM ghcr.io/qvest-digital/go-mxl-builder:latest AS builder), code:sh (docker build -f docker/Dockerfile --target builder -t go-mxl), Container images, Dev containers, Tags, Usage, Why trixie-slim and not distroless

### Community 21 - "Community 21"
Cohesion: 0.15
Nodes (13): code:bash ($(cat graphify-out/.graphify_python) -c "), code:block11 ([Agent tool call 1: files 1-15, subagent_type="general-purpo), code:bash (PROJECT_ROOT=$(cat graphify-out/.graphify_root)), code:block13 (You are a graphify extraction subagent. Read the files liste), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c ") (+5 more)

### Community 22 - "Community 22"
Cohesion: 0.22
Nodes (9): host, name, owner, workspace, added_dirs, current_dir, git_worktree, project_dir (+1 more)

### Community 23 - "Community 23"
Cohesion: 0.29
Nodes (7): resets_at, used_percentage, rate_limits, five_hour, seven_day, resets_at, used_percentage

### Community 24 - "Community 24"
Cohesion: 0.29
Nodes (6): customManagers, extends, minimumReleaseAge, packageRules, prHourlyLimit, $schema

### Community 25 - "Community 25"
Cohesion: 0.33
Nodes (5): error, retry_count, timestamp, tool_input_preview, tool_name

### Community 26 - "Community 26"
Cohesion: 0.40
Nodes (4): backgroundTasks, sessionId, sessionStartTimestamp, timestamp

### Community 29 - "Community 29"
Cohesion: 0.33
Nodes (5): agents, last_updated, total_completed, total_failed, total_spawned

### Community 30 - "Community 30"
Cohesion: 0.33
Nodes (6): context_window, context_window_size, remaining_percentage, total_input_tokens, total_output_tokens, used_percentage

### Community 38 - "Community 38"
Cohesion: 0.33
Nodes (6): cost, total_api_duration_ms, total_cost_usd, total_duration_ms, total_lines_added, total_lines_removed

### Community 39 - "Community 39"
Cohesion: 0.40
Nodes (5): current_usage, cache_creation_input_tokens, cache_read_input_tokens, input_tokens, output_tokens

### Community 41 - "Community 41"
Cohesion: 0.67
Nodes (3): model, display_name, id

## Knowledge Gaps
- **198 isolated node(s):** `$schema`, `extends`, `minimumReleaseAge`, `prHourlyLimit`, `customManagers` (+193 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `statusErr()` connect `Community 10` to `Community 2`, `Community 6`?**
  _High betweenness centrality (0.086) - this node is a cross-community bridge._
- **Why does `fabricsStatusErr()` connect `Community 3` to `Community 2`, `Community 10`, `Community 11`, `Community 5`?**
  _High betweenness centrality (0.067) - this node is a cross-community bridge._
- **Why does `CurrentIndex()` connect `Community 4` to `Community 1`, `Community 2`, `Community 11`, `Community 5`?**
  _High betweenness centrality (0.060) - this node is a cross-community bridge._
- **Are the 35 inferred relationships involving `statusErr()` (e.g. with `LibVersion()` and `IsTmpFs()`) actually correct?**
  _`statusErr()` has 35 INFERRED edges - model-reasoned connections that need verification._
- **Are the 23 inferred relationships involving `newTestInstance()` (e.g. with `TestNewReaderMissingFlow()` and `TestReaderCloseIdempotent()`) actually correct?**
  _`newTestInstance()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **Are the 24 inferred relationships involving `fabricsStatusErr()` (e.g. with `.NewInitiator()` and `.Setup()`) actually correct?**
  _`fabricsStatusErr()` has 24 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newTestWriter()` (e.g. with `TestReaderCloseIdempotent()` and `TestReaderHandleNilAfterClose()`) actually correct?**
  _`newTestWriter()` has 10 INFERRED edges - model-reasoned connections that need verification._