package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// DefaultGenesis returns the default txfee genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Basedenom: sdk.DefaultBondDenom,
		Feetokens: []FeeToken{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure. It does not verify that the corresponding pool IDs actually exist.
// This is done in InitGenesis.
func (gs GenesisState) Validate() error {
	err := sdk.ValidateDenom(gs.Basedenom)
	if err != nil {
		return err
	}

	for _, feeToken := range gs.Feetokens {
		err := sdk.ValidateDenom(feeToken.Denom)
		if err != nil {
			return err
		}
	}

	return nil
}
