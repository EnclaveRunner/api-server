//nolint:paralleltest // Currently not supported (should do so in the future)
package api

import (
	"errors"
	"testing"

	pb "api-server/proto_gen"

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

func TestAnyToProtoVal(t *testing.T) {
	tests := []struct {
		name  string
		input any
		check func(t *testing.T, got *pb.Val)
	}{
		{
			name:  "bool true",
			input: true,
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_BoolVal)
				require.True(t, ok, "expected BoolVal")
				assert.True(t, v.BoolVal)
			},
		},
		{
			name:  "bool false",
			input: false,
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_BoolVal)
				require.True(t, ok, "expected BoolVal")
				assert.False(t, v.BoolVal)
			},
		},
		{
			name:  "int",
			input: int(42),
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_S64Val)
				require.True(t, ok, "expected S64Val")
				assert.Equal(t, int64(42), v.S64Val)
			},
		},
		{
			name:  "int64",
			input: int64(-100),
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_S64Val)
				require.True(t, ok, "expected S64Val")
				assert.Equal(t, int64(-100), v.S64Val)
			},
		},
		{
			name:  "float64",
			input: float64(3.14),
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_F64Val)
				require.True(t, ok, "expected F64Val")
				assert.InDelta(t, 3.14, v.F64Val, 1e-9)
			},
		},
		{
			name:  "string",
			input: "hello",
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_StringVal)
				require.True(t, ok, "expected StringVal")
				assert.Equal(t, "hello", v.StringVal)
			},
		},
		{
			name:  "nil",
			input: nil,
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				_, ok := got.Value.(*pb.Val_OptionVal)
				require.True(t, ok, "expected OptionVal for nil")
			},
		},
		{
			name:  "unknown type falls back to OptionVal",
			input: struct{ x int }{x: 1},
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				_, ok := got.Value.(*pb.Val_OptionVal)
				require.True(t, ok, "expected OptionVal for unknown type")
			},
		},
		{
			name:  "empty slice",
			input: []interface{}{},
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_ListVal)
				require.True(t, ok, "expected ListVal")
				assert.Empty(t, v.ListVal.Values)
			},
		},
		{
			name:  "slice with mixed types",
			input: []interface{}{true, int(1), "x"},
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_ListVal)
				require.True(t, ok, "expected ListVal")
				require.Len(t, v.ListVal.Values, 3)
				_, ok = v.ListVal.Values[0].Value.(*pb.Val_BoolVal)
				assert.True(t, ok, "first element should be BoolVal")
				_, ok = v.ListVal.Values[1].Value.(*pb.Val_S64Val)
				assert.True(t, ok, "second element should be S64Val")
				_, ok = v.ListVal.Values[2].Value.(*pb.Val_StringVal)
				assert.True(t, ok, "third element should be StringVal")
			},
		},
		{
			name:  "nested slice",
			input: []any{[]any{int(1), int(2)}},
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				outer, ok := got.Value.(*pb.Val_ListVal)
				require.True(t, ok, "expected outer ListVal")
				require.Len(t, outer.ListVal.Values, 1)
				inner, ok := outer.ListVal.Values[0].Value.(*pb.Val_ListVal)
				require.True(t, ok, "expected inner ListVal")
				require.Len(t, inner.ListVal.Values, 2)
			},
		},
		{
			name:  "empty map",
			input: map[string]any{},
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_RecordVal)
				require.True(t, ok, "expected RecordVal")
				assert.Empty(t, v.RecordVal.Fields)
			},
		},
		{
			name:  "map with single string field",
			input: map[string]interface{}{"key": "value"},
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_RecordVal)
				require.True(t, ok, "expected RecordVal")
				require.Len(t, v.RecordVal.Fields, 1)
				assert.Equal(t, "key", v.RecordVal.Fields[0].Name)
				sv, ok := v.RecordVal.Fields[0].Value.Value.(*pb.Val_StringVal)
				require.True(t, ok, "expected StringVal for field value")
				assert.Equal(t, "value", sv.StringVal)
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"inner": map[string]interface{}{"n": int(7)},
			},
			check: func(t *testing.T, got *pb.Val) {
				t.Helper()
				v, ok := got.Value.(*pb.Val_RecordVal)
				require.True(t, ok, "expected outer RecordVal")
				require.Len(t, v.RecordVal.Fields, 1)
				assert.Equal(t, "inner", v.RecordVal.Fields[0].Name)
				inner, ok := v.RecordVal.Fields[0].Value.Value.(*pb.Val_RecordVal)
				require.True(t, ok, "expected inner RecordVal")
				require.Len(t, inner.RecordVal.Fields, 1)
				assert.Equal(t, "n", inner.RecordVal.Fields[0].Name)
				nv, ok := inner.RecordVal.Fields[0].Value.Value.(*pb.Val_S64Val)
				require.True(t, ok, "expected S64Val for nested field")
				assert.Equal(t, int64(7), nv.S64Val)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := anyToProtoVal(tt.input)
			require.NotNil(t, got)
			tt.check(t, got)
		})
	}
}

