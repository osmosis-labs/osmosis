package osmoutils

import (
	"math/rand"
	"reflect"
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

// ContainsDuplicateDeepEqual returns true if there are duplicates
// in the slice by performing deep comparison. This is useful
// for comparing matrices or slices of pointers.
// Returns false if there are no deep equal duplicates.
func ContainsDuplicateDeepEqual[T any](multihops []T) bool {
	for i := 0; i < len(multihops)-1; i++ {
		if reflect.DeepEqual(multihops[i], multihops[i+1]) {
			return true
		}
	}
	return false
}

type LessFunc[T any] func(a, b T) bool

// MergeSlices efficiently merges two sorted slices into a single sorted slice.
// The resulting slice contains all elements from slice1 and slice2, sorted according to the less function.
// The input slices must be sorted in ascending order according to the less function.
// The less function takes two elements of type T and returns a boolean value indicating whether the first element is less than the second element.
// The function returns a new slice containing all elements from slice1 and slice2, sorted according to the less function.
// The function does not modify the input slices.
func MergeSlices[T any](slice1, slice2 []T, less LessFunc[T]) []T {
	result := make([]T, 0, len(slice1)+len(slice2))
	i, j := 0, 0

	for i < len(slice1) && j < len(slice2) {
		if less(slice1[i], slice2[j]) {
			result = append(result, slice1[i])
			i++
		} else {
			result = append(result, slice2[j])
			j++
		}
	}

	// Append any remaining elements from slice1 and slice2
	result = append(result, slice1[i:]...)
	result = append(result, slice2[j:]...)

	return result
}

// Contains returns true if the slice contains the item, false otherwise.
func Contains[T comparable](slice []T, item T) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// GetRandomSubset returns a random subset of the given slice
func GetRandomSubset[T any](slice []T) []T {
	if len(slice) == 0 {
		return []T{}
	}

	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})

	n := rand.Intn(len(slice))
	return slice[:n]
}
