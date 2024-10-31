package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

var emptyCoins = sdk.NewCoins()

// CreateGroup creates a new group. The group is 1:1 mapped to a group gauge that allocates rewards dynamically across its internal pool gauges based on
// the volume splitting policy.
// For each pool ID in the given slice, its main internal gauge is used to create gauge records to be associated with the Group.
// Note, that implies that only perpetual pool gauges can be associated with the Group.
// For Group's own distribution policy, a 1:1 group Gauge is created. This is the Gauge that receives incentives at the end of an epoch
// in the pool incentives as defined by the DistrRecord. The Group's Gauge can either be perpetual or non-perpetual.
// If numEpochPaidOver is 0, then the Group's Gauge is perpetual. Otherwise, it is non-perpetual.
// It syncs the group's weights at the time of creation. This is useful for validating that all the pools
// in the group are valid and have the associated volume at group creation time.
// Charges group creation fee, unless incentives module account.
// Returns nil on success.
// Returns error if:
// - given pool IDs slice is empty or has 1 pool only
// - fails to initialize gauge information for every pool ID
// - fails to send coins from owner to the incentives module for the Group's Gauge
// - fails to charge group creation fee
// - fails to set the Group's Gauge to state
func (k Keeper) CreateGroup(ctx sdk.Context, coins sdk.Coins, numEpochPaidOver uint64, owner sdk.AccAddress, poolIDs []uint64) (uint64, error) {
	newGroup, err := k.createGroup(ctx, coins, numEpochPaidOver, owner, poolIDs)
	if err != nil {
		return 0, err
	}

	// Note: we rely on the syncing logic to persist the group to state
	// if updated successfully.
	// The reason we sync is to make sure that all pools in the group are valid
	// and have the associated volume at group creation time. This prevents
	// creating groups of pools that are invalid.
	// Contrary to distribution logic that silently skips the error, we bubble it up here
	// to fail the creation message.
	// The group is saved to state upon successful sync.
	if err := k.syncGroupWeights(ctx, newGroup); err != nil {
		return 0, err
	}

	return newGroup.GroupGaugeId, nil
}

// CreateGroupAsIncentivesModuleAcc creates a group as incentives module account.
// The group is 1:1 mapped to a group gauge that allocates rewards dynamically across its internal pool gauges based on
// the volume splitting policy.
// For each pool ID in the given slice, its main internal gauge is used to create gauge records to be associated with the Group.
// Note, that implies that only perpetual pool gauges can be associated with the Group.
// For Group's own distribution policy, a 1:1 group Gauge is created. This is the Gauge that receives incentives at the end of an epoch
// in the pool incentives as defined by the DistrRecord. The Group's Gauge can either be perpetual or non-perpetual.
// If numEpochPaidOver is 0, then the Group's Gauge is perpetual. Otherwise, it is non-perpetual.
// The group is created with empty coins. It does not sync weights at the time
// of creation. This is useful for creating groups of pools in a privileged way
// For example, in the upgrade handler.
// Use with care since it is possible to create a group of pools that are invalid (have no volume) due to lack of syncing.
// Stores the group to state.
// The group creation fee is not charged on the incentives module account.
// See other details of group creation by reviewing createGroup() spec.
// Returns group gauge ID on success.
// Returns error if:
// - fails to create Group
func (k Keeper) CreateGroupAsIncentivesModuleAcc(ctx sdk.Context, numEpochPaidOver uint64, poolIDs []uint64) (uint64, error) {
	incentivesModuleAddress := k.ak.GetModuleAddress(types.ModuleName)
	newGroup, err := k.createGroup(ctx, emptyCoins, numEpochPaidOver, incentivesModuleAddress, poolIDs)
	if err != nil {
		return 0, err
	}

	// Store the group to state.
	k.SetGroup(ctx, newGroup)

	return newGroup.GroupGaugeId, nil
}

