package keeper

import (
	"fmt"

	"cosmossdk.io/log"

	errorsmod "cosmossdk.io/errors"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v31/x/mint/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v31/x/pool-incentives/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v31/x/txfees/types"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper of the mint store.
type Keeper struct {
	storeKey            storetypes.StoreKey
	paramSpace          paramtypes.Subspace
	accountKeeper       types.AccountKeeper
	bankKeeper          types.BankKeeper
	communityPoolKeeper types.CommunityPoolKeeper
	epochKeeper         types.EpochKeeper
	stakingKeeper       types.StakingKeeper
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
	sk types.StakingKeeper,
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
		stakingKeeper:       sk,
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

// GetInflation calculates the current inflation rate.
// Formula: ((Epoch provisions * (1 - Community Pool proportion)) * 365) / Circulating Supply
//
// The denominator is circulating supply (minted - burned - restricted), the
// public-float base used by Coingecko/CMC, rather than offset-adjusted total
// supply. This is a query-only change (GetInflation is called only from the
// gRPC querier, never from minting/BeginBlock/EndBlock), so it is point-release
// safe. Note for integrators: the reported inflation value is higher than under
// the previous total-supply denominator, since the denominator is now smaller.
func (k Keeper) GetInflation(ctx sdk.Context) (osmomath.Dec, error) {
	// Get current epoch provisions
	minter := k.GetMinter(ctx)
	epochProvisions := minter.EpochProvisions

	// Get distribution parameters
	params := k.GetParams(ctx)
	communityPoolProportion := params.DistributionProportions.CommunityPool

	// Get circulating supply of the mint denom, reusing the params read above.
	circulatingSupply, err := k.getCirculatingSupply(ctx, params)
	if err != nil {
		return osmomath.ZeroDec(), fmt.Errorf("failed to get circulating supply: %w", err)
	}

	// Calculate circulating provisions: epoch provisions * (1 - community pool proportion)
	oneMinusCommunityPool := osmomath.OneDec().Sub(communityPoolProportion)
	circulatingProvisions := epochProvisions.Mul(oneMinusCommunityPool)

	// Calculate annualized provisions: circulating provisions * 365
	annualizedProvisions := circulatingProvisions.Mul(osmomath.NewDec(365))

	// Calculate inflation rate: annualized provisions / circulating supply
	if circulatingSupply.IsPositive() {
		circulatingSupplyDec := circulatingSupply.ToLegacyDec()
		return annualizedProvisions.Quo(circulatingSupplyDec), nil
	}

	return osmomath.ZeroDec(), nil
}

// The exported supply getters below are thin wrappers over unexported variants
// that take the mint params (or just the mint denom) as an argument. Params are
// threaded through so that one query reads the params subspace exactly once,
// instead of each nested getter re-reading it.

// GetBurnedSupply returns the amount of mint-denom held in the null/burn
// address (txfeestypes.DefaultNullAddress, the all-zero account). These coins
// are permanently removed from circulation.
func (k Keeper) GetBurnedSupply(ctx sdk.Context) osmomath.Int {
	return k.getBurnedSupply(ctx, k.GetParams(ctx).MintDenom)
}

func (k Keeper) getBurnedSupply(ctx sdk.Context, mintDenom string) osmomath.Int {
	burned := k.bankKeeper.GetBalance(ctx, txfeestypes.DefaultNullAddress, mintDenom)
	return burned.Amount
}

// GetTotalSupply returns the total supply of the mint denom (minted - burned).
//
// "minted" is the raw bank supply (GetSupply), which includes the unvested
// developer-vesting balance exactly once. Total supply is reported before
// netting out restricted holdings, matching the Coingecko/CMC total-supply
// methodology.
func (k Keeper) GetTotalSupply(ctx sdk.Context) osmomath.Int {
	return k.getTotalSupply(ctx, k.GetParams(ctx).MintDenom)
}

func (k Keeper) getTotalSupply(ctx sdk.Context, mintDenom string) osmomath.Int {
	mintedSupply := k.bankKeeper.GetSupply(ctx, mintDenom)
	burnedSupply := k.getBurnedSupply(ctx, mintDenom)
	return mintedSupply.Amount.Sub(burnedSupply)
}

// GetCirculatingSupply returns the circulating supply, equivalent to the public
// float: minted - burned - restricted.
//
// All three terms are computed on the same raw-supply base (GetSupply). The
// developer-vesting balance is included in minted exactly once and subtracted in
// restricted exactly once, so it nets to zero in circulating supply. Do NOT base
// this on GetSupplyWithOffset: that already nets out developer vesting, which
// combined with the restricted-supply subtraction would double-count it.
func (k Keeper) GetCirculatingSupply(ctx sdk.Context) (osmomath.Int, error) {
	return k.getCirculatingSupply(ctx, k.GetParams(ctx))
}

func (k Keeper) getCirculatingSupply(ctx sdk.Context, params types.Params) (osmomath.Int, error) {
	restrictedSupply, err := k.getRestrictedSupply(ctx, params)
	if err != nil {
		return osmomath.ZeroInt(), fmt.Errorf("failed to get restricted supply: %w", err)
	}

	return k.getTotalSupply(ctx, params.MintDenom).Sub(restrictedSupply), nil
}

// GetRestrictedSupply returns the amount of mint-denom held by known restricted
// entities and therefore excluded from circulating supply:
//   - the developer-vesting module account balance (still-unvested dev tokens);
//   - the community pool;
//   - the developer-vested reward receiver addresses (balance + staked + unbonding);
//   - the curated foundation/investor restricted addresses (balance + staked + unbonding).
//
// Liquid balances, bonded delegations, and unbonding amounts are counted.
// OSMO a restricted address moves into other modules is NOT tracked: x/lockup
// locks, superfluid positions, LP shares, and CL positions all leave the
// address's bank/staking footprint and would be reported as circulating.
// Restricted entities are expected to hold OSMO only liquid or (un)staked; if
// one ever locks or deploys OSMO, this accounting needs a corresponding term.
func (k Keeper) GetRestrictedSupply(ctx sdk.Context) (osmomath.Int, error) {
	return k.getRestrictedSupply(ctx, k.GetParams(ctx))
}

func (k Keeper) getRestrictedSupply(ctx sdk.Context, params types.Params) (osmomath.Int, error) {
	restrictedSupply := osmomath.ZeroInt()

	// seen tracks addresses already counted so that an address appearing in both
	// the governance param receivers and the curated constant is not
	// double-counted.
	seen := make(map[string]struct{})

	// 1. Developer vesting module account balance (unvested dev tokens). Seed
	// the dedup set with the module address: param validation does not stop
	// governance from listing it as a rewards receiver, which would otherwise
	// double-count it below.
	devVestingAddr := k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName)
	if devVestingAddr != nil {
		devVestingBalance := k.bankKeeper.GetBalance(ctx, devVestingAddr, params.MintDenom)
		restrictedSupply = restrictedSupply.Add(devVestingBalance.Amount)
		seen[devVestingAddr.String()] = struct{}{}
	}

	// 2. Community pool balance. Read from the distribution FeePool (the
	// authoritative community-pool accounting), not the distribution module
	// account balance. CommunityPool is DecCoins; truncate to Int (rounds down,
	// so restricted is at most 1 uosmo under-counted per denom).
	feePool, err := k.communityPoolKeeper.GetFeePool(ctx)
	if err != nil {
		return osmomath.ZeroInt(), fmt.Errorf("failed to get community pool: %w", err)
	}
	for _, coin := range feePool.GetCommunityPool() {
		if coin.Denom == params.MintDenom {
			restrictedSupply = restrictedSupply.Add(coin.Amount.TruncateInt())
			break
		}
	}

	// 3. Developer-vested reward receiver addresses (governance param). Skip
	// empty addresses (those route to the community pool, already counted).
	for _, devAddr := range params.WeightedDeveloperRewardsReceivers {
		if devAddr.Address == emptyWeightedAddressReceiver {
			continue
		}
		addr, err := sdk.AccAddressFromBech32(devAddr.Address)
		if err != nil {
			return osmomath.ZeroInt(), fmt.Errorf("failed to parse developer rewards receiver %q: %w", devAddr.Address, err)
		}
		if _, ok := seen[addr.String()]; ok {
			continue
		}
		seen[addr.String()] = struct{}{}
		holdings, err := k.addressHoldings(ctx, addr, params.MintDenom)
		if err != nil {
			return osmomath.ZeroInt(), err
		}
		restrictedSupply = restrictedSupply.Add(holdings)
	}

	// 4. Curated foundation/investor/strategic restricted addresses (compiled-in
	// constant). Parsed here (not at package init) because the bech32 "osmo"
	// prefix is set by app params init, which may not have run when the types
	// package loads. De-duplicated against the param receivers above.
	for _, addrStr := range types.RestrictedAddresses {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			return osmomath.ZeroInt(), fmt.Errorf("failed to parse restricted address %q: %w", addrStr, err)
		}
		if _, ok := seen[addr.String()]; ok {
			continue
		}
		seen[addr.String()] = struct{}{}
		holdings, err := k.addressHoldings(ctx, addr, params.MintDenom)
		if err != nil {
			return osmomath.ZeroInt(), err
		}
		restrictedSupply = restrictedSupply.Add(holdings)
	}

	return restrictedSupply, nil
}

