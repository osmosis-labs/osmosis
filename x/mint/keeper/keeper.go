package keeper

import (
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v7/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	errAmountCannotBeNilOrZero               = errors.New("amount cannot be nil or zero")
	errDevVestingModuleAccountAlreadyCreated = fmt.Errorf("%s module account already exists", types.DeveloperVestingModuleAcctName)
	errDevVestingModuleAccountNotCreated     = fmt.Errorf("%s module account does not exist", types.DeveloperVestingModuleAcctName)
)

// Keeper of the mint store.
type Keeper struct {
	cdc              codec.BinaryCodec
	storeKey         sdk.StoreKey
	paramSpace       paramtypes.Subspace
	accountKeeper    types.AccountKeeper
	bankKeeper       types.BankKeeper
	distrKeeper      types.DistrKeeper
	epochKeeper      types.EpochKeeper
	hooks            types.MintHooks
	feeCollectorName string
}

// NewKeeper creates a new mint Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak types.AccountKeeper, bk types.BankKeeper, dk types.DistrKeeper, epochKeeper types.EpochKeeper,
	feeCollectorName string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         key,
		paramSpace:       paramSpace,
		accountKeeper:    ak,
		bankKeeper:       bk,
		distrKeeper:      dk,
		epochKeeper:      epochKeeper,
		feeCollectorName: feeCollectorName,
	}
}

// SetInitialSupplyOffsetDuringMigration sets the supply offset based on the balance of the
// developer vesting module account. CreateDeveloperVestingModuleAccount must be called
// prior to calling this method. That is, developer vesting module account must exist when
// SetInitialSupplyOffsetDuringMigration is called. Also, SetInitialSupplyOffsetDuringMigration
// should only be called one time during the initial migration to v7. This is done so because
// we would like to ensure that unvested developer tokens are not returned as part of the supply
// queries. The method returns an error if current height in ctx is greater than the v7 upgrade height.
func (k Keeper) SetInitialSupplyOffsetDuringMigration(ctx sdk.Context) error {
	if !k.accountKeeper.HasAccount(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)) {
		return errDevVestingModuleAccountNotCreated
	}

	moduleAccBalance := k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), k.GetParams(ctx).MintDenom)
	k.bankKeeper.AddSupplyOffset(ctx, moduleAccBalance.Denom, moduleAccBalance.Amount.Neg())
	return nil
}

// CreateDeveloperVestingModuleAccount creates the developer vesting module account
// and mints amount of tokens to it.
// Should only be called during the initial genesis creation, never again. Returns nil on success.
// Returns error in the following cases:
// - amount is nil or zero.
// - if ctx has block height greater than 0.
// - developer vesting module account is already created prior to calling this method.
func (k Keeper) CreateDeveloperVestingModuleAccount(ctx sdk.Context, amount sdk.Coin) error {
	if amount.IsNil() || amount.Amount.IsZero() {
		return errAmountCannotBeNilOrZero
	}
	if k.accountKeeper.HasAccount(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)) {
		return errDevVestingModuleAccountAlreadyCreated
	}

	moduleAcc := authtypes.NewEmptyModuleAccount(
		types.DeveloperVestingModuleAcctName, authtypes.Minter)
	k.accountKeeper.SetModuleAccount(ctx, moduleAcc)

	err := k.bankKeeper.MintCoins(ctx, types.DeveloperVestingModuleAcctName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}
	return nil
}

// _____________________________________________________________________

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Set the mint hooks.
func (k *Keeper) SetHooks(h types.MintHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set mint hooks twice")
	}

	k.hooks = h

	return k
}

// GetLastReductionEpochNum returns last reduction epoch number.
func (k Keeper) GetLastReductionEpochNum(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.LastReductionEpochKey)
	if b == nil {
		return 0
	}

	return int64(sdk.BigEndianToUint64(b))
}

// SetLastReductionEpochNum set last reduction epoch number.
func (k Keeper) SetLastReductionEpochNum(ctx sdk.Context, epochNum int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastReductionEpochKey, sdk.Uint64ToBigEndian(uint64(epochNum)))
}

// get the minter.
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshal(b, &minter)
	return
}

// set the minter.
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&minter)
	store.Set(types.MinterKey, b)
}

// _____________________________________________________________________

