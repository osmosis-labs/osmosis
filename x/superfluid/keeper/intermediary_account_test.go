package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestIntermediaryAccountsSetGetDeleteFlow() {
	suite.SetupTest()

	// initial check
	accs := suite.app.SuperfluidKeeper.GetAllIntermediaryAccounts(suite.ctx)
	suite.Require().Len(accs, 0)

	// set account
	valAddr := sdk.ValAddress([]byte("addr1---------------"))
	acc := types.SuperfluidIntermediaryAccount{
		Denom:   "gamm/pool/1",
		ValAddr: valAddr.String(),
		GaugeId: 1,
	}
	suite.app.SuperfluidKeeper.SetIntermediaryAccount(suite.ctx, acc)

	// get account
	gacc := suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, acc.GetAddress())
	suite.Require().Equal(gacc.Denom, "gamm/pool/1")
	suite.Require().Equal(gacc.ValAddr, valAddr.String())
	suite.Require().Equal(gacc.GaugeId, uint64(1))

	// check accounts
	accs = suite.app.SuperfluidKeeper.GetAllIntermediaryAccounts(suite.ctx)
	suite.Require().Len(accs, 1)

	// delete asset
	suite.app.SuperfluidKeeper.DeleteIntermediaryAccount(suite.ctx, acc.GetAddress())

	// get account
	gacc = suite.app.SuperfluidKeeper.GetIntermediaryAccount(suite.ctx, acc.GetAddress())
	suite.Require().Equal(gacc.Denom, "")
	suite.Require().Equal(gacc.ValAddr, "")
	suite.Require().Equal(gacc.GaugeId, uint64(0))

	// check accounts
	accs = suite.app.SuperfluidKeeper.GetAllIntermediaryAccounts(suite.ctx)
	suite.Require().Len(accs, 0)
}

func (suite *KeeperTestSuite) TestLockIdIntermediaryAccountConnection() {
	suite.SetupTest()

	// get account
	addr := suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, 1)
	suite.Require().Equal(addr.String(), "")

	// set account
	valAddr := sdk.ValAddress([]byte("addr1---------------"))
	acc := types.SuperfluidIntermediaryAccount{
		Denom:   "gamm/pool/1",
		ValAddr: valAddr.String(),
		GaugeId: 1,
	}
	suite.app.SuperfluidKeeper.SetLockIdIntermediaryAccountConnection(suite.ctx, 1, acc)

	// get account
	addr = suite.app.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.ctx, 1)
	suite.Require().Equal(addr.String(), acc.GetAddress().String())
}
