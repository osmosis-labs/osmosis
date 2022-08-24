package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
	"github.com/osmosis-labs/osmosis/v11/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v11/x/pool-incentives/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper of the mint store.
type Keeper struct {
	cdc                 codec.BinaryCodec
	storeKey            sdk.StoreKey
	paramSpace          paramtypes.Subspace
	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	communityPoolKeeper types.CommunityPoolKeeper
	epochKeeper         types.EpochKeeper
	hooks               types.MintHooks
	feeCollectorName    string
}

type invalidRatioError struct {
	ActualRatio sdk.Dec
}

func (e invalidRatioError) Error() string {
	return fmt.Sprintf("mint allocation ratio (%s) is greater than 1", e.ActualRatio)
}

type insufficientDevVestingBalanceError struct {
	ActualBalance         sdk.Int
	AttemptedDistribution sdk.Int
}

func (e insufficientDevVestingBalanceError) Error() string {
	return fmt.Sprintf("developer vesting balance (%s) is smaller than requested distribution of (%s)", e.ActualBalance, e.AttemptedDistribution)
}

const emptyWeightedAddressReceiver = ""

// NewKeeper creates a new mint Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak types.AccountKeeper, bk types.BankKeeper, ck types.CommunityPoolKeeper, epochKeeper types.EpochKeeper,
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
		cdc:                 cdc,
		storeKey:            key,
		paramSpace:          paramSpace,
		accountKeeper:       ak,
		bankKeeper:          bk,
		communityPoolKeeper: ck,
		epochKeeper:         epochKeeper,
		feeCollectorName:    feeCollectorName,
	}
}

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

// GetMinter gets the minter.
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	osmoutils.MustGet(ctx.KVStore(k.storeKey), types.MinterKey, &minter)
	return
}

// SetMinter sets the minter.
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	osmoutils.MustSet(ctx.KVStore(k.storeKey), types.MinterKey, &minter)
}

// GetParams returns the total set of minting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the total set of minting parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetInflationTruncationDelta returns the truncation delta.
// Panics if key is invalid.
func (k Keeper) GetTruncationDelta(ctx sdk.Context, moduleAccountName string) (sdk.Dec, error) {
	storeKey, err := getTruncationStoreKeyFromModuleAccount(moduleAccountName)
	if err != nil {
		return sdk.Dec{}, err
	}
	return osmoutils.MustGetDec(ctx.KVStore(k.storeKey), storeKey), nil
}

// SetTruncationDelta sets the truncation delta, returning an error if any. nil otherwise.
func (k Keeper) SetTruncationDelta(ctx sdk.Context, moduleAccountName string, truncationDelta sdk.Dec) error {
	storeKey, err := getTruncationStoreKeyFromModuleAccount(moduleAccountName)
	if err != nil {
		return err
	}

	if truncationDelta.LT(sdk.ZeroDec()) {
		return sdkerrors.Wrapf(types.ErrInvalidAmount, "truncation delta must be positive, was (%s)", truncationDelta)
	}
	osmoutils.MustSetDec(ctx.KVStore(k.storeKey), storeKey, truncationDelta)
	return nil
}

// DistributeMintedCoin implements distribution of a minted coin from mint to external modules.
func (k Keeper) DistributeMintedCoin(ctx sdk.Context, mintedCoin sdk.Coin) error {
	params := k.GetParams(ctx)
	proportions := params.DistributionProportions

	// allocate staking incentives into fee collector account to be moved to on next begin blocker by staking module account.
	stakingIncentivesAmount, err := k.distributeToModule(ctx, k.feeCollectorName, mintedCoin, proportions.Staking)
	if err != nil {
		return err
	}

	// allocate pool allocation ratio to pool-incentives module account.
	poolIncentivesAmount, err := k.distributeToModule(ctx, poolincentivestypes.ModuleName, mintedCoin, proportions.PoolIncentives)
	if err != nil {
		return err
	}

	// allocate dev rewards to respective accounts from developer vesting module account.
	devRewardAmount, err := k.distributeDeveloperRewards(ctx, mintedCoin, proportions.DeveloperRewards, params.WeightedDeveloperRewardsReceivers)
	if err != nil {
		return err
	}

	// subtract from original provision to ensure no coins left over after the allocations
	communityPoolAmount := mintedCoin.Amount.Sub(stakingIncentivesAmount).Sub(poolIncentivesAmount).Sub(devRewardAmount)
	err = k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(params.MintDenom, communityPoolAmount)), k.accountKeeper.GetModuleAddress(types.ModuleName))
	if err != nil {
		return err
	}

	// call an hook after the minting and distribution of new coins
	k.hooks.AfterDistributeMintedCoin(ctx)

	return err
}

