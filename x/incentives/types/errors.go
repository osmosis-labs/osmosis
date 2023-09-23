package types

import (
	fmt "fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
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
