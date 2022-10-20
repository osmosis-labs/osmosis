package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

func (gs GenesisState) Validate() error {
	return nil
}

func InitGenesis(ctx sdk.Context, genState GenesisState) {

}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context) *GenesisState {
	return &GenesisState{}
}
