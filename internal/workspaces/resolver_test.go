package workspaces

import (
	"testing"
)

func TestRepoResolver_Resolve(t *testing.T) {
	t.Parallel()

	t.Run("empty string returns false", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		_, ok, err := resolver.Resolve("", true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if ok {
			t.Error("expected ok=false for empty string")
		}
	})

	t.Run("whitespace only returns false", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		_, ok, err := resolver.Resolve("   ", true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if ok {
			t.Error("expected ok=false for whitespace")
		}
	})

	t.Run("https URL extracts repo name", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		repo, ok, err := resolver.Resolve("https://github.com/org/my-repo.git", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected ok=true for URL")
		}

		if repo.Name != "my-repo" {
			t.Errorf("expected name 'my-repo', got %q", repo.Name)
		}

		if repo.URL != "https://github.com/org/my-repo.git" {
			t.Errorf("expected URL to match input, got %q", repo.URL)
		}
	})

	t.Run("scp style URL extracts repo name", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		repo, ok, err := resolver.Resolve("git@github.com:org/my-repo.git", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected ok=true for scp URL")
		}

		if repo.Name != "my-repo" {
			t.Errorf("expected name 'my-repo', got %q", repo.Name)
		}
	})

	t.Run("GitHub shorthand creates URL", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		repo, ok, err := resolver.Resolve("myorg/myrepo", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !ok {
			t.Fatal("expected ok=true for GitHub shorthand")
		}

		if repo.Name != "myrepo" {
			t.Errorf("expected name 'myrepo', got %q", repo.Name)
		}

		if repo.URL != "https://github.com/myorg/myrepo" {
			t.Errorf("expected GitHub URL, got %q", repo.URL)
		}
	})

	t.Run("GitHub shorthand with empty owner returns error", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		_, _, err := resolver.Resolve("/myrepo", true)
		if err == nil {
			t.Fatal("expected error for shorthand with empty owner")
		}
	})

	t.Run("GitHub shorthand with empty repo returns error", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		_, _, err := resolver.Resolve("myorg/", true)
		if err == nil {
			t.Fatal("expected error for shorthand with empty repo")
		}
	})

	t.Run("unknown identifier returns error when user requested", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		_, _, err := resolver.Resolve("unknown-repo", true)
		if err == nil {
			t.Fatal("expected error for unknown identifier")
		}
	})

	t.Run("unknown identifier returns error when not user requested", func(t *testing.T) {
		t.Parallel()

		resolver := NewRepoResolver(nil)

		_, _, err := resolver.Resolve("unknown-repo", false)
		if err == nil {
			t.Fatal("expected error for unknown identifier")
		}
	})
}
