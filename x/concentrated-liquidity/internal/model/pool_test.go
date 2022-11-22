package model_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/internal/model"
)

type ConcentratedPoolTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestConcentratedPoolTestSuite(t *testing.T) {
	suite.Run(t, new(ConcentratedPoolTestSuite))
}

func (s *ConcentratedPoolTestSuite) TestSpotPrice() {
	defaultSpotPrice := sdk.MustNewDecFromStr("0.2")
	defaultSqrtSpotPrice, err := defaultSpotPrice.ApproxSqrt()
	s.Require().NoError(err)
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

	mock_pool := model.Pool{
		CurrentSqrtPrice: defaultSqrtSpotPrice,
		Token0:           defaultToken0,
		Token1:           defaultToken1,
	}

	for _, tc := range testCases {
		sp, err := mock_pool.SpotPrice(sdk.Context{}, tc.baseDenom, tc.quoteDenom)
		if tc.expectedErr {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)

			// we use elipson due to sqrt approximation
			elipson := sdk.MustNewDecFromStr("0.0000000000000001")
			s.Require().True(sp.Sub(tc.expectedSp).Abs().LT(elipson))
		}
	}
}

func (s *ConcentratedPoolTestSuite) TestUpdateLiquidity() {
	defaultLiquidity := sdk.NewDec(100)
	mock_pool := model.Pool{
		Liquidity: defaultLiquidity,
	}

	// try updating it with zero dec
	mock_pool.UpdateLiquidity(sdk.ZeroDec())

	s.Require().Equal(defaultLiquidity, mock_pool.Liquidity)

	// try adding 10 to pool liquidity
	mock_pool.UpdateLiquidity(sdk.NewDec(10))
	s.Require().Equal(defaultLiquidity.Add(sdk.NewDec(10)), mock_pool.Liquidity)
}

func (s *ConcentratedPoolTestSuite) TestApplySwap() {
	s.Setup()

	defaultLiquidity := sdk.NewDec(100)
	defaultCurrTick := sdk.NewInt(1)
	defaultCurrSqrtPrice := sdk.NewDec(5)

	mock_pool := model.Pool{
		Liquidity:        defaultLiquidity,
		CurrentTick:      defaultCurrTick,
		CurrentSqrtPrice: defaultCurrSqrtPrice,
	}

	newLiquidity := defaultLiquidity.Mul(sdk.NewDec(2))
	newCurrTick := defaultCurrTick.Mul(sdk.NewInt(2))
	newCurrSqrtPrice := defaultCurrSqrtPrice.Mul(sdk.NewDec(2))

	gasBeforeSwap := s.Ctx.GasMeter().GasConsumed()
	mock_pool.ApplySwap(newLiquidity, newCurrTick, newCurrSqrtPrice)
	gasAfterSwap := s.Ctx.GasMeter().GasConsumed()

	s.Require().Equal(gasAfterSwap-gasBeforeSwap, uint64(gammtypes.BalancerGasFeeForSwap))
	s.Require().Equal(mock_pool.Liquidity, newLiquidity)
	s.Require().Equal(mock_pool.CurrentTick, newCurrTick)
	s.Require().Equal(mock_pool.CurrentSqrtPrice, newCurrSqrtPrice)

}
