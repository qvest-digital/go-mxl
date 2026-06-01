# Graph Report - go-mxl  (2026-06-01)

## Corpus Check
- 79 files · ~33,627 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 887 nodes · 1289 edges · 64 communities (58 shown, 6 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 190 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d398895b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

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
- [[_COMMUNITY_Community 16|Community 16]]
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
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 56|Community 56]]
- [[_COMMUNITY_Community 57|Community 57]]
- [[_COMMUNITY_Community 58|Community 58]]

## God Nodes (most connected - your core abstractions)
1. `statusErr()` - 39 edges
2. `newTestInstance()` - 30 edges
3. `fabricsStatusErr()` - 29 edges
4. `newTestWriter()` - 21 edges
5. `CurrentIndex()` - 19 edges
6. `Initiator` - 17 edges
7. `newTestFabrics()` - 16 edges
8. `ErrClosed()` - 16 edges
9. `/graphify` - 16 edges
10. `What You Must Do When Invoked` - 16 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `ParseTargetInfo()`  [INFERRED]
  examples/fabrics-initiator/main.go → /home/daniel/work/projects/go-mxl/.claude/worktrees/claude-md-graphify-54cda722/fabrics/target_info.go
- `main()` --calls--> `CurrentIndex()`  [INFERRED]
  examples/fabrics-initiator/main.go → /home/daniel/work/projects/go-mxl/.claude/worktrees/claude-md-graphify-54cda722/mxl/time.go
- `ParseProvider()` --calls--> `fabricsStatusErr()`  [INFERRED]
  /home/daniel/work/projects/go-mxl/.claude/worktrees/claude-md-graphify-54cda722/fabrics/provider.go → fabrics/status.go
- `fabricsStatusErr()` --calls--> `Status`  [INFERRED]
  fabrics/status.go → mxl/status.go
- `fabricsStatusErr()` --calls--> `StatusErrFromInt32()`  [INFERRED]
  fabrics/status.go → mxl/status.go

## Import Cycles
- None detected.

## Communities (64 total, 6 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.16
Nodes (16): agent, name, agent_type, cwd, effort, level, exceeds_200k_tokens, fast_mode (+8 more)

### Community 1 - "Community 1"
Cohesion: 0.14
Nodes (36): newTestInstance(), TestInstanceCloseIdempotent(), TestInstanceFlowDefMissing(), TestInstanceGarbageCollect(), TestInstanceIsFlowActiveMissing(), TestInstanceMethodsAfterClose(), TestIsTmpFsDevShm(), TestIsTmpFsMissingPath() (+28 more)

### Community 2 - "Community 2"
Cohesion: 0.09
Nodes (19): EndpointAddress, endpointCBuf, ParseTargetInfo(), T, TargetInfo, newTestTargetInfo(), TestParseTargetInfoEmpty(), TestParseTargetInfoInvalid() (+11 more)

### Community 3 - "Community 3"
Cohesion: 0.06
Nodes (34): Initiator, Duration, EndpointAddress, Instance, Instance, Mutex, Provider, TargetInfo (+26 more)

### Community 4 - "Community 4"
Cohesion: 0.10
Nodes (30): freePort(), Instance, T, newDomain(), TestFabricsGrainTransferTCP(), TestFabricsSampleTransferTCP(), newDomain(), TestAudioSamplesRoundTrip() (+22 more)

### Community 5 - "Community 5"
Cohesion: 0.17
Nodes (19): Instance, T, Writer, newTestFabrics(), TestInitiatorAddRemoveTargetNil(), TestInitiatorAddTargetClosedInfo(), TestInitiatorCloseIdempotent(), TestInitiatorMethodsAfterClose() (+11 more)

### Community 6 - "Community 6"
Cohesion: 0.14
Nodes (16): CommonFlowConfig, ContinuousFlowConfig, DataFormat, DiscreteFlowConfig, FlowConfig, FlowInfo, goFlowConfig(), goFlowInfo() (+8 more)

