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
	AttemptedDistribution sdk.Dec
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
func (k Keeper) GetTruncationDelta(ctx sdk.Context, key []byte) sdk.Dec {
	resultProto := sdk.DecProto{}
	osmoutils.MustGet(ctx.KVStore(k.storeKey), key, &resultProto)
	return resultProto.Dec
}

// SetInflationTruncationDelta sets the truncation delta.
func (k Keeper) SetTruncationDelta(ctx sdk.Context, key []byte, truncationDelta sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&sdk.DecProto{
		Dec: truncationDelta,
	})
	store.Set(key, b)
}

func (k Keeper) distributeEpochProvisions(ctx sdk.Context) (sdk.Int, error) {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// Mint and distribute inflation provisions from mint module account.
	// These exclude developer vesting rewards.
	inflationAmount, err := k.distributeInflationProvisions(ctx, minter.InflationProvisions(params))
	if err != nil {
		return sdk.Int{}, err
	}

	// Allocate dev rewards to respective accounts from developer vesting module account.
	developerVestingAmount, err := k.distributeDeveloperVestingProvisions(ctx, minter.DeveloperVestingEpochProvisions(params), params.WeightedDeveloperRewardsReceivers)
	if err != nil {
		return sdk.Int{}, err
	}

	totalDistributed := inflationAmount.Add(developerVestingAmount)

	// call a hook after the minting and distribution of new coins
	k.hooks.AfterDistributeMintedCoin(ctx)
	return totalDistributed, nil
}

// distributeInflationProvisions implements distribution of a minted coin from mint to external modules.
// inflation component incluedes all proportions from the parameters other than developer rewards.
func (k Keeper) distributeInflationProvisions(ctx sdk.Context, inflationCoin sdk.DecCoin) (sdk.Int, error) {
	params := k.GetParams(ctx)
	proportions := params.DistributionProportions

	// mint coins, update supply
	err := k.mintInflationCoins(ctx, sdk.NewCoins(sdk.NewCoin(inflationCoin.Denom, inflationCoin.Amount.TruncateInt())))
	if err != nil {
		return sdk.Int{}, err
	}

	// The mint coins are created from the mint module account exclusive of developer
	// rewards. Developer rewards are distributed from the developer vesting module account.
	// As a result, we exclude the developer proportions from calculations of mint distributions.
	nonDeveloperRewardsProportion := sdk.OneDec().Sub(proportions.DeveloperRewards)

	// allocate staking incentives into fee collector account to be moved to on next begin blocker by staking module account.
	stakingIncentivesAmount, err := k.distributeToModule(ctx, k.feeCollectorName, inflationCoin, proportions.Staking.Quo(nonDeveloperRewardsProportion))
	if err != nil {
		return sdk.Int{}, err
	}

	// allocate pool allocation ratio to pool-incentives module account.
	poolIncentivesAmount, err := k.distributeToModule(ctx, poolincentivestypes.ModuleName, inflationCoin, proportions.PoolIncentives.Quo(nonDeveloperRewardsProportion))
	if err != nil {
		return sdk.Int{}, err
	}

	// subtract from original provision to ensure no coins left over after the allocations
	inflationAmount := inflationCoin.Amount.TruncateInt()
	communityPoolAmount := inflationAmount.Sub(stakingIncentivesAmount).Sub(poolIncentivesAmount)
	err = k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(params.MintDenom, communityPoolAmount)), k.accountKeeper.GetModuleAddress(types.ModuleName))
	if err != nil {
		return sdk.Int{}, err
	}

	inflationTruncationMintedAndDistributed, err := k.handleTruncationDelta(ctx, types.TruncatedInflationDeltaKey, types.ModuleName, inflationCoin, inflationCoin.Amount.TruncateInt())
	if err != nil {
		return sdk.Int{}, err
	}

	inflationAmount = inflationAmount.Add(inflationTruncationMintedAndDistributed)

	if inflationAmount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(inflationAmount.Int64()), "mint_inflation_tokens")
	}

	return inflationAmount, nil
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

