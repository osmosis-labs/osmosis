package types

import (
	fmt "fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// This code was copied from bech32 library
// https://github.com/enigmampc/btcutil/blob/master/bech32/bech32.go#L26
func ValidateHrp(hrp string) error {
	if hrp == "" {
		return sdkerrors.Wrap(ErrInvalidHRP, "empty HRP")
	}

	// Only	ASCII characters between 33 and 126 are allowed.
	for i := 0; i < len(hrp); i++ {
		if hrp[i] < 33 || hrp[i] > 126 {
			return fmt.Errorf("invalid character in "+
				"string: '%c'", hrp[i])
		}
	}

	// The characters must be all lowercase
	lower := strings.ToLower(hrp)
	if hrp != lower {
		return fmt.Errorf("string not all lowercase")
	}

	return nil
}
