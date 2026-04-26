// Package skill provides helpers for working with skill metadata.
package skill

import (
	"bufio"
	"io"
	"strings"

	"go.yaml.in/yaml/v4"
)

// ParseFrontmatter reads YAML-style frontmatter from r and returns the
// key-value pairs. Frontmatter is delimited by lines containing only "---".
// If no frontmatter delimiters are found, an empty map is returned.
// Keys are normalised to lower case.
func ParseFrontmatter(r io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(r)
	foundStart := false
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !foundStart {
				foundStart = true
				continue
			}
			// End delimiter
			break
		}
		if foundStart {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !foundStart || len(lines) == 0 {
		return map[string]string{}, nil
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &raw); err != nil {
		return map[string]string{}, nil
	}

	result := make(map[string]string, len(raw))
	for k, v := range raw {
		key := strings.ToLower(k)
		if s, ok := v.(string); ok {
			result[key] = s
		}
	}
	return result, nil
}
