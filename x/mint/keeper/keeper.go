package keeper

import (
	"fmt"

	"cosmossdk.io/log"

	errorsmod "cosmossdk.io/errors"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/mint/types"
	stablestakingincentivestypes "github.com/osmosis-labs/osmosis/v27/x/stable-staking-incentives/types"
)

// Keeper of the mint store.
type Keeper struct {
	storeKey            storetypes.StoreKey
	paramSpace          paramtypes.Subspace
	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	communityPoolKeeper types.CommunityPoolKeeper
	epochKeeper         types.EpochKeeper
	hooks               types.MintHooks
	feeCollectorName    string
}

type invalidRatioError struct {
	ActualRatio osmomath.Dec
}

func (e invalidRatioError) Error() string {
	return fmt.Sprintf("mint allocation ratio (%s) is greater than 1", e.ActualRatio)
}

type insufficientDevVestingBalanceError struct {
	ActualBalance         osmomath.Int
	AttemptedDistribution osmomath.Int
}

func (e insufficientDevVestingBalanceError) Error() string {
	return fmt.Sprintf("developer vesting balance (%s) is smaller than requested distribution of (%s)", e.ActualBalance, e.AttemptedDistribution)
}

const emptyWeightedAddressReceiver = ""

// NewKeeper creates a new mint Keeper instance.
func NewKeeper(
	key storetypes.StoreKey, paramSpace paramtypes.Subspace,
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
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.MinterKey, &minter)
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

// SetParam sets a specific mint module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
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

	// allocate pool allocation ratio to stable-staking-incentives module account.
	poolIncentivesAmount, err := k.distributeToModule(ctx, stablestakingincentivestypes.ModuleName, mintedCoin, proportions.PoolIncentives)
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

// distributeToModule distributes mintedCoin multiplied by proportion to the recepientModule account.osmomath.Dec
func (k Keeper) distributeToModule(ctx sdk.Context, recipientModule string, mintedCoin sdk.Coin, proportion osmomath.Dec) (osmomath.Int, error) {
	distributionCoin, err := getProportions(mintedCoin, proportion)
	if err != nil {
		return osmomath.Int{}, err
	}
	ctx.Logger().Info("distributeToModule", "module", types.ModuleName, "recepientModule", recipientModule, "distributionCoin", distributionCoin, "height", ctx.BlockHeight())
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, recipientModule, sdk.NewCoins(distributionCoin)); err != nil {
		return osmomath.Int{}, err
	}
	return distributionCoin.Amount, nil
}

