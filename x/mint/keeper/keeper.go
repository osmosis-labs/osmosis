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

// IncreaseTruncationDelta increases the truncation delta at key by increaseByDelta.
func (k Keeper) IncreaseTruncationDelta(ctx sdk.Context, key []byte, increaseByDelta sdk.Dec) sdk.Dec {
	currentTruncationDelta := k.GetTruncationDelta(ctx, key)
	newTruncationDelta := currentTruncationDelta.Add(increaseByDelta)
	k.SetTruncationDelta(ctx, key, newTruncationDelta)
	return newTruncationDelta
}

// DecreaseTruncationDelta decreases the truncation delta at key by increaseByDelta.
func (k Keeper) DecreaseTruncationDelta(ctx sdk.Context, key []byte, decreaseByDelta sdk.Dec) sdk.Dec {
	currentTruncationDelta := k.GetTruncationDelta(ctx, key)
	newTruncationDelta := currentTruncationDelta.Sub(decreaseByDelta)
	k.SetTruncationDelta(ctx, key, newTruncationDelta)
	return newTruncationDelta
}

func (k Keeper) distributeEpochProvisions(ctx sdk.Context) (sdk.Int, error) {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// mint coins, update supply
	inflationCoin := minter.InflationProvision(params)
	err := k.mintInflationCoins(ctx, sdk.NewCoins(inflationCoin))
	if err != nil {
		return sdk.Int{}, err
	}

	// send the minted coins to the fee collector account
	err = k.distributeInflationCoin(ctx, inflationCoin)
	if err != nil {
		return sdk.Int{}, err
	}

	developerVestingCoin := minter.DeveloperVestingEpochProvision(params)
	// allocate dev rewards to respective accounts from developer vesting module account.
	developerVestingAmount, err := k.distributeDeveloperRewards(ctx, developerVestingCoin, params.WeightedDeveloperRewardsReceivers)
	if err != nil {
		return sdk.Int{}, err
	}

	inflationTruncationDelta, developerVestingTruncationDelta := k.updateTruncationDeltaAccumulators(ctx, &minter, inflationCoin.Amount, developerVestingAmount, params.DistributionProportions.DeveloperRewards)

	inflationTruncationDistributed, err := k.handleInflationTruncationDelta(ctx, inflationCoin.Denom, inflationTruncationDelta)
	if err != nil {
		return sdk.Int{}, err
	}

	developerVestingTruncationDistributed, err := k.handleDeveloperVestingTruncationDelta(ctx, inflationCoin.Denom, developerVestingTruncationDelta)
	if err != nil {
		return sdk.Int{}, err
	}

	totalDistributed := inflationCoin.Amount.Add(developerVestingAmount).Add(developerVestingTruncationDistributed).Add(inflationTruncationDistributed)

	// call a hook after the minting and distribution of new coins
	k.hooks.AfterDistributeMintedCoin(ctx)
	return totalDistributed, nil
}

// distributeInflationCoin implements distribution of a minted coin from mint to external modules.
// inflation component incluedes all proportions from the parameters other than developer rewards.
func (k Keeper) distributeInflationCoin(ctx sdk.Context, mintedCoin sdk.Coin) error {
	params := k.GetParams(ctx)
	proportions := params.DistributionProportions

	// The mint coins are created from the mint module account exclusive of developer
	// rewards. Developer rewards are distributed from the developer vesting module account.
	// As a result, we exclude the developer proportions from calculations of mint distributions.
	nonDeveloperRewardsProportion := sdk.OneDec().Sub(proportions.DeveloperRewards)

	// allocate staking incentives into fee collector account to be moved to on next begin blocker by staking module account.
	stakingIncentivesAmount, err := k.distributeToModule(ctx, k.feeCollectorName, mintedCoin, proportions.Staking.Quo(nonDeveloperRewardsProportion))
	if err != nil {
		return err
	}

	// allocate pool allocation ratio to pool-incentives module account.
	poolIncentivesAmount, err := k.distributeToModule(ctx, poolincentivestypes.ModuleName, mintedCoin, proportions.PoolIncentives.Quo(nonDeveloperRewardsProportion))
	if err != nil {
		return err
	}

	// subtract from original provision to ensure no coins left over after the allocations
	communityPoolAmount := mintedCoin.Amount.Sub(stakingIncentivesAmount).Sub(poolIncentivesAmount)
	err = k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(params.MintDenom, communityPoolAmount)), k.accountKeeper.GetModuleAddress(types.ModuleName))
	if err != nil {
		return err
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "mint_inflation_tokens")
	}

	return nil
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
func (k Keeper) distributeToModule(ctx sdk.Context, recipientModule string, mintedCoin sdk.Coin, proportion sdk.Dec) (sdk.Int, error) {
	distributionAmount, err := getProportions(mintedCoin.Amount.ToDec(), proportion)
	if err != nil {
		return sdk.Int{}, err
	}
	truncatedDistributionAmount := distributionAmount.TruncateInt()
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, recipientModule, sdk.NewCoins(sdk.NewCoin(mintedCoin.Denom, truncatedDistributionAmount))); err != nil {
		return sdk.Int{}, err
	}
	return truncatedDistributionAmount, nil
}

