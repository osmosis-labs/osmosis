package osmoutils

import "sort"

// Max returns the maximum value among the given values of any type that supports comparison.
func Max(values ...interface{}) interface{} {
	if len(values) == 0 {
		return nil
	}

	max := values[0]
	for _, value := range values[1:] {
		switch value := value.(type) {
		case int:
			if intValue, ok := max.(int); ok && value > intValue {
				max = value
			}
		case int8:
			if int8Value, ok := max.(int8); ok && value > int8Value {
				max = value
			}
		case int16:
			if int16Value, ok := max.(int16); ok && value > int16Value {
				max = value
			}
		case int32:
			if int32Value, ok := max.(int32); ok && value > int32Value {
				max = value
			}
		case int64:
			if int64Value, ok := max.(int64); ok && value > int64Value {
				max = value
			}
		case uint:
			if uintValue, ok := max.(uint); ok && value > uintValue {
				max = value
			}
		case uint8:
			if uint8Value, ok := max.(uint8); ok && value > uint8Value {
				max = value
			}
		case uint16:
			if uint16Value, ok := max.(uint16); ok && value > uint16Value {
				max = value
			}
		case uint32:
			if uint32Value, ok := max.(uint32); ok && value > uint32Value {
				max = value
			}
		case uint64:
			if uint64Value, ok := max.(uint64); ok && value > uint64Value {
				max = value
			}
		case uintptr:
			if uintptrValue, ok := max.(uintptr); ok && value > uintptrValue {
				max = value
			}
		}
	}
	return max
}

// DifferenceBetweenUint64Arrays takes two slices of uint64, 'a' and 'b', as input.
// It returns a new slice containing the elements that are unique to either 'a' or 'b'.
// The function uses two maps for efficient lookup of elements.
//
// Example:
// a := []uint64{1, 2, 3, 4, 5}
// b := []uint64{4, 5, 6, 7, 8}
// result := DisjointArrays(a, b)
// result will be []uint64{1, 2, 3, 6, 7, 8}
//
// Note: This function returns the difference between the two arrays in ascending order,
// and does not preserve the order of the elements in the original arrays.
func DisjointArrays(a, b []uint64) []uint64 {
	if len(a) == 0 && len(b) == 0 {
		return []uint64{}
	}

	m1 := make(map[uint64]bool)
	m2 := make(map[uint64]bool)

	for _, item := range a {
		m1[item] = true
	}

	for _, item := range b {
		m2[item] = true
	}

	var result []uint64
	for item := range m1 {
		if !m2[item] {
			result = append(result, item)
		}
	}

	for item := range m2 {
		if !m1[item] {
			result = append(result, item)
		}
	}

	if len(result) == 0 {
		return []uint64{}
	}

	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })

	return result
}