// createGroup creates a new group. The group is 1:1 mapped to a group gauge that allocates rewards dynamically across its internal pool gauges based on
// the volume splitting policy.
// For each pool ID in the given slice, its main internal gauge is used to create gauge records to be associated with the Group.
// Note, that implies that only perpetual pool gauges can be associated with the Group.
// For Group's own distribution policy, a 1:1 group Gauge is created. This is the Gauge that receives incentives at the end of an epoch
// in the pool incentives as defined by the DistrRecord. The Group's Gauge can either be perpetual or non-perpetual.
// If numEpochPaidOver is 0, then the Group's Gauge is perpetual. Otherwise, it is non-perpetual.
// It does not attempt to sync the group's weights. Use with care since it is
// possible to create a group of pools that are invalid (have no volume).
// Returns nil on success.
// Returns error if:
// - given pool IDs slice is empty or has 1 pool only
// - fails to initialize gauge information for every pool ID
// - fails to send coins from owner to the incentives module for the Group's Gauge
// - fails to charge group creation fee
// - fails to set the Group's Gauge to state
//
// Notes:
// - does not persist the group to state
// - persists group's Gauge to state
// - does not charge group creation fee if sender is the incentives module account
func (k Keeper) createGroup(ctx sdk.Context, coins sdk.Coins, numEpochPaidOver uint64, owner sdk.AccAddress, poolIDs []uint64) (types.Group, error) {
	if len(poolIDs) == 0 {
		return types.Group{}, types.ErrNoPoolIDsGiven
	}
	if len(poolIDs) == 1 {
		return types.Group{}, types.OnePoolIDGroupError{PoolID: poolIDs[0]}
	}

	if !osmoassert.Uint64ArrayValuesAreUnique(poolIDs) {
		return types.Group{}, types.DuplicatePoolIDError{PoolIDs: poolIDs}
	}

	// Initialize gauge information for every pool ID.
	initialInternalGaugeInfo, err := k.initGaugeInfo(ctx, poolIDs)
	if err != nil {
		return types.Group{}, err
	}

	// Charge group creation fee.
	_, err = k.chargeGroupCreationFeeIfNotWhitelisted(ctx, owner)
	if err != nil {
		return types.Group{}, err
	}

	groupGaugeID, err := k.CreateGauge(ctx, numEpochPaidOver == types.PerpetualNumEpochsPaidOver, owner, coins, byGroupQueryCondition, ctx.BlockTime(), numEpochPaidOver, 0)
	if err != nil {
		return types.Group{}, err
	}

	newGroup := types.Group{
		GroupGaugeId:      groupGaugeID,
		InternalGaugeInfo: initialInternalGaugeInfo,
		// Note: only Volume splitting exists today.
		// We allow for other splitting policies to be added in the future
		// by extending the enum.
		SplittingPolicy: types.ByVolume,
	}

	return newGroup, nil
}

// initGaugeInfo takes in a list of pool IDs and returns a InternalGaugeInfo struct with weights initialized to zero.
// Returns error if fails to retrieve gauge ID for a pool.
func (k Keeper) initGaugeInfo(ctx sdk.Context, poolIds []uint64) (types.InternalGaugeInfo, error) {
	gaugeRecords := make([]types.InternalGaugeRecord, 0, len(poolIds))
	for _, poolID := range poolIds {
		gaugeID, err := k.pik.GetInternalGaugeIDForPool(ctx, poolID)
		if err != nil {
			return types.InternalGaugeInfo{}, err
		}

		gaugeRecords = append(gaugeRecords, types.InternalGaugeRecord{
			GaugeId:          gaugeID,
			CurrentWeight:    osmomath.ZeroInt(),
			CumulativeWeight: osmomath.ZeroInt(),
		})
	}

	return types.InternalGaugeInfo{
		TotalWeight:  osmomath.ZeroInt(),
		GaugeRecords: gaugeRecords,
	}, nil
}

