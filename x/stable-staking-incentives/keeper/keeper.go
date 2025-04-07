package keeper

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/osmoutils/cosmwasm"

	"cosmossdk.io/log"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	storetypes "cosmossdk.io/store/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey

	paramSpace paramtypes.Subspace

	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	contractKeeper cosmwasm.ContractKeeper
}

func NewKeeper(storeKey storetypes.StoreKey, paramSpace paramtypes.Subspace, accountKeeper types.AccountKeeper, bankKeeper types.BankKeeper) Keeper {
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
		bankKeeper:    bankKeeper,
	}
}

func (k *Keeper) SetContractKeeper(contractKeeper cosmwasm.ContractKeeper) {
	k.contractKeeper = contractKeeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// AllocateAsset allocates and distributes coin according a gauge’s proportional weight that is recorded in the record.
func (k Keeper) AllocateAsset(ctx sdk.Context) error {
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	asset := k.bankKeeper.GetBalance(ctx, moduleAddr, appparams.BaseCoinUnit)
	if asset.Amount.IsZero() {
		// when allocating asset is zero, skip execution
		return nil
	}

	cosmwasmContractAddress, err := sdk.AccAddressFromBech32(k.GetParams(ctx).DistributionContractAddress)
	if err != nil {
		ctx.Logger().Error("AllocateAsset failed to parse contract address",
			"module", types.ModuleName,
			"height", ctx.BlockHeight(),
			"raw_address", k.GetParams(ctx).DistributionContractAddress,
			"error", err.Error())
		return nil
	}

	ctx.Logger().Info(
		"AllocateAsset minted amount",
		"module", types.ModuleName,
		"totalMintedAmount", asset.Amount,
		"height", ctx.BlockHeight(),
		"destination", cosmwasmContractAddress.String(),
	)

	coins := sdk.NewCoins(asset)
	distributeMsg := map[string]interface{}{"distribute_rewards": struct{}{}}
	_, err = cosmwasm.Execute[any, any](ctx, k.contractKeeper, cosmwasmContractAddress.String(), moduleAddr, coins, distributeMsg)
	if err != nil {
		ctx.Logger().Error("AllocateAsset failed to distribute rewards",
			"module", types.ModuleName,
			"height", ctx.BlockHeight(),
			"error", err.Error())
	}
	return nil
}
