# Contributor notes (Claude and humans)

Rules for working in this repository. Read them before opening a PR or
running an automated assistant against this tree.

## STOP. Use a git worktree.

Before ANY mutation of this repository -- edit, write, commit,
branch create, push, rebase, `gh pr create` -- the very first
action of the session is to set up a dedicated `git worktree`.
This rule has no exceptions. There is no change small enough to
skip it: typos, single-line fixes, doc-only PRs, even edits to
this file itself all require the same worktree dance.

The repository is worked on by multiple parallel sessions and
editors at once. Two writers in the same tree corrupt staging
state, step on each other's branches, and lose work without
warning -- the rule exists to make that physically impossible,
not to be polite about it.

Worktrees ALWAYS live under `<repo>/.claude/worktrees/`. Not
next to the repo, not in `/tmp`, not anywhere else -- that path
is the project's established convention and the location the
harness manages.

The preferred path is to delegate the mutation to a sub-agent
with `isolation: "worktree"` on the Agent call. The harness
then creates `<repo>/.claude/worktrees/agent-<id>/` automatically
with a unique id, so two concurrent sub-agents never share a
path. Do not assume a worktree from earlier in the session is
still mounted.

For a manual worktree (no sub-agent), pick a short random tag
so two sessions on the same topic never collide, and place the
worktree under `.claude/worktrees/`:

```sh
git fetch origin
id=$(openssl rand -hex 4)
git worktree add .claude/worktrees/<topic>-$id -b <topic> origin/main
cd .claude/worktrees/<topic>-$id
```

All subsequent edits, commits, and pushes happen from the
worktree.

The only thing allowed in the main checkout is read-only
inspection: `git log`, `git diff`, `git status`, `gh pr view`,
reading files, grep / find. Anything that touches the index, the
working tree, the branch list, or the remote is out.

## Documentation

- Keep `README.md` and code comments tight. State facts; don't speculate.
- Do not write about how the binding "should be used" or "is ideal for"
  any particular system. The binding wraps libmxl; downstream choice is
  not ours to characterize.
- Do not reference other projects (mediamtx, gstreamer, ffmpeg, …) as
  motivating consumers in docs, comments, or commit messages. They are
  not part of the public contract.
- Don't invent API behavior. If you can't verify it by reading libmxl's
  headers or testing against a real install, leave it out.
- Don't add SPDX headers or copyright lines to new files unless you're
  preserving existing ones from an external source.

## Branches and PRs

- Direct commits to `main` are off by default. Every change opens a
  feature branch and a PR against `main`. Commit directly to `main`
  only when the maintainer has explicitly approved it for that
  specific change.
- Force-pushes are off by default. Force-pushing to `main` is
  prohibited. Force-pushing to a feature branch is only permitted
  with explicit approval, because another editor may be reviewing
  the branch or checked out against it.
- Merge PRs with **Squash and merge**. release-please derives version
  bumps and changelog entries from the resulting single commit on
  `main`, and a noisy merge of dozens of intermediate commits would
  bury the release-relevant ones.
- Delete the feature branch on the remote as soon as the PR is
  merged (GitHub's "Delete branch" button on the merged PR, or the
  repo-level "Automatically delete head branches" setting). Stale
  remote branches confuse the next contributor's `git fetch` and
  inflate `git branch -r` output.

### Squash commit format for release-please

GitHub's squash-merge uses the PR title as the resulting commit
subject (with the PR number appended) and the PR body as the commit
body. release-please parses that commit on `main` to decide what
gets a changelog entry and how the version bumps. Two consequences:

1. **PR title is Conventional Commits.** Write the PR title in
   `<type>(<scope>): <subject>` form just as if it were a single
   commit subject. Subject `<= 72` chars, imperative mood.
2. **Multiple release-relevant changes go in the PR body, at the
   bottom, one per line.** release-please reads additional
   conventional-commit footer lines and emits one changelog entry
   per line. Add them after the prose explanation, separated by a
   blank line. Example:

   ```
   feat(docker): support linux/arm64 in published images

   CI matrix now builds both linux/amd64 and linux/arm64 per tag, so
   Apple Silicon devs get a native pull under Docker Desktop with no
   QEMU tax.

   fix(docker): drop TARGETARCH default so buildx auto-fills it
   ```

   That single squash commit produces two release-relevant entries:
   the `feat(docker)` (driving a minor bump) and the `fix(docker)`
   (a patch entry under the same release). Use `BREAKING CHANGE:`
   or `BREAKING-CHANGE:` (release-please accepts both) and
   `Release-As: X.Y.Z` for explicit overrides.

