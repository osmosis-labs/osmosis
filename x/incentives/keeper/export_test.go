package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

var ByGroupQueryCondition = byGroupQueryCondition

// AddGaugeRefByKey appends the provided gauge ID into an array associated with the provided key.
func (k Keeper) AddGaugeRefByKey(ctx sdk.Context, key []byte, gaugeID uint64) error {
	return k.addGaugeRefByKey(ctx, key, gaugeID)
}

// DeleteGaugeRefByKey removes the provided gauge ID from an array associated with the provided key.
func (k Keeper) DeleteGaugeRefByKey(ctx sdk.Context, key []byte, guageID uint64) error {
	return k.deleteGaugeRefByKey(ctx, key, guageID)
}

// GetGaugeRefs returns the gauge IDs specified by the provided key.
func (k Keeper) GetGaugeRefs(ctx sdk.Context, key []byte) []uint64 {
	return k.getGaugeRefs(ctx, key)
}

// GetAllGaugeIDsByDenom returns all active gauge-IDs associated with lockups of the provided denom.
func (k Keeper) GetAllGaugeIDsByDenom(ctx sdk.Context, denom string) []uint64 {
	return k.getAllGaugeIDsByDenom(ctx, denom)
}

// MoveUpcomingGaugeToActiveGauge moves a gauge that has reached it's start time from an upcoming to an active status.
func (k Keeper) MoveUpcomingGaugeToActiveGauge(ctx sdk.Context, gauge types.Gauge) error {
	return k.moveUpcomingGaugeToActiveGauge(ctx, gauge)
}

// MoveActiveGaugeToFinishedGauge moves a gauge that has completed its distribution from an active to a finished status.
func (k Keeper) MoveActiveGaugeToFinishedGauge(ctx sdk.Context, gauge types.Gauge) error {
	return k.moveActiveGaugeToFinishedGauge(ctx, gauge)
}

// ChargeFeeIfSufficientFeeDenomBalance see chargeFeeIfSufficientFeeDenomBalance spec.
func (k Keeper) ChargeFeeIfSufficientFeeDenomBalance(ctx sdk.Context, address sdk.AccAddress, fee osmomath.Int, gaugeCoins sdk.Coins) error {
	return k.chargeFeeIfSufficientFeeDenomBalance(ctx, address, fee, gaugeCoins)
}

// SyncGroupWeights updates the individual and total weights of the gauge records based on the splitting policy.
func (k Keeper) SyncGroupWeights(ctx sdk.Context, group types.Group) error {
	return k.syncGroupWeights(ctx, group)
}

// SetGauge sets the regular gauge to state.
func (k Keeper) SetGauge(ctx sdk.Context, gauge *types.Gauge) error {
	return k.setGauge(ctx, gauge)
}

// exporting an internal helper for testing
func (k Keeper) AddToGaugeRewardsInternal(ctx sdk.Context, coins sdk.Coins, gaugeID uint64) error {
	return k.addToGaugeRewards(ctx, coins, gaugeID)
}

// SyncVolumeSplitGroup updates the individual and total weights of the gauge records based on the volume splitting policy.
func (k Keeper) SyncVolumeSplitGroup(ctx sdk.Context, volumeSplitGauge types.Group) error {
	return k.syncVolumeSplitGroup(ctx, volumeSplitGauge)
}

func (k Keeper) HandleGroupPostDistribute(ctx sdk.Context, groupGauge types.Gauge, coinsDistributed sdk.Coins) error {
	return k.handleGroupPostDistribute(ctx, groupGauge, coinsDistributed)
}

func (k Keeper) InitGaugeInfo(ctx sdk.Context, poolIds []uint64) (types.InternalGaugeInfo, error) {
	return k.initGaugeInfo(ctx, poolIds)
}

func RegularGaugeStoreKey(ID uint64) []byte {
	return gaugeStoreKey(ID)
}

func CombineKeys(keys ...[]byte) []byte {
	return combineKeys(keys...)
}

func GetTimeKeys(timestamp time.Time) []byte {
	return getTimeKey(timestamp)
}

func (k Keeper) ChargeGroupCreationFeeIfNotWhitelisted(ctx sdk.Context, sender sdk.AccAddress) (chargedFee bool, err error) {
	return k.chargeGroupCreationFeeIfNotWhitelisted(ctx, sender)
}

func (k Keeper) CreateGroupInternal(ctx sdk.Context, coins sdk.Coins, numEpochPaidOver uint64, owner sdk.AccAddress, poolIDs []uint64) (types.Group, error) {
	return k.createGroup(ctx, coins, numEpochPaidOver, owner, poolIDs)
}

func (k Keeper) CalculateGroupWeights(ctx sdk.Context, group types.Group) (types.Group, error) {
	return k.calculateGroupWeights(ctx, group)
}

func (k Keeper) GetNoLockGaugeUptime(ctx sdk.Context, gauge types.Gauge, poolId uint64) time.Duration {
	return k.getNoLockGaugeUptime(ctx, gauge, poolId)
}

func (k Keeper) SkipSpamGaugeDistribute(ctx sdk.Context, locks []*lockuptypes.PeriodLock, gauge types.Gauge, totalDistrCoins sdk.Coins, remainCoins sdk.Coins) (bool, sdk.Coins, error) {
	return k.skipSpamGaugeDistribute(ctx, locks, gauge, totalDistrCoins, remainCoins)
}

func (k Keeper) CheckIfDenomsAreDistributable(ctx sdk.Context, coins sdk.Coins) error {
	return k.checkIfDenomsAreDistributable(ctx, coins)
}
