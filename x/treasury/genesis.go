package treasury

import (
	"fmt"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/treasury/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/treasury/types"
)

// InitGenesis initializes default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, data *types.GenesisState) {
	keeper.SetParams(ctx, data.Params)
	keeper.SetTaxRate(ctx, data.TaxRate)

	// check if the module account exists
	moduleAcc := keeper.GetTreasuryModuleAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// fund reserve
	treasuryCoins := sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseCoinUnit, 10_000_000*appparams.MicroUnit))
	err := keeper.BankKeeper.MintCoins(ctx, types.ModuleName, treasuryCoins) // 10 mil to treasury
	if err != nil {
		panic("could not mint genesis treasury coins")
	}
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) (data *types.GenesisState) {
	params := keeper.GetParams(ctx)

	taxRate := keeper.GetTaxRate(ctx)
	return types.NewGenesisState(params, taxRate)
}
