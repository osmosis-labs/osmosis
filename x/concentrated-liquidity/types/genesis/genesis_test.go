package genesis_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types/genesis"
)

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name           string
		genesis        genesis.GenesisState
		exepectedError bool
	}{
		{
			name:           "default genesis",
			genesis:        *genesis.DefaultGenesis(),
			exepectedError: false,
		},
		{
			name: "invalid params",
			genesis: *&genesis.GenesisState{
				Params:                types.Params{},
				PoolData:              genesis.DefaultGenesis().PoolData,
				NextPositionId:        genesis.DefaultGenesis().GetNextPositionId(),
				NextIncentiveRecordId: genesis.DefaultGenesis().GetNextIncentiveRecordId(),
			},
			exepectedError: true,
		},
		{
			name: "next position id is zero",
			genesis: *&genesis.GenesisState{
				Params:                genesis.DefaultGenesis().GetParams(),
				PoolData:              genesis.DefaultGenesis().PoolData,
				NextPositionId:        0,
				NextIncentiveRecordId: genesis.DefaultGenesis().GetNextIncentiveRecordId(),
			},
			exepectedError: true,
		},
		{
			name: "next incentive record id is zero",
			genesis: *&genesis.GenesisState{
				Params:                genesis.DefaultGenesis().GetParams(),
				PoolData:              genesis.DefaultGenesis().PoolData,
				NextPositionId:        genesis.DefaultGenesis().GetNextPositionId(),
				NextIncentiveRecordId: 0,
			},
			exepectedError: true,
		},
	}

	for _, test := range tests {
		err := test.genesis.Validate()
		if test.exepectedError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
