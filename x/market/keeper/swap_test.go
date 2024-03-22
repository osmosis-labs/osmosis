package keeper

//func TestApplySwapToPool(t *testing.T) {
//	input := CreateTestInput(t)
//
//	lunaPriceInSDR := sdk.NewDecWithPrec(17, 1)
//	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, appParams.MicroSDRDenom, lunaPriceInSDR)
//
//	offerCoin := sdk.NewCoin(appParams.BaseCoinUnit, sdk.NewInt(1000))
//	askCoin := sdk.NewDecCoin(appParams.MicroSDRDenom, sdk.NewInt(1700))
//	oldSDRPoolDelta := input.MarketKeeper.GetTerraPoolDelta(input.Ctx)
//	input.MarketKeeper.ApplySwapToPool(input.Ctx, offerCoin, askCoin)
//	newSDRPoolDelta := input.MarketKeeper.GetTerraPoolDelta(input.Ctx)
//	sdrDiff := newSDRPoolDelta.Sub(oldSDRPoolDelta)
//	require.Equal(t, sdk.NewDec(-1700), sdrDiff)
//
//	// reverse swap
//	offerCoin = sdk.NewCoin(appParams.MicroSDRDenom, sdk.NewInt(1700))
//	askCoin = sdk.NewDecCoin(appParams.BaseCoinUnit, sdk.NewInt(1000))
//	oldSDRPoolDelta = input.MarketKeeper.GetTerraPoolDelta(input.Ctx)
//	input.MarketKeeper.ApplySwapToPool(input.Ctx, offerCoin, askCoin)
//	newSDRPoolDelta = input.MarketKeeper.GetTerraPoolDelta(input.Ctx)
//	sdrDiff = newSDRPoolDelta.Sub(oldSDRPoolDelta)
//	require.Equal(t, sdk.NewDec(1700), sdrDiff)
//
//	// TERRA <> TERRA, no pool changes are expected
//	offerCoin = sdk.NewCoin(appParams.MicroSDRDenom, sdk.NewInt(1700))
//	askCoin = sdk.NewDecCoin(appParams.MicroKRWDenom, sdk.NewInt(3400))
//	oldSDRPoolDelta = input.MarketKeeper.GetTerraPoolDelta(input.Ctx)
//	input.MarketKeeper.ApplySwapToPool(input.Ctx, offerCoin, askCoin)
//	newSDRPoolDelta = input.MarketKeeper.GetTerraPoolDelta(input.Ctx)
//	sdrDiff = newSDRPoolDelta.Sub(oldSDRPoolDelta)
//	require.Equal(t, sdk.NewDec(0), sdrDiff)
//}
//
//func TestComputeSwap(t *testing.T) {
//	input := CreateTestInput(t)
//
//	// Set Oracle Price
//	lunaPriceInSDR := sdk.NewDecWithPrec(17, 1)
//	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, appParams.MicroSDRDenom, lunaPriceInSDR)
//
//	for i := 0; i < 100; i++ {
//		swapAmountInSDR := lunaPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
//		offerCoin := sdk.NewCoin(appParams.MicroSDRDenom, swapAmountInSDR)
//		retCoin, spread, err := input.MarketKeeper.ComputeSwap(input.Ctx, offerCoin, appParams.BaseCoinUnit)
//
//		require.NoError(t, err)
//		require.True(t, spread.GTE(input.MarketKeeper.MinStabilitySpread(input.Ctx)))
//		require.Equal(t, sdk.NewDecFromInt(offerCoin.Amount).Quo(lunaPriceInSDR), retCoin.Amount)
//	}
//
//	offerCoin := sdk.NewCoin(appParams.MicroSDRDenom, lunaPriceInSDR.QuoInt64(2).TruncateInt())
//	_, _, err := input.MarketKeeper.ComputeSwap(input.Ctx, offerCoin, appParams.BaseCoinUnit)
//	require.Error(t, err)
//}
//
//func TestComputeInternalSwap(t *testing.T) {
//	input := CreateTestInput(t)
//
//	// Set Oracle Price
//	lunaPriceInSDR := sdk.NewDecWithPrec(17, 1)
//	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, appParams.MicroSDRDenom, lunaPriceInSDR)
//
//	for i := 0; i < 100; i++ {
//		offerCoin := sdk.NewDecCoin(appParams.MicroSDRDenom, lunaPriceInSDR.MulInt64(rand.Int63()+1).TruncateInt())
//		retCoin, err := input.MarketKeeper.ComputeInternalSwap(input.Ctx, offerCoin, appParams.BaseCoinUnit)
//		require.NoError(t, err)
//		require.Equal(t, offerCoin.Amount.Quo(lunaPriceInSDR), retCoin.Amount)
//	}
//
//	offerCoin := sdk.NewDecCoin(appParams.MicroSDRDenom, lunaPriceInSDR.QuoInt64(2).TruncateInt())
//	_, err := input.MarketKeeper.ComputeInternalSwap(input.Ctx, offerCoin, appParams.BaseCoinUnit)
//	require.Error(t, err)
//}
//
//func TestIlliquidTobinTaxListParams(t *testing.T) {
//	input := CreateTestInput(t)
//
//	// Set Oracle Price
//	lunaPriceInSDR := sdk.OneDec() //sdk.NewDecWithPrec(1, 1)
//	lunaPriceInMNT := sdk.OneDec() //sdk.NewDecWithPrec(1, 1)
//	//input.OracleKeeper.SetLunaExchangeRate(input.Ctx, appParams.MicroSDRDenom, lunaPriceInSDR)
//	//input.OracleKeeper.SetLunaExchangeRate(input.Ctx, appParams.MicroMNTDenom, lunaPriceInMNT)
//
//	tobinTax := sdk.NewDecWithPrec(25, 4)
//	params := input.MarketKeeper.GetParams(input.Ctx)
//	input.MarketKeeper.SetParams(input.Ctx, params)
//
//	illiquidFactor := sdk.NewDec(2)
//	input.OracleKeeper.SetTobinTax(input.Ctx, appParams.MicroSDRDenom, tobinTax)
//	input.OracleKeeper.SetTobinTax(input.Ctx, appParams.MicroMNTDenom, tobinTax.Mul(illiquidFactor))
//
//	swapAmountInSDR := lunaPriceInSDR.MulInt64(rand.Int63()%10000 + 2).TruncateInt()
//	offerCoin := sdk.NewCoin(appParams.MicroSDRDenom, swapAmountInSDR)
//	_, spread, err := input.MarketKeeper.ComputeSwap(input.Ctx, offerCoin, appParams.MicroMNTDenom)
//	require.NoError(t, err)
//	require.Equal(t, tobinTax.Mul(illiquidFactor), spread)
//}