// mintInflationCoins mints tokens for inflation from the mint module accounts
//. It is meant to be used internally by the mint module.
// CONTRACT: minter's expected minter amount is updated separately
// CONTRACT: only called with the mint denom, never other coins.
func (k Keeper) mintInflationCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// distributeToModule distributes mintedCoin multiplied by proportion to the recepientModule account.
// If the minted coin amount multiplied by proportion is not whole, rounds down to the nearest integer.
// Returns the distributed rounded down amount, or error.
func (k Keeper) distributeToModule(ctx sdk.Context, recipientModule string, mintedCoin sdk.DecCoin, proportion sdk.Dec) (sdk.Int, error) {
	distributionAmount, err := getProportions(mintedCoin.Amount, proportion)
	if err != nil {
		return sdk.Int{}, err
	}
	truncatedDistributionAmount := distributionAmount.TruncateInt()
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, recipientModule, sdk.NewCoins(sdk.NewCoin(mintedCoin.Denom, truncatedDistributionAmount))); err != nil {
		return sdk.Int{}, err
	}
	return truncatedDistributionAmount, nil
}

// distributeDeveloperVestingProvisions distributes developer rewards from developer vesting module account
// to the respective account receivers by weight (developerRewardsReceivers).
// If no developer reward receivers given, funds the community pool instead.
// If developer reward receiver address is empty, funds the community pool.
// Distributes any delta resulting from truncating the amount to a whole integer to the community pool.
// Returns the total amount distributed from the developer vesting module account rounded down to the nearest integer.
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
func (k Keeper) distributeDeveloperVestingProvisions(ctx sdk.Context, developerRewardsCoin sdk.DecCoin, developerRewardsReceivers []types.WeightedAddress) (sdk.Int, error) {
	devRewardsAmount := developerRewardsCoin.Amount

	developerRewardsModuleAccountAddress := k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	oldDeveloperAccountBalance := k.bankKeeper.GetBalance(ctx, developerRewardsModuleAccountAddress, developerRewardsCoin.Denom)
	if oldDeveloperAccountBalance.Amount.ToDec().LT(devRewardsAmount) {
		return sdk.Int{}, insufficientDevVestingBalanceError{ActualBalance: oldDeveloperAccountBalance.Amount, AttemptedDistribution: devRewardsAmount}
	}

	truncatedDevRewardsAmount := devRewardsAmount.TruncateInt()
	devRewardCoins := sdk.NewCoins(sdk.NewCoin(developerRewardsCoin.Denom, truncatedDevRewardsAmount))

	// If no developer rewards receivers provided, fund the community pool from
	// the developer vesting module account.
	if len(developerRewardsReceivers) == 0 {
		err := k.communityPoolKeeper.FundCommunityPool(ctx, devRewardCoins, developerRewardsModuleAccountAddress)
		if err != nil {
			return sdk.Int{}, err
		}
	} else {
		// allocate developer rewards to addresses by weight
		for _, w := range developerRewardsReceivers {
			devPortionAmount, err := getProportions(devRewardsAmount, w.Weight)
			if err != nil {
				return sdk.Int{}, err
			}
			devRewardPortionCoins := sdk.NewCoins(sdk.NewCoin(developerRewardsCoin.Denom, devPortionAmount.TruncateInt()))
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

	// Take the new balance of the developer rewards pool to esitimate the truncation delta
	// stemming from the distribution of developer rewards to each of the accounts.
	newDeveloperAccountBalance := k.bankKeeper.GetBalance(ctx, developerRewardsModuleAccountAddress, developerRewardsCoin.Denom)
	distributedDuringCurrentEpochAmount := oldDeveloperAccountBalance.Sub(newDeveloperAccountBalance).Amount
	developerVestingTruncationDistributed, err := k.handleTruncationDelta(ctx, types.TruncatedDeveloperVestingDeltaKey, types.DeveloperVestingModuleAcctName, developerRewardsCoin, distributedDuringCurrentEpochAmount)
	if err != nil {
		return sdk.Int{}, err
	}

	// Take the current balance of the developer rewards pool and remove it from the supply offset
	// We re-introduce the new updated supply offset based on all amount that has been sent out
	// from the developer rewards module account address.
	k.bankKeeper.AddSupplyOffset(ctx, developerRewardsCoin.Denom, oldDeveloperAccountBalance.Amount)
	// Re-introduce the new supply offset
	k.bankKeeper.AddSupplyOffset(ctx, developerRewardsCoin.Denom, newDeveloperAccountBalance.Amount.Sub(developerVestingTruncationDistributed).Neg())

	if truncatedDevRewardsAmount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(truncatedDevRewardsAmount.Int64()), "mint_developer_vested_tokens")
	}

	// Return the amount of coins distributed to the developer rewards module account.
	// We truncate because the same is done to the delta that is distributed to the community pool.
	return truncatedDevRewardsAmount, nil
}

