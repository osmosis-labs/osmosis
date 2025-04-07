package authenticator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/smart-account/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// if there is no NextAuthenticatorId set in genstate, go will set it to 0
	// we need to set it to FirstAuthenticatorId in that case
	if genState.NextAuthenticatorId == 0 {
		genState.NextAuthenticatorId = keeper.FirstAuthenticatorId
	}

	k.SetNextAuthenticatorId(ctx, genState.NextAuthenticatorId)

	for i := range genState.AuthenticatorData {
		accAddr, err := sdk.AccAddressFromBech32(genState.AuthenticatorData[i].Address)
		if err != nil {
			panic(err)
		}

		for j := range genState.AuthenticatorData[i].Authenticators {
			err := k.AddAuthenticatorWithId(
				ctx,
				accAddr,
				genState.AuthenticatorData[i].Authenticators[j].Type,
				genState.AuthenticatorData[i].Authenticators[j].Config,
				genState.AuthenticatorData[i].Authenticators[j].Id,
			)
			if err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.NextAuthenticatorId = k.InitializeOrGetNextAuthenticatorId(ctx)
	allAuthenticators, err := k.GetAllAuthenticatorData(ctx)
	if err != nil {
		panic(err)
	}
	genesis.AuthenticatorData = allAuthenticators

	return genesis
}
