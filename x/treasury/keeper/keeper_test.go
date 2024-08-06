package keeper

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"

	core "github.com/osmosis-labs/osmosis/v23/types"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/types"
)

func TestRewardWeight(t *testing.T) {
	input := CreateTestInput(t)

	// See that we can get and set reward weights
	for i := int64(0); i < 10; i++ {
		input.TreasuryKeeper.SetRewardWeight(input.Ctx, sdk.NewDecWithPrec(i, 2))
		require.Equal(t, sdk.NewDecWithPrec(i, 2), input.TreasuryKeeper.GetRewardWeight(input.Ctx))
	}
}

func TestTaxRate(t *testing.T) {
	input := CreateTestInput(t)

	// See that we can get and set tax rate
	for i := int64(0); i < 10; i++ {
		input.TreasuryKeeper.SetTaxRate(input.Ctx, sdk.NewDecWithPrec(i, 2))
		require.Equal(t, sdk.NewDecWithPrec(i, 2), input.TreasuryKeeper.GetTaxRate(input.Ctx))
	}
}

func TestTaxCap(t *testing.T) {
	input := CreateTestInput(t)

	for i := int64(0); i < 10; i++ {
		input.TreasuryKeeper.SetTaxCap(input.Ctx, core.MicroCNYDenom, sdk.NewInt(i))
		require.Equal(t, sdk.NewInt(i), input.TreasuryKeeper.GetTaxCap(input.Ctx, core.MicroCNYDenom))
	}
}

func TestIterateTaxCap(t *testing.T) {
	input := CreateTestInput(t)

	cnyCap := sdk.NewInt(123)
	usdCap := sdk.NewInt(13)
	krwCap := sdk.NewInt(1300)
	input.TreasuryKeeper.SetTaxCap(input.Ctx, core.MicroCNYDenom, cnyCap)
	input.TreasuryKeeper.SetTaxCap(input.Ctx, core.MicroUSDDenom, usdCap)
	input.TreasuryKeeper.SetTaxCap(input.Ctx, core.MicroKRWDenom, krwCap)

	input.TreasuryKeeper.IterateTaxCap(input.Ctx, func(denom string, taxCap math.Int) bool {
		switch denom {
		case core.MicroCNYDenom:
			require.Equal(t, cnyCap, taxCap)
		case core.MicroUSDDenom:
			require.Equal(t, usdCap, taxCap)
		case core.MicroKRWDenom:
			require.Equal(t, krwCap, taxCap)
		}

		return false
	})
}

func TestTaxProceeds(t *testing.T) {
	input := CreateTestInput(t)

	for i := int64(0); i < 10; i++ {
		proceeds := sdk.NewCoins(sdk.NewCoin(core.MicroSDRDenom, sdk.NewInt(100+i)))
		input.TreasuryKeeper.RecordEpochTaxProceeds(input.Ctx, proceeds)
		input.TreasuryKeeper.RecordEpochTaxProceeds(input.Ctx, proceeds)
		input.TreasuryKeeper.RecordEpochTaxProceeds(input.Ctx, proceeds)

		require.Equal(t, proceeds.Add(proceeds...).Add(proceeds...), input.TreasuryKeeper.PeekEpochTaxProceeds(input.Ctx))

		input.TreasuryKeeper.SetEpochTaxProceeds(input.Ctx, sdk.Coins{})
		require.True(t, input.TreasuryKeeper.PeekEpochTaxProceeds(input.Ctx).IsZero())
	}
}

func TestMicroLunaIssuance(t *testing.T) {
	input := CreateTestInput(t)

	initialSupply := input.BankKeeper.GetSupply(input.Ctx, core.MicroLunaDenom)
	// See that we can get and set luna issuance
	blocksPerEpoch := core.BlocksPerWeek
	for i := int64(0); i < 10; i++ {
		input.Ctx = input.Ctx.WithBlockHeight(i * int64(blocksPerEpoch))

		input.TreasuryKeeper.RecordEpochInitialIssuance(input.Ctx)
		require.Equal(t, initialSupply.Amount.Add(sdk.NewInt(i)), input.TreasuryKeeper.GetEpochInitialIssuance(input.Ctx).AmountOf(core.MicroLunaDenom))

		input.BankKeeper.MintCoins(input.Ctx, faucetAccountName, sdk.NewCoins(sdk.NewCoin(core.MicroLunaDenom, sdk.OneInt())))
	}
}

