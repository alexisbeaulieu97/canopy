package workspaces

import (
	"testing"
)

func TestURLStrategy_Name(t *testing.T) {
	t.Parallel()

	s := NewURLStrategy(nil)
	if s.Name() != "url" {
		t.Errorf("expected name 'url', got %q", s.Name())
	}
}

func TestURLStrategy_Resolve(t *testing.T) {
	t.Parallel()

	t.Run("https URL without registry", func(t *testing.T) {
		t.Parallel()

		s := NewURLStrategy(nil)

		repo, ok := s.Resolve("https://github.com/org/my-repo.git")
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

	t.Run("scp style URL", func(t *testing.T) {
		t.Parallel()

		s := NewURLStrategy(nil)

		repo, ok := s.Resolve("git@github.com:org/my-repo.git")
		if !ok {
			t.Fatal("expected ok=true for scp URL")
		}

		if repo.Name != "my-repo" {
			t.Errorf("expected name 'my-repo', got %q", repo.Name)
		}
	})

	t.Run("URL with registry lookup", func(t *testing.T) {
		t.Parallel()

		lookup := func(url string) (string, string, bool) {
			if url == "https://github.com/org/repo.git" {
				return "my-alias", "https://github.com/org/repo.git", true
			}

			return "", "", false
		}

		s := NewURLStrategy(lookup)

		repo, ok := s.Resolve("https://github.com/org/repo.git")
		if !ok {
			t.Fatal("expected ok=true")
		}

		if repo.Name != "my-alias" {
			t.Errorf("expected name 'my-alias' from registry, got %q", repo.Name)
		}
	})

	t.Run("non-URL returns false", func(t *testing.T) {
		t.Parallel()

		s := NewURLStrategy(nil)

		_, ok := s.Resolve("org/repo")
		if ok {
			t.Error("expected ok=false for non-URL")
		}
	})

	t.Run("simple name returns false", func(t *testing.T) {
		t.Parallel()

		s := NewURLStrategy(nil)

		_, ok := s.Resolve("myrepo")
		if ok {
			t.Error("expected ok=false for simple name")
		}
	})
}

func TestRegistryStrategy_Name(t *testing.T) {
	t.Parallel()

	s := NewRegistryStrategy(nil)
	if s.Name() != "registry" {
		t.Errorf("expected name 'registry', got %q", s.Name())
	}
}

func TestRegistryStrategy_Resolve(t *testing.T) {
	t.Parallel()

	t.Run("found in registry", func(t *testing.T) {
		t.Parallel()

		lookup := func(alias string) (string, string, bool) {
			if alias == "myrepo" {
				return "myrepo", "https://github.com/org/myrepo.git", true
			}

			return "", "", false
		}

		s := NewRegistryStrategy(lookup)

		repo, ok := s.Resolve("myrepo")
		if !ok {
			t.Fatal("expected ok=true for registered alias")
		}

		if repo.Name != "myrepo" {
			t.Errorf("expected name 'myrepo', got %q", repo.Name)
		}

		if repo.URL != "https://github.com/org/myrepo.git" {
			t.Errorf("expected URL from registry, got %q", repo.URL)
		}
	})

	t.Run("not found in registry", func(t *testing.T) {
		t.Parallel()

		lookup := func(_ string) (string, string, bool) {
			return "", "", false
		}

		s := NewRegistryStrategy(lookup)

		_, ok := s.Resolve("unknown")
		if ok {
			t.Error("expected ok=false for unknown alias")
		}
	})

	t.Run("nil lookup returns false", func(t *testing.T) {
		t.Parallel()

		s := NewRegistryStrategy(nil)

		_, ok := s.Resolve("myrepo")
		if ok {
			t.Error("expected ok=false with nil lookup")
		}
	})
}

func TestGitHubShorthandStrategy_Name(t *testing.T) {
	t.Parallel()

	s := NewGitHubShorthandStrategy()
	if s.Name() != "github-shorthand" {
		t.Errorf("expected name 'github-shorthand', got %q", s.Name())
	}
}

func TestGitHubShorthandStrategy_Resolve(t *testing.T) {
	t.Parallel()

	t.Run("valid shorthand", func(t *testing.T) {
		t.Parallel()

		s := NewGitHubShorthandStrategy()

		repo, ok := s.Resolve("myorg/myrepo")
		if !ok {
			t.Fatal("expected ok=true for valid shorthand")
		}

		if repo.Name != "myrepo" {
			t.Errorf("expected name 'myrepo', got %q", repo.Name)
		}

		if repo.URL != "https://github.com/myorg/myrepo" {
			t.Errorf("expected GitHub URL, got %q", repo.URL)
		}
	})

	t.Run("empty owner returns false", func(t *testing.T) {
		t.Parallel()

		s := NewGitHubShorthandStrategy()

		_, ok := s.Resolve("/myrepo")
		if ok {
			t.Error("expected ok=false for empty owner")
		}
	})

	t.Run("empty repo returns false", func(t *testing.T) {
		t.Parallel()

		s := NewGitHubShorthandStrategy()

		_, ok := s.Resolve("myorg/")
		if ok {
			t.Error("expected ok=false for empty repo")
		}
	})

	t.Run("no slash returns false", func(t *testing.T) {
		t.Parallel()

		s := NewGitHubShorthandStrategy()

		_, ok := s.Resolve("myrepo")
		if ok {
			t.Error("expected ok=false for no slash")
		}
	})

	t.Run("multiple slashes returns false", func(t *testing.T) {
		t.Parallel()

		s := NewGitHubShorthandStrategy()

		_, ok := s.Resolve("org/group/repo")
		if ok {
			t.Error("expected ok=false for multiple slashes")
		}
	})

	t.Run("whitespace trimmed", func(t *testing.T) {
		t.Parallel()

		s := NewGitHubShorthandStrategy()

		repo, ok := s.Resolve(" myorg / myrepo ")
		if !ok {
			t.Fatal("expected ok=true with whitespace")
		}

		if repo.Name != "myrepo" {
			t.Errorf("expected name 'myrepo', got %q", repo.Name)
		}
	})
}
