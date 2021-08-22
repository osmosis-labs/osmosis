package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

func ValidateHRP(hrp string) error {
	if hrp == "" {
		return sdkerrors.Wrap(ErrInvalidHRP, "empty HRP")
	}
	return nil
}
