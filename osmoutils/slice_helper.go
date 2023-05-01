package osmoutils

import (
	"fmt"
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

// iThSmallest returns the ith smallest element in the slice.
// The notion of ordering is defined by the given less function.
// It uses a divide and conquer approach without relying on
// randomization for selecting a pivot. This makes this function deterministic.
// Instead of randomization for pivot selection,
// it uses a median of medians algorithm for selecting the pivot.
// The results in the following reccurence relation:
// T(n) <= T(n/5) + T(7n/10) + O(n)
// By master's theorem, this results in O(n) time complexity.
// Allocates additional space and results in 0(n) space complexity.
// In place.
// Parameters:
// s: slice of type T and size n
// i: index of the ith smallest element to return where i is in the range [1, n]
// less: a function that defines the ordering of the elements in the slice.
// Example:
// Input: [5, 7, 1, 4, 9] 2, less = func(a, b int) bool { return a < b }
// Output: 5
// More information about the algorithm:
// https://brilliant.org/wiki/median-finding-algorithm/
// https://www.youtube.com/watch?v=EzeYI7p9MjU
func iThSmallest[T constraints.Ordered](s []T, i int, less LessFunc[T]) T {
	if i < 0 || i > len(s) {
		panic(fmt.Sprintf("i (%d) is out of bounds (%d)", i, len(s)))
	}

	// Select a pivot by dividing the original slice into subgroups of 5 elements.

	originalLength := len(s)

	// Pre-allocate enough buffer for the medians in case input is large.
	mediansOfSubSlices := make([]T, 0, originalLength/5+originalLength%5)
	for i := 0; i < originalLength; i += 5 {
		// choose either 5 elements, or everything that is remaining
		// if the last slice ends up being smaller than 5 elements.
		sliceOfFive := s[i:Min(i+5, originalLength)]

		// sort the subslice of five. Note, that this is sort
		// is O(1) since only 5 elements.
		sort.Slice(sliceOfFive, func(i, j int) bool {
			return less(sliceOfFive[i], sliceOfFive[j])
		})

		// We append the median of the slice of 5 elements into the median
		// medians of sub slices.
		mediansOfSubSlices = append(mediansOfSubSlices, sliceOfFive[len(sliceOfFive)/2])
	}

	var pivot T
	numberOfMedians := len(mediansOfSubSlices)
	if numberOfMedians <= 5 {
		// Base case of the pivot finding divide and conquer.
		sort.Slice(mediansOfSubSlices, func(i, j int) bool {
			return less(mediansOfSubSlices[i], mediansOfSubSlices[j])
		})
		pivot = mediansOfSubSlices[numberOfMedians/2]
	} else {
		// Use the same algorithm to find the median of the medians of subslices
		// to pivot on for solving the original problem.
		pivot = iThSmallest(mediansOfSubSlices, numberOfMedians/2, less)
	}

	// Note, that partitioning below takes O(n)
	// By the master's theorem the overall result
	// gets amortized to O(n) regardless.

	// partition elements
	smallerThanPivot := make([]T, 0)
	greaterThanPivot := make([]T, 0)
	numEqualToPivot := 0
	for _, cur := range s {
		if cur == pivot {
			numEqualToPivot++
			continue
		}
		if less(cur, pivot) {
			smallerThanPivot = append(smallerThanPivot, cur)
		} else {
			greaterThanPivot = append(greaterThanPivot, cur)
		}
	}
	numberOfElementsSmallerThanPivot := len(smallerThanPivot)

	if i < numberOfElementsSmallerThanPivot {
		return iThSmallest(smallerThanPivot, i, less)
	} else if i > numberOfElementsSmallerThanPivot {
		// Handle edge case where there are duplicate pivots and it is the result.
		if len(greaterThanPivot) == 0 && numEqualToPivot > 1 {
			return pivot
		}

		// Note that if we are searching in the second half, we must re-scale i to
		// not account for the current elements smaller than pivot and the pivot itself.
		return iThSmallest(greaterThanPivot, i-numberOfElementsSmallerThanPivot-numEqualToPivot, less)
	}

	// i == numberOfElementsSmallerThanPivot
	return pivot
}
