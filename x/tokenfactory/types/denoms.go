package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

const (
	ModuleDenomPrefix = "factory"
)

// GetTokenDenom constructs a denom string for tokens created by tokenfactory
// based on an input creator address and a nonce
// The denom constructed is factory/{creator}/{nonce}
func GetTokenDenom(creator, nonce string) (string, error) {
	if strings.Contains(creator, "/") {
		return "", ErrInvalidCreator
	}
	denom := strings.Join([]string{ModuleDenomPrefix, creator, nonce}, "/")
	return denom, sdk.ValidateDenom(denom)
}

// DeconstructDenom takes a token denom string and verifies that it is a valid
// denom of the tokenfactory module, and is of the form `factory/{creator}/{nonce}`
// If valid, it returns the creator address and nonce
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

	// Handle the case where a denom has a slash in its nonce. For example,
	// when we did the split, we'd turn factory/sunnyaddr/atomderivative/sikka into ["factory", "sunnyaddr", "atomderivative", "sikka"]
	// So we have to join [2:] with a "/" as the delimiter to get back the correct nonce which should be "atomderivative/sikka"
	nonce = strings.Join(strParts[2:], "/")

	return creator, nonce, nil
}

// NewTokenFactoryDenomMintCoinsRestriction creates and returns a BankMintingRestrictionFn that only allows minting of
// valid tokenfactory denoms
func NewTokenFactoryDenomMintCoinsRestriction() bankkeeper.BankMintingRestrictionFn {
	return func(ctx sdk.Context, coinsToMint sdk.Coins) error {
		for _, coin := range coinsToMint {
			_, _, err := DeconstructDenom(coin.Denom)
			if err != nil {
				return fmt.Errorf("does not have permission to mint %s", coin.Denom)
			}
		}
		return nil
	}
}
