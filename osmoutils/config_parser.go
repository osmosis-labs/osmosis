package osmoutils

import (
	"fmt"
	"strconv"
	"strings"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"

	"github.com/spf13/cast"
)

// ParseBool parses a boolean value from a server type option.
func ParseBool(opts servertypes.AppOptions, groupOptName, optName string, defaultValue bool) bool {
	fullOptName := groupOptName + "." + optName
	valueInterface := opts.Get(fullOptName)
	value := defaultValue
	if valueInterface != nil {
		valueStr, ok := valueInterface.(string)
		if !ok {
			panic("invalidly configured " + fullOptName)
		}
		valueStr = strings.TrimSpace(valueStr)
		v, err := strconv.ParseBool(valueStr)
		if err != nil {
			fmt.Println("error in parsing" + optName + " as bool, setting to false")
			return false
		}
		return v
	}

	return value
}

// ParseInt parses an integer value from a server type option.
func ParseInt(opts servertypes.AppOptions, groupOptName, optName string) int {
	fullOptName := groupOptName + "." + optName
	valueInterface := opts.Get(fullOptName)
	if valueInterface == nil {
		panic("missing config for " + fullOptName)
	}
	value := cast.ToInt(valueInterface)
	return value
}

// ParseUint64 parses a uint64 value from a server type option.
func ParseUint64Slice(opts servertypes.AppOptions, groupOptName, optName string) []uint64 {
	stringSlice := ParseString(opts, groupOptName, optName)

	valueUint64Slice, err := ParseStringToUint64Slice(stringSlice)
	if err != nil {
		panic(fmt.Sprintf("invalidly configured osmosis-sqs.%s, err= %v", optName, err))
	}

	return valueUint64Slice
}

// ParseStringToUint64Slice parses a string to a slice of uint64 values.
func ParseStringToUint64Slice(input string) ([]uint64, error) {
	// Remove "[" and "]" from the input string
	input = strings.Trim(input, "[]")

	// Split the string into individual values
	values := strings.Split(input, ",")

	// Initialize a slice to store uint64 values
	var result []uint64

	if len(values) == 1 && values[0] == "" {
		return result, nil
	}

	// Iterate over the values and convert them to uint64
	for _, v := range values {
		// Parse the string value to uint64
		u, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return nil, err
		}

		// Append the uint64 value to the result slice
		result = append(result, u)
	}

	return result, nil
}

// ParseString parses a string value from a server type option.
func ParseString(opts servertypes.AppOptions, groupOptName, optName string) string {
	fullOptName := groupOptName + "." + optName
	valueInterface := opts.Get(fullOptName)
	if valueInterface == nil {
		panic("missing config for " + fullOptName)
	}
	value := cast.ToString(valueInterface)
	return value
}
