package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"

	"github.com/osmosis-labs/osmosis/v26/x/stable-staking-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	storetypes "cosmossdk.io/store/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	paramSpace paramtypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	wasmKeeper    types.ContractKeeper
}

func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, accountKeeper types.AccountKeeper) Keeper {
	// ensure pool-incentives module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		accountKeeper: accountKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) ExecuteWasmContract(ctx sdk.Context, contractAddr sdk.AccAddress, msg interface{}, funds sdk.Coins) error {
	msgBz, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = k.wasmKeeper.Execute(ctx, contractAddr, k.accountKeeper.GetModuleAddress(types.ModuleName), msgBz, funds)
	return err
}

// AllocateAsset allocates and distributes coin according a gauge’s proportional weight that is recorded in the record.
func (k Keeper) AllocateAsset(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	asset := k.bankKeeper.GetBalance(ctx, moduleAddr, params.MintedDenom)
	if asset.Amount.IsZero() {
		// when allocating asset is zero, skip execution
		return nil
	}

	ctx.Logger().Info("AllocateAsset minted amount", "module", types.ModuleName, "totalMintedAmount", asset.Amount, "height", ctx.BlockHeight())

	coins := sdk.NewCoins(asset)
	distributeMsg := map[string]interface{}{"distribute": struct{}{}}
	return k.ExecuteWasmContract(ctx, moduleAddr, distributeMsg, coins)
}
