// Package skill provides helpers for working with skill metadata.
package skill

import (
	"bufio"
	"io"
	"strings"
)

// ParseFrontmatter reads YAML-style frontmatter from r and returns the
// key-value pairs. Frontmatter is delimited by lines containing only "---".
// If no frontmatter delimiters are found, an empty map is returned.
// Keys are normalised to lower case.
func ParseFrontmatter(r io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(r)
	inFrontmatter := false
	foundStart := false
	result := make(map[string]string)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "---" {
			if !foundStart {
				foundStart = true
				inFrontmatter = true
				continue
			}
			// End delimiter
			break
		}
		if inFrontmatter {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.ToLower(strings.TrimSpace(parts[0]))
				val := strings.TrimSpace(parts[1])
				result[key] = val
			}
		}
	}

	return result, scanner.Err()
}
