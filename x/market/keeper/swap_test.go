package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
)

func (s *KeeperTestSuite) TestApplySwapToPool() {
	lunaPriceInSDR := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetLunaExchangeRate(s.Ctx, appparams.MicroSDRDenom, lunaPriceInSDR)

	offerCoin := sdk.NewCoin(appparams.BaseCoinUnit, sdk.NewInt(1000))
	askCoin := sdk.NewDecCoin(appparams.MicroSDRDenom, sdk.NewInt(1700))
	oldSDRPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.App.MarketKeeper.ApplySwapToPool(s.Ctx, offerCoin, askCoin)
	newSDRPoolDelta := s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	sdrDiff := newSDRPoolDelta.Sub(oldSDRPoolDelta)
	s.Require().Equal(sdk.NewDec(-1700), sdrDiff)

	// reverse swap
	offerCoin = sdk.NewCoin(appparams.MicroSDRDenom, sdk.NewInt(1700))
	askCoin = sdk.NewDecCoin(appparams.BaseCoinUnit, sdk.NewInt(1000))
	oldSDRPoolDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.App.MarketKeeper.ApplySwapToPool(s.Ctx, offerCoin, askCoin)
	newSDRPoolDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	sdrDiff = newSDRPoolDelta.Sub(oldSDRPoolDelta)
	s.Require().Equal(sdk.NewDec(1700), sdrDiff)

	// no pool changes are expected
	offerCoin = sdk.NewCoin(appparams.MicroSDRDenom, sdk.NewInt(1700))
	askCoin = sdk.NewDecCoin(appparams.MicroKRWDenom, sdk.NewInt(3400))
	oldSDRPoolDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	s.App.MarketKeeper.ApplySwapToPool(s.Ctx, offerCoin, askCoin)
	newSDRPoolDelta = s.App.MarketKeeper.GetOsmosisPoolDelta(s.Ctx)
	sdrDiff = newSDRPoolDelta.Sub(oldSDRPoolDelta)
	s.Require().Equal(sdk.NewDec(0), sdrDiff)
}

func (s *KeeperTestSuite) TestComputeSwap() {
	// Set Oracle Price
	lunaPriceInSDR := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetLunaExchangeRate(s.Ctx, appparams.MicroSDRDenom, lunaPriceInSDR)

	for i := 0; i < 100; i++ {
		swapAmountInSDR := lunaPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
		offerCoin := sdk.NewCoin(appparams.MicroSDRDenom, swapAmountInSDR)
		retCoin, spread, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)

		s.Require().NoError(err)
		s.Require().True(spread.GTE(s.App.MarketKeeper.MinStabilitySpread(s.Ctx)))
		s.Require().Equal(sdk.NewDecFromInt(offerCoin.Amount).Quo(lunaPriceInSDR), retCoin.Amount)
	}

	offerCoin := sdk.NewCoin(appparams.MicroSDRDenom, lunaPriceInSDR.QuoInt64(2).TruncateInt())
	_, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestComputeInternalSwap() {
	// Set Oracle Price
	lunaPriceInSDR := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetLunaExchangeRate(s.Ctx, appparams.MicroSDRDenom, lunaPriceInSDR)

	for i := 0; i < 100; i++ {
		offerCoin := sdk.NewDecCoin(appparams.MicroSDRDenom, lunaPriceInSDR.MulInt64(rand.Int63()+1).TruncateInt())
		retCoin, err := s.App.MarketKeeper.ComputeInternalSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
		s.Require().NoError(err)
		s.Require().Equal(offerCoin.Amount.Quo(lunaPriceInSDR), retCoin.Amount)
	}

	offerCoin := sdk.NewDecCoin(appparams.MicroSDRDenom, lunaPriceInSDR.QuoInt64(2).TruncateInt())
	_, err := s.App.MarketKeeper.ComputeInternalSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestIlliquidTobinTaxListParams() {
	// Set Oracle Price
	lunaPriceInSDR := sdk.NewDecWithPrec(17, 1)
	lunaPriceInMNT := sdk.NewDecWithPrec(7652, 1)
	s.App.OracleKeeper.SetLunaExchangeRate(s.Ctx, appparams.MicroSDRDenom, lunaPriceInSDR)
	s.App.OracleKeeper.SetLunaExchangeRate(s.Ctx, appparams.MicroMNTDenom, lunaPriceInMNT)

	tobinTax := sdk.NewDecWithPrec(25, 4)
	params := s.App.MarketKeeper.GetParams(s.Ctx)
	s.App.MarketKeeper.SetParams(s.Ctx, params)

	illiquidFactor := sdk.NewDec(2)
	s.App.OracleKeeper.SetTobinTax(s.Ctx, appparams.MicroSDRDenom, tobinTax)
	s.App.OracleKeeper.SetTobinTax(s.Ctx, appparams.MicroMNTDenom, tobinTax.Mul(illiquidFactor))

	swapAmountInSDR := lunaPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(appparams.MicroSDRDenom, swapAmountInSDR)
	_, spread, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.MicroMNTDenom)
	s.Require().NoError(err)
	s.Require().Equal(tobinTax.Mul(illiquidFactor), spread)
}
