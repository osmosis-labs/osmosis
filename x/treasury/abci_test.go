package treasury

import (
	"testing"

	"github.com/stretchr/testify/require"

	core "github.com/osmosis-labs/osmosis/v23/types"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/keeper"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"
)

func TestBurnAddress(t *testing.T) {
	input := keeper.CreateTestInput(t)

	burnAddress := input.AccountKeeper.GetModuleAddress(types.BurnModuleName)
	require.Equal(t, keeper.InitCoins, input.BankKeeper.GetAllBalances(input.Ctx, burnAddress))

	EndBlocker(input.Ctx, input.TreasuryKeeper)
	require.True(t, input.BankKeeper.GetAllBalances(input.Ctx, burnAddress).IsZero())
}

func TestEndBlockerIssuanceUpdateWithBurnModule(t *testing.T) {
	input := keeper.CreateTestInput(t)

	supply := input.BankKeeper.GetSupply(input.Ctx, core.MicroLunaDenom)

	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek) - 1)
	EndBlocker(input.Ctx, input.TreasuryKeeper)

	issuance := input.TreasuryKeeper.GetEpochInitialIssuance(input.Ctx)
	require.Equal(t,
		// subtract due to burn module account burning
		supply.Amount.Sub(keeper.InitCoins.AmountOf(core.MicroLunaDenom)),
		issuance.AmountOf(core.MicroLunaDenom))
}

func TestUpdate(t *testing.T) {
	input := keeper.CreateTestInput(t)

	windowProbation := input.TreasuryKeeper.WindowProbation(input.Ctx)

	targetEpoch := int64(windowProbation + 1)
	for epoch := int64(0); epoch < targetEpoch; epoch++ {
		input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek)*epoch - 1)
		EndBlocker(input.Ctx, input.TreasuryKeeper)
	}

	// load old tax rate & reward weight
	taxRate := input.TreasuryKeeper.GetTaxRate(input.Ctx)
	rewardWeight := input.TreasuryKeeper.GetRewardWeight(input.Ctx)

	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek)*targetEpoch - 1)
	EndBlocker(input.Ctx, input.TreasuryKeeper)

	// zero tax proceeds will increase tax rate with change max amount
	newTaxRate := input.TreasuryKeeper.GetTaxRate(input.Ctx)
	require.Equal(t, taxRate.Add(input.TreasuryKeeper.TaxPolicy(input.Ctx).ChangeRateMax), newTaxRate)

	// zero mining rewards will increase reward weight with change max amount
	newRewardWeight := input.TreasuryKeeper.GetRewardWeight(input.Ctx)
	require.Equal(t, rewardWeight.Add(input.TreasuryKeeper.RewardPolicy(input.Ctx).ChangeRateMax), newRewardWeight)
}

func TestEmptyIndicator(t *testing.T) {
	input := keeper.CreateTestInput(t)

	windowProbation := input.TreasuryKeeper.WindowProbation(input.Ctx)

	targetEpoch := int64(windowProbation + 1)
	for epoch := int64(0); epoch < targetEpoch; epoch++ {
		// skip last epoch end blocker
		// to make indicators empty
		if epoch == targetEpoch-1 {
			continue
		}

		input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek)*epoch - 1)
		EndBlocker(input.Ctx, input.TreasuryKeeper)
	}

	// Must result in the same output as normal zero tax & mining rewards
	// load old tax rate & reward weight
	taxRate := input.TreasuryKeeper.GetTaxRate(input.Ctx)
	rewardWeight := input.TreasuryKeeper.GetRewardWeight(input.Ctx)

	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek)*targetEpoch - 1)
	EndBlocker(input.Ctx, input.TreasuryKeeper)

	// zero tax proceeds will increase tax rate with change max amount
	newTaxRate := input.TreasuryKeeper.GetTaxRate(input.Ctx)
	require.Equal(t, taxRate.Add(input.TreasuryKeeper.TaxPolicy(input.Ctx).ChangeRateMax), newTaxRate)

	// zero mining rewards will increase reward weight with change max amount
	newRewardWeight := input.TreasuryKeeper.GetRewardWeight(input.Ctx)
	require.Equal(t, rewardWeight.Add(input.TreasuryKeeper.RewardPolicy(input.Ctx).ChangeRateMax), newRewardWeight)
}