func TestPeekEpochSeigniorage(t *testing.T) {
	input := CreateTestInput(t)

	for i := int64(0); i < 10; i++ {
		input.Ctx = input.Ctx.WithBlockHeight(i * int64(core.BlocksPerWeek))
		faucetBalance := input.BankKeeper.GetBalance(input.Ctx, input.AccountKeeper.GetModuleAddress(faucetAccountName), core.MicroLunaDenom)

		input.TreasuryKeeper.RecordEpochInitialIssuance(input.Ctx)

		issueAmount := sdk.NewInt(rand.Int63()%1000000 + 1)
		err := input.BankKeeper.MintCoins(input.Ctx, faucetAccountName, sdk.NewCoins(sdk.NewCoin(core.MicroLunaDenom, issueAmount)))
		require.NoError(t, err)

		burnAmount := sdk.NewInt(rand.Int63()%(faucetBalance.Amount.Int64()+issueAmount.Int64()) + 1)
		err = input.BankKeeper.BurnCoins(input.Ctx, faucetAccountName, sdk.NewCoins(sdk.NewCoin(core.MicroLunaDenom, burnAmount)))
		require.NoError(t, err)

		targetSeigniorage := burnAmount.Sub(issueAmount)
		if targetSeigniorage.IsNegative() {
			targetSeigniorage = sdk.ZeroInt()
		}

		require.Equal(t, targetSeigniorage, input.TreasuryKeeper.PeekEpochSeigniorage(input.Ctx))
	}
}

func TestIndicatorGetterSetter(t *testing.T) {
	input := CreateTestInput(t)

	for e := int64(0); e < 10; e++ {
		randomVal := sdk.NewDec(rand.Int63() + 1)
		input.TreasuryKeeper.SetTR(input.Ctx, e, randomVal)
		require.Equal(t, randomVal, input.TreasuryKeeper.GetTR(input.Ctx, e))
		input.TreasuryKeeper.SetSR(input.Ctx, e, randomVal)
		require.Equal(t, randomVal, input.TreasuryKeeper.GetSR(input.Ctx, e))
		input.TreasuryKeeper.SetTSL(input.Ctx, e, randomVal.TruncateInt())
		require.Equal(t, randomVal.TruncateInt(), input.TreasuryKeeper.GetTSL(input.Ctx, e))
	}

	input.TreasuryKeeper.ClearTRs(input.Ctx)
	input.TreasuryKeeper.ClearSRs(input.Ctx)
	input.TreasuryKeeper.ClearTSLs(input.Ctx)

	for e := int64(0); e < 10; e++ {
		require.Equal(t, sdk.ZeroDec(), input.TreasuryKeeper.GetTR(input.Ctx, e))
		require.Equal(t, sdk.ZeroDec(), input.TreasuryKeeper.GetSR(input.Ctx, e))
		require.Equal(t, sdk.ZeroInt(), input.TreasuryKeeper.GetTSL(input.Ctx, e))
	}
}

func TestParams(t *testing.T) {
	input := CreateTestInput(t)

	defaultParams := types.DefaultParams()
	input.TreasuryKeeper.SetParams(input.Ctx, defaultParams)

	retrievedParams := input.TreasuryKeeper.GetParams(input.Ctx)
	require.Equal(t, defaultParams, retrievedParams)
}

func TestBurnTaxExemptionList(t *testing.T) {
	input := CreateTestInput(t)

	require.False(t, input.TreasuryKeeper.HasBurnTaxExemptionAddress(input.Ctx, ""))
	require.Panics(t, func() { input.TreasuryKeeper.AddBurnTaxExemptionAddress(input.Ctx, "") })
	require.Panics(t, func() { input.TreasuryKeeper.RemoveBurnTaxExemptionAddress(input.Ctx, "") })

	pubKey := secp256k1.GenPrivKey().PubKey()
	address := sdk.AccAddress(pubKey.Address())

	// add an address
	input.TreasuryKeeper.AddBurnTaxExemptionAddress(input.Ctx, address.String())
	require.True(t, input.TreasuryKeeper.HasBurnTaxExemptionAddress(input.Ctx, address.String()))

	// add it again
	input.TreasuryKeeper.AddBurnTaxExemptionAddress(input.Ctx, address.String())
	require.True(t, input.TreasuryKeeper.HasBurnTaxExemptionAddress(input.Ctx, address.String()))

	// remove it
	input.TreasuryKeeper.RemoveBurnTaxExemptionAddress(input.Ctx, address.String())
	require.False(t, input.TreasuryKeeper.HasBurnTaxExemptionAddress(input.Ctx, address.String()))
}