func TestProtoValToAny(t *testing.T) {
	tests := []struct {
		name  string
		input *pb.Val
		want  any
	}{
		{
			name:  "nil proto value",
			input: nil,
			want:  nil,
		},
		{
			name:  "bool",
			input: &pb.Val{Value: &pb.Val_BoolVal{BoolVal: true}},
			want:  true,
		},
		{
			name:  "s64",
			input: &pb.Val{Value: &pb.Val_S64Val{S64Val: -42}},
			want:  int64(-42),
		},
		{
			name:  "f64",
			input: &pb.Val{Value: &pb.Val_F64Val{F64Val: 3.5}},
			want:  float64(3.5),
		},
		{
			name:  "string",
			input: &pb.Val{Value: &pb.Val_StringVal{StringVal: "hello"}},
			want:  "hello",
		},
		{
			name: "list",
			input: &pb.Val{
				Value: &pb.Val_ListVal{ListVal: &pb.ListVal{Values: []*pb.Val{
					{Value: &pb.Val_BoolVal{BoolVal: false}},
					{Value: &pb.Val_S64Val{S64Val: 1}},
					{Value: &pb.Val_StringVal{StringVal: "x"}},
				}}},
			},
			want: []interface{}{false, int64(1), "x"},
		},
		{
			name: "record",
			input: &pb.Val{
				Value: &pb.Val_RecordVal{
					RecordVal: &pb.RecordVal{Fields: []*pb.RecordField{
						{Name: "ok", Value: &pb.Val{Value: &pb.Val_BoolVal{BoolVal: true}}},
						{Name: "n", Value: &pb.Val{Value: &pb.Val_S64Val{S64Val: 7}}},
					}},
				},
			},
			want: map[string]interface{}{"ok": true, "n": int64(7)},
		},
		{
			name:  "option none",
			input: &pb.Val{Value: &pb.Val_OptionVal{OptionVal: &pb.OptionVal{}}},
			want:  nil,
		},
		{
			name: "nested structure",
			input: &pb.Val{
				Value: &pb.Val_RecordVal{
					RecordVal: &pb.RecordVal{Fields: []*pb.RecordField{
						{
							Name: "items",
							Value: &pb.Val{
								Value: &pb.Val_ListVal{ListVal: &pb.ListVal{Values: []*pb.Val{
									{Value: &pb.Val_S64Val{S64Val: 1}},
									{Value: &pb.Val_OptionVal{OptionVal: &pb.OptionVal{}}},
								}}},
							},
						},
					}},
				},
			},
			want: map[string]interface{}{"items": []interface{}{int64(1), nil}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := protoValToAny(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAnyProtoValRoundTripSubset(t *testing.T) {
	input := map[string]interface{}{
		"flag": true,
		"num":  int64(11),
		"txt":  "abc",
		"list": []interface{}{float64(1.5), nil, map[string]interface{}{"k": "v"}},
	}

	got := protoValToAny(anyToProtoVal(input))
	assert.Equal(t, input, got)
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
