package market

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v23/x/market/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, data *types.GenesisState) {
	keeper.SetParams(ctx, data.Params)
	keeper.SetOsmosisPoolDelta(ctx, data.OsmosisPoolDelta)

	// check if the module account exists
	moduleAcc := keeper.GetMarketAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// check if the reserve module account exists
	reserveModuleAcc := keeper.GetReserveMarketAccount(ctx)
	if reserveModuleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ReserveModuleName))
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) (data *types.GenesisState) {
	params := keeper.GetParams(ctx)
	sPoolDelta := keeper.GetOsmosisPoolDelta(ctx)

	return types.NewGenesisState(sPoolDelta, params)
}
