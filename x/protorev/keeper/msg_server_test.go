package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// TestMsgSetHotRoutes tests the MsgSetHotRoutes message.
func (suite *KeeperTestSuite) TestMsgSetHotRoutes() {
	validStepSize := sdk.NewInt(1_000_000)
	invalidStepSize := sdk.NewInt(0)

	testCases := []struct {
		description       string
		admin             string
		hotRoutes         []types.TokenPairArbRoutes
		passValidateBasic bool
		pass              bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			[]types.TokenPairArbRoutes{},
			false,
			false,
		},
		{
			"Invalid message (mismatch admin)",
			apptesting.CreateRandomAccounts(1)[0].String(),
			[]types.TokenPairArbRoutes{},
			true,
			false,
		},
		{
			"Valid message (with proper hot routes)",
			suite.adminAccount.String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
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
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
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
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
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
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
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
									TokenOut: "Atom",
								},
							},
							StepSize: validStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			false,
			false,
		},
		{
			"Invalid message (with proper hot routes)",
			suite.adminAccount.String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
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
									TokenOut: "Atom",
								},
							},
							StepSize: invalidStepSize,
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			false,
			false,
		},
		{
			"Invalid message with nil step size (with proper hot routes)",
			suite.adminAccount.String(),
			[]types.TokenPairArbRoutes{
				{
					ArbRoutes: []types.Route{
						{
							Trades: []types.Trade{
								{
									Pool:     1,
									TokenIn:  "Atom",
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
									TokenOut: "Atom",
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
			msg := types.NewMsgSetHotRoutes(tc.admin, tc.hotRoutes)

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
			apptesting.CreateRandomAccounts(1)[0].String(),
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
			msg := types.NewMsgSetDeveloperAccount(testCase.admin, testCase.developerAccount)

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

// TestMsgSetMaxPoolPointsPerTx tests the MsgSetMaxPoolPointsPerTx message.
func (suite *KeeperTestSuite) TestMsgSetMaxPoolPointsPerTx() {
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
			apptesting.CreateRandomAccounts(1)[0].String(),
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
			"Valid message (correct admin, valid max pool points per tx)",
			suite.adminAccount.String(),
			types.MaxPoolPointsPerTx - 1,
			true,
			true,
		},
		{
			"Invalid message (correct admin, too many max pool points per tx)",
			suite.adminAccount.String(),
			types.MaxPoolPointsPerTx + 1,
			false,
			false,
		},
	}

	for _, testCase := range cases {
		suite.Run(testCase.description, func() {
			msg := types.NewMsgSetMaxPoolPointsPerTx(testCase.admin, testCase.maxPoolPointsPerTx)

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

// TestMsgSetMaxPoolPointsPerBlock tests the MsgSetMaxPoolPointsPerBlock message.
func (suite *KeeperTestSuite) TestMsgSetMaxPoolPointsPerBlock() {
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
			apptesting.CreateRandomAccounts(1)[0].String(),
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
			"Valid message (correct admin, valid max pool points per block)",
			suite.adminAccount.String(),
			types.MaxPoolPointsPerBlock - 1,
			true,
			true,
		},
		{
			"Invalid message (correct admin, too many max routes per block)",
			suite.adminAccount.String(),
			types.MaxPoolPointsPerBlock + 1,
			false,
			false,
		},
	}

	for _, testCase := range cases {
		suite.Run(testCase.description, func() {
			msg := types.NewMsgSetMaxPoolPointsPerBlock(testCase.admin, testCase.maxPoolPointsPerBlock)

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
			apptesting.CreateRandomAccounts(1)[0].String(),
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
			msg := types.NewMsgSetPoolWeights(testCase.admin, testCase.poolWeights)

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
				suite.Require().Equal(testCase.poolWeights, poolWeights)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestMsgSetBaseDenoms tests the MsgSetBaseDenoms message.
func (suite *KeeperTestSuite) TestMsgSetBaseDenoms() {
	cases := []struct {
		description       string
		admin             string
		baseDenoms        []types.BaseDenom
		passValidateBasic bool
		pass              bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(1_000_000),
				},
			},
			false,
			false,
		},
		{
			"Invalid message (invalid base denoms must start with osmo)",
			suite.adminAccount.String(),
			[]types.BaseDenom{
				{
					Denom:    "Atom",
					StepSize: sdk.NewInt(1_000_000),
				},
			},
			false,
			false,
		},
		{
			"Invalid message (invalid step size)",
			suite.adminAccount.String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(0),
				},
			},
			false,
			false,
		},
		{
			"Invalid message (wrong admin)",
			apptesting.CreateRandomAccounts(1)[0].String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(1_000_000),
				},
			},
			true,
			false,
		},
		{
			"Valid message (correct admin)",
			suite.adminAccount.String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(1_000_000),
				},
			},
			true,
			true,
		},
	}

	for _, testCase := range cases {
		suite.Run(testCase.description, func() {
			msg := types.NewMsgSetBaseDenoms(testCase.admin, testCase.baseDenoms)

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*suite.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := sdk.WrapSDKContext(suite.Ctx)
			response, err := server.SetBaseDenoms(wrappedCtx, msg)
			if testCase.pass {
				suite.Require().NoError(err)
				suite.Require().Equal(response, &types.MsgSetBaseDenomsResponse{})

				baseDenoms, err := suite.App.AppKeepers.ProtoRevKeeper.GetAllBaseDenoms(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(testCase.baseDenoms, baseDenoms)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
