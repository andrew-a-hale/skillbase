# Agent Context

## Project

- **Name:** skillbase
- **Language:** Go
- **Module:** `github.com/andrew-a-hale/skillbase`
- **Go version:** 1.26.2

## Architecture

A small CLI with three packages:

- `main.go` — Entrypoint. Delegates to `cli.Dispatch()`.
- `cli/` — All command logic.
  - `commands.go` — Command dispatcher and business logic (`list`, `find`, `get`, `remove`, `update`).
  - `repository.go` — `Repository` interface and `GitRepository` implementation. Clones Git repos, discovers skills by `SKILL.md`, parses frontmatter descriptions.
  - `scope.go` — `ScopeResolver` interface and `FileSystemScopeResolver`. Detects `.claude/` and `.agents/` directories, resolves symlink targets.
- `internal/fsutil/` — Simple filesystem helpers (`CopyDir`, `CopyFile`).

## Key Behaviors

- `SKILLS_DEFAULT_REPO` environment variable is **required**. There is no hardcoded fallback.
- `get` copies skills to `~/.skills/<name>/` and symlinks into agent scope directories.
- `remove -g` deletes from `~/.skills/` and removes all symlinks.
- `remove` (without `-g`) only removes symlinks from the project scope.
- `update` re-runs `get` for an installed skill.

## Testing

- Table-driven unit tests in `*_test.go` files.
- `fakeExecutor` mocks Git commands.
- Tests use `t.TempDir()` for filesystem isolation.

## Build & Release

- `Makefile` targets: `build`, `test`, `coverage`, `lint`, `clean`, `install`, `release`, `release-snapshot`.
- GitHub Actions CI runs tests on push/PR.
- GoReleaser builds cross-platform binaries on tag push.

## Conventions

- Prefer minimal, explicit error handling.
- Keep packages small and focused.
- Use interfaces for external dependencies (Git executor, scope resolver) to enable testing.
