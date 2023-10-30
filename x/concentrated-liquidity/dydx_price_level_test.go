package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

func (s *KeeperTestSuite) TestDyDx() {
	var (
		base  = "dydx"
		quote = "usdc"
	)

	clPool := s.PrepareConcentratedPoolWithCoins(base, quote)

	// Create position setting the price to $2.46 assuming that
	// - base exponent (dydx) is 18
	// - quote expoennet (usdc) is 6

	initialCoins := sdk.NewCoins(
		sdk.NewCoin(base, sdk.NewInt(1_000_000_000_000_000_000)),
		sdk.NewCoin(quote, sdk.NewInt(2_460_000)),
	)

	s.FundAcc(s.TestAccs[0], initialCoins)
	_, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, clPool.GetId(), s.TestAccs[0], initialCoins)
	s.Require().NoError(err)

	// Check that the price is $2.46
	price, err := s.App.ConcentratedLiquidityKeeper.CalculateSpotPrice(s.Ctx, clPool.GetId(), quote, base)
	s.Require().NoError(err)

	s.Require().Equal(osmomath.NewBigDecWithPrec(246, 12+2).String(), price.String())
}
