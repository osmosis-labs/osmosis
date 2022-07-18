package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestIntermediaryAccountCreation() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		delegatorNumber  int64
		superDelegations []superfluidDelegation
	}{
		{
			"test intermediary account with single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
		},
		{
			"test multiple intermediary accounts with multiple superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			valAddrs := suite.SetupValidators(tc.validatorStats)
			delAddrs := CreateRandomAccounts(int(tc.delegatorNumber))

			// we create two additional pools: total three pools, 10 gauges
			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			var interAccs []types.SuperfluidIntermediaryAccount

			for _, superDelegation := range tc.superDelegations {
				delAddr := delAddrs[superDelegation.delIndex]
				valAddr := valAddrs[superDelegation.valIndex]
				denom := denoms[superDelegation.lpIndex]

				// check intermediary Account prior to superfluid delegation, should have nil Intermediary Account
				expAcc := types.NewSuperfluidIntermediaryAccount(denom, valAddr.String(), 0)
				interAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, expAcc.GetAccAddress())
				suite.Require().NotEqual(expAcc.GetAccAddress(), interAcc.GetAccAddress())
				suite.Require().Equal("", interAcc.Denom)
				suite.Require().Equal(uint64(0), interAcc.GaugeId)
				suite.Require().Equal("", interAcc.ValAddr)

				lock := suite.SetupSuperfluidDelegate(delAddr, valAddr, denom, superDelegation.lpAmount)

				// check that intermediary Account connection is established
				interAccConnection := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lock.ID)
				suite.Require().Equal(expAcc.GetAccAddress(), interAccConnection)

				interAcc = suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, interAccConnection)
				suite.Require().Equal(expAcc.GetAccAddress(), interAcc.GetAccAddress())

				// check on interAcc that has been created
				suite.Require().Equal(denom, interAcc.Denom)
				suite.Require().Equal(valAddr.String(), interAcc.ValAddr)

				interAccs = append(interAccs, interAcc)
			}
			suite.checkIntermediaryAccountDelegations(interAccs)
		})
	}
}

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
	suite.App.SuperfluidKeeper.DeleteIntermediaryAccountIfNoDelegation(suite.Ctx, acc)

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
