package concentrated_pool_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"

	concentrated_pool "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/concentrated-pool"
)

type ConcentratedPoolTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestConcentratedPoolTestSuite(t *testing.T) {
	suite.Run(t, new(ConcentratedPoolTestSuite))
}

func TestSpotPrice(t *testing.T) {
	defaultSpotPrice := sdk.MustNewDecFromStr("0.2")
	defaultSqrtSpotPrice, err := defaultSpotPrice.ApproxSqrt()
	require.NoError(t, err)
	reverseSpotPirce := sdk.NewDec(1).Quo(defaultSpotPrice)
	defaultToken0 := "tokenA"
	defaultToken1 := "tokenB"
	randomToken := "random"

	testCases := []struct {
		baseDenom   string
		quoteDenom  string
		expectedSp  sdk.Dec
		expectedErr bool
	}{
		{defaultToken0, defaultToken1, defaultSpotPrice, false},
		{defaultToken1, defaultToken0, reverseSpotPirce, false},
		{defaultToken0, defaultToken0, reverseSpotPirce, true},
		{defaultToken0, randomToken, sdk.ZeroDec(), true},
	}

	mock_pool := concentrated_pool.Pool{
		CurrentSqrtPrice: defaultSqrtSpotPrice,
		Token0:           defaultToken0,
		Token1:           defaultToken1,
	}

	for _, tc := range testCases {
		sp, err := mock_pool.SpotPrice(sdk.Context{}, tc.baseDenom, tc.quoteDenom)
		if tc.expectedErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)

			// we use elipson due to sqrt approximation
			elipson := sdk.MustNewDecFromStr("0.0000000000000001")
			require.True(t, sp.Sub(tc.expectedSp).Abs().LT(elipson))
		}
	}
}

func TestUpdateLiquidity(t *testing.T) {
	defaultLiquidity := sdk.NewDec(100)
	mock_pool := concentrated_pool.Pool{
		Liquidity: defaultLiquidity,
	}

	// try updating it with zero dec
	mock_pool.UpdateLiquidity(sdk.ZeroDec())
	require.Equal(t, defaultLiquidity, mock_pool.Liquidity)

	// try adding 10 to pool liquidity
	mock_pool.UpdateLiquidity(sdk.NewDec(10))
	require.Equal(t, defaultLiquidity.Add(sdk.NewDec(10)), mock_pool.Liquidity)
}

func (s *ConcentratedPoolTestSuite) TestApplySwap() {
	s.Setup()

	defaultLiquidity := sdk.NewDec(100)
	defaultCurrTick := sdk.NewInt(1)
	defaultCurrSqrtPrice := sdk.NewDec(5)

	mock_pool := concentrated_pool.Pool{
		Liquidity:        defaultLiquidity,
		CurrentTick:      defaultCurrTick,
		CurrentSqrtPrice: defaultCurrSqrtPrice,
	}

	newLiquidity := defaultLiquidity.Mul(sdk.NewDec(2))
	newCurrTick := defaultCurrTick.Mul(sdk.NewInt(2))
	newCurrSqrtPrice := defaultCurrSqrtPrice.Mul(sdk.NewDec(2))

	gasBeforeSwap := s.Ctx.GasMeter().GasConsumed()
	mock_pool.ApplySwap(s.Ctx, mock_pool.GetId(), newLiquidity, newCurrTick, newCurrSqrtPrice)
	gasAfterSwap := s.Ctx.GasMeter().GasConsumed()

	s.Require().Equal(gasAfterSwap-gasBeforeSwap, uint64(gammtypes.BalancerGasFeeForSwap))
	s.Require().Equal(mock_pool.Liquidity, newLiquidity)
	s.Require().Equal(mock_pool.CurrentTick, newCurrTick)
	s.Require().Equal(mock_pool.CurrentSqrtPrice, newCurrSqrtPrice)

}
