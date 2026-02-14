package main

import (
	"testing"
)

func TestParseCoAuthors(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		seenAuthors map[string]bool
		want        []string
	}{
		{
			name:        "basic co-author",
			body:        "Some commit message\n\nCo-Authored-By: Alice <alice@example.com>\n",
			seenAuthors: make(map[string]bool),
			want:        []string{"Alice <alice@example.com>"},
		},
		{
			name:        "case insensitive prefix",
			body:        "msg\n\nco-authored-by: Bob <bob@example.com>\n",
			seenAuthors: make(map[string]bool),
			want:        []string{"Bob <bob@example.com>"},
		},
		{
			name:        "mixed case prefix",
			body:        "msg\n\nCO-AUTHORED-BY: Carol <carol@example.com>\n",
			seenAuthors: make(map[string]bool),
			want:        []string{"Carol <carol@example.com>"},
		},
		{
			name:        "multiple co-authors",
			body:        "msg\n\nCo-Authored-By: Alice <a@x.com>\nCo-Authored-By: Bob <b@x.com>\n",
			seenAuthors: make(map[string]bool),
			want:        []string{"Alice <a@x.com>", "Bob <b@x.com>"},
		},
		{
			name:        "skip already seen author",
			body:        "msg\n\nCo-Authored-By: Alice <alice@example.com>\n",
			seenAuthors: map[string]bool{"alice <alice@example.com>": true},
			want:        nil,
		},
		{
			name:        "dedup case insensitive",
			body:        "msg\n\nCo-Authored-By: Alice <ALICE@EXAMPLE.COM>\n",
			seenAuthors: map[string]bool{"alice <alice@example.com>": true},
			want:        nil,
		},
		{
			name:        "skip empty author after prefix",
			body:        "msg\n\nCo-Authored-By:   \n",
			seenAuthors: make(map[string]bool),
			want:        nil,
		},
		{
			name:        "skip lines shorter than prefix",
			body:        "short\nCo-Auth\n",
			seenAuthors: make(map[string]bool),
			want:        nil,
		},
		{
			name:        "no co-authors",
			body:        "Just a normal commit message\nwith multiple lines\n",
			seenAuthors: make(map[string]bool),
			want:        nil,
		},
		{
			name:        "dedup within body",
			body:        "msg\n\nCo-Authored-By: Alice <a@x.com>\nCo-Authored-By: Alice <a@x.com>\n",
			seenAuthors: make(map[string]bool),
			want:        []string{"Alice <a@x.com>"},
		},
		{
			name:        "leading whitespace on line",
			body:        "msg\n\n  Co-Authored-By: Alice <a@x.com>\n",
			seenAuthors: make(map[string]bool),
			want:        []string{"Alice <a@x.com>"},
		},
		{
			name:        "updates seenAuthors map",
			body:        "msg\n\nCo-Authored-By: New Person <new@x.com>\n",
			seenAuthors: make(map[string]bool),
			want:        []string{"New Person <new@x.com>"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCoAuthors([]byte(tt.body), tt.seenAuthors)
			if len(got) != len(tt.want) {
				t.Fatalf("parseCoAuthors() returned %d authors, want %d\ngot:  %v\nwant: %v", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseCoAuthors()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}

	// Verify seenAuthors side effect separately.
	t.Run("seenAuthors side effect", func(t *testing.T) {
		seen := make(map[string]bool)
		parseCoAuthors([]byte("Co-Authored-By: New Person <new@x.com>\n"), seen)
		if !seen["new person <new@x.com>"] {
			t.Error("expected seenAuthors to contain the new author after parsing")
		}
	})
}
