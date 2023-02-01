package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/app/apptesting"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

// TestMsgSetHotRoutes tests the MsgSetHotRoutes message.
func (suite *KeeperTestSuite) TestMsgSetHotRoutes() {
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
			[]*types.TokenPairArbRoutes{
				{
					ArbRoutes: []*types.Route{
						{
							Trades: []*types.Trade{
								{
									Pool:     1,
									TokenIn:  types.AtomDenomination,
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: types.AtomDenomination,
								},
							},
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			true,
			true,
		},
		{
			"Invalid message (with duplicate hot routes)",
			suite.adminAccount.String(),
			[]*types.TokenPairArbRoutes{
				{
					ArbRoutes: []*types.Route{
						{
							Trades: []*types.Trade{
								{
									Pool:     1,
									TokenIn:  types.AtomDenomination,
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: types.AtomDenomination,
								},
							},
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
				{
					ArbRoutes: []*types.Route{
						{
							Trades: []*types.Trade{
								{
									Pool:     1,
									TokenIn:  types.AtomDenomination,
									TokenOut: "Juno",
								},
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     3,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: types.AtomDenomination,
								},
							},
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
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
		description        string
		admin              string
		maxPoolPointsPerTx uint64
		passValidateBasic  bool
		pass               bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			1,
			false,
			false,
		},
		{
			"Invalid message (invalid max pool points per tx)",
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
			"Valid message (correct admin, max pool points per tx = 15)",
			suite.adminAccount.String(),
			15,
			true,
			true,
		},
		{
			"Invalid message (correct admin, max pool points per tx = 1000)",
			suite.adminAccount.String(),
			1000,
			false,
			false,
		},
	}

	for _, testCase := range cases {
		suite.Run(testCase.description, func() {
			suite.SetupTest()
			msg := types.NewMsgSetMaxPoolPointsPerTx(testCase.admin, testCase.maxPoolPointsPerTx)
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
			response, err := server.SetMaxPoolPointsPerTx(wrappedCtx, msg)
			if testCase.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, &types.MsgSetMaxPoolPointsPerTxResponse{})

				maxRoutesPerTx, err := suite.App.AppKeepers.ProtoRevKeeper.GetMaxPointsPerTx(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(testCase.maxPoolPointsPerTx, maxRoutesPerTx)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestMsgSetMaxRoutesPerBlock tests the MsgSetMaxRoutesPerBlock message.
func (suite *KeeperTestSuite) TestMsgSetMaxRoutesPerBlock() {
	cases := []struct {
		description           string
		admin                 string
		maxPoolPointsPerBlock uint64
		passValidateBasic     bool
		pass                  bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			1,
			false,
			false,
		},
		{
			"Invalid message (invalid max pool points per block)",
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
			"Valid message (correct admin, max pool points per block = 150)",
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
			msg := types.NewMsgSetMaxPoolPointsPerBlock(testCase.admin, testCase.maxPoolPointsPerBlock)
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
			response, err := server.SetMaxPoolPointsPerBlock(wrappedCtx, msg)
			if testCase.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, &types.MsgSetMaxPoolPointsPerBlockResponse{})

				maxRoutesPerBlock, err := suite.App.AppKeepers.ProtoRevKeeper.GetMaxPointsPerBlock(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(testCase.maxPoolPointsPerBlock, maxRoutesPerBlock)
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

				poolWeights := suite.App.AppKeepers.ProtoRevKeeper.GetPoolWeights(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(testCase.poolWeights, *poolWeights)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
