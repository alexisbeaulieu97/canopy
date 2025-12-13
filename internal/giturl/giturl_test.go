package giturl

import (
	"testing"
)

func TestIsURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "https", input: "https://github.com/org/repo", want: true},
		{name: "http", input: "http://github.com/org/repo", want: true},
		{name: "ssh scheme", input: "ssh://git@github.com/org/repo", want: true},
		{name: "git protocol", input: "git://github.com/org/repo", want: true},
		{name: "scp style", input: "git@github.com:org/repo", want: true},
		{name: "file", input: "file:///tmp/repo", want: true},
		{name: "simple name", input: "repo-name", want: false},
		{name: "org/repo", input: "org/repo", want: false},
		{name: "empty", input: "", want: false},
		{name: "whitespace", input: "   ", want: false},
		{name: "https with port", input: "https://github.com:443/org/repo", want: true},
		{name: "ssh with port", input: "ssh://git@github.com:22/org/repo", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := IsURL(tt.input)
			if got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractRepoName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "standard https", url: "https://github.com/org/repo.git", want: "repo"},
		{name: "scp style", url: "git@github.com:org/repo.git", want: "repo"},
		{name: "trailing slash", url: "https://github.com/org/repo/", want: "repo"},
		{name: "multiple trailing slashes", url: "https://github.com/org/repo///", want: "repo"},
		{name: "empty input", url: "", want: ""},
		{name: "slash only", url: "///", want: ""},
		{name: "file scheme", url: "file:///tmp/repo.git", want: "repo"},
		{name: "ssh scheme", url: "ssh://git@example.com/org/repo.git", want: "repo"},
		{name: "https with user info", url: "https://user:token@github.com/org/repo.git", want: "repo"},
		{name: "no .git suffix", url: "https://github.com/org/repo", want: "repo"},
		{name: "git protocol", url: "git://github.com/org/repo.git", want: "repo"},
		{name: "complex path", url: "https://gitlab.com/group/subgroup/project.git", want: "project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ExtractRepoName(tt.url)
			if got != tt.want {
				t.Errorf("ExtractRepoName(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestDeriveAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "standard https", url: "https://github.com/org/Repo.git", want: "repo"},
		{name: "scp style", url: "git@github.com:org/MY-REPO.git", want: "my-repo"},
		{name: "empty input", url: "", want: ""},
		{name: "whitespace only", url: "   ", want: ""},
		{name: "mixed case", url: "https://github.com/org/MyProject.git", want: "myproject"},
		{name: "already lowercase", url: "https://github.com/org/myrepo.git", want: "myrepo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := DeriveAlias(tt.url)
			if got != tt.want {
				t.Errorf("DeriveAlias(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}