// GetParams returns the total set of minting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of minting parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// _____________________________________________________________________

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// GetProportions gets the balance of the `MintedDenom` from minted coins and returns coins according to the `AllocationRatio`.
func (k Keeper) GetProportions(ctx sdk.Context, mintedCoin sdk.Coin, ratio sdk.Dec) (sdk.Coin, error) {
	if ratio.GT(sdk.OneDec()) {
		return sdk.Coin{}, fmt.Errorf("mint allocation ratio %s is greater than 1", ratio)
	}
	return sdk.NewCoin(mintedCoin.Denom, mintedCoin.Amount.ToDec().Mul(ratio).TruncateInt()), nil
}

// DistributeMintedCoins implements distribution of minted coins from mint to external modules.
func (k Keeper) DistributeMintedCoin(ctx sdk.Context, mintedCoin sdk.Coin) error {
	params := k.GetParams(ctx)
	proportions := params.DistributionProportions

	// allocate staking incentives into fee collector account to be moved to on next begin blocker by staking module
	stakingIncentivesCoin, err := k.distributeToModule(ctx, k.feeCollectorName, mintedCoin, proportions.Staking)
	if err != nil {
		return err
	}

	// allocate pool allocation ratio to pool-incentives module account account
	poolIncentivesCoin, err := k.distributeToModule(ctx, poolincentivestypes.ModuleName, mintedCoin, proportions.PoolIncentives)
	if err != nil {
		return err
	}

	devRewardCoin, err := k.GetProportions(ctx, mintedCoin, proportions.DeveloperRewards)
	if err != nil {
		return err
	}
	devRewardCoins := sdk.NewCoins(devRewardCoin)
	// This is supposed to come from the developer vesting module address, not the mint module address
	// we over-allocated to the mint module address earlier though, so we burn it right here.
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, devRewardCoins); err != nil {
		return err
	}

	// Take the current balance of the developer rewards pool and remove it from the supply offset
	// We re-introduce the new supply at the end, in order to avoid any rounding discrepancies.
	developerAccountBalance := k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), mintedCoin.Denom)
	k.bankKeeper.AddSupplyOffset(ctx, mintedCoin.Denom, developerAccountBalance.Amount)

	if len(params.WeightedDeveloperRewardsReceivers) == 0 {
		// fund community pool when rewards address is empty
		if err := k.distrKeeper.FundCommunityPool(ctx, devRewardCoins, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)); err != nil {
			return err
		}
	} else {
		// allocate developer rewards to addresses by weight
		for _, w := range params.WeightedDeveloperRewardsReceivers {
			devPortionCoin, err := k.GetProportions(ctx, devRewardCoin, w.Weight)
			if err != nil {
				return err
			}
			devRewardPortionCoins := sdk.NewCoins(devPortionCoin)
			if w.Address == "" {
				err := k.distrKeeper.FundCommunityPool(ctx, devRewardPortionCoins,
					k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName))
				if err != nil {
					return err
				}
			} else {
				devRewardsAddr, err := sdk.AccAddressFromBech32(w.Address)
				if err != nil {
					return err
				}
				// If recipient is vesting account, pay to account according to its vesting condition
				err = k.bankKeeper.SendCoinsFromModuleToAccount(
					ctx, types.DeveloperVestingModuleAcctName, devRewardsAddr, devRewardPortionCoins)
				if err != nil {
					return err
				}
			}
		}
	}

	// Take the new balance of the developer rewards pool and add it back to the supply offset deduction
	developerAccountBalance = k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), mintedCoin.Denom)
	k.bankKeeper.AddSupplyOffset(ctx, mintedCoin.Denom, developerAccountBalance.Amount.Neg())

	// subtract from original provision to ensure no coins left over after the allocations
	communityPoolCoin := mintedCoin.Sub(stakingIncentivesCoin).Sub(poolIncentivesCoin).Sub(devRewardCoin)
	err = k.distrKeeper.FundCommunityPool(ctx, sdk.NewCoins(communityPoolCoin), k.accountKeeper.GetModuleAddress(types.ModuleName))
	if err != nil {
		return err
	}

	// call an hook after the minting and distribution of new coins
	k.hooks.AfterDistributeMintedCoin(ctx, mintedCoin)

	return err
}

func (k Keeper) distributeToModule(ctx sdk.Context, recipientModule string, mintedCoin sdk.Coin, proportion sdk.Dec) (sdk.Coin, error) {
	distributionCoin, err := k.GetProportions(ctx, mintedCoin, proportion)
	if err != nil {
		return sdk.Coin{}, err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, recipientModule, sdk.NewCoins(distributionCoin)); err != nil {
		return sdk.Coin{}, err
	}
	return distributionCoin, nil
}
