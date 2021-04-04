package keeper

import (
	"time"

	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/c-osmosis/osmosis/x/gamm/types"
)

type Hooks struct {
	k Keeper
}

var _ gammtypes.GammHooks = Hooks{}
var _ lockuptypes.LockupHooks = Hooks{}

// Create new pool yield hooks
func (k Keeper) Hooks() Hooks { return Hooks{k} }

// creates a farm for each pool’s lockable duration
func (h Hooks) AfterPoolCreated(ctx sdk.Context, poolId uint64) {
	err := h.k.CreatePoolFarms(ctx, poolId)
	if err != nil {
		panic(err)
	}
}

func (h Hooks) OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	// If the locked token in the lockup module is a pool’s share, attempt to add/remove the share to the farm’s pool
	for _, coin := range amount {
		poolId, ok := gammtypes.GetPoolIdFromShareDenom(coin.Denom)
		if ok {
			farmId, err := h.k.GetPoolFarmId(ctx, poolId, lockDuration)

			// The duration of the lockup module doesn’t have to match the pool-yield keeper’s lockableDurations.
			// Therefore, it is possible to not be able to retrieve the farm id from the pool id and lock duration.
			if err == nil {
				farm, err := h.k.farmKeeper.GetFarm(ctx, farmId)
				if err != nil {
					panic(err)
				}

				// Note that the Farm module doesn’t custody shares within the module, and leaves other modules to manage the balance.
				// In this case, the shares are not managed in the pool-yield module, but the lockup module.
				_, err = h.k.farmKeeper.DepositShareToFarm(ctx, farm.FarmId, address, coin.Amount)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func (h Hooks) OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time) {
	// If the locked token in the lockup module is a pool’s share, attempt to add/remove the share to the farm’s pool
	for _, coin := range amount {
		poolId, ok := gammtypes.GetPoolIdFromShareDenom(coin.Denom)
		if ok {
			farmId, err := h.k.GetPoolFarmId(ctx, poolId, lockDuration)

			// The duration of the lockup module doesn’t have to match the pool-yield keeper’s lockableDurations.
			// Therefore, it is possible to not be able to retrieve the farm id from the pool id and lock duration.
			if err == nil {
				farm, err := h.k.farmKeeper.GetFarm(ctx, farmId)
				if err != nil {
					panic(err)
				}

				// Note that the Farm module doesn’t custody shares within the module, and leaves other modules to manage the balance.
				// In this case, the shares are not managed in the pool-yield module, but the lockup module.
				_, err = h.k.farmKeeper.WithdrawShareFromFarm(ctx, farm.FarmId, address, coin.Amount)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
