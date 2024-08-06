package treasury

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	core "github.com/osmosis-labs/osmosis/v23/types"
	"github.com/osmosis-labs/osmosis/v23/x/treasury/keeper"
)

func TestExportInitGenesis(t *testing.T) {
	input := keeper.CreateTestInput(t)
	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek) * 3)

	input.TreasuryKeeper.RecordEpochInitialIssuance(input.Ctx)
	input.TreasuryKeeper.SetRewardWeight(input.Ctx, sdk.NewDec(1123))
	input.TreasuryKeeper.SetTaxCap(input.Ctx, "foo", sdk.NewInt(1234))
	input.TreasuryKeeper.SetTaxRate(input.Ctx, sdk.NewDec(5435))
	input.TreasuryKeeper.SetEpochTaxProceeds(input.Ctx, sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(923))))
	input.TreasuryKeeper.SetTR(input.Ctx, int64(0), sdk.NewDec(123))
	input.TreasuryKeeper.SetTR(input.Ctx, int64(1), sdk.NewDec(345))
	input.TreasuryKeeper.SetTR(input.Ctx, int64(2), sdk.NewDec(567))
	input.TreasuryKeeper.SetSR(input.Ctx, int64(0), sdk.NewDec(123))
	input.TreasuryKeeper.SetSR(input.Ctx, int64(1), sdk.NewDec(345))
	input.TreasuryKeeper.SetSR(input.Ctx, int64(2), sdk.NewDec(567))
	input.TreasuryKeeper.SetTSL(input.Ctx, int64(0), sdk.NewInt(123))
	input.TreasuryKeeper.SetTSL(input.Ctx, int64(1), sdk.NewInt(345))
	input.TreasuryKeeper.SetTSL(input.Ctx, int64(2), sdk.NewInt(567))
	genesis := ExportGenesis(input.Ctx, input.TreasuryKeeper)

	newInput := keeper.CreateTestInput(t)
	newInput.Ctx = newInput.Ctx.WithBlockHeight(int64(core.BlocksPerWeek) * 3)
	InitGenesis(newInput.Ctx, newInput.TreasuryKeeper, genesis)
	newGenesis := ExportGenesis(newInput.Ctx, newInput.TreasuryKeeper)

	require.Equal(t, genesis, newGenesis)

	// Make epoch initial issuance to zero
	tmp := genesis.EpochInitialIssuance
	genesis.EpochInitialIssuance = sdk.Coins{}

	newInput = keeper.CreateTestInput(t)
	newInput.Ctx = newInput.Ctx.WithBlockHeight(int64(core.BlocksPerWeek) * 3)
	InitGenesis(newInput.Ctx, newInput.TreasuryKeeper, genesis)
	newGenesis = ExportGenesis(newInput.Ctx, newInput.TreasuryKeeper)

	// Return back epoch initial issuance
	genesis.EpochInitialIssuance = tmp
	require.Equal(t, genesis, newGenesis)
}
