package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestIntermediaryAccountsSetGetDeleteFlow() {
	suite.SetupTest()

	// initial check
	accs := suite.App.SuperfluidKeeper.GetAllIntermediaryAccounts(suite.Ctx)
	suite.Require().Len(accs, 0)

	// set account
	valAddr := sdk.ValAddress([]byte("addr1---------------"))
	acc := types.NewSuperfluidIntermediaryAccount("gamm/pool/1", valAddr.String(), 1)
	suite.App.SuperfluidKeeper.SetIntermediaryAccount(suite.Ctx, acc)

	// get account
	gacc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, acc.GetAccAddress())
	suite.Require().Equal(gacc.Denom, "gamm/pool/1")
	suite.Require().Equal(gacc.ValAddr, valAddr.String())
	suite.Require().Equal(gacc.GaugeId, uint64(1))

	// check accounts
	accs = suite.App.SuperfluidKeeper.GetAllIntermediaryAccounts(suite.Ctx)
	suite.Require().Equal(accs, []types.SuperfluidIntermediaryAccount{acc})

	// delete asset
	suite.App.SuperfluidKeeper.DeleteIntermediaryAccount(suite.Ctx, acc.GetAccAddress())

	// get account
	gacc = suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, acc.GetAccAddress())
	suite.Require().Equal(gacc.Denom, "")
	suite.Require().Equal(gacc.ValAddr, "")
	suite.Require().Equal(gacc.GaugeId, uint64(0))

	// check accounts
	accs = suite.App.SuperfluidKeeper.GetAllIntermediaryAccounts(suite.Ctx)
	suite.Require().Len(accs, 0)
}

func (suite *KeeperTestSuite) TestLockIdIntermediaryAccountConnection() {
	suite.SetupTest()

	// get account
	addr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, 1)
	suite.Require().Equal(addr.String(), "")

	// set account
	valAddr := sdk.ValAddress([]byte("addr1---------------"))
	acc := types.NewSuperfluidIntermediaryAccount("gamm/pool/1", valAddr.String(), 1)
	suite.App.SuperfluidKeeper.SetLockIdIntermediaryAccountConnection(suite.Ctx, 1, acc)

	// get account
	addr = suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, 1)
	suite.Require().Equal(addr.String(), acc.GetAccAddress().String())

	// check get all
	conns := suite.App.SuperfluidKeeper.GetAllLockIdIntermediaryAccountConnections(suite.Ctx)
	suite.Require().Len(conns, 1)

	// delete account
	suite.App.SuperfluidKeeper.DeleteLockIdIntermediaryAccountConnection(suite.Ctx, 1)

	// get account
	addr = suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, 1)
	suite.Require().Equal(addr.String(), "")

}
