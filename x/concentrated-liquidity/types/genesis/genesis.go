package genesis

import (
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

// DefaultGenesis returns the default GenesisState for the concentrated-liquidity module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		PoolData:              []PoolData{},
		Params:                types.DefaultParams(),
		NextPositionId:        1,
		NextIncentiveRecordId: 1,
		// By default, the migration threshold is set to 0, which means all pools are migrated.
		IncentivesAccumulatorPoolIdMigrationThreshold: 0,
		SpreadFactorPoolIdMigrationThreshold:          0,
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if gs.NextPositionId == 0 {
		return types.InvalidNextPositionIdError{NextPositionId: gs.NextPositionId}
	}
	if gs.NextIncentiveRecordId == 0 {
		return types.InvalidNextIncentiveRecordIdError{NextIncentiveRecordId: gs.NextIncentiveRecordId}
	}
	return nil
}
