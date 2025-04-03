package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v27/x/market"
)

func (s *KeeperTestSuite) TestExportInitGenesis() {
	genesis := market.ExportGenesis(s.Ctx, *s.App.MarketKeeper)

	market.InitGenesis(s.Ctx, *s.App.MarketKeeper, genesis)
	newGenesis := market.ExportGenesis(s.Ctx, *s.App.MarketKeeper)

	s.Require().Equal(genesis, newGenesis)
}
