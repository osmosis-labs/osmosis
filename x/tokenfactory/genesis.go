package tokenfactory

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/x/tokenfactory/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, genDenom := range genState.GetFactoryDenoms() {
		creator, nonce, err := types.DeconstructDenom(genDenom.GetDenom())
		if err != nil {
			panic(err.Error())
		}
		k.CreateDenom(ctx, creator, nonce)
		k.SetAuthorityMetadata(ctx, genDenom.GetDenom(), genDenom.GetAuthorityMetadata())
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genDenoms := []types.GenesisDenom{}
	iterator := k.GetAllDenomsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Value())

		authorityMetadata, err := k.GetAuthorityMetadata(ctx, denom)
		if err != nil {
			panic(err.Error())
		}

		genDenoms = append(genDenoms, types.GenesisDenom{
			Denom:             denom,
			AuthorityMetadata: authorityMetadata,
		})
	}

	return &types.GenesisState{
		FactoryDenoms: genDenoms,
	}
}
