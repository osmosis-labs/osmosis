package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

var (
	KeySigners     = []byte("Signers")
	KeyAssets      = []byte("Assets")
	KeyVotesNeeded = []byte("VotesNeeded")
	KeyFee         = []byte("Fee")
)

func NewParams(
	signers []string,
	assets []Asset,
	votesNeeded uint64,
	fee math.LegacyDec,
) Params {
	return Params{
		Signers:     signers,
		Assets:      assets,
		VotesNeeded: votesNeeded,
		Fee:         fee,
	}
}

// DefaultParams creates default x/bridge params.
func DefaultParams() Params {
	return Params{
		Signers:     []string{},
		Assets:      DefaultAssets(),
		VotesNeeded: DefaultVotesNeeded,
		Fee:         math.LegacyZeroDec(),
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
	if osmoutils.ContainsDuplicate(p.Signers) {
		return errorsmod.Wrapf(ErrInvalidSigners, "Signers are duplicated")
	}

	if len(p.Assets) == 0 {
		return errorsmod.Wrapf(ErrInvalidAssets, "Assets are empty")
	}
	for _, asset := range p.Assets {
		err := asset.Validate()
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidAsset, err.Error())
		}
	}
	// check if p.Assets contains duplicated assets by ID
	assetIDs := Map(p.Assets, func(v Asset) AssetID { return v.Id })
	if osmoutils.ContainsDuplicate(assetIDs) {
		return errorsmod.Wrapf(ErrInvalidAssets, "Assets are duplicated")
	}

	if p.Fee.IsNegative() || p.Fee.GT(math.LegacyOneDec()) {
		return errorsmod.Wrapf(ErrInvalidAsset, "Fee should be between 0 and 1")
	}

	// don't p.VotesNeeded since it's always valid

	return nil
}

func (p Params) GetAsset(id AssetID) (Asset, bool) {
	for i := range p.Assets {
		if p.Assets[i].Id == id {
			return p.Assets[i], true
		}
	}
	return Asset{}, false
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
		paramtypes.NewParamSetPair(KeyVotesNeeded, &p.VotesNeeded, validateVotesNeeded),
		paramtypes.NewParamSetPair(KeyFee, &p.Fee, validateFee),
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

func validateVotesNeeded(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateFee(i interface{}) error {
	fee, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if fee.IsNegative() || fee.GT(math.LegacyOneDec()) {
		return errorsmod.Wrapf(ErrInvalidAsset, "Fee should be between 0 and 1")
	}

	return nil
}