// distributeDeveloperRewards distributes developer rewards from developer vesting module account
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
func (k Keeper) distributeDeveloperRewards(ctx sdk.Context, developerRewardsCoin sdk.Coin, developerRewardsReceivers []types.WeightedAddress) (sdk.Int, error) {
	devRewardsAmount := developerRewardsCoin.Amount.ToDec()

	developerRewardsModuleAccountAddress := k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	oldDeveloperAccountBalance := k.bankKeeper.GetBalance(ctx, developerRewardsModuleAccountAddress, developerRewardsCoin.Denom)
	if oldDeveloperAccountBalance.Amount.ToDec().LT(devRewardsAmount) {
		return sdk.Int{}, insufficientDevVestingBalanceError{ActualBalance: oldDeveloperAccountBalance.Amount, AttemptedDistribution: devRewardsAmount}
	}

	devRewardCoins := sdk.NewCoins(developerRewardsCoin)

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
	truncationDelta := devRewardCoins.Sub(sdk.NewCoins(oldDeveloperAccountBalance.Sub(newDeveloperAccountBalance)))
	if err := k.communityPoolKeeper.FundCommunityPool(ctx, truncationDelta, developerRewardsModuleAccountAddress); err != nil {
		return sdk.Int{}, err
	}

	// Take the current balance of the developer rewards pool and remove it from the supply offset
	// We re-introduce the new updated supply offset based on all amount that has been sent out
	// from the developer rewards module account address.
	k.bankKeeper.AddSupplyOffset(ctx, developerRewardsCoin.Denom, oldDeveloperAccountBalance.Amount)
	// Re-introduce the new supply offset
	k.bankKeeper.AddSupplyOffset(ctx, developerRewardsCoin.Denom, newDeveloperAccountBalance.Amount.Sub(truncationDelta.AmountOf(developerRewardsCoin.Denom)).Neg())

	if developerRewardsCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(developerRewardsCoin.Amount.Int64()), "mint_developer_vested_tokens")
	}

	// Return the amount of coins distributed to the developer rewards module account.
	// We truncate because the same is done to the delta that is distributed to the community pool.
	return devRewardsAmount.TruncateInt(), nil
}

// TODO: fix
// distributeTruncationDelta distributes any truncation delta to the community pool.
// Due to limitations of some SDK interfaces that operate on integers, there are known truncation differences
// from the expected total epoch mint provisions.
// To use these interfaces, we always round down to the nearest integer by truncating decimals.
// As a result, it is possible to undermint. To mitigate that, we distribute any delta to the community pool.
// The delta is calculated by subtracting the actual distributions from the given expected total distributions.
func (k Keeper) handleDeveloperVestingTruncationDelta(ctx sdk.Context, mintedDenom string, developerVestingTruncationDelta sdk.Dec) (sdk.Int, error) {
	// N.B: Truncation is acceptable because we check delta at the end of every epoch.
	// As a result, actual minted distributions always approach the expected value.
	// For distributing delta from mint module account, we have to pre-mint first.
	if developerVestingTruncationDelta.IsNegative() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidAmount, "developer rewards delta was negative (%s)", developerVestingTruncationDelta)
	}
	if developerVestingTruncationDelta.GT(sdk.OneDec()) {
		truncationDevVestingDeltaToDistribute := developerVestingTruncationDelta.TruncateInt()
		if err := k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(mintedDenom, truncationDevVestingDeltaToDistribute)), k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)); err != nil {
			return sdk.Int{}, err
		}

		k.DecreaseTruncationDelta(ctx, types.TruncatedDeveloperVestingDeltaKey, truncationDevVestingDeltaToDistribute.ToDec())

		if truncationDevVestingDeltaToDistribute.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(truncationDevVestingDeltaToDistribute.Int64()), "mint_distributed_developer_rewards_truncation_delta")
		}

		return truncationDevVestingDeltaToDistribute, nil
	}
	return sdk.ZeroInt(), nil
}

