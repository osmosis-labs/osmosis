package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	ModuleDenomPrefix = "factory"
)

func GetTokenDenom(creator, nonce string) (string, error) {
	denom := strings.Join([]string{"factory", creator, nonce}, "/")
	return denom, sdk.ValidateDenom(denom)
}

func DeconstructDenom(denom string) (creator string, nonce string, err error) {
	err = sdk.ValidateDenom(denom)
	if err != nil {
		return "", "", err
	}

	strParts := strings.Split(denom, "/")
	if len(strParts) < 3 {
		return "", "", sdkerrors.Wrapf(ErrInvalidDenom, "not enough parts of denom %s", denom)
	}

	if strParts[0] != ModuleDenomPrefix {
		return "", "", sdkerrors.Wrapf(ErrInvalidDenom, "denom prefix is incorrect. Is: %s.  Should be: %s", strParts[0], ModuleDenomPrefix)
	}

	creator = strParts[1]
	_, err = sdk.AccAddressFromBech32(creator)
	if err != nil {
		return "", "", sdkerrors.Wrapf(ErrInvalidDenom, "Invalid creator address (%s)", err)
	}

	nonce = strings.Join(strParts[2:], "/")

	return creator, nonce, nil
}
