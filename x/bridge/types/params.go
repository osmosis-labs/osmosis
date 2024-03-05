package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeySigners = []byte("Signers")
	KeyAssets  = []byte("Assets")
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

// ParamKeyTable for the x/bridge module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeySigners, &p.Signers, validateSigners),
		paramtypes.NewParamSetPair(KeyAssets, &p.Assets, validateAssets),
	}
}

func validateSigners(i interface{}) error {
	signers, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, signer := range signers {
		_, err := sdk.AccAddressFromBech32(signer)
		if err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid signer address (%s)", err)
		}
	}

	return nil
}

func validateAssets(i interface{}) error {
	assets, ok := i.([]Asset)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, asset := range assets {
		err := asset.Validate()
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
		}
	}

	return nil
}
