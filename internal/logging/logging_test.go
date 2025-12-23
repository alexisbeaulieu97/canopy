package logging

import (
	"testing"
)

func TestRedactSensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no sensitive data",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "API key in command",
			input:    "curl -H api_key=secret123 http://example.com",
			expected: "curl -H [REDACTED] http://example.com",
		},
		{
			name:     "password in output",
			input:    "password=mysecretpassword",
			expected: "[REDACTED]",
		},
		{
			name:     "HTTPS URL with credentials",
			input:    "git clone https://user:token@github.com/org/repo.git",
			expected: "git clone [REDACTED]github.com/org/repo.git",
		},
		{
			name:     "SSH URL with user",
			input:    "git clone ssh://git@github.com/org/repo.git",
			expected: "git clone [REDACTED]github.com/org/repo.git",
		},
		{
			name:     "bearer token",
			input:    "Authorization: Bearer abc123xyz",
			expected: "Authorization: [REDACTED]",
		},
		{
			name:     "auth token",
			input:    "auth_token: my-secret-token-value",
			expected: "[REDACTED]",
		},
		{
			name:     "secret key",
			input:    "secret_key=verysecretvalue",
			expected: "[REDACTED]",
		},
		{
			name:     "mixed content with secrets",
			input:    "Starting service with api_key=secret123 on port 8080",
			expected: "Starting service with [REDACTED] on port 8080",
		},
		{
			name:     "password with colon",
			input:    "password: mypassword123",
			expected: "[REDACTED]",
		},
		{
			name:     "multiple secrets",
			input:    "api_key=key1 password=pass1",
			expected: "[REDACTED] [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSensitive(tt.input)
			if result != tt.expected {
				t.Errorf("RedactSensitive(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
