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

// ReverseSlice reverses the input slice in place.
// Does mutate argument.
func ReverseSlice[T any](s []T) []T {
	maxIndex := len(s)
	for i := 0; i < maxIndex/2; i++ {
		temp := s[i]
		s[i] = s[maxIndex-i-1]
		s[maxIndex-1-i] = temp
	}
	return s
}

// ContainsDuplicate checks if there are any duplicate
// elements in the slice.
func ContainsDuplicate[T any](arr []T) bool {
	visited := make(map[any]bool, 0)
	for i := 0; i < len(arr); i++ {
		if visited[arr[i]] {
			return true
		} else {
			visited[arr[i]] = true
		}
	}
	return false
}