// TODO: update spec and tests
// distributeTruncationDelta distributes any truncation delta to the community pool.
// Due to limitations of some SDK interfaces that operate on integers, there are known truncation differences
// from the expected total epoch mint provisions.
// To use these interfaces, we always round down to the nearest integer by truncating decimals.
// As a result, it is possible to undermint. To mitigate that, we distribute any delta to the community pool.
// The delta is calculated by subtracting the actual distributions from the given expected total distributions.
func (k Keeper) handleTruncationDelta(ctx sdk.Context, key []byte, moduleAccountName string, provisions sdk.DecCoin, amountDistributed sdk.Int) (sdk.Int, error) {
	deltaAmount := k.calculateTotalTruncationDelta(ctx, key, provisions, amountDistributed)
	if deltaAmount.LT(sdk.OneDec()) {
		k.SetTruncationDelta(ctx, key, deltaAmount)
		return sdk.ZeroInt(), nil
	}

	// N.B: Truncation is acceptable because we check delta at the end of every epoch.
	// As a result, actual minted distributions always approach the expected value.
	truncationDeltaToDistribute := deltaAmount.TruncateInt()
	// For funding from mint module account, we must pre-mint first.
	if moduleAccountName == types.ModuleName {
		if err := k.mintInflationCoins(ctx, sdk.NewCoins(sdk.NewCoin(provisions.Denom, truncationDeltaToDistribute))); err != nil {
			return sdk.Int{}, err
		}
	}
	if err := k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(provisions.Denom, truncationDeltaToDistribute)), k.accountKeeper.GetModuleAddress(moduleAccountName)); err != nil {
		return sdk.Int{}, err
	}

	newDelta := deltaAmount.Sub(truncationDeltaToDistribute.ToDec())

	if newDelta.IsNegative() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidAmount, "developer rewards delta was negative (%s)", newDelta)
	}

	k.SetTruncationDelta(ctx, key, newDelta)

	// TODO: move telemetry to the caller
	// if truncationDeltaToDistribute.IsInt64() {
	// 	defer telemetry.ModuleSetGauge(types.ModuleName, float32(truncationDeltaToDistribute.Int64()), "mint_distributed_developer_rewards_truncation_delta")
	// }

	return truncationDeltaToDistribute, nil
}

// TODO: spec and test
func (k Keeper) calculateTotalTruncationDelta(ctx sdk.Context, key []byte, provisions sdk.DecCoin, amountDistributed sdk.Int) sdk.Dec {
	currentEpochRewardsDelta := provisions.Amount.Sub(amountDistributed.ToDec())
	return k.GetTruncationDelta(ctx, key).Add(currentEpochRewardsDelta)
}

// getProportions gets the balance of the `MintedDenom` from minted coins and returns coins according to the
// allocation ratio. Returns error if ratio is greater than 1.
func getProportions(value sdk.Dec, ratio sdk.Dec) (sdk.Dec, error) {
	if ratio.GT(sdk.OneDec()) {
		return sdk.Dec{}, invalidRatioError{ratio}
	}
	return value.Mul(ratio), nil
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
