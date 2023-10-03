package osmoutils

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
// It returns a new slice containing the elements that are in 'a' but not in 'b'.
// The function uses a map for efficient lookup of elements.
//
// Example:
// a := []uint64{1, 2, 3, 4, 5}
// b := []uint64{4, 5, 6, 7, 8}
// result := DifferenceBetweenUint64Arrays(a, b)
// result will be []uint64{1, 2, 3}
//
// Note: This function does not preserve the order of the elements.
func DifferenceBetweenUint64Arrays(a, b []uint64) []uint64 {
	m := make(map[uint64]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		delete(m, item)
	}

	var result []uint64
	for item := range m {
		result = append(result, item)
	}

	return result
}
