# Agent Context

## Project

- **Name:** skillbase
- **Language:** Go
- **Module:** `github.com/andrew-a-hale/skillbase`
- **Go version:** 1.26.2

## Architecture

A small CLI / TUI with four packages:

- `main.go` — Entrypoint. Delegates to `cli.Dispatch()`.
- `cli/` — All command logic.
  - `commands.go` — Command dispatcher and shared business logic helpers.
  - `cmd_*.go` — Per-command handlers (`help`, `list`, `find`, `get`, `remove`, `update`).
  - `skillstore.go` — `SkillStore` interface and `FileSystemSkillStore`. Manages installation, linking, removal, and provenance tracking.
  - `repository.go` — `Repository` interface and `GitRepository` implementation. Clones Git repos, discovers skills by `SKILL.md`, parses frontmatter descriptions.
  - `scope.go` — `ScopeResolver` interface and `FileSystemScopeResolver`. Detects `.claude/` and `.agents/` directories, resolves symlink targets.
- `tui/` — Interactive Terminal UI using Charm Bubble Tea.
  - `list.go`, `find.go`, `get.go`, `remove.go`, `update.go` — Bubble Tea models for each command.
  - `component.go` — Reusable cursor list with mouse-wheel support.
  - `style.go` — Lipgloss styles and colour palette.
  - `keys.go` — Keymap definitions (vim-style + arrow keys).
- `internal/fsutil/` — Simple filesystem helpers (`CopyDir`, `CopyFile`).
- `internal/skill/` — Frontmatter parser for `SKILL.md` metadata.

## Key Behaviors

- `SKILLBASE_DEFAULT_REPO` environment variable is **required**. There is no hardcoded fallback.
- `get` copies skills to `~/.skillbase/<name>/` and symlinks into agent scope directories.
- `remove -g` deletes from `~/.skillbase/` and removes all symlinks.
- `remove` (without `-g`) only removes symlinks from the project scope.
- `update` re-runs `get` for an installed skill.
- Commands without required arguments launch an **interactive TUI** using Bubble Tea (`tea.WithAltScreen()`, `tea.WithMouseCellMotion()`).
  - `list`, `find` — always interactive.
  - `get`, `remove`, `update` — interactive when the skill name (or required flags) is omitted.

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
