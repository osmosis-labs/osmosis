package keeper_test

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v23/app/apptesting/assets"
	appparams "github.com/osmosis-labs/osmosis/v23/app/params"
)

func (s *KeeperTestSuite) TestComputeSwap() {
	// Set Oracle Price
	melodyPriceInSDR := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, melodyPriceInSDR)

	for i := 0; i < 100; i++ {
		swapAmountInSDR := melodyPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
		offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)
		retCoin, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)

		s.Require().NoError(err)
		//s.Require().True(spread.GTE(s.App.MarketKeeper.MinStabilitySpread(s.Ctx)))
		s.Require().Equal(sdk.NewDecFromInt(offerCoin.Amount).Quo(melodyPriceInSDR), retCoin.Amount)
	}

	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, melodyPriceInSDR.QuoInt64(2).TruncateInt())
	_, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestComputeInternalSwap() {
	// Set Oracle Price
	melodyPriceInSDR := sdk.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, melodyPriceInSDR)

	for i := 0; i < 100; i++ {
		offerCoin := sdk.NewDecCoin(assets.MicroSDRDenom, melodyPriceInSDR.MulInt64(rand.Int63()+1).TruncateInt())
		retCoin, err := s.App.MarketKeeper.ComputeInternalSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
		s.Require().NoError(err)
		s.Require().Equal(offerCoin.Amount.Quo(melodyPriceInSDR), retCoin.Amount)
	}

	offerCoin := sdk.NewDecCoin(assets.MicroSDRDenom, melodyPriceInSDR.QuoInt64(2).TruncateInt())
	_, err := s.App.MarketKeeper.ComputeInternalSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestIlliquidTobinTaxListParams() {
	// Set Oracle Price
	melodyPriceInSDR := sdk.NewDecWithPrec(17, 1)
	melodyPriceInMNT := sdk.NewDecWithPrec(7652, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, melodyPriceInSDR)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroMNTDenom, melodyPriceInMNT)

	tobinTax := sdk.NewDecWithPrec(25, 4)
	params := s.App.MarketKeeper.GetParams(s.Ctx)
	s.App.MarketKeeper.SetParams(s.Ctx, params)

	illiquidFactor := sdk.NewDec(2)
	s.App.OracleKeeper.SetTobinTax(s.Ctx, assets.MicroSDRDenom, tobinTax)
	s.App.OracleKeeper.SetTobinTax(s.Ctx, assets.MicroMNTDenom, tobinTax.Mul(illiquidFactor))

	swapAmountInSDR := melodyPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)
	_, spread, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, assets.MicroMNTDenom)
	s.Require().NoError(err)
	s.Require().Equal(tobinTax.Mul(illiquidFactor), spread)
}
