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
