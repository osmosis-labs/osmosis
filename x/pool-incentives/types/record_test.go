package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

<<<<<<< HEAD
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v18/x/pool-incentives/types"
=======
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/pool-incentives/types"
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
)

// TestDistrRecord is a test on the weights of distribution gauges.
func TestDistrRecord(t *testing.T) {
	zeroWeight := types.DistrRecord{
		GaugeId: 1,
		Weight:  osmomath.NewInt(0),
	}

	require.NoError(t, zeroWeight.ValidateBasic())

	negativeWeight := types.DistrRecord{
		GaugeId: 1,
		Weight:  osmomath.NewInt(-1),
	}

	require.Error(t, negativeWeight.ValidateBasic())

	positiveWeight := types.DistrRecord{
		GaugeId: 1,
		Weight:  osmomath.NewInt(1),
	}

	require.NoError(t, positiveWeight.ValidateBasic())
}