### Community 7 - "Community 7"
Cohesion: 0.12
Nodes (16): 1.0.0-rc.0 (2026-05-16), [1.0.0-rc.1](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.0...v1.0.0-rc.1) (2026-05-16), [1.0.0-rc.2](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.1...v1.0.0-rc.2) (2026-05-16), [1.0.0-rc.3](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.2...v1.0.0-rc.3) (2026-05-16), [1.0.0-rc.4](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.3...v1.0.0-rc.4) (2026-05-17), [1.0.0-rc.5](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.4...v1.0.0-rc.5) (2026-05-17), Build System, Changelog (+8 more)

### Community 8 - "Community 8"
Cohesion: 0.10
Nodes (19): Branches and PRs, Build, code:sh (git fetch origin), code:block2 (feat(docker): support linux/arm64 in published images), code:sh (git worktree remove .claude/worktrees/<topic>-<id>), code:sh (go test ./...), code:sh (go test -tags mxl_integration ./...), Commits (+11 more)

### Community 9 - "Community 9"
Cohesion: 0.12
Nodes (16): Build requirements, code:sh (go get github.com/qvest-digital/go-mxl), code:sh (pkg-config --cflags --libs libmxl), code:sh (# .env), code:go (package main), code:sh (go get github.com/qvest-digital/go-mxl@v1.0.0-rc.1), Container images, Examples (+8 more)

### Community 10 - "Community 10"
Cohesion: 0.06
Nodes (20): Instance, Reader, Reader, Instance, Instance, GrainWriteAccess, durationNs(), Instance (+12 more)

### Community 11 - "Community 11"
Cohesion: 0.17
Nodes (9): main(), Provider, ParseProvider(), TestParseProviderAuto(), TestParseProviderEmpty(), TestParseProviderUnknown(), TestProviderRoundTrip(), TestProviderStringNonEmpty() (+1 more)

### Community 12 - "Community 12"
Cohesion: 0.18
Nodes (13): build, context, dockerfile, target, customizations, vscode, name, runArgs (+5 more)

### Community 13 - "Community 13"
Cohesion: 0.05
Nodes (38): code:block1 (/graphify                                             # full), code:bash (if [ ! -f graphify-out/.graphify_python ]; then), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash (if [ ! -f graphify-out/.graphify_extract.json ]; then), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c ") (+30 more)

### Community 14 - "Community 14"
Cohesion: 0.27
Nodes (11): changelog-sections, include-component-in-tag, include-v-in-tag, initial-version, package-name, packages, prerelease, prerelease-type (+3 more)

### Community 15 - "Community 15"
Cohesion: 0.24
Nodes (10): customizations, vscode, image, name, runArgs, go.buildTags, extensions, settings (+2 more)

### Community 16 - "Community 16"
Cohesion: 0.30
Nodes (10): mxl_go_config_continuous(), mxl_go_config_discrete(), mxl_go_wrapped_count_ro(), mxl_go_wrapped_count_rw(), mxl_go_wrapped_fragment_ptr_ro(), mxl_go_wrapped_fragment_ptr_rw(), mxl_go_wrapped_fragment_size_ro(), mxl_go_wrapped_fragment_size_rw() (+2 more)

### Community 17 - "Community 17"
Cohesion: 0.18
Nodes (9): code:sh (write-grain -domain /dev/shm/mxl -flow-def flow.json), code:sh (write-samples -domain /dev/shm/mxl -flow-def flow.json), code:sh (sync-group -domain /dev/shm/mxl -rate 30000/1001 \), code:sh (# 1. produce into the source domain), examples, Fabric transport, Local grain (video/data) flow, Local samples (audio) flow (+1 more)

### Community 18 - "Community 18"
Cohesion: 0.29
Nodes (9): changelog-sections, extra-files, include-component-in-tag, include-v-in-tag, package-name, packages, prerelease, release-type (+1 more)

### Community 19 - "Community 19"
Cohesion: 0.04
Nodes (49): code:bash ($(cat graphify-out/.graphify_python) -c "), code:block11 ([Agent tool call 1: files 1-15, subagent_type="general-purpo), code:bash (PROJECT_ROOT=$(cat graphify-out/.graphify_root)), code:block13 (You are a graphify extraction subagent. Read the files liste), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c "), code:bash ($(cat graphify-out/.graphify_python) -c ") (+41 more)

### Community 20 - "Community 20"
Cohesion: 0.17
Nodes (10): Building locally, code:dockerfile (FROM ghcr.io/qvest-digital/go-mxl-builder:latest AS builder), code:sh (docker build -f docker/Dockerfile --target builder -t go-mxl), Container images, Dev containers, Provider and hardware coverage, RDMA stack source, Tags (+2 more)

### Community 21 - "Community 21"
Cohesion: 0.05
Nodes (37): directoryMap, docker, docs, examples, fabrics, graphify-out, mxl, fileCount (+29 more)

### Community 22 - "Community 22"
Cohesion: 0.22
Nodes (9): host, name, owner, workspace, added_dirs, current_dir, git_worktree, project_dir (+1 more)

### Community 23 - "Community 23"
Cohesion: 0.29
Nodes (7): resets_at, used_percentage, rate_limits, five_hour, seven_day, resets_at, used_percentage

### Community 24 - "Community 24"
Cohesion: 0.43
Nodes (6): customManagers, extends, minimumReleaseAge, packageRules, prHourlyLimit, $schema

### Community 25 - "Community 25"
Cohesion: 0.48
Nodes (5): error, retry_count, timestamp, tool_input_preview, tool_name

### Community 26 - "Community 26"
Cohesion: 0.07
Nodes (29): build, buildCommand, devCommand, lintCommand, scripts, testCommand, conventions, fileOrganization (+21 more)

### Community 28 - "Community 28"
Cohesion: 0.53
Nodes (4): TestGrainComplete(), TestGrainCopyEmpty(), TestGrainCopyIndependent(), TestGrainInvalid()

### Community 29 - "Community 29"
Cohesion: 0.48
Nodes (5): agents, last_updated, total_completed, total_failed, total_spawned

### Community 30 - "Community 30"
Cohesion: 0.18
Nodes (11): context_window, context_window_size, current_usage, remaining_percentage, total_input_tokens, total_output_tokens, used_percentage, cache_creation_input_tokens (+3 more)

### Community 38 - "Community 38"
Cohesion: 0.33
Nodes (6): cost, total_api_duration_ms, total_cost_usd, total_duration_ms, total_lines_added, total_lines_removed

### Community 39 - "Community 39"
Cohesion: 0.07
Nodes (27): 1.0.0-rc.0 (2026-05-16), [1.0.0-rc.1](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.0...v1.0.0-rc.1) (2026-05-16), [1.0.0-rc.2](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.1...v1.0.0-rc.2) (2026-05-16), [1.0.0-rc.3](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.2...v1.0.0-rc.3) (2026-05-16), [1.0.0-rc.4](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.3...v1.0.0-rc.4) (2026-05-17), [1.0.0-rc.5](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.4...v1.0.0-rc.5) (2026-05-17), [1.0.0-rc.6](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.5...v1.0.0-rc.6) (2026-05-27), [1.0.0-rc.7](https://github.com/qvest-digital/go-mxl/compare/v1.0.0-rc.6...v1.0.0-rc.7) (2026-05-27) (+19 more)

### Community 40 - "Community 40"
Cohesion: 0.19
Nodes (12): Status, StatusErrFromInt32(), T, TestErrClosedDistinctFromStatus(), TestStatusErrAs(), TestStatusErrIs(), TestStatusErrNilOnOK(), TestStatusErrorStrings() (+4 more)

### Community 41 - "Community 41"
Cohesion: 0.67
Nodes (3): model, display_name, id

### Community 43 - "Community 43"
Cohesion: 0.15
Nodes (12): active_modes, background_jobs, active, recent, stats, created_at, todo_summary, completed (+4 more)

### Community 44 - "Community 44"
Cohesion: 0.15
Nodes (12): active_modes, background_jobs, active, recent, stats, created_at, todo_summary, completed (+4 more)

### Community 45 - "Community 45"
Cohesion: 0.17
Nodes (11): Build requirements, Container images, Examples, go-mxl, Graphify, Install, License, Memory safety (+3 more)

### Community 46 - "Community 46"
Cohesion: 0.25
Nodes (7): 20:44 | main, 20:51 | main, 21:09 | main, 21:17 | main, 21:25 | main, 21:32 | main, 21:47 | main

### Community 47 - "Community 47"
Cohesion: 0.29
Nodes (6): agents_completed, agents_spawned, ended_at, modes_used, reason, session_id

### Community 48 - "Community 48"
Cohesion: 0.29
Nodes (6): agents_completed, agents_spawned, ended_at, modes_used, reason, session_id

### Community 49 - "Community 49"
Cohesion: 0.33
Nodes (5): boot_id, cwd, pid, session_id, started_at

### Community 50 - "Community 50"
Cohesion: 0.33
Nodes (5): boot_id, cwd, pid, session_id, started_at

### Community 51 - "Community 51"
Cohesion: 0.33
Nodes (5): boot_id, cwd, pid, session_id, started_at

### Community 52 - "Community 52"
Cohesion: 0.33
Nodes (5): boot_id, cwd, pid, session_id, started_at

### Community 53 - "Community 53"
Cohesion: 0.40
Nodes (4): backgroundTasks, sessionId, sessionStartTimestamp, timestamp

### Community 54 - "Community 54"
Cohesion: 0.40
Nodes (4): backgroundTasks, sessionId, sessionStartTimestamp, timestamp

### Community 55 - "Community 55"
Cohesion: 0.40
Nodes (4): backgroundTasks, sessionId, sessionStartTimestamp, timestamp

### Community 56 - "Community 56"
Cohesion: 0.40
Nodes (4): backgroundTasks, sessionId, sessionStartTimestamp, timestamp

### Community 57 - "Community 57"
Cohesion: 0.67
Nodes (3): T, TestErrNotReadyMatchesInverse(), TestErrNotReadyMatchesMxlStatusErrNotReady()

## Knowledge Gaps
- **318 isolated node(s):** `extensions`, `go.buildTags`, `dockerfile`, `context`, `target` (+313 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `statusErr()` connect `Community 10` to `Community 40`, `Community 2`?**
  _High betweenness centrality (0.056) - this node is a cross-community bridge._
- **Why does `CurrentIndex()` connect `Community 4` to `Community 1`, `Community 2`, `Community 11`?**
  _High betweenness centrality (0.047) - this node is a cross-community bridge._
- **Why does `fabricsStatusErr()` connect `Community 3` to `Community 40`, `Community 2`, `Community 11`?**
  _High betweenness centrality (0.043) - this node is a cross-community bridge._
- **Are the 35 inferred relationships involving `statusErr()` (e.g. with `.Cancel()` and `.Commit()`) actually correct?**
  _`statusErr()` has 35 INFERRED edges - model-reasoned connections that need verification._
- **Are the 23 inferred relationships involving `newTestInstance()` (e.g. with `TestNewReaderMissingFlow()` and `TestReaderCloseIdempotent()`) actually correct?**
  _`newTestInstance()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **Are the 25 inferred relationships involving `fabricsStatusErr()` (e.g. with `.AddTarget()` and `.Close()`) actually correct?**
  _`fabricsStatusErr()` has 25 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newTestWriter()` (e.g. with `TestReaderCloseIdempotent()` and `TestReaderGetMaxReadLengthSamplesAudio()`) actually correct?**
  _`newTestWriter()` has 10 INFERRED edges - model-reasoned connections that need verification._