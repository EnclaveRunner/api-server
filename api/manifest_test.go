//nolint:paralleltest // Currently not supported (should do so in the future)
package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedFuncIdentifier holds the fields we want to assert after parsing a
// source string.
type expectedFuncIdentifier struct {
	namespace   string
	name        string
	iface       string
	funcName    string
	tag         string
	versionHash string
}

func TestParseSource(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    expectedFuncIdentifier
		expectError bool
	}{
		{
			name:  "valid source with tag",
			input: "myns:myapp/myiface/myfunc@latest",
			expected: expectedFuncIdentifier{
				namespace: "myns",
				name:      "myapp",
				iface:     "myiface",
				funcName:  "myfunc",
				tag:       "latest",
			},
		},
		{
			name:  "valid source with version hash",
			input: "myns:myapp/myiface/myfunc@hash:abc123def456",
			expected: expectedFuncIdentifier{
				namespace:   "myns",
				name:        "myapp",
				iface:       "myiface",
				funcName:    "myfunc",
				versionHash: "abc123def456",
			},
		},
		{
			name:  "valid source with numeric tag",
			input: "local:myservice/compute/run@v1.2.3",
			expected: expectedFuncIdentifier{
				namespace: "local",
				name:      "myservice",
				iface:     "compute",
				funcName:  "run",
				tag:       "v1.2.3",
			},
		},
		{
			name:        "invalid format - missing namespace colon",
			input:       "myapp/myiface/myfunc@latest",
			expectError: true,
		},
		{
			name:        "invalid format - missing interface and function",
			input:       "myns:myapp@latest",
			expectError: true,
		},
		{
			name:        "invalid format - missing version separator",
			input:       "myns:myapp/myiface/myfunc",
			expectError: true,
		},
		{
			name:        "invalid format - empty string",
			input:       "",
			expectError: true,
		},
		{
			name:  "valid source with complex hash",
			input: "enclave:nginx/http/serve@hash:sha256abc123",
			expected: expectedFuncIdentifier{
				namespace:   "enclave",
				name:        "nginx",
				iface:       "http",
				funcName:    "serve",
				versionHash: "sha256abc123",
			},
		},
		{
			name:  "edge case - single character components",
			input: "a:b/c/d@e",
			expected: expectedFuncIdentifier{
				namespace: "a",
				name:      "b",
				iface:     "c",
				funcName:  "d",
				tag:       "e",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSource(tt.input)

			if tt.expectError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
				assert.True(
					t,
					errors.Is(err, ErrInvalidIdentifier),
					"Expected ErrInvalidIdentifier, got: %v",
					err,
				)
			} else {
				require.NoError(t, err, "Unexpected error for input: %s", tt.input)
				require.NotNil(t, result)
				require.NotNil(t, result.Artifact)
				require.NotNil(t, result.Artifact.Package)
				assert.Equal(t, tt.expected.namespace, result.Artifact.Package.Namespace, "namespace mismatch")
				assert.Equal(t, tt.expected.name, result.Artifact.Package.Name, "name mismatch")
				assert.Equal(t, tt.expected.iface, result.Interface, "interface mismatch")
				assert.Equal(t, tt.expected.funcName, result.Name, "function name mismatch")
				if tt.expected.versionHash != "" {
					assert.Equal(t, tt.expected.versionHash, result.Artifact.GetVersionHash(), "version hash mismatch")
				} else {
					assert.Equal(t, tt.expected.tag, result.Artifact.GetTag(), "tag mismatch")
				}
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
			name: "valid manifest YAML",
			input: `apiVersion: v1
kind: Blueprint
metadata:
  name: test-blueprint
spec:
  artifact:
    source: github.com/user/repo:latest`,
			expected: BaseManifest{
				APIVersion: "v1",
				Kind:       "Blueprint",
				Metadata:   map[string]any{"name": "test-blueprint"},
				Spec: map[string]any{
					"artifact": map[string]any{
						"source": "github.com/user/repo:latest",
					},
				},
			},
			expectError: false,
		},
		{
			name: "minimal valid manifest YAML",
			input: `apiVersion: v1
kind: Task
metadata: {}
spec: {}`,
			expected: BaseManifest{
				APIVersion: "v1",
				Kind:       "Task",
				Metadata:   map[string]any{},
				Spec:       map[string]any{},
			},
			expectError: false,
		},
		{
			name:        "invalid YAML",
			input:       `apiVersion: v1\nkind: [invalid yaml structure`,
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
			name:        "not YAML object",
			input:       `just a string without structure`,
			expected:    BaseManifest{},
			expectError: true,
		},
		{
			name: "complex blueprint YAML",
			input: `apiVersion: v1
kind: Blueprint
metadata:
  name: my-blueprint
  namespace: default
spec:
  artifact:
    source: github.com/user/myapp:v1.0.0
    function: process
    input: SGVsbG8gV29ybGQ=
status:
  healthy: true
  created: "2023-01-01T00:00:00Z"
  events: []
  revisions: 1`,
			expected: BaseManifest{
				APIVersion: "v1",
				Kind:       "Blueprint",
				Metadata: map[string]any{
					"name":      "my-blueprint",
					"namespace": "default",
				},
				Spec: map[string]any{
					"artifact": map[string]any{
						"source":   "github.com/user/myapp:v1.0.0",
						"function": "process",
						"input":    "SGVsbG8gV29ybGQ=",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := unmarshalManifest([]byte(tt.input))

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
		"myns:myapp/myiface/myfunc@latest",
		"myns:nginx/http/serve@hash:sha256abc123",
		"myns:myservice/compute/run@v1.2.3",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for b.Loop() {
				_, _ = parseSource(tc)
			}
		})
	}
}
