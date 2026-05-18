# Contributor notes (Claude and humans)

Rules for working in this repository. Read them before opening a PR or
running an automated assistant against this tree.

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
- Use a separate `git worktree` per branch when working alongside
  other editors on the repository. Worktrees keep the main checkout
  clean while sharing the object database, so parallel sessions do
  not collide over staged changes or the working tree.
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
branch tracking `origin/main`:

```sh
git fetch origin
git worktree add ../go-mxl.<topic> -b <topic> origin/main
cd ../go-mxl.<topic>
```

When the PR has merged, drop the worktree, the local branch, and
the now-stale remote tracking ref:

```sh
git worktree remove ../go-mxl.<topic>
git branch -D <topic>
git fetch --prune origin
```

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

## When in doubt

Ask the maintainer before changing the public Go API, the module path,
or the release/tagging strategy.
