// nolint
package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSource(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Identifier
		expectError bool
	}{
		{
			name:  "valid source with tag",
			input: "github.com/user/myartifact:latest",
			expected: Identifier{
				Source: "github.com",
				Author: "user",
				Name:   "myartifact",
				Tag:    "latest",
				Hash:   "",
			},
			expectError: false,
		},
		{
			name:  "valid source with version hash",
			input: "github.com/user/myartifact:hash:abc123def456",
			expected: Identifier{
				Source: "github.com",
				Author: "user",
				Name:   "myartifact",
				Tag:    "",
				Hash:   "abc123def456",
			},
			expectError: false,
		},
		{
			name:  "valid source with numeric tag",
			input: "registry.local/company/service:v1.2.3",
			expected: Identifier{
				Source: "registry.local",
				Author: "company",
				Name:   "service",
				Tag:    "v1.2.3",
				Hash:   "",
			},
			expectError: false,
		},
		{
			name:        "invalid format - missing parts",
			input:       "github.com/user",
			expected:    Identifier{},
			expectError: true,
		},
		{
			name:        "invalid format - too many slashes",
			input:       "github.com/user/repo/extra:tag",
			expected:    Identifier{},
			expectError: true,
		},
		{
			name:        "invalid format - missing colon",
			input:       "github.com/user/repo",
			expected:    Identifier{},
			expectError: true,
		},
		{
			name:        "invalid format - empty string",
			input:       "",
			expected:    Identifier{},
			expectError: true,
		},
		{
			name:        "invalid format - only slashes",
			input:       "///",
			expected:    Identifier{},
			expectError: true,
		},
		{
			name:  "valid source with complex hash format",
			input: "docker.io/library/nginx:hash:sha2562a25e1f8f0aa9571689513d5b68c8bb94b9bc8f5a9229a8c0250482cfb1c8a99",
			expected: Identifier{
				Source: "docker.io",
				Author: "library",
				Name:   "nginx",
				Tag:    "",
				Hash:   "sha2562a25e1f8f0aa9571689513d5b68c8bb94b9bc8f5a9229a8c0250482cfb1c8a99",
			},
			expectError: false,
		},
		{
			name:  "edge case - single character components",
			input: "a/b/c:d",
			expected: Identifier{
				Source: "a",
				Author: "b",
				Name:   "c",
				Tag:    "d",
				Hash:   "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSource(tt.input)

			if tt.expectError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
				assert.Equal(
					t,
					ErrInvalidIdentifier,
					err,
					"Expected ErrInvalidIdentifier",
				)
			} else {
				assert.NoError(t, err, "Unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "Result mismatch for input: %s", tt.input)
			}
		})
	}
}

func TestUnmarshalManifest(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    BaseManifest
		expectError bool
	}{
		{
			name: "valid manifest JSON",
			input: `{
				"apiVersion": "v1",
				"kind": "Blueprint",
				"metadata": {"name": "test-blueprint"},
				"spec": {"artifact": {"source": "github.com/user/repo:latest"}}
			}`,
			expected: BaseManifest{
				APIVersion: "v1",
				Kind:       "Blueprint",
				Metadata:   map[string]interface{}{"name": "test-blueprint"},
				Spec: map[string]interface{}{
					"artifact": map[string]interface{}{
						"source": "github.com/user/repo:latest",
					},
				},
			},
			expectError: false,
		},
		{
			name: "minimal valid manifest",
			input: `{
				"apiVersion": "v1",
				"kind": "Task",
				"metadata": {},
				"spec": {}
			}`,
			expected: BaseManifest{
				APIVersion: "v1",
				Kind:       "Task",
				Metadata:   map[string]interface{}{},
				Spec:       map[string]interface{}{},
			},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			input:       `{"apiVersion": "v1", "kind":}`,
			expected:    BaseManifest{},
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    BaseManifest{},
			expectError: true,
		},
		{
			name:        "not JSON object",
			input:       `"just a string"`,
			expected:    BaseManifest{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err := unmarshalManifest(reader)

			if tt.expectError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
			} else {
				require.NoError(t, err, "Unexpected error for input: %s", tt.input)
				assert.Equal(t, tt.expected.APIVersion, result.APIVersion)
				assert.Equal(t, tt.expected.Kind, result.Kind)
				assert.Equal(t, tt.expected.Metadata, result.Metadata)
				assert.Equal(t, tt.expected.Spec, result.Spec)
			}
		})
	}
}

func TestIdentifierStruct(t *testing.T) {
	t.Run("Identifier struct fields", func(t *testing.T) {
		id := Identifier{
			Source: "github.com",
			Author: "testuser",
			Name:   "testartifact",
			Hash:   "abc123",
			Tag:    "v1.0.0",
		}

		assert.Equal(t, "github.com", id.Source)
		assert.Equal(t, "testuser", id.Author)
		assert.Equal(t, "testartifact", id.Name)
		assert.Equal(t, "abc123", id.Hash)
		assert.Equal(t, "v1.0.0", id.Tag)
	})
}

func TestErrInvalidIdentifier(t *testing.T) {
	t.Run("ErrInvalidIdentifier is properly defined", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidIdentifier)
		assert.Contains(
			t,
			ErrInvalidIdentifier.Error(),
			"invalid identifier format",
		)
	})
}

// Benchmark tests for parseSource function
func BenchmarkParseSource(b *testing.B) {
	testCases := []string{
		"github.com/user/repo:latest",
		"docker.io/library/nginx:hash:sha256abc123",
		"registry.local/company/service:v1.2.3",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for range b.N {
				_, _ = parseSource(tc)
			}
		})
	}
}