// distributeDeveloperRewards distributes developer rewards from developer vesting module account
// to the respective account receivers by weight (developerRewardsReceivers).
// If no developer reward receivers given, funds the community pool instead.
// Returns the total amount distributed from the developer vesting module account.
// Updates supply offsets to reflect the amount of coins distributed. This is done so because the developer rewards distributions are
// allocated from its own module account, not the mint module account (TODO: next step in https://github.com/osmosis-labs/osmosis/issues/1916).
// Returns nil on success, error otherwise.
// With respect to input parameters, errors occur when:
// - developerRewardsProportion is greater than 1.
// - invalid address in developer rewards receivers.
// - the balance of developer module account is less than totalMintedCoin * developerRewardsProportion.
// - the balance of mint module is less than totalMintedCoin * developerRewardsProportion.
// CONTRACT:
// - weights in developerRewardsReceivers add up to 1.
// - addresses in developerRewardsReceivers are valid or empty string.osmomath.Dec
func (k Keeper) distributeDeveloperRewards(ctx sdk.Context, totalMintedCoin sdk.Coin, developerRewardsProportion osmomath.Dec, developerRewardsReceivers []types.WeightedAddress) (osmomath.Int, error) {
	devRewardCoin, err := getProportions(totalMintedCoin, developerRewardsProportion)
	if err != nil {
		return osmomath.Int{}, err
	}

	developerRewardsModuleAccountAddress := k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	developerAccountBalance := k.bankKeeper.GetBalance(ctx, developerRewardsModuleAccountAddress, totalMintedCoin.Denom)
	if developerAccountBalance.Amount.LT(devRewardCoin.Amount) {
		return osmomath.Int{}, insufficientDevVestingBalanceError{ActualBalance: developerAccountBalance.Amount, AttemptedDistribution: devRewardCoin.Amount}
	}

	devRewardCoins := sdk.NewCoins(devRewardCoin)
	// TODO: https://github.com/osmosis-labs/osmosis/issues/2025
	// Avoid over-allocating from the mint module address and have to later burn it here:
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, devRewardCoins); err != nil {
		return osmomath.Int{}, err
	}

	// Take the current balance of the developer rewards pool and remove it from the supply offset
	// We re-introduce the new supply at the end, in order to avoid any rounding discrepancies.
	k.bankKeeper.AddSupplyOffset(ctx, totalMintedCoin.Denom, developerAccountBalance.Amount)

	// If no developer rewards receivers provided, fund the community pool from
	// the developer vesting module account.
	if len(developerRewardsReceivers) == 0 {
		err = k.communityPoolKeeper.FundCommunityPool(ctx, devRewardCoins, developerRewardsModuleAccountAddress)
		if err != nil {
			return osmomath.Int{}, err
		}
	} else {
		// allocate developer rewards to addresses by weight
		for _, w := range developerRewardsReceivers {
			devPortionCoin, err := getProportions(devRewardCoin, w.Weight)
			if err != nil {
				return osmomath.Int{}, err
			}
			devRewardPortionCoins := sdk.NewCoins(devPortionCoin)
			// fund community pool when rewards address is empty.
			if w.Address == emptyWeightedAddressReceiver {
				err := k.communityPoolKeeper.FundCommunityPool(ctx, devRewardPortionCoins,
					k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName))
				if err != nil {
					return osmomath.Int{}, err
				}
			} else {
				devRewardsAddr, err := sdk.AccAddressFromBech32(w.Address)
				if err != nil {
					return osmomath.Int{}, err
				}
				// If recipient is vesting account, pay to account according to its vesting condition
				err = k.bankKeeper.SendCoinsFromModuleToAccount(
					ctx, types.DeveloperVestingModuleAcctName, devRewardsAddr, devRewardPortionCoins)
				if err != nil {
					return osmomath.Int{}, err
				}
			}
		}
	}

	// Take the new balance of the developer rewards pool and add it back to the supply offset deduction
	developerAccountBalance = k.bankKeeper.GetBalance(ctx, developerRewardsModuleAccountAddress, totalMintedCoin.Denom)
	k.bankKeeper.AddSupplyOffset(ctx, totalMintedCoin.Denom, developerAccountBalance.Amount.Neg())

	return devRewardCoin.Amount, nil
}

// getProportions gets the balance of the `MintedDenom` from minted coins and returns coins according to the
// allocation ratio. Returns error if ratio is greater than 1.
// TODO: this currently rounds down and is the cause of rounding discrepancies.
// To be fixed in: https://github.com/osmosis-losmomath.Decosis/issues/1917
func getProportions(mintedCoin sdk.Coin, ratio osmomath.Dec) (sdk.Coin, error) {
	if ratio.GT(osmomath.OneDec()) {
		return sdk.Coin{}, invalidRatioError{ratio}
	}
	return sdk.NewCoin(mintedCoin.Denom, mintedCoin.Amount.ToLegacyDec().Mul(ratio).TruncateInt()), nil
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
		return errorsmod.Wrap(types.ErrAmountNilOrZero, "amount cannot be nil or zero")
	}
	if k.accountKeeper.HasAccount(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)) {
		return errorsmod.Wrapf(types.ErrModuleAccountAlreadyExist, "%s vesting module account already exist", types.DeveloperVestingModuleAcctName)
	}

	moduleAcc := authtypes.NewEmptyModuleAccount(
		types.DeveloperVestingModuleAcctName, authtypes.Minter)
	maccI, ok := (k.accountKeeper.NewAccount(ctx, moduleAcc)).(sdk.ModuleAccountI) // this sets the account number
	if !ok {
		return fmt.Errorf("account of type %T doesn't implement sdk.ModuleAccountI", moduleAcc)
	}

	k.accountKeeper.SetModuleAccount(ctx, maccI)

	err := k.bankKeeper.MintCoins(ctx, types.DeveloperVestingModuleAcctName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}
	return nil
}
