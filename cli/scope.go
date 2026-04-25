package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallationTarget represents a location where a skill should be linked
type InstallationTarget struct {
	Agent    string // "claude" or "agents"
	Path     string // full path to the skillbase directory
	IsGlobal bool
}

// ScopeResolver determines where skills should be installed
type ScopeResolver interface {
	// Resolve returns all installation targets for a skill
	// skillName: name of the skill (for symlink naming)
	// global: use global agent scope instead of project scope
	// agentFilter: specific agent ("claude", "agents") or empty for all detected
	Resolve(skillName string, global bool, agentFilter string) ([]InstallationTarget, error)

	// DetectAgents returns which agents are available in the current context
	DetectAgents(global bool) []string
}

// FileSystemScopeResolver implements ScopeResolver using filesystem detection
type FileSystemScopeResolver struct {
	home string
	cwd  string
}

// NewFileSystemScopeResolver creates a resolver for the current environment
func NewFileSystemScopeResolver() (*FileSystemScopeResolver, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	return &FileSystemScopeResolver{
		home: os.Getenv("HOME"),
		cwd:  cwd,
	}, nil
}

// DetectAgents returns which agent directories exist
func (f *FileSystemScopeResolver) DetectAgents(global bool) []string {
	if global {
		return f.detectGlobalAgents()
	}
	return f.detectProjectAgents()
}

func (f *FileSystemScopeResolver) detectProjectAgents() []string {
	var agents []string
	if f.dirExists(filepath.Join(f.cwd, ".claude")) {
		agents = append(agents, "claude")
	}
	if f.dirExists(filepath.Join(f.cwd, ".agents")) {
		agents = append(agents, "agents")
	}
	return agents
}

func (f *FileSystemScopeResolver) detectGlobalAgents() []string {
	var agents []string
	if f.dirExists(filepath.Join(f.home, ".claude")) {
		agents = append(agents, "claude")
	}
	if f.dirExists(filepath.Join(f.home, ".agents")) {
		agents = append(agents, "agents")
	}
	return agents
}

func (f *FileSystemScopeResolver) dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Resolve returns installation targets based on scope and filter
func (f *FileSystemScopeResolver) Resolve(skillName string, global bool, agentFilter string) ([]InstallationTarget, error) {
	agents := f.selectAgents(global, agentFilter)

	if len(agents) == 0 {
		scope := "project"
		if global {
			scope = "global"
		}
		return nil, fmt.Errorf("no %s agent scopes found", scope)
	}

	var targets []InstallationTarget
	for _, agent := range agents {
		path := f.getAgentPath(agent, global)
		targets = append(targets, InstallationTarget{
			Agent:    agent,
			Path:     path,
			IsGlobal: global,
		})
	}

	return targets, nil
}

func (f *FileSystemScopeResolver) selectAgents(global bool, agentFilter string) []string {
	// If specific agent requested, use it
	if agentFilter != "" {
		return []string{agentFilter}
	}

	// Otherwise auto-detect
	return f.DetectAgents(global)
}

func (f *FileSystemScopeResolver) getAgentPath(agent string, global bool) string {
	if global {
		return filepath.Join(f.home, "."+agent, "skills")
	}
	return filepath.Join(f.cwd, "."+agent, "skills")
}
