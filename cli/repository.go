package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CleanupFunc removes temporary resources
type CleanupFunc func() error

// Skill represents a discovered skill in a repository
type Skill struct {
	Name        string
	Path        string // relative path in repo
	Description string
}

// Repository defines operations for accessing skill repositories
type Repository interface {
	// Clone downloads the repository to a temporary location.
	// Returns the local path and a cleanup function.
	Clone(ctx context.Context) (string, CleanupFunc, error)

	// ListSkills finds all skills in a cloned repository.
	ListSkills(clonePath string) ([]Skill, error)

	// GetSkill retrieves a specific skill by path from a cloned repository.
	GetSkill(clonePath, skillPath string) (Skill, error)
}

// Executor runs external commands.
type Executor interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type execExecutor struct{}

func (e *execExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// Repository errors
var (
	ErrCleanupFailed  = fmt.Errorf("failed to cleanup")
	ErrCloneFailed    = fmt.Errorf("failed to clone repository")
	ErrSkillNotFound  = fmt.Errorf("skill not found")
	ErrInvalidRepoURL = fmt.Errorf("invalid repository URL")
)

// GitRepository implements Repository for Git-backed skill repositories
type GitRepository struct {
	url      string
	executor Executor
}

// NewGitRepository creates a Git repository adapter
func NewGitRepository(url string) (*GitRepository, error) {
	if url == "" {
		return nil, ErrInvalidRepoURL
	}
	return &GitRepository{url: url, executor: &execExecutor{}}, nil
}

// Clone downloads the Git repository to a temporary directory
func (g *GitRepository) Clone(ctx context.Context) (string, CleanupFunc, error) {
	tempDir, err := os.MkdirTemp("", "skillbase-clone-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cleanup := func() error {
		return os.RemoveAll(tempDir)
	}

	output, err := g.executor.Run(ctx, "git", "clone", "--depth", "1", g.url, tempDir)
	if err != nil {
		if err := cleanup(); err != nil {
			return "", nil, fmt.Errorf("%w: %v", ErrCleanupFailed, err)
		}
		return "", nil, fmt.Errorf("%w: %v (output: %s)", ErrCloneFailed, err, string(output))
	}

	return tempDir, cleanup, nil
}

// ListSkills discovers all skills (directories containing SKILL.md) in cloned repo
func (g *GitRepository) ListSkills(clonePath string) ([]Skill, error) {
	var skills []Skill
	seenNames := make(map[string]bool)

	err := filepath.Walk(clonePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() != "SKILL.md" {
			return nil
		}

		dir := filepath.Dir(path)
		name := filepath.Base(dir)

		// Check for name collision
		if seenNames[name] {
			return fmt.Errorf("skill name collision: %q found at multiple paths", name)
		}
		seenNames[name] = true

		relPath, _ := filepath.Rel(clonePath, dir)
		desc := extractSkillDescription(path)

		skills = append(skills, Skill{
			Name:        name,
			Path:        relPath,
			Description: desc,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return skills, nil
}

// GetSkill retrieves a specific skill from the cloned repository
func (g *GitRepository) GetSkill(clonePath, skillPath string) (Skill, error) {
	skillDir := filepath.Join(clonePath, skillPath)
	skillFile := filepath.Join(skillDir, "SKILL.md")

	if _, err := os.Stat(skillFile); err != nil {
		if os.IsNotExist(err) {
			return Skill{}, fmt.Errorf("%w: %q", ErrSkillNotFound, skillPath)
		}
		return Skill{}, err
	}

	name := filepath.Base(skillDir)
	desc := extractSkillDescription(skillFile)

	return Skill{
		Name:        name,
		Path:        skillPath,
		Description: desc,
	}, nil
}

// extractSkillDescription parses SKILL.md frontmatter for description field
func extractSkillDescription(skillFile string) string {
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return ""
	}

	// Simple regex for description: field in frontmatter
	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "description:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}
