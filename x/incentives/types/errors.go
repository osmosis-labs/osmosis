package types

import (
	fmt "fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

var (
	ErrNoPoolIDsGiven        = fmt.Errorf("no pool IDs given")
	ErrZeroNumEpochsPaidOver = fmt.Errorf("num epochs paid over must be greater than zero for non-perpetual gauges")
	ErrUnauthorized          = fmt.Errorf("unauthorized to perform this action. Must be an incentives module account")
)

type UnsupportedSplittingPolicyError struct {
	GroupGaugeId    uint64
	SplittingPolicy SplittingPolicy
}

func (e UnsupportedSplittingPolicyError) Error() string {
	return fmt.Sprintf("Attempted to sync group gauge (%d) with unsupported splitting policy: %s", e.GroupGaugeId, e.SplittingPolicy)
}

type NoPoolVolumeError struct {
	PoolId uint64
}

func (e NoPoolVolumeError) Error() string {
	return fmt.Sprintf("Pool %d has no volume.", e.PoolId)
}

type CumulativeVolumeDecreasedError struct {
	PoolId         uint64
	PreviousVolume osmomath.Int
	NewVolume      osmomath.Int
}

func (e CumulativeVolumeDecreasedError) Error() string {
	return fmt.Sprintf("Cumulative volume should not be able to decrease. Pool id (%d), previous volume (%s), new volume (%s)", e.PoolId, e.PreviousVolume, e.NewVolume)
}

type UnexpectedFinishedGaugeError struct {
	GaugeId uint64
}

func (e UnexpectedFinishedGaugeError) Error() string {
	return fmt.Sprintf("gauge with ID (%d) is already finished", e.GaugeId)
}

type GroupNotFoundError struct {
	GroupGaugeId uint64
}

func (e GroupNotFoundError) Error() string {
	return fmt.Sprintf("group with gauge ID (%d) not found", e.GroupGaugeId)
}

type GaugeNotFoundError struct {
	GaugeID uint64
}

func (e GaugeNotFoundError) Error() string {
	return fmt.Sprintf("gauge with ID (%d) not found", e.GaugeID)
}

type OnePoolIDGroupError struct {
	PoolID uint64
}

func (e OnePoolIDGroupError) Error() string {
	return fmt.Sprintf("one pool ID %d given. Need at least two to create valid Group", e.PoolID)
}

type GroupTotalWeightZeroError struct {
	GroupID uint64
}

func (e GroupTotalWeightZeroError) Error() string {
	return fmt.Sprintf("Group with ID %d has total weight of zero", e.GroupID)
}

type InvalidGaugeTypeError struct {
	GaugeType lockuptypes.LockQueryType
}

func (e InvalidGaugeTypeError) Error() string {
	return fmt.Sprintf("invalid gauge type: %s", e.GaugeType)
}

type NoVolumeSinceLastSyncError struct {
	PoolID uint64
}

func (e NoVolumeSinceLastSyncError) Error() string {
	return fmt.Sprintf("Pool %d has no volume since last sync", e.PoolID)
}

type DuplicatePoolIDError struct {
	PoolIDs []uint64
}

func (e DuplicatePoolIDError) Error() string {
	return fmt.Sprintf("one or more pool IDs provided in the pool ID array contains a duplicate: %d", e.PoolIDs)
}

type NoRouteForDenomError struct {
	Denom string
}

func (e NoRouteForDenomError) Error() string {
	return fmt.Sprintf("denom %s does not exist as a protorev hot route, therefore, the value of rewards at time of epoch distribution will not be able to be determined", e.Denom)
}
