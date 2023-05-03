package types

import (
<<<<<<< HEAD
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
=======
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
>>>>>>> 560224f5 (refactor: use cosmossdk.io/errors (#5065))
)

// The following regiisters various pool-incentives errors.
var (
	ErrNoGaugeIdExist                = errorsmod.Register(ModuleName, 1, "no gauge id exist")
	ErrDistrRecordNotPositiveWeight  = errorsmod.Register(ModuleName, 2, "weight in record should be positive")
	ErrDistrRecordNotRegisteredGauge = errorsmod.Register(ModuleName, 3, "gauge was not registered")
	ErrDistrRecordRegisteredGauge    = errorsmod.Register(ModuleName, 4, "gauge was already registered")
	ErrDistrRecordNotSorted          = errorsmod.Register(ModuleName, 5, "gauges are not sorted")

	ErrEmptyProposalRecords  = errorsmod.Register(ModuleName, 10, "records are empty")
	ErrEmptyProposalGaugeIds = errorsmod.Register(ModuleName, 11, "gauge ids are empty")
)
