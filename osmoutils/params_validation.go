package osmoutils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateAddressList validates a slice of addresses.
//
// Parameters:
// - i: The parameter to validate.
//
// Returns:
// - An error if any of the strings are not addresses
func ValidateAddressList(i interface{}) error {
	whitelist, ok := i.([]string)

	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, a := range whitelist {
		if _, err := sdk.AccAddressFromBech32(a); err != nil {
			return fmt.Errorf("invalid address")
		}
	}

	return nil
}
