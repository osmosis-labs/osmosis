package types

import (
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
	return Params{}
}

// default twap module parameters.
func DefaultParams() Params {
	return Params{}
}

// Validate validate params.
func (p Params) Validate() error {
	return nil
}

// Implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validate genesis state.
func (g GenesisState) Validate() error {
	return g.Params.Validate()
}
