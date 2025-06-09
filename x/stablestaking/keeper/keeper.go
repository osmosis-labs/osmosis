package keeper

import (
	"fmt"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/osmosis-labs/osmosis/v27/x/stablestaking/types"
)

type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.Codec
	paramSpace paramstypes.Subspace

	epochKeeper   types.EpochKeeper
	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	OracleKeeper  types.OracleKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	paramstore paramstypes.Subspace,
	epochKeeper types.EpochKeeper,
	accKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	oracleKeeper types.OracleKeeper,
) Keeper {
	// ensure stable staking module account is set
	if addr := accKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		paramSpace:    paramstore,
		epochKeeper:   epochKeeper,
		BankKeeper:    bankKeeper,
		AccountKeeper: accKeeper,
		OracleKeeper:  oracleKeeper,
	}
}

// GetParams return module params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSetIfExists(ctx, &params)
	return params
}

// SetParams set up module params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

func (k Keeper) SetEpochSnapshot(ctx sdk.Context, snapshot types.EpochSnapshot, denom string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.SnapshotKey))
	bz := k.cdc.MustMarshal(&snapshot)

	// Store both by epoch and as latest
	epoch := k.epochKeeper.GetEpochInfo(ctx, k.GetParams(ctx).RewardEpochIdentifier).CurrentEpoch
	epochKey := sdk.Uint64ToBigEndian(uint64(epoch))
	store.Set(epochKey, bz)

	// Also store as latest for quick access
	store.Set([]byte(fmt.Sprintf("latest:%s", denom)), bz)
}

func (k Keeper) GetEpochSnapshot(ctx sdk.Context, denom string) (types.EpochSnapshot, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.SnapshotKey))

	// Try to get the latest snapshot first
	bz := store.Get([]byte(fmt.Sprintf("latest:%s", denom)))
	if bz == nil {
		return types.EpochSnapshot{}, fmt.Errorf("epoch snapshot not found")
	}

	var snapshot types.EpochSnapshot
	k.cdc.MustUnmarshal(bz, &snapshot)
	return snapshot, nil
}

func (k Keeper) SnapshotCurrentEpoch(ctx sdk.Context) {
	params := k.GetParams(ctx)
	if len(params.SupportedTokens) == 0 {
		return
	}

	// Get the current epoch
	currentEpoch := k.epochKeeper.GetEpochInfo(ctx, params.RewardEpochIdentifier).CurrentEpoch
	pools := k.GetPools(ctx)

	// For each supported token, create a snapshot
	for _, pool := range pools {
		var stakers []*types.UserStake

		// Iterate through all stakers and collect their stakes for this denom
		k.IterateActiveStakers(ctx, func(addr sdk.AccAddress, stake types.UserStake) {
			// Only include stakers for this specific denom
			if stake.Epoch <= currentEpoch-1 { // Only include stakers from previous epochs
				stakers = append(stakers, &stake)
			}
		})

		// Create a snapshot for this pool
		snapshot := types.EpochSnapshot{
			TotalShares: pool.TotalShares,
			TotalStaked: pool.TotalStaked,
			Stakers:     stakers,
		}

		// Store the snapshot
		k.SetEpochSnapshot(ctx, snapshot, pool.Denom)
	}
}

func (k Keeper) IterateActiveStakers(ctx sdk.Context, cb func(addr sdk.AccAddress, stake types.UserStake)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UserStakeKey))

	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var stake types.UserStake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)

		addr, err := sdk.AccAddressFromBech32(stake.Address)
		if err != nil {
			panic(fmt.Sprintf("invalid address in active staker store: %s", err))
		}

		cb(addr, stake)
	}
}

