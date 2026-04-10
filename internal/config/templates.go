package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	cerrors "github.com/alexisbeaulieu97/canopy/internal/errors"
	"github.com/alexisbeaulieu97/canopy/internal/ports"
)

// GetReposForWorkspace returns default repos for a given workspace ID based on patterns.
func (c *Config) GetReposForWorkspace(workspaceID string) []string {
	for _, p := range c.Defaults.WorkspacePatterns {
		matched, err := regexp.MatchString(p.Pattern, workspaceID)
		if err == nil && matched {
			return p.Repos
		}
	}

	return nil
}

// GetTemplates returns a copy of the configured templates keyed by name.
func (c *Config) GetTemplates() map[string]ports.WorkspaceTemplate {
	if len(c.Templates) == 0 {
		return map[string]ports.WorkspaceTemplate{}
	}

	templates := make(map[string]ports.WorkspaceTemplate, len(c.Templates))
	for name, tmpl := range c.Templates {
		templates[name] = toPortTemplate(name, tmpl)
	}

	return templates
}

// ResolveTemplate returns a template by name with helpful errors if missing.
func (c *Config) ResolveTemplate(name string) (ports.WorkspaceTemplate, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ports.WorkspaceTemplate{}, cerrors.NewInvalidArgument("template", "template name is required")
	}

	templates := c.GetTemplates()
	if tmpl, ok := templates[name]; ok {
		return tmpl, nil
	}

	if len(templates) == 0 {
		return ports.WorkspaceTemplate{}, cerrors.NewInvalidArgument("template", "no templates are defined")
	}

	names := make([]string, 0, len(templates))
	for tmplName := range templates {
		names = append(names, tmplName)
	}
	sort.Strings(names)

	return ports.WorkspaceTemplate{}, cerrors.NewInvalidArgument("template", fmt.Sprintf("unknown template %q (available: %s)", name, strings.Join(names, ", ")))
}
