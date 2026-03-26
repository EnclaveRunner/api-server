package api

import (
	"cmp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		list     []int
		limit    int
		offset   int
		expected []int
	}{
		{
			name:     "sorts input and returns window",
			list:     []int{5, 1, 3, 2, 4},
			limit:    2,
			offset:   1,
			expected: []int{2, 3},
		},
		{
			name:     "offset at list length returns empty",
			list:     []int{3, 1, 2},
			limit:    2,
			offset:   3,
			expected: []int{},
		},
		{
			name:     "offset beyond list length returns empty",
			list:     []int{3, 1, 2},
			limit:    2,
			offset:   4,
			expected: []int{},
		},
		{
			name:     "limit larger than remaining elements",
			list:     []int{9, 7, 8, 6},
			limit:    10,
			offset:   2,
			expected: []int{8, 9},
		},
		{
			name:     "zero limit returns empty",
			list:     []int{2, 1, 3},
			limit:    0,
			offset:   1,
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := paginate(tt.list, tt.limit, tt.offset, cmp.Compare[int])

			assert.Equal(t, tt.expected, result)
		})
	}
}