// addressHoldings returns an address's liquid balance plus its bonded (staked)
// and unbonding amounts in the given denom.
//
// Bonded amounts come from the staking keeper's GetDelegatorBonded, which
// converts delegation shares to tokens via each validator's current exchange
// rate (accounting for slashing), accumulates truncated per-validator values,
// and rounds once. It skips a delegation whose validator record is missing,
// matching the SDK's own reporting convention; that state is structurally
// unreachable (validators are not removed while they have delegations).
// Unbonding amounts are slash-adjusted entry balances. Tokens mid-redelegation
// remain represented by the destination delegation shares.
func (k Keeper) addressHoldings(ctx sdk.Context, addr sdk.AccAddress, denom string) (osmomath.Int, error) {
	balance := k.bankKeeper.GetBalance(ctx, addr, denom)
	bonded, err := k.stakingKeeper.GetDelegatorBonded(ctx, addr)
	if err != nil {
		return osmomath.ZeroInt(), fmt.Errorf("failed to get bonded amount for %s: %w", addr, err)
	}
	unbonding, err := k.stakingKeeper.GetDelegatorUnbonding(ctx, addr)
	if err != nil {
		return osmomath.ZeroInt(), fmt.Errorf("failed to get unbonding amount for %s: %w", addr, err)
	}

	return balance.Amount.Add(bonded).Add(unbonding), nil
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
