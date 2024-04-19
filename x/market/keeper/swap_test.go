package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
)

func (s *KeeperTestSuite) TestComputeSwap() {
	// Set Oracle Price
	osmoPriceInSDR := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetOsmoExchangeRate(s.Ctx, appparams.MicroSDRDenom, osmoPriceInSDR)

	for i := 0; i < 100; i++ {
		swapAmountInSDR := osmoPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
		offerCoin := sdk.NewCoin(appparams.MicroSDRDenom, swapAmountInSDR)
		retCoin, spread, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)

		s.Require().NoError(err)
		s.Require().True(spread.GTE(s.App.MarketKeeper.MinStabilitySpread(s.Ctx)))
		s.Require().Equal(sdk.NewDecFromInt(offerCoin.Amount).Quo(osmoPriceInSDR), retCoin.Amount)
	}

	offerCoin := sdk.NewCoin(appparams.MicroSDRDenom, osmoPriceInSDR.QuoInt64(2).TruncateInt())
	_, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestComputeInternalSwap() {
	// Set Oracle Price
	osmoPriceInSDR := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetOsmoExchangeRate(s.Ctx, appparams.MicroSDRDenom, osmoPriceInSDR)

	for i := 0; i < 100; i++ {
		offerCoin := sdk.NewDecCoin(appparams.MicroSDRDenom, osmoPriceInSDR.MulInt64(rand.Int63()+1).TruncateInt())
		retCoin, err := s.App.MarketKeeper.ComputeInternalSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
		s.Require().NoError(err)
		s.Require().Equal(offerCoin.Amount.Quo(osmoPriceInSDR), retCoin.Amount)
	}

	offerCoin := sdk.NewDecCoin(appparams.MicroSDRDenom, osmoPriceInSDR.QuoInt64(2).TruncateInt())
	_, err := s.App.MarketKeeper.ComputeInternalSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestIlliquidTobinTaxListParams() {
	// Set Oracle Price
	osmoPriceInSDR := sdk.NewDecWithPrec(17, 1)
	osmoPriceInMNT := sdk.NewDecWithPrec(7652, 1)
	s.App.OracleKeeper.SetOsmoExchangeRate(s.Ctx, appparams.MicroSDRDenom, osmoPriceInSDR)
	s.App.OracleKeeper.SetOsmoExchangeRate(s.Ctx, appparams.MicroMNTDenom, osmoPriceInMNT)

	tobinTax := sdk.NewDecWithPrec(25, 4)
	params := s.App.MarketKeeper.GetParams(s.Ctx)
	s.App.MarketKeeper.SetParams(s.Ctx, params)

	illiquidFactor := sdk.NewDec(2)
	s.App.OracleKeeper.SetTobinTax(s.Ctx, appparams.MicroSDRDenom, tobinTax)
	s.App.OracleKeeper.SetTobinTax(s.Ctx, appparams.MicroMNTDenom, tobinTax.Mul(illiquidFactor))

	swapAmountInSDR := osmoPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(appparams.MicroSDRDenom, swapAmountInSDR)
	_, spread, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.MicroMNTDenom)
	s.Require().NoError(err)
	s.Require().Equal(tobinTax.Mul(illiquidFactor), spread)
}