// getLastReductionEpochNum returns last reduction epoch number.
func (k Keeper) getLastReductionEpochNum(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.LastReductionEpochKey)
	if b == nil {
		return 0
	}

	return int64(sdk.BigEndianToUint64(b))
}

// setLastReductionEpochNum set last reduction epoch number.
func (k Keeper) setLastReductionEpochNum(ctx sdk.Context, epochNum int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastReductionEpochKey, sdk.Uint64ToBigEndian(uint64(epochNum)))
}

// mintCoins implements an alias call to the underlying bank keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) mintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// distributeToModule distributes mintedCoin multiplied by proportion to the recepientModule account.
func (k Keeper) distributeToModule(ctx sdk.Context, recipientModule string, mintedCoin sdk.Coin, proportion sdk.Dec) (sdk.Int, error) {
	distributionCoin, err := getProportions(mintedCoin, proportion)
	if err != nil {
		return sdk.Int{}, err
	}
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, recipientModule, sdk.NewCoins(distributionCoin)); err != nil {
		return sdk.Int{}, err
	}
	return distributionCoin.Amount, nil
}

// distributeDeveloperRewards distributes developer rewards from developer vesting module account
// to the respective account receivers by weight (developerRewardsReceivers).
// If no developer reward receivers given, funds the community pool instead.
// Returns the total amount distributed from the developer vesting module account.
// Updates supply offsets to reflect the amount of coins distributed. This is done so because the developer rewards distributions are
// allocated from its own module account, not the mint module accont (TODO: next step in https://github.com/osmosis-labs/osmosis/issues/1916).
// Returns nil on success, error otherwise.
// With respect to input parameters, errors occur when:
// - developerRewardsProportion is greater than 1.
// - invalid address in developer rewards receivers.
// - the balance of developer module account is less than totalMintedCoin * developerRewardsProportion.
// - the balance of mint module is less than totalMintedCoin * developerRewardsProportion.
// CONTRACT:
// - weights in developerRewardsReceivers add up to 1.
// - addresses in developerRewardsReceivers are valid or empty string.
func (k Keeper) distributeDeveloperRewards(ctx sdk.Context, totalMintedCoin sdk.Coin, developerRewardsProportion sdk.Dec, developerRewardsReceivers []types.WeightedAddress) (sdk.Int, error) {
	devRewardCoin, err := getProportions(totalMintedCoin, developerRewardsProportion)
	if err != nil {
		return sdk.Int{}, err
	}

	developerRewardsModuleAccountAddress := k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	developerAccountBalance := k.bankKeeper.GetBalance(ctx, developerRewardsModuleAccountAddress, totalMintedCoin.Denom)
	if developerAccountBalance.Amount.LT(devRewardCoin.Amount) {
		return sdk.Int{}, insufficientDevVestingBalanceError{ActualBalance: developerAccountBalance.Amount, AttemptedDistribution: devRewardCoin.Amount}
	}

	devRewardCoins := sdk.NewCoins(devRewardCoin)
	// TODO: https://github.com/osmosis-labs/osmosis/issues/2025
	// Avoid over-allocating from the mint module address and have to later burn it here:
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, devRewardCoins); err != nil {
		return sdk.Int{}, err
	}

	// Take the current balance of the developer rewards pool and remove it from the supply offset
	// We re-introduce the new supply at the end, in order to avoid any rounding discrepancies.
	k.bankKeeper.AddSupplyOffset(ctx, totalMintedCoin.Denom, developerAccountBalance.Amount)

	// If no developer rewards receivers provided, fund the community pool from
	// the developer vesting module account.
	if len(developerRewardsReceivers) == 0 {
		err = k.communityPoolKeeper.FundCommunityPool(ctx, devRewardCoins, developerRewardsModuleAccountAddress)
		if err != nil {
			return sdk.Int{}, err
		}
	} else {
		// allocate developer rewards to addresses by weight
		for _, w := range developerRewardsReceivers {
			devPortionCoin, err := getProportions(devRewardCoin, w.Weight)
			if err != nil {
				return sdk.Int{}, err
			}
			devRewardPortionCoins := sdk.NewCoins(devPortionCoin)
			// fund community pool when rewards address is empty.
			if w.Address == emptyWeightedAddressReceiver {
				err := k.communityPoolKeeper.FundCommunityPool(ctx, devRewardPortionCoins,
					k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName))
				if err != nil {
					return sdk.Int{}, err
				}
			} else {
				devRewardsAddr, err := sdk.AccAddressFromBech32(w.Address)
				if err != nil {
					return sdk.Int{}, err
				}
				// If recipient is vesting account, pay to account according to its vesting condition
				err = k.bankKeeper.SendCoinsFromModuleToAccount(
					ctx, types.DeveloperVestingModuleAcctName, devRewardsAddr, devRewardPortionCoins)
				if err != nil {
					return sdk.Int{}, err
				}
			}
		}
	}

	// Take the new balance of the developer rewards pool and add it back to the supply offset deduction
	developerAccountBalance = k.bankKeeper.GetBalance(ctx, developerRewardsModuleAccountAddress, totalMintedCoin.Denom)
	k.bankKeeper.AddSupplyOffset(ctx, totalMintedCoin.Denom, developerAccountBalance.Amount.Neg())

	return devRewardCoin.Amount, nil
}

