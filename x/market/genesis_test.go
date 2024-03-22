package market

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v23/x/market/keeper"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestExportInitGenesis(t *testing.T) {
	input := keeper.CreateTestInput(t)
	input.MarketKeeper.SetTerraPoolDelta(input.Ctx, sdk.NewDec(1123))
	genesis := ExportGenesis(input.Ctx, input.MarketKeeper)

	newInput := keeper.CreateTestInput(t)
	InitGenesis(newInput.Ctx, newInput.MarketKeeper, genesis)
	newGenesis := ExportGenesis(newInput.Ctx, newInput.MarketKeeper)

	require.Equal(t, genesis, newGenesis)
}
