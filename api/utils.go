package api

import (
	"slices"
)

func paginate[S ~[]E, E any](
	list S,
	limit, offset int,
	cmp func(a, b E) int,
) S {
	if offset >= len(list) {
		return S{}
	}

	slices.SortFunc(list, cmp)

	start := offset
	end := min(len(list), offset+limit)

	return list[start:end]
}
