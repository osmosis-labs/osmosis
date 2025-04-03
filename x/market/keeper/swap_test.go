package keeper_test

import (
	"github.com/osmosis-labs/osmosis/osmomath"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
)

func (s *KeeperTestSuite) TestComputeSwap() {
	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1) // 1 SDR -> 1.7 Melody
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)

	s.Run("Swap SDR to Melody", func() {
		for i := 0; i < 100; i++ {
			swapAmountInSDR := sdrPriceInMelody.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
			offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)
			retCoin, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
			s.Require().NoError(err)
			//s.Require().True(spread.GTE(s.App.MarketKeeper.MinStabilitySpread(s.Ctx)))
			s.Require().Equal(osmomath.NewDecFromInt(offerCoin.Amount).Mul(sdrPriceInMelody), retCoin.Amount)
		}

		offerCoin := sdk.NewCoin(assets.MicroSDRDenom, sdrPriceInMelody.QuoInt64(2).TruncateInt())
		_, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
		s.Require().Error(err)

		offerCoin = sdk.NewCoin(assets.MicroSDRDenom, osmomath.NewDec(1).TruncateInt())
		retCoin, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
		s.Require().NoError(err)
		s.Require().Equal(sdrPriceInMelody, retCoin.Amount)
	})
	s.Run("Swap Melody to SDR", func() {
		for i := 0; i < 100; i++ {
			swapAmountInMelody := sdrPriceInMelody.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
			offerCoin := sdk.NewCoin(appparams.BaseCoinUnit, swapAmountInMelody)
			retCoin, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, assets.MicroSDRDenom)
			s.Require().NoError(err)
			//s.Require().True(spread.GTE(s.App.MarketKeeper.MinStabilitySpread(s.Ctx)))
			s.Require().Equal(osmomath.NewDecFromInt(offerCoin.Amount).Quo(sdrPriceInMelody), retCoin.Amount)
		}

		offerCoin := sdk.NewCoin(appparams.BaseCoinUnit, sdrPriceInMelody.MulInt64(10).TruncateInt())
		retCoin, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, assets.MicroSDRDenom)
		s.Require().NoError(err)
		s.Require().Equal(osmomath.NewDec(10), retCoin.Amount)
	})
}

func (s *KeeperTestSuite) TestComputeInternalSwap() {
	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)

	for i := 0; i < 100; i++ {
		offerCoin := sdk.NewDecCoin(assets.MicroSDRDenom, sdrPriceInMelody.MulInt64(rand.Int63()+1).TruncateInt())
		retCoin, err := s.App.MarketKeeper.ComputeInternalSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
		s.Require().NoError(err)
		s.Require().Equal(offerCoin.Amount.Mul(sdrPriceInMelody), retCoin.Amount)
	}

	offerCoin := sdk.NewDecCoin(assets.MicroSDRDenom, sdrPriceInMelody.QuoInt64(2).TruncateInt())
	_, err := s.App.MarketKeeper.ComputeInternalSwap(s.Ctx, offerCoin, appparams.BaseCoinUnit)
	s.Require().Error(err)
}

// TestComputeInternalSwapStableAndStable tests the case where the offer coin and the return coin are both stable coins.
// In this case the conversion should go through Melody price.
func (s *KeeperTestSuite) TestComputeInternalSwapStableAndStable() {
	// Set Oracle Price
	sdrPriceInMelody := osmomath.NewDecWithPrec(17, 1)
	mntPriceInMelody := osmomath.NewDecWithPrec(7652, 1)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, sdrPriceInMelody)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroMNTDenom, mntPriceInMelody)

	params := s.App.MarketKeeper.GetParams(s.Ctx)
	s.App.MarketKeeper.SetParams(s.Ctx, params)

	swapAmountInSDR := sdrPriceInMelody.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)
	swapCoin, _, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, assets.MicroMNTDenom)
	s.Require().NoError(err)
	s.Require().Equal(swapAmountInSDR.ToLegacyDec().Mul(sdrPriceInMelody).Quo(mntPriceInMelody), swapCoin.Amount)
}

//func (s *KeeperTestSuite) TestIlliquidTobinTaxListParams() {
//	// Set Oracle Price
//	melodyPriceInSDR := osmomath.NewDecWithPrec(17, 1)
//	melodyPriceInMNT := osmomath.NewDecWithPrec(7652, 1)
//	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroSDRDenom, melodyPriceInSDR)
//	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroMNTDenom, melodyPriceInMNT)
//
//	tobinTax := osmomath.NewDecWithPrec(25, 4)
//	params := s.App.MarketKeeper.GetParams(s.Ctx)
//	s.App.MarketKeeper.SetParams(s.Ctx, params)
//
//	illiquidFactor := osmomath.NewDec(2)
//	s.App.OracleKeeper.SetTobinTax(s.Ctx, assets.MicroSDRDenom, tobinTax)
//	s.App.OracleKeeper.SetTobinTax(s.Ctx, assets.MicroMNTDenom, tobinTax.Mul(illiquidFactor))
//
//	swapAmountInSDR := melodyPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
//	offerCoin := sdk.NewCoin(assets.MicroSDRDenom, swapAmountInSDR)
//	_, spread, err := s.App.MarketKeeper.ComputeSwap(s.Ctx, offerCoin, assets.MicroMNTDenom)
//	s.Require().NoError(err)
//	s.Require().Equal(tobinTax.Mul(illiquidFactor), spread)
//}
