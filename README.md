# skillbase

A lightweight CLI for managing AI agent skills across projects.

Skillbase fetches skill definitions from Git repositories and links them into agent scopes (e.g., `.claude/skills/`, `.agents/skills/`), letting you share and version-control your agent capabilities separately from any single project.

---

## Installation

### From source

```bash
go install github.com/andrew-a-hale/skillbase@latest
```

### Pre-built binaries

Download from the [releases page](https://github.com/andrew-a-hale/skillbase/releases).

---

## Quick start

Set your default skills repository:

```bash
export SKILLBASE_DEFAULT_REPO="https://github.com/your-org/skills"
```

> **Required:** `SKILLBASE_DEFAULT_REPO` must be set for commands that interact with a repository (`find`, `get`, `update`).

List available skills:

```bash
skillbase find
```

Install a skill to the current project:

```bash
skillbase get my-skill
```

Install globally:

```bash
skillbase get -g my-skill
```

List installed skills:

```bash
skillbase ls          # project scope
skillbase ls -g       # global scope
```

Remove a skill:

```bash
skillbase rm my-skill
```

Update a skill:

```bash
skillbase update my-skill
```

---

## Usage

```
skillbase - Manage agent skills

Commands:
  help              Print this help message
  list, ls          List installed skills
    -g              List global skills
  find [filter]     Find available skills in repository
  get [skill|url]   Download skill(s) and link to agent scope
    -p path         Skill path within repository
    -a agent        Target agent (claude|agents)
    -g              Install to global agent scope
  rm, remove <name> Remove skill
    -a agent        Remove from specific agent only
    -g              Remove from global storage
  update <name>     Update existing skill
```

### Install from a specific repository

```bash
skillbase get https://github.com/user/repo/skill-path
```

Or use `-p` to specify the skill path separately:

```bash
skillbase get -p skill-path https://github.com/user/repo
```

### Target a specific agent

```bash
skillbase get -a claude my-skill
skillbase rm -a agents my-skill
```

---

## How it works

1. **Skills repository** — A Git repo containing directories with `SKILL.md` files.
2. **Install** — `skillbase` clones the repo, copies the skill to `~/.skillbase/<name>/`, and creates symlinks in the detected agent scope directories.
3. **Detect agents** — Automatically finds `.claude/` or `.agents/` directories in the current project (or home directory for `-g`).

### Skill repository layout

```
skills-repo/
  skill-one/
    SKILL.md
  skill-two/
    SKILL.md
    helpers/
      script.sh
```

`SKILL.md` frontmatter is parsed for the `description` field, which appears in `skillbase find`.

---

## Development

```bash
# Build
make build

# Test
make test

# Coverage report
make coverage

# Lint
make lint

# Release snapshot (local)
make release-snapshot
```

---

## License

MIT — see [LICENSE](LICENSE).