// chargeGroupCreationFeeIfNotWhitelisted charges fee as defined in the params if the sender is not whitelisted.
// Does not charge fee if sender is the incentives module account or if sender is whitelisted.
// Returns true if charged fee, false otherwise.
// Returns error if:
// - One if the addresses in params is invalid
// - fails to send coins from sender to the community pool
func (k Keeper) chargeGroupCreationFeeIfNotWhitelisted(ctx sdk.Context, sender sdk.AccAddress) (chargedFee bool, err error) {
	params := k.GetParams(ctx)
	incentivesModuleAddress := k.ak.GetModuleAddress(types.ModuleName)

	// don't charge fee if sender is the incentives module account
	if sender.Equals(incentivesModuleAddress) {
		return false, nil
	}

	for _, unrestrictedAddressStr := range params.UnrestrictedCreatorWhitelist {
		unrestrictedAddress, err := sdk.AccAddressFromBech32(unrestrictedAddressStr)
		if err != nil {
			return false, err
		}

		// don't charge fee if sender is in the whitelist
		if unrestrictedAddress.Equals(sender) {
			return false, nil
		}
	}

	// Charge fee
	groupCreationFee := params.GroupCreationFee
	if err := k.bk.SendCoinsFromAccountToModule(ctx, sender, distrtypes.ModuleName, groupCreationFee); err != nil {
		return false, err
	}
	return true, nil
}

// GetPoolIdsAndDurationsFromGaugeRecords retrieves the pool IDs and their associated durations from a group's gauge records
// It iterates over each record and retrieves the pool ID and duration.
// The function returns two slices: one for the pool IDs and one for the durations. The indices in these slices correspond to each other.
// If there is an error retrieving the pool ID and duration for any gauge record, the function returns an error.
func (k Keeper) GetPoolIdsAndDurationsFromGaugeRecords(ctx sdk.Context, gaugeRecords []types.InternalGaugeRecord) ([]uint64, []time.Duration, error) {
	poolIds := make([]uint64, 0, len(gaugeRecords))
	durations := make([]time.Duration, 0, len(gaugeRecords))
	for _, gaugeRecord := range gaugeRecords {
		poolId, gaugeDuration, err := k.GetPoolIdAndDurationFromGaugeRecord(ctx, gaugeRecord)
		if err != nil {
			return nil, nil, err
		}
		poolIds = append(poolIds, poolId)
		durations = append(durations, gaugeDuration)
	}
	return poolIds, durations, nil
}

// GetPoolIdAndDurationFromGaugeRecord retrieves the pool ID and duration associated with a given gauge record.
// The function first retrieves the gauge associated with the gauge record.
// If the gauge's lock query type is NoLock, the function sets the gauge duration to the epoch duration.
// Otherwise, it sets the gauge duration to the longest lockable duration.
// The function then retrieves the pool ID associated with the gauge ID and the gauge duration.
// The function returns the pool ID and the gauge duration.
// If there is an error retrieving the gauge, the longest lockable duration, or the pool ID, the function returns an error.
func (k Keeper) GetPoolIdAndDurationFromGaugeRecord(ctx sdk.Context, gaugeRecord types.InternalGaugeRecord) (uint64, time.Duration, error) {
	gauge, err := k.GetGaugeByID(ctx, gaugeRecord.GaugeId)
	if err != nil {
		return 0, 0, err
	}
	gaugeType := gauge.DistributeTo.LockQueryType
	gaugeDuration := time.Duration(0)

	if gaugeType == lockuptypes.NoLock {
		// If NoLock, it's a CL pool, so we set the "lockableDuration" to epoch duration
		gaugeDuration = k.GetEpochInfo(ctx).Duration
	} else if gaugeType == lockuptypes.ByDuration {
		// Otherwise, it's a balancer pool so we set it to longest lockable duration
		// TODO: add support for CW pools once there's clarity around default gauge type.
		// Tracked in issue https://github.com/osmosis-labs/osmosis/issues/6403
		gaugeDuration, err = k.pik.GetLongestLockableDuration(ctx)
		if err != nil {
			return 0, 0, err
		}
	} else {
		return 0, 0, types.InvalidGaugeTypeError{GaugeType: gaugeType}
	}

	// Retrieve pool ID using GetPoolIdFromGaugeId(gaugeId, lockableDuration)
	poolId, err := k.pik.GetPoolIdFromGaugeId(ctx, gaugeRecord.GaugeId, gaugeDuration)
	if err != nil {
		return 0, 0, err
	}
	return poolId, gaugeDuration, nil
}
