package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/app/apptesting"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

// TestMsgSetHotRoutes tests the MsgSetHotRoutes message.
func (suite *KeeperTestSuite) TestMsgSetHotRoutes() {
	validArbRoutes := types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.AtomDenomination)

	notThreePoolArbRoutes := types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.AtomDenomination)
	extraTrade := types.NewTrade(100000, "a", "b")
	notThreePoolArbRoutes.ArbRoutes = append(notThreePoolArbRoutes.ArbRoutes, &types.Route{Trades: []*types.Trade{&extraTrade}})
	testCases := []struct {
		description       string
		admin             string
		hotRoutes         []*types.TokenPairArbRoutes
		passValidateBasic bool
		pass              bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			[]*types.TokenPairArbRoutes{},
			false,
			false,
		},
		{
			"Invalid message (with no token pair arb routes)",
			suite.adminAccount.String(),
			[]*types.TokenPairArbRoutes{},
			true,
			false,
		},
		{
			"Invalid message (mismatch admin)",
			apptesting.CreateRandomAccounts(1)[0].String(),
			[]*types.TokenPairArbRoutes{},
			true,
			false,
		},
		{
			"Valid message (with proper hot routes)",
			"",
			[]*types.TokenPairArbRoutes{&validArbRoutes},
			true,
			true,
		},
		{
			"Invalid message (with duplicate hot routes)",
			suite.adminAccount.String(),
			[]*types.TokenPairArbRoutes{&validArbRoutes, &validArbRoutes},
			false,
			false,
		},
		{
			"Invalid message (with a 4 hop hot route)",
			suite.adminAccount.String(),
			[]*types.TokenPairArbRoutes{&notThreePoolArbRoutes},
			false,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.description, func() {
			suite.SetupTest()
			msg := types.NewMsgSetHotRoutes(tc.admin, tc.hotRoutes)
			if tc.pass {
				msg.Admin = suite.adminAccount.String()
			}

			err := msg.ValidateBasic()
			if tc.passValidateBasic {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*suite.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := sdk.WrapSDKContext(suite.Ctx)
			response, err := server.SetHotRoutes(wrappedCtx, msg)
			if tc.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, &types.MsgSetHotRoutesResponse{})

				hotRoutes, err := suite.App.AppKeepers.ProtoRevKeeper.GetAllTokenPairArbRoutes(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.hotRoutes, hotRoutes)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

// TestMsgSetDeveloperAccount tests the MsgSetDeveloperAccount message.
func (suite *KeeperTestSuite) TestMsgSetDeveloperAccount() {
	cases := []struct {
		description       string
		admin             string
		developerAccount  string
		passValidateBasic bool
		pass              bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			apptesting.CreateRandomAccounts(1)[0].String(),
			false,
			false,
		},
		{
			"Invalid message (invalid developer account)",
			suite.adminAccount.String(),
			"developer",
			false,
			false,
		},
		{
			"Invalid message (wrong admin)",
			suite.adminAccount.String(),
			apptesting.CreateRandomAccounts(1)[0].String(),
			true,
			false,
		},
		{
			"Valid message (correct admin)",
			suite.adminAccount.String(),
			apptesting.CreateRandomAccounts(1)[0].String(),
			true,
			true,
		},
	}

	for _, testCase := range cases {
		suite.Run(testCase.description, func() {
			suite.SetupTest()
			msg := types.NewMsgSetDeveloperAccount(testCase.admin, testCase.developerAccount)
			if testCase.pass {
				msg.Admin = suite.adminAccount.String()
			}

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*suite.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := sdk.WrapSDKContext(suite.Ctx)
			response, err := server.SetDeveloperAccount(wrappedCtx, msg)
			if testCase.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, &types.MsgSetDeveloperAccountResponse{})

				developerAccount, err := suite.App.AppKeepers.ProtoRevKeeper.GetDeveloperAccount(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(sdk.MustAccAddressFromBech32(testCase.developerAccount), developerAccount)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestMsgSetMaxRoutesPerTx tests the MsgSetMaxRoutesPerTx message.
func (suite *KeeperTestSuite) TestMsgSetMaxRoutesPerTx() {
	cases := []struct {
		description       string
		admin             string
		maxRoutesPerTx    uint64
		passValidateBasic bool
		pass              bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			1,
			false,
			false,
		},
		{
			"Invalid message (invalid max routes per tx)",
			suite.adminAccount.String(),
			0,
			false,
			false,
		},
		{
			"Invalid message (wrong admin)",
			suite.adminAccount.String(),
			1,
			true,
			false,
		},
		{
			"Valid message (correct admin)",
			suite.adminAccount.String(),
			1,
			true,
			true,
		},
		{
			"Valid message (correct admin, max routes per tx = 15)",
			suite.adminAccount.String(),
			15,
			true,
			true,
		},
		{
			"Invalid message (correct admin, max routes per tx = 1000)",
			suite.adminAccount.String(),
			1000,
			false,
			false,
		},
	}

	for _, testCase := range cases {
		suite.Run(testCase.description, func() {
			suite.SetupTest()
			msg := types.NewMsgSetMaxRoutesPerTx(testCase.admin, testCase.maxRoutesPerTx)
			if testCase.pass {
				msg.Admin = suite.adminAccount.String()
			}

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*suite.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := sdk.WrapSDKContext(suite.Ctx)
			response, err := server.SetMaxRoutesPerTx(wrappedCtx, msg)
			if testCase.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, &types.MsgSetMaxRoutesPerTxResponse{})

				maxRoutesPerTx, err := suite.App.AppKeepers.ProtoRevKeeper.GetMaxRoutesPerTx(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(testCase.maxRoutesPerTx, maxRoutesPerTx)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestMsgSetMaxRoutesPerBlock tests the MsgSetMaxRoutesPerBlock message.
func (suite *KeeperTestSuite) TestMsgSetMaxRoutesPerBlock() {
	cases := []struct {
		description       string
		admin             string
		maxRoutesPerBlock uint64
		passValidateBasic bool
		pass              bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			1,
			false,
			false,
		},
		{
			"Invalid message (invalid max routes per block)",
			suite.adminAccount.String(),
			0,
			false,
			false,
		},
		{
			"Invalid message (wrong admin)",
			suite.adminAccount.String(),
			1,
			true,
			false,
		},
		{
			"Valid message (correct admin)",
			suite.adminAccount.String(),
			1,
			true,
			true,
		},
		{
			"Valid message (correct admin, max routes per block = 150)",
			suite.adminAccount.String(),
			150,
			true,
			true,
		},
		{
			"Invalid message (correct admin, max routes per block = 10000)",
			suite.adminAccount.String(),
			10000,
			false,
			false,
		},
	}

	for _, testCase := range cases {
		suite.Run(testCase.description, func() {
			suite.SetupTest()
			msg := types.NewMsgSetMaxRoutesPerBlock(testCase.admin, testCase.maxRoutesPerBlock)
			if testCase.pass {
				msg.Admin = suite.adminAccount.String()
			}

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*suite.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := sdk.WrapSDKContext(suite.Ctx)
			response, err := server.SetMaxRoutesPerBlock(wrappedCtx, msg)
			if testCase.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, &types.MsgSetMaxRoutesPerBlockResponse{})

				maxRoutesPerBlock, err := suite.App.AppKeepers.ProtoRevKeeper.GetMaxRoutesPerBlock(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(testCase.maxRoutesPerBlock, maxRoutesPerBlock)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestMsgSetPoolWeights tests the MsgSetPoolWeights message.
func (suite *KeeperTestSuite) TestMsgSetPoolWeights() {
	cases := []struct {
		description       string
		admin             string
		poolWeights       types.PoolWeights
		passValidateBasic bool
		pass              bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			types.PoolWeights{
				StableWeight:       1,
				BalancerWeight:     2,
				ConcentratedWeight: 3,
			},
			false,
			false,
		},
		{
			"Invalid message (invalid pool weight)",
			suite.adminAccount.String(),
			types.PoolWeights{
				StableWeight:       0,
				BalancerWeight:     2,
				ConcentratedWeight: 1,
			},
			false,
			false,
		},
		{
			"Invalid message (unset pool weight)",
			suite.adminAccount.String(),
			types.PoolWeights{
				StableWeight: 1,
			},
			false,
			false,
		},
		{
			"Invalid message (wrong admin)",
			suite.adminAccount.String(),
			types.PoolWeights{
				StableWeight:       1,
				BalancerWeight:     2,
				ConcentratedWeight: 3,
			},
			true,
			false,
		},
		{
			"Valid message (correct admin)",
			suite.adminAccount.String(),
			types.PoolWeights{
				StableWeight:       1,
				BalancerWeight:     2,
				ConcentratedWeight: 3,
			},
			true,
			true,
		},
	}

	for _, testCase := range cases {
		suite.Run(testCase.description, func() {
			suite.SetupTest()
			msg := types.NewMsgSetPoolWeights(testCase.admin, testCase.poolWeights)
			if testCase.pass {
				msg.Admin = suite.adminAccount.String()
			}

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*suite.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := sdk.WrapSDKContext(suite.Ctx)
			response, err := server.SetPoolWeights(wrappedCtx, msg)
			if testCase.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, &types.MsgSetPoolWeightsResponse{})

				poolWeights, err := suite.App.AppKeepers.ProtoRevKeeper.GetPoolWeights(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(testCase.poolWeights, *poolWeights)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
