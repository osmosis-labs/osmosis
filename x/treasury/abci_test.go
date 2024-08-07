package treasury

//func TestEndBlockerIssuanceUpdateWithBurnModule(t *testing.T) {
//	input := keeper.CreateTestInput(t)
//
//	supply := input.BankKeeper.GetSupply(input.Ctx, core.MicroLunaDenom)
//
//	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek) - 1)
//	EndBlocker(input.Ctx, input.TreasuryKeeper)
//
//	issuance := input.TreasuryKeeper.GetEpochInitialIssuance(input.Ctx)
//	require.Equal(t,
//		// subtract due to burn module account burning
//		supply.Amount.Sub(keeper.InitCoins.AmountOf(core.MicroLunaDenom)),
//		issuance.AmountOf(core.MicroLunaDenom))
//}
//
//func TestUpdate(t *testing.T) {
//	input := keeper.CreateTestInput(t)
//
//	windowProbation := input.TreasuryKeeper.WindowProbation(input.Ctx)
//
//	targetEpoch := int64(windowProbation + 1)
//	for epoch := int64(0); epoch < targetEpoch; epoch++ {
//		input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek)*epoch - 1)
//		EndBlocker(input.Ctx, input.TreasuryKeeper)
//	}
//
//	// load old tax rate & reward weight
//	taxRate := input.TreasuryKeeper.GetTaxRate(input.Ctx)
//	rewardWeight := input.TreasuryKeeper.GetRewardWeight(input.Ctx)
//
//	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek)*targetEpoch - 1)
//	EndBlocker(input.Ctx, input.TreasuryKeeper)
//
//	// zero tax proceeds will increase tax rate with change max amount
//	newTaxRate := input.TreasuryKeeper.GetTaxRate(input.Ctx)
//	require.Equal(t, taxRate.Add(input.TreasuryKeeper.TaxPolicy(input.Ctx).ChangeRateMax), newTaxRate)
//
//	// zero mining rewards will increase reward weight with change max amount
//	newRewardWeight := input.TreasuryKeeper.GetRewardWeight(input.Ctx)
//	require.Equal(t, rewardWeight.Add(input.TreasuryKeeper.RewardPolicy(input.Ctx).ChangeRateMax), newRewardWeight)
//}