// calculateTotalTruncationDelta returns the total truncation delta that has not been distributed yet
// for the given key given provisions and amount distributed. Both of the given values are epoch specific.
// The returned delta might include the delta from previous epochs if they have not been distributed due to truncations.
// For example, assume that for some number of epochs our expected provisions
// are 100.6 and the actual amount distributed is 100 every epoch due to truncations.
// Then, at epoch 1, we have a delta of 0.6. 0.6 cannot be distributed because it is not an integer.
// So we persist it until the next epoch. Then, at at epoch 2, we get a delta of 1.2 (0.6 from epoch 1 and 0.6 from epoch 2).
// Now, 1 can be distributed and 0.2 gets persisted until the next epoch.
// CONTRACT: provisions are greater than or equal to amountDistributed.
func (k Keeper) calculateTotalTruncationDelta(ctx sdk.Context, modulAccountName string, provisions sdk.Dec, amountDistributed sdk.Int) (sdk.Dec, error) {
	truncationDelta, err := k.GetTruncationDelta(ctx, modulAccountName)
	if err != nil {
		return sdk.Dec{}, err
	}
	currentEpochRewardsDelta := provisions.Sub(amountDistributed.ToDec())
	return truncationDelta.Add(currentEpochRewardsDelta), err
}

// createDeveloperVestingModuleAccount creates the developer vesting module account
// and mints amount of tokens to it.
// Should only be called during the initial genesis creation, never again. Returns nil on success.
// Returns error in the following cases:
// - amount is nil or zero.
// - if ctx has block height greater than 0.
// - developer vesting module account is already created prior to calling this method.
func (k Keeper) createDeveloperVestingModuleAccount(ctx sdk.Context, amount sdk.Coin) error {
	if amount.IsNil() || amount.Amount.IsZero() {
		return sdkerrors.Wrap(types.ErrInvalidAmount, "amount cannot be nil or zero")
	}
	if k.accountKeeper.HasAccount(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)) {
		return sdkerrors.Wrapf(types.ErrModuleAccountAlreadyExist, "%s vesting module account already exist", types.DeveloperVestingModuleAcctName)
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

// handleTruncationDelta estimates and distributes truncation delta from either mint module account or
// developer vesting module account. Returns the total amount distributed from truncations.
// If truncations are estimated to be less than one, persists them in store until the next epoch without
// any distributions during the current epoch.
// More on why this handling truncations is necessary: due to limitations of some SDK interfaces that operate on integers,
// there are known truncation differences from the expected total epoch mint provisions.
// To use these interfaces, we always round down to the nearest integer by truncating decimals.
// As a result, it is possible to undermint. To mitigate that, we distribute any delta to the community pool.
// The delta is calculated by subtracting the actual distributions from the given expected total distributions
// and adding it to any left overs from the previous epoch. The left overs might be stemming from the inability
// to distribute decimal truncations less than 1. As a result, we store them in the store until the next epoch.
// These truncation distributions have eventual guarantees. That is, they are guaranteed to be distributed
// eventually but not necessarily during the same epoch.
// Returns error if module account name is other than mint or developer vesting is given.
// The truncation delta is calculated by subtracting amountDistributed from probisions and adding to
// any leftover truncations from the previous epoch.
// Therefore, provisions must be greater than or equal to the amount distributed. Errors if not.
// For any amount to be distributed from the mint module account, it mints the estimated truncation amount
// before distributing it to the community pool.
// Additionally, it errors in the following cases:
// - unable to mint tokens
// - unable to fund the community pool
func (k Keeper) handleTruncationDelta(ctx sdk.Context, moduleAccountName string, provisions sdk.DecCoin, amountDistributed sdk.Int) (sdk.Int, error) {
	if moduleAccountName != types.DeveloperVestingModuleAcctName && moduleAccountName != types.ModuleName {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidModuleAccountGiven, "truncation delta can only be handled by (%s) or (%s) module accounts but (%s) was given", types.DeveloperVestingModuleAcctName, types.ModuleName, moduleAccountName)
	}
	if provisions.Amount.LT(amountDistributed.ToDec()) {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidAmount, "provisions (%s) must be greater than or equal to amount disributed (%s)", provisions.Amount, amountDistributed)
	}

	deltaAmount, err := k.calculateTotalTruncationDelta(ctx, moduleAccountName, provisions.Amount, amountDistributed)
	if err != nil {
		return sdk.Int{}, err
	}
	if deltaAmount.LT(sdk.OneDec()) {
		if err := k.SetTruncationDelta(ctx, moduleAccountName, deltaAmount); err != nil {
			return sdk.Int{}, err
		}
		return sdk.ZeroInt(), nil
	}

	// N.B: Truncation is acceptable because we check delta at the end of every epoch.
	// As a result, actual minted distributions always approach the expected value.
	truncationDeltaToDistribute := deltaAmount.TruncateInt()
	// For funding from mint module account, we must pre-mint first.
	if moduleAccountName == types.ModuleName {
		if err := k.mintCoins(ctx, sdk.NewCoins(sdk.NewCoin(provisions.Denom, truncationDeltaToDistribute))); err != nil {
			return sdk.Int{}, err
		}
	}
	if err := k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(provisions.Denom, truncationDeltaToDistribute)), k.accountKeeper.GetModuleAddress(moduleAccountName)); err != nil {
		return sdk.Int{}, err
	}

	newDelta := deltaAmount.Sub(truncationDeltaToDistribute.ToDec())

	if err := k.SetTruncationDelta(ctx, moduleAccountName, newDelta); err != nil {
		return sdk.Int{}, err
	}

	if truncationDeltaToDistribute.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(truncationDeltaToDistribute.Int64()), fmt.Sprintf("mint_truncation_distributed_%s_delta", moduleAccountName))
	}

	return truncationDeltaToDistribute, nil
}

