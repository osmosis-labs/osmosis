package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewParams(signers []string, assets []Asset) Params {
	return Params{
		Signers: signers,
		Assets:  assets,
	}
}

// DefaultParams creates default x/bridge params.
func DefaultParams() Params {
	return Params{
		Signers: []string{}, // TODO: what to use as the default?
		Assets:  DefaultAssets(),
	}
}

// Validate x/bridge params.
func (p Params) Validate() error {
	for _, signer := range p.Signers {
		_, err := sdk.AccAddressFromBech32(signer)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid signer address (%s)", err)
		}
	}

	for _, asset := range p.Assets {
		err := asset.Validate()
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
		}
	}

	return nil
}
