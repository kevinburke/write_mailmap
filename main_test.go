package main

import (
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func identityAuthorMapper(authors []string) ([]string, error) {
	return authors, nil
}

func TestParseCoAuthors(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		seenAuthors map[string]bool
		mapAuthors  authorMapper
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
		{
			name:        "mailmaps co-author before dedup",
			body:        "msg\n\nCo-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>\n",
			seenAuthors: make(map[string]bool),
			mapAuthors: func(authors []string) ([]string, error) {
				if len(authors) != 1 || authors[0] != "Claude Opus 4.8 (1M context) <noreply@anthropic.com>" {
					t.Fatalf("mapper got %v, want Claude Opus co-author", authors)
				}
				return []string{"Claude <noreply@anthropic.com>"}, nil
			},
			want: []string{"Claude <noreply@anthropic.com>"},
		},
		{
			name: "dedups mailmapped co-author",
			body: "msg\n\nCo-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>\n",
			seenAuthors: map[string]bool{
				"claude <noreply@anthropic.com>": true,
			},
			mapAuthors: func(authors []string) ([]string, error) {
				return []string{"Claude <noreply@anthropic.com>"}, nil
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mapAuthors == nil {
				tt.mapAuthors = identityAuthorMapper
			}
			got, err := parseCoAuthors([]byte(tt.body), tt.seenAuthors, tt.mapAuthors)
			if err != nil {
				t.Fatal(err)
			}
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
		if _, err := parseCoAuthors([]byte("Co-Authored-By: New Person <new@x.com>\n"), seen, identityAuthorMapper); err != nil {
			t.Fatal(err)
		}
		if !seen["new person <new@x.com>"] {
			t.Error("expected seenAuthors to contain the new author after parsing")
		}
	})
}

func TestParseCoAuthorsInvalidTrailer(t *testing.T) {
	_, err := parseCoAuthors([]byte("Co-Authored-By: not an address\n"), make(map[string]bool), identityAuthorMapper)
	if err == nil {
		t.Fatal("expected invalid Co-Authored-By trailer to return an error")
	}
}

func TestMailmapAuthors(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if out, err := exec.Command("git", "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	err := os.WriteFile(".mailmap", []byte("Claude <noreply@anthropic.com> Claude Opus 4.8 (1M context) <noreply@anthropic.com>\n"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	got, err := mailmapAuthors([]string{"Claude Opus 4.8 (1M context) <noreply@anthropic.com>"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Claude <noreply@anthropic.com>"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mailmapAuthors() = %v, want %v", got, want)
	}
}
