package types

import (
	"fmt"
)

func ValidateEpochIdentifierInterface(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	err := ValidateEpochIdentifierString(v)
	if err != nil {
		return fmt.Errorf("Invalid epoch identifier: %s", err)
	}
	return nil
}

func ValidateEpochIdentifierString(s string) error {
	if s == "" {
		return fmt.Errorf("empty distribution epoch identifier: %+v", s)
	}
	return nil
}
