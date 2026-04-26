package skill

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name:     "frontmatter description",
			content:  "---\ndescription: hello world\n---\n",
			expected: map[string]string{"description": "hello world"},
		},
		{
			name:     "multiple fields",
			content:  "---\nname: foo\ndescription: bar\n---\n",
			expected: map[string]string{"name": "foo", "description": "bar"},
		},
		{
			name:     "no description",
			content:  "---\nname: foo\n---\n",
			expected: map[string]string{"name": "foo"},
		},
		{
			name:     "no frontmatter",
			content:  "# Hello\n",
			expected: map[string]string{},
		},
		{
			name:     "case insensitive keys",
			content:  "---\nDESCRIPTION: UPPER\n---\n",
			expected: map[string]string{"description": "UPPER"},
		},
		{
			name:     "empty",
			content:  "",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFrontmatter(strings.NewReader(tt.content))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d keys, want %d", len(got), len(tt.expected))
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Fatalf("key %q: got %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
