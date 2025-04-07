package types

import (
	"fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	epochtypes "github.com/osmosis-labs/osmosis/v27/x/epochs/types"
)

// Parameter store keys.
var (
	KeyPruneEpochIdentifier    = []byte("PruneEpochIdentifier")
	KeyRecordHistoryKeepPeriod = []byte("RecordHistoryKeepPeriod")

	_ paramtypes.ParamSet = &Params{}
)

const (
	defaultPruneEpochIdentifier    = "day"
	defaultRecordHistoryKeepPeriod = 48 * time.Hour
)

// ParamTable for twap module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(pruneEpochIdentifier string, recordHistoryKeepPeriod time.Duration) Params {
	return Params{
		PruneEpochIdentifier:    pruneEpochIdentifier,
		RecordHistoryKeepPeriod: recordHistoryKeepPeriod,
	}
}

// default twap module parameters.
func DefaultParams() Params {
	return Params{
		PruneEpochIdentifier:    defaultPruneEpochIdentifier,
		RecordHistoryKeepPeriod: defaultRecordHistoryKeepPeriod,
	}
}

// validate params.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierString(p.PruneEpochIdentifier); err != nil {
		return err
	}

	if err := validatePeriod(p.RecordHistoryKeepPeriod); err != nil {
		return err
	}

	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPruneEpochIdentifier, &p.PruneEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(KeyRecordHistoryKeepPeriod, &p.RecordHistoryKeepPeriod, validatePeriod),
	}
}

func validatePeriod(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("time must be positive: %d", v)
	}

	return nil
}
