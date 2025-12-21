package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestTemplateParsing(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `projects_root: /projects
workspaces_root: /workspaces
closed_root: /closed

templates:
  backend:
    description: "Backend workspace defaults"
    repos: ["backend", "common"]
    default_branch: "main"
    setup_commands:
      - "echo setup"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	tmpl, ok := cfg.Templates["backend"]
	if !ok {
		t.Fatalf("expected backend template to be parsed")
	}

	if tmpl.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", tmpl.DefaultBranch, "main")
	}

	if len(tmpl.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(tmpl.Repos))
	}
}

func TestResolveTemplateUnknown(t *testing.T) {
	cfg := &Config{Templates: map[string]Template{"backend": {Repos: []string{"repo"}}}}

	_, err := cfg.ResolveTemplate("missing")
	if err == nil {
		t.Fatal("expected error for unknown template")
	}

	if !strings.Contains(err.Error(), "backend") {
		t.Fatalf("expected available template names in error, got %q", err.Error())
	}
}

func TestValidateTemplatesMissingRepos(t *testing.T) {
	cfg := &Config{Templates: map[string]Template{"bad": {Description: "missing repos"}}}

	err := cfg.ValidateTemplates()
	if err == nil {
		t.Fatal("expected validation error for template without repos")
	}

	if !strings.Contains(err.Error(), "templates.bad.repos") {
		t.Fatalf("unexpected error: %v", err)
	}
}
