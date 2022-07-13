package osmoutils

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// SortSlice sorts a slice of type T elements that implement constraints.Ordered.
func SortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func Filter[T interface{}](filter func(T) bool, s []T) []T {
	filteredSlice := []T{}
	for _, s := range s {
		if filter(s) {
			filteredSlice = append(filteredSlice)
		}
	}
	return filteredSlice
}