### Working in a worktree

From the main checkout, create a worktree pinned to a feature
branch tracking `origin/main`. The worktree lives under
`.claude/worktrees/`, the project's canonical worktree location,
and gets a short random tag so two sessions on the same topic
never share a path:

```sh
git fetch origin
id=$(openssl rand -hex 4)
git worktree add .claude/worktrees/<topic>-$id -b <topic> origin/main
cd .claude/worktrees/<topic>-$id
```

When the PR has merged, drop the worktree, the local branch, and
the now-stale remote tracking ref:

```sh
git worktree remove .claude/worktrees/<topic>-<id>
git branch -D <topic>
git fetch --prune origin
```

`git worktree remove` is rarely needed when sub-agents manage the
worktree: the harness cleans up its own `agent-<id>` directories.
The teardown block above is for the manual `git worktree add`
case.

## Commits

- Use Conventional Commits: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`,
  `test:`, `ci:`, `build:`, `perf:`, `style:`. Breaking changes get `!`
  (`feat!:`) or a `BREAKING CHANGE:` footer.
- Prefer small, focused commits. The release tooling derives version
  bumps and the changelog from commit subjects.
- Subject line ≤ 72 chars, imperative mood ("add", "fix", not "added",
  "fixes"). Body wraps at 72.

### Message content

A commit message documents why a change exists, in terms that stay
useful when read alone, years later, by someone with no memory of
the work that produced it. The same rules apply to PR descriptions.

- Explain *why*. The diff shows *what*; don't restate it.
- Stay scoped to this repository and this change. No speculation
  about upstream, downstream, future work, or follow-ups. If
  something was deliberately left out of the diff, name it and the
  reason — only when that omission matters for understanding the
  present change.
- Reference another repository or project only when its state is
  the direct reason for the change (a dependency bump, a vendored
  fix, an API contract pinned to a published version). Context for
  reviewers, gratitude, or cross-linking belongs in the PR thread
  or an issue, not the commit.
- Write declarative facts. No personal pronouns ("I", "we", "you").
  Don't address a reader: no "note that…", "as you can see…", "we
  decided to…", "this should help…".
- Don't narrate. No history of what was tried first, what failed,
  or what alternatives were considered.
- No filler verbs without specifics. "Clean up", "improve",
  "refactor" alone tell nothing; either name the actual change or
  drop the line.
- No checklists, "Summary" / "Test plan" sections, marketing
  phrasing, or emojis. Those belong in the PR description if
  anywhere.
- No tool-authored `Co-Authored-By:` trailers — the message
  describes the change, not the process that produced it.
- Cross-reference an issue or PR only when its content is itself
  the reason for the change (`closes #N` where the issue is the
  why). Vague "see #N for context" pointers do not belong here.

## Versioning and tags

- The module path is `github.com/qvest-digital/go-mxl`. While the major
  version is 0 or 1, tags are `vMAJOR.MINOR.PATCH` at the repo root.
- For a future v2+, the module path must include the `/v2` suffix and
  tags become `v2.0.0` etc. — see <https://go.dev/ref/mod#major-version-suffixes>.
- Releases are automated by `release-please` (see `.github/workflows/`).
  Don't hand-tag or hand-edit `CHANGELOG.md` — let the workflow do it.

## Build

- The package is cgo. `libmxl` must be installed with headers and a
  pkg-config file before `go build` works. See `README.md`.
- Tests that exercise a writer↔reader round-trip live under build tag
  `mxl_integration` and need a tmpfs-mounted `/dev/shm`. CI runs them
  in the builder image alongside the unit tests; the build tag keeps
  them out of a plain `go test ./...` for callers without that mount.

## Graphify

A [Graphify](https://github.com/safishamsi/graphify) knowledge graph
of this repo lives at `graphify-out/`. The graph is committed, so a
fresh clone already has it; `.graphifyignore` controls what gets
indexed. See the "Graphify" section in `README.md` for the install
and rebuild recipe.

When the `graphify` Skill is available in the current Claude session,
use it before reaching for `grep` / `find` / wide `Read` sweeps:
query the committed graph to locate symbols, callers, and cross-file
references, then read only the files the graph points at. The graph
makes drilling down from a name to its definition, usages, and
neighbouring types a single query instead of a search-and-prune
pass. Fall back to `grep` / `find` when the Skill is not loaded or
when the working tree has drifted past the last committed graph
(check `graphify-out/GRAPH_REPORT.md` for the indexed snapshot).

## When in doubt

Ask the maintainer before changing the public Go API, the module path,
or the release/tagging strategy.
