package parsing_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/json"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mocks"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/routertesting"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/router/usecase/routertesting/parsing"
	concentratedmodel "github.com/osmosis-labs/osmosis/v21/x/concentrated-liquidity/model"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

const testFileName = "pools.json"

var (
	zeroMinOSMOTVL                 = osmomath.ZeroInt()
	testPoolToMarshal domain.PoolI = &mocks.MockRoutablePool{
		ChainPoolModel: &concentratedmodel.Pool{
			Id:                   1,
			Token0:               routertesting.Denom0,
			Token1:               routertesting.Denom1,
			CurrentTickLiquidity: routertesting.DefaultLiquidityAmt,
			CurrentTick:          routertesting.DefaultCurrentTick,
			TickSpacing:          1,
			LastLiquidityUpdate:  time.Unix(1, 1).UTC(),
			SpreadFactor:         routertesting.DefaultSpreadFactor,
			CurrentSqrtPrice:     osmomath.OneBigDec(),
		},
		TotalValueLockedUSDC: osmomath.OneInt(),
		Balances:             routertesting.DefaultPoolBalances,
		Denoms:               []string{routertesting.Denom0, routertesting.Denom1},
		SpreadFactor:         routertesting.DefaultSpreadFactor,
		PoolType:             poolmanagertypes.Concentrated,
	}

	defaultTickModel = domain.TickModel{
		Ticks: []domain.LiquidityDepthsWithRange{
			{
				LiquidityAmount: osmomath.OneDec(),
				LowerTick:       1,
				UpperTick:       2,
			},
		},
		CurrentTickIndex: 0,
		HasNoLiquidity:   false,
	}
)

// This test validates that ReadPools can read a file from the state.
func TestReadPoolsFileFromState(t *testing.T) {
	t.Skip("This test is not meant to be run in CI. Use for debugging only")

	pools, _, err := parsing.ReadPools(testFileName)
	require.NoError(t, err)

	require.NotEmpty(t, pools)
	require.Greater(t, len(pools), 500)

	for _, pool := range pools {
		err := pool.Validate(zeroMinOSMOTVL)
		if err != nil {
			t.Logf("pool %d failed validation: %s", pool.GetId(), err)
		}
	}
}

// This test validates that StorePools succesfull stores pools to a file
// that ReadPools can then read back into the system.
func TestStoreFilesAndReadBack(t *testing.T) {

	t.Skip("This test is not meant to be run in CI. Use for debugging only")

	// Delete test file if exists since the system under test recreates it.
	_, err := os.Stat(testFileName)
	if err == nil {
		err = os.Remove(testFileName)
		require.NoError(t, err)
	}

	err = parsing.StorePools([]domain.PoolI{testPoolToMarshal}, map[uint64]domain.TickModel{
		testPoolToMarshal.GetId(): defaultTickModel,
	}, testFileName)
	require.NoError(t, err)

	pools, _, err := parsing.ReadPools(testFileName)
	require.NoError(t, err)

	require.Equal(t, 1, len(pools))
	for _, pool := range pools {
		require.NoError(t, pool.Validate(zeroMinOSMOTVL))
	}
}

// This test validates that unmarshalling and marshalling a pool works as expected.
func TestMarshalUnmarshalPool(t *testing.T) {
	serializedPools, err := parsing.MarshalPool(testPoolToMarshal)
	require.NoError(t, err)

	var interimPools parsing.SerializedPool
	err = json.Unmarshal(serializedPools, &interimPools)
	require.NoError(t, err)

	unmarshalledPool, err := parsing.UnmarshalPool(interimPools)
	require.NoError(t, err)

	require.Equal(t, testPoolToMarshal.GetUnderlyingPool(), unmarshalledPool.GetUnderlyingPool())
	require.Equal(t, testPoolToMarshal.GetSQSPoolModel(), unmarshalledPool.GetSQSPoolModel())
}

// This test validates that unmarshalling and marshalling a taker fee map works as expected.
func TestMarshalUnmarshalTakerFeeMap(t *testing.T) {
	takerFeeMap := domain.TakerFeeMap{
		domain.DenomPair{
			Denom0: routertesting.Denom0,
			Denom1: routertesting.Denom1,
		}: osmomath.OneDec(),
	}

	takerFeeMapBz, err := json.Marshal(takerFeeMap)
	require.NoError(t, err)

	unmarshalledTakerFeeMap := domain.TakerFeeMap{}
	err = json.Unmarshal(takerFeeMapBz, &unmarshalledTakerFeeMap)
	require.NoError(t, err)

	require.Equal(t, takerFeeMap, unmarshalledTakerFeeMap)
}
