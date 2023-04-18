package types

import (
	"fmt"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// The following regiisters various pool-incentives errors.
var (
	ErrNoGaugeIdExist                = sdkerrors.Register(ModuleName, 1, "no gauge id exist")
	ErrDistrRecordNotPositiveWeight  = sdkerrors.Register(ModuleName, 2, "weight in record should be positive")
	ErrDistrRecordNotRegisteredGauge = sdkerrors.Register(ModuleName, 3, "gauge was not registered")
	ErrDistrRecordRegisteredGauge    = sdkerrors.Register(ModuleName, 4, "gauge was already registered")
	ErrDistrRecordNotSorted          = sdkerrors.Register(ModuleName, 5, "gauges are not sorted")

	ErrEmptyProposalRecords  = sdkerrors.Register(ModuleName, 10, "records are empty")
	ErrEmptyProposalGaugeIds = sdkerrors.Register(ModuleName, 11, "gauge ids are empty")
)

type NoGaugeAssociatedWithPoolError struct {
	PoolId   uint64
	Duration time.Duration
}

func (e NoGaugeAssociatedWithPoolError) Error() string {
	return fmt.Sprintf("no gauge associated with pool id (%d) and duration (%d)", e.PoolId, e.Duration)
}

type NoPoolAssociatedWithGaugeError struct {
	GaugeId  uint64
	Duration time.Duration
}

func (e NoPoolAssociatedWithGaugeError) Error() string {
	return fmt.Sprintf("no pool associated with gauge id (%d) and duration (%d)", e.GaugeId, e.Duration)
}
