package keeper_test

import "github.com/osmosis-labs/osmosis/v12/x/protorev/types"

func (suite *KeeperTestSuite) TestEpochHook() {
	// All of the pools initialized in the setup function are available in keeper_test.go
	// akash <-> types.OsmosisDenomination
	// juno <-> types.OsmosisDenomination
	// ethereum <-> types.OsmosisDenomination
	// bitcoin <-> types.OsmosisDenomination
	// canto <-> types.OsmosisDenomination
	// akash <-> types.AtomDenomination
	// juno <-> types.AtomDenomination
	// ethereum <-> types.AtomDenomination
	// bitcoin <-> types.AtomDenomination
	// canto <-> types.AtomDenomination
	// types.OsmosisDenomination <-> types.AtomDenomination
	// akash <-> juno
	// akash <-> ethereum
	// akash <-> bitcoin
	// akash <-> canto
	// juno <-> ethereum
	// juno <-> bitcoin
	// juno <-> canto
	// ethereum <-> bitcoin
	// ethereum <-> canto
	// bitcoin <-> canto

	// As such, we should expect the following pools to be created:
	// akash <-> types.OsmosisDenomination
	// juno <-> types.OsmosisDenomination
	// ethereum <-> types.OsmosisDenomination
	// bitcoin <-> types.OsmosisDenomination
	// canto <-> types.OsmosisDenomination
	// akash <-> types.AtomDenomination
	// juno <-> types.AtomDenomination
	// ethereum <-> types.AtomDenomination
	// bitcoin <-> types.AtomDenomination
	// canto <-> types.AtomDenomination
	// types.OsmosisDenomination <-> types.AtomDenomination

	// The epoch hook is run after the pools are set and committed so all that must be done is the stores must be checked if they are correctly set
	for index, pool := range suite.pools {
		// Check if there is a match with osmo
		if otherDenom, match := types.CheckMatch(pool.Asset1, pool.Asset2, types.OsmosisDenomination); match {
			poolId, err := suite.App.AppKeepers.ProtoRevKeeper.GetOsmoPool(suite.Ctx, otherDenom)

			// pool ID must exist
			suite.Require().NoError(err)
			suite.Require().Equal(uint64(index+1), poolId)
		}

		// Check if there is a match with atom
		if otherDenom, match := types.CheckMatch(pool.Asset1, pool.Asset2, types.AtomDenomination); match {
			poolId, err := suite.App.AppKeepers.ProtoRevKeeper.GetAtomPool(suite.Ctx, otherDenom)

			// pool ID must exist
			suite.Require().NoError(err)
			suite.Require().Equal(uint64(index+1), poolId)
		}
	}
}