func (k Keeper) DistributeRewardsToLastEpochStakers(ctx sdk.Context) {
	params := k.GetParams(ctx)
	if len(params.SupportedTokens) == 0 {
		return
	}

	// Get total rewards available
	moduleRewardsAddr := k.AccountKeeper.GetModuleAddress(types.NativeRewardsCollectorName)
	totalReward := k.BankKeeper.GetBalance(ctx, moduleRewardsAddr, appparams.BaseCoinUnit)
	if totalReward.IsZero() {
		return // No rewards to distribute
	}

	// Calculate the total staked amount across all pools
	var totalStakedAcrossPools math.LegacyDec
	poolSnapshots := make(map[string]types.EpochSnapshot)

	for _, denom := range params.SupportedTokens {
		snapshot, err := k.GetEpochSnapshot(ctx, denom)
		if err != nil {
			continue
		}
		if snapshot.TotalStaked.IsZero() {
			continue
		}
		totalStakedAcrossPools = totalStakedAcrossPools.Add(snapshot.TotalStaked)
		poolSnapshots[denom] = snapshot
	}

	if totalStakedAcrossPools.IsZero() {
		return // No staked amount across any pool
	}

	// Distribute rewards proportionally to each pool based on their total staked amount
	for _, snapshot := range poolSnapshots {
		// Calculate pool's share of total rewards based on its total staked amount
		poolRewardShare := snapshot.TotalStaked.Quo(totalStakedAcrossPools)
		poolReward := poolRewardShare.MulInt(totalReward.Amount).TruncateInt()

		if poolReward.IsZero() {
			continue
		}

		// Distribute pool's rewards to its stakers
		for _, staker := range snapshot.Stakers {
			if staker.Shares.IsZero() {
				continue
			}

			// Calculate staker's share of pool rewards
			stakerReward := staker.Shares.Quo(snapshot.TotalShares).MulInt(poolReward).TruncateInt()
			if stakerReward.IsZero() {
				continue
			}

			addr, err := sdk.AccAddressFromBech32(staker.Address)
			if err != nil {
				panic(fmt.Sprintf("invalid address in snapshot: %s", err))
			}

			// Send reward tokens
			rewardCoin := sdk.NewCoin(appparams.BaseCoinUnit, stakerReward)
			err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.NativeRewardsCollectorName, addr, sdk.NewCoins(rewardCoin))
			if err != nil {
				panic(fmt.Sprintf("failed to send rewards: %s", err))
			}
		}
	}
}

func (k Keeper) GetEpochReward(ctx sdk.Context) math.Int {
	params := k.GetParams(ctx)
	if len(params.SupportedTokens) == 0 {
		return math.ZeroInt()
	}

	// Get the total staked amount for the first supported token
	pool, found := k.GetPool(ctx, params.SupportedTokens[0])
	if !found || pool.TotalStaked.IsZero() {
		return math.ZeroInt()
	}

	// Parse reward rate from params
	rewardRate, err := math.LegacyNewDecFromStr(params.RewardRate)
	if err != nil {
		panic(fmt.Sprintf("invalid reward rate: %s", err))
	}

	// Calculate reward: total_staked * reward_rate
	reward := pool.TotalStaked.Mul(rewardRate).TruncateInt()
	return reward
}

func (k Keeper) CompleteUnbonding(ctx sdk.Context, currentEpoch int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte(types.UnbondingKey))
	iterator := storetypes.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	//currentEpoch := k.epochKeeper.GetEpochInfo(ctx, "day").CurrentEpoch

	for ; iterator.Valid(); iterator.Next() {
		var unbondingInfo types.UnbondingInfo
		k.cdc.MustUnmarshal(iterator.Value(), &unbondingInfo)

		// Check if unbonding period has passed
		if unbondingInfo.UnbondEpoch <= currentEpoch {
			// Convert address string to AccAddress
			addr, err := sdk.AccAddressFromBech32(unbondingInfo.Address)
			if err != nil {
				panic(fmt.Sprintf("invalid address in unbonding info: %s", err))
			}

			// Create coin from unbonding amount
			unbondAmount := sdk.NewCoin(unbondingInfo.Denom, unbondingInfo.Amount.TruncateInt())

			// Send tokens back to user
			err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(unbondAmount))
			if err != nil {
				panic(fmt.Sprintf("failed to send unbonded tokens: %s", err))
			}

			// Delete the unbonding info
			store.Delete(iterator.Key())
		}
	}
}
