package osmoutils

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// SortSlice sorts a slice of type T elements that implement constraints.Ordered.
// Mutates input slice s
func SortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func Filter[T interface{}](filter func(T) bool, s []T) []T {
	filteredSlice := []T{}
	for _, s := range s {
		if filter(s) {
			filteredSlice = append(filteredSlice, s)
		}
	}
	return filteredSlice
}

// ReverseSlice returns a reversed copy of the input slice.
// Does not mutate argument.
func ReverseSlice[T any](s []T) []T {
	newSlice := make([]T, len(s))
	maxIndex := len(s) - 1
	for i := 0; i < len(s); i++ {
		newSlice[maxIndex-i] = s[i]
	}
	return newSlice
}
