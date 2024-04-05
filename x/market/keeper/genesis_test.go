package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v23/x/market"
)

func (s *KeeperTestSuite) TestExportInitGenesis() {
	s.App.MarketKeeper.SetOsmosisPoolDelta(s.Ctx, sdk.NewDec(1123))
	genesis := market.ExportGenesis(s.Ctx, *s.App.MarketKeeper)

	market.InitGenesis(s.Ctx, *s.App.MarketKeeper, genesis)
	newGenesis := market.ExportGenesis(s.Ctx, *s.App.MarketKeeper)

	s.Require().Equal(genesis, newGenesis)
}