// getProportions gets the balance of the `MintedDenom` from minted coins and returns coins according to the
// allocation ratio. Returns error if ratio is greater than 1.
// TODO: this currently rounds down and is the cause of rounding discrepancies.
// To be fixed in: https://github.com/osmosis-labs/osmosis/issues/1917
func getProportions(mintedCoin sdk.Coin, ratio sdk.Dec) (sdk.Coin, error) {
	if ratio.GT(sdk.OneDec()) {
		return sdk.Coin{}, invalidRatioError{ratio}
	}
	return sdk.NewCoin(mintedCoin.Denom, mintedCoin.Amount.ToDec().Mul(ratio).TruncateInt()), nil
}

// getTruncationStoreKeyFromModuleAccount returns a truncation store key for the given module account name.
// moduleAccountName can either be developer vesting or mint module account.
// returns error if moduleAccountName is not either of the above.
func getTruncationStoreKeyFromModuleAccount(moduleAccountName string) ([]byte, error) {
	switch moduleAccountName {
	case types.DeveloperVestingModuleAcctName:
		return types.TruncatedDeveloperVestingDeltaKey, nil
	case types.ModuleName:
		return types.TruncatedInflationDeltaKey, nil
	default:
		return nil, sdkerrors.Wrapf(types.ErrInvalidModuleAccountGiven, "truncation delta can only be handled by (%s) or (%s) module accounts but (%s) was given", types.DeveloperVestingModuleAcctName, types.ModuleName, moduleAccountName)
	}
}