// TODO: spec and tests
func (k Keeper) handleInflationTruncationDelta(ctx sdk.Context, mintedDenom string, inflationTruncationDelta sdk.Dec) (sdk.Int, error) {
	// N.B: Similarly to developer vesting delta, truncation here is acceptable.
	if inflationTruncationDelta.IsNegative() {
		return sdk.Int{}, sdkerrors.Wrapf(types.ErrInvalidAmount, "inflation delta was negative (%s)", inflationTruncationDelta)
	}
	if inflationTruncationDelta.IsPositive() {
		truncatedInflationDeltaToDistribute := inflationTruncationDelta.TruncateInt()
		if err := k.mintInflationCoins(ctx, sdk.NewCoins(sdk.NewCoin(mintedDenom, truncatedInflationDeltaToDistribute))); err != nil {
			return sdk.Int{}, err
		}

		if err := k.communityPoolKeeper.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(mintedDenom, truncatedInflationDeltaToDistribute)), k.accountKeeper.GetModuleAddress(types.ModuleName)); err != nil {
			return sdk.Int{}, err
		}

		k.DecreaseTruncationDelta(ctx, types.TruncatedInflationDeltaKey, truncatedInflationDeltaToDistribute.ToDec())

		if truncatedInflationDeltaToDistribute.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(truncatedInflationDeltaToDistribute.Int64()), "mint_distributed_inflation_truncation_delta")
		}

		return truncatedInflationDeltaToDistribute, nil
	}
	return sdk.ZeroInt(), nil
}

// getDeveloperVestedAmount returns the vestes amount from the developer vesting module account.
func (k Keeper) getDeveloperVestedAmount(ctx sdk.Context, denom string) sdk.Int {
	unvestedAmount := k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), denom).Amount
	vestedAmount := sdk.NewInt(developerVestingAmount).Sub(unvestedAmount)
	return vestedAmount
}

// getInflationAmount returns the amount minted by the mint module account
// without considering the developer rewards module account.
// The developer rewards were pre-minted to its own module account at genesis.
// Therefore, the developer rewards can be distributed separately.
// As a result, we should not consider the original developer
// vesting amount when calculating the minted amount.
func (k Keeper) getInflationAmount(ctx sdk.Context, denom string) sdk.Int {
	totalSupply := k.bankKeeper.GetSupply(ctx, denom).Amount
	return totalSupply.Sub(sdk.NewInt(developerVestingAmount))
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

// updateTruncationDeltaAccumulators updates the truncation delta accumulators by mutating minter
// It does not persist minter to the store. The caller is responsible for persisting the minter.
func (k Keeper) updateTruncationDeltaAccumulators(ctx sdk.Context, minter *types.Minter, distributedInflationAmount, distributedDeveloperVestingAmount sdk.Int, developerRewardsProportion sdk.Dec) (sdk.Dec, sdk.Dec) {
	devRewardsProportion := minter.EpochProvisions.Mul(developerRewardsProportion)
	inflationProportion := minter.EpochProvisions.Sub(devRewardsProportion)

	inflationDelta := inflationProportion.Sub(distributedInflationAmount.ToDec())
	devRewardsDelta := devRewardsProportion.Sub(distributedDeveloperVestingAmount.ToDec())

	newInflationDelta := k.IncreaseTruncationDelta(ctx, types.TruncatedInflationDeltaKey, inflationDelta)
	newDevRewardsDelta := k.IncreaseTruncationDelta(ctx, types.TruncatedDeveloperVestingDeltaKey, devRewardsDelta)
	return newInflationDelta, newDevRewardsDelta
}
