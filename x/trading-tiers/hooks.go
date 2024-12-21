package tradingtiers

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gogoproto/proto"

	appparams "github.com/osmosis-labs/osmosis/v28/app/params"
	"github.com/osmosis-labs/osmosis/v28/x/trading-tiers/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v28/x/txfees/types"
	epochstypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	// Determine the osmo usd twap over the epoch

	// Get the pool id for the osmo usd pool
	var osmoUsdPoolId uint64
	k.paramSpace.Get(ctx, types.KeyOsmoUsdPoolId, &osmoUsdPoolId)

	// Get the twap for the last 24 hours
	twentyFourHourAgo := ctx.BlockTime().Add(-24 * time.Hour)
	osmoUsdTwapDec, err := k.twapKeeper.GetArithmeticTwapToNow(ctx, osmoUsdPoolId, appparams.BaseCoinUnit, "usd", twentyFourHourAgo)
	if err != nil {
		return err
	}
	osmoUsdTwap := osmoUsdTwapDec.TruncateInt()

	// Get the current epoch number
	epochInfos := k.epochsKeeper.GetEpochInfo(ctx, "day")
	lastEpochNum := epochInfos.CurrentEpoch - 1

	// Set this value in the store for the epoch number
	err = k.SetOsmoUsdValueForEpoch(ctx, lastEpochNum, osmoUsdTwap)
	if err != nil {
		return err
	}

	// Iterate over lastEpochNum for AccountDailyOsmoVolumePrefix
	// For each entry, use the current osmo usd value to determine the usd volume the account made.
	// Add this volume to the value of the respective AccountRollingWindowUSDVolumePrefix entry.
	// If this summation results in a tier increase, change the key accordingly.
	store := ctx.KVStore(k.storeKey)
	prefix := types.FormatAccountDailyOsmoVolumeDayOnly(lastEpochNum)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		account := types.AccountDailyOsmoVolumeKey(iterator.Key()).GetAccount()
		osmoVolumeProto := sdk.IntProto{}
		err := proto.Unmarshal(iterator.Value(), &osmoVolumeProto)
		if err != nil {
			return err
		}
		osmoVolume := osmoVolumeProto.Int

		// Convert osmo volume to usd volume
		usdVolume := osmoVolume.Mul(osmoUsdTwap)

		// Get the current rolling window usd volume
		rollingWindowUsdVolumeKey := types.GetAccountRollingWindowUSDVolumeKey(account)
		rollingWindowUsdVolume := types.MustUnmarshalUSDVolume(k.cdc, store.Get(rollingWindowUsdVolumeKey))

		// Add the usd volume to the rolling window usd volume
		rollingWindowUsdVolume = rollingWindowUsdVolume.Add(usdVolume)

		// Set the new rolling window usd volume
		store.Set(rollingWindowUsdVolumeKey, types.MustMarshalUSDVolume(k.cdc, rollingWindowUsdVolume))

		// Check if the new rolling window usd volume results in a tier increase
		newTier := k.GetTierFromUSDVolume(ctx, rollingWindowUsdVolume)
		oldTier := k.GetTierFromAccount(ctx, account)

		if newTier > oldTier {
			// Change the tier of the account
			k.SetTierForAccount(ctx, account, newTier)
		}
	}

	// Iterate over lastEpochNum - 31 for AccountDailyOsmoVolumePrefix
	// For each entry, use the current osmo usd value to determine the usd volume the account made.
	// Subtract this volume from the value of the respective AccountRollingWindowUSDVolumePrefix entry.
	// If this subtraction results in a tier decrease, change the key accordingly.
	// If the subtraction results in a zero value, delete the key.

	// Set the cached value for the epoch number
	return nil
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// GetModuleName implements types.EpochHooks.
func (Hooks) GetModuleName() string {
	return txfeestypes.ModuleName
}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
