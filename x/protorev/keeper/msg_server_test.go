package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// TestMsgSetHotRoutes tests the MsgSetHotRoutes message.
func (s *KeeperTestSuite) TestMsgSetHotRoutes() {
	validStepSize := osmomath.NewInt(1_000_000)
	invalidStepSize := osmomath.NewInt(0)

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
			s.adminAccount.String(),
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
			s.adminAccount.String(),
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
			s.adminAccount.String(),
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
			s.adminAccount.String(),
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
		s.Run(tc.description, func() {
			msg := types.NewMsgSetHotRoutes(tc.admin, tc.hotRoutes)

			err := msg.ValidateBasic()
			if tc.passValidateBasic {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*s.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := s.Ctx
			response, err := server.SetHotRoutes(wrappedCtx, msg)
			if tc.pass {
				s.Require().NoError(err)
				s.Require().Equal(response, &types.MsgSetHotRoutesResponse{})

				hotRoutes, err := s.App.AppKeepers.ProtoRevKeeper.GetAllTokenPairArbRoutes(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(tc.hotRoutes, hotRoutes)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestMsgSetDeveloperAccount tests the MsgSetDeveloperAccount message.
func (s *KeeperTestSuite) TestMsgSetDeveloperAccount() {
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
			s.adminAccount.String(),
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
			s.adminAccount.String(),
			apptesting.CreateRandomAccounts(1)[0].String(),
			true,
			true,
		},
	}

	for _, testCase := range cases {
		s.Run(testCase.description, func() {
			msg := types.NewMsgSetDeveloperAccount(testCase.admin, testCase.developerAccount)

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*s.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := s.Ctx
			response, err := server.SetDeveloperAccount(wrappedCtx, msg)
			if testCase.pass {
				s.Require().NoError(err)
				s.Require().Equal(response, &types.MsgSetDeveloperAccountResponse{})

				developerAccount, err := s.App.AppKeepers.ProtoRevKeeper.GetDeveloperAccount(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(sdk.MustAccAddressFromBech32(testCase.developerAccount), developerAccount)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestMsgSetMaxPoolPointsPerTx tests the MsgSetMaxPoolPointsPerTx message.
func (s *KeeperTestSuite) TestMsgSetMaxPoolPointsPerTx() {
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
			s.adminAccount.String(),
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
			s.adminAccount.String(),
			1,
			true,
			true,
		},
		{
			"Valid message (correct admin, valid max pool points per tx)",
			s.adminAccount.String(),
			types.MaxPoolPointsPerTx - 1,
			true,
			true,
		},
		{
			"Invalid message (correct admin, too many max pool points per tx)",
			s.adminAccount.String(),
			types.MaxPoolPointsPerTx + 1,
			false,
			false,
		},
	}

	for _, testCase := range cases {
		s.Run(testCase.description, func() {
			msg := types.NewMsgSetMaxPoolPointsPerTx(testCase.admin, testCase.maxPoolPointsPerTx)

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*s.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := s.Ctx
			response, err := server.SetMaxPoolPointsPerTx(wrappedCtx, msg)
			if testCase.pass {
				s.Require().NoError(err)
				s.Require().Equal(response, &types.MsgSetMaxPoolPointsPerTxResponse{})

				maxRoutesPerTx, err := s.App.AppKeepers.ProtoRevKeeper.GetMaxPointsPerTx(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(testCase.maxPoolPointsPerTx, maxRoutesPerTx)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestMsgSetMaxPoolPointsPerBlock tests the MsgSetMaxPoolPointsPerBlock message.
func (s *KeeperTestSuite) TestMsgSetMaxPoolPointsPerBlock() {
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
			s.adminAccount.String(),
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
			s.adminAccount.String(),
			50,
			true,
			true,
		},
		{
			"Invalid message (correct admin but less points than max pool points per tx)",
			s.adminAccount.String(),
			17,
			true,
			false,
		},
		{
			"Valid message (correct admin, valid max pool points per block)",
			s.adminAccount.String(),
			types.MaxPoolPointsPerBlock - 1,
			true,
			true,
		},
		{
			"Invalid message (correct admin, too many max routes per block)",
			s.adminAccount.String(),
			types.MaxPoolPointsPerBlock + 1,
			false,
			false,
		},
	}

	for _, testCase := range cases {
		s.Run(testCase.description, func() {
			msg := types.NewMsgSetMaxPoolPointsPerBlock(testCase.admin, testCase.maxPoolPointsPerBlock)

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*s.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := s.Ctx
			response, err := server.SetMaxPoolPointsPerBlock(wrappedCtx, msg)
			if testCase.pass {
				s.Require().NoError(err)
				s.Require().Equal(response, &types.MsgSetMaxPoolPointsPerBlockResponse{})

				maxRoutesPerBlock, err := s.App.AppKeepers.ProtoRevKeeper.GetMaxPointsPerBlock(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(testCase.maxPoolPointsPerBlock, maxRoutesPerBlock)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestMsgSetPoolTypeInfo tests the MsgSetInfoByPoolType message.
func (s *KeeperTestSuite) TestMsgSetPoolTypeInfo() {
	cases := []struct {
		description       string
		admin             string
		poolInfo          types.InfoByPoolType
		passValidateBasic bool
		pass              bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			types.InfoByPoolType{
				Stable:       types.StablePoolInfo{Weight: 1},
				Balancer:     types.BalancerPoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm:     types.CosmwasmPoolInfo{WeightMaps: nil},
			},
			false,
			false,
		},
		{
			"Invalid message (invalid pool weight)",
			s.adminAccount.String(),
			types.InfoByPoolType{
				Stable:       types.StablePoolInfo{Weight: 0},
				Balancer:     types.BalancerPoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm:     types.CosmwasmPoolInfo{WeightMaps: nil},
			},
			false,
			false,
		},
		{
			"Invalid message (unset pool weight)",
			s.adminAccount.String(),
			types.InfoByPoolType{
				Stable:       types.StablePoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm:     types.CosmwasmPoolInfo{WeightMaps: nil},
			},
			false,
			false,
		},
		{
			"Invalid message (wrong admin)",
			apptesting.CreateRandomAccounts(1)[0].String(),
			types.InfoByPoolType{
				Stable:       types.StablePoolInfo{Weight: 1},
				Balancer:     types.BalancerPoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm:     types.CosmwasmPoolInfo{WeightMaps: nil},
			},
			true,
			false,
		},
		{
			"Valid message (correct admin)",
			s.adminAccount.String(),
			types.InfoByPoolType{
				Stable:       types.StablePoolInfo{Weight: 1},
				Balancer:     types.BalancerPoolInfo{Weight: 1},
				Concentrated: types.ConcentratedPoolInfo{Weight: 1, MaxTicksCrossed: 1},
				Cosmwasm:     types.CosmwasmPoolInfo{WeightMaps: nil},
			},
			true,
			true,
		},
	}

	for _, testCase := range cases {
		s.Run(testCase.description, func() {
			msg := types.NewMsgSetPoolTypeInfo(testCase.admin, testCase.poolInfo)

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*s.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := s.Ctx
			response, err := server.SetInfoByPoolType(wrappedCtx, msg)
			if testCase.pass {
				s.Require().NoError(err)
				s.Require().Equal(response, &types.MsgSetInfoByPoolTypeResponse{})

				poolWeights := s.App.AppKeepers.ProtoRevKeeper.GetInfoByPoolType(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(testCase.poolInfo, poolWeights)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// TestMsgSetBaseDenoms tests the MsgSetBaseDenoms message.
func (s *KeeperTestSuite) TestMsgSetBaseDenoms() {
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
					StepSize: osmomath.NewInt(1_000_000),
				},
			},
			false,
			false,
		},
		{
			"Invalid message (invalid base denoms must start with osmo)",
			s.adminAccount.String(),
			[]types.BaseDenom{
				{
					Denom:    "Atom",
					StepSize: osmomath.NewInt(1_000_000),
				},
			},
			false,
			false,
		},
		{
			"Invalid message (invalid step size)",
			s.adminAccount.String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(0),
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
					StepSize: osmomath.NewInt(1_000_000),
				},
			},
			true,
			false,
		},
		{
			"Valid message (correct admin)",
			s.adminAccount.String(),
			[]types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: osmomath.NewInt(1_000_000),
				},
			},
			true,
			true,
		},
	}

	for _, testCase := range cases {
		s.Run(testCase.description, func() {
			msg := types.NewMsgSetBaseDenoms(testCase.admin, testCase.baseDenoms)

			err := msg.ValidateBasic()
			if testCase.passValidateBasic {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				return
			}

			server := keeper.NewMsgServer(*s.App.AppKeepers.ProtoRevKeeper)
			wrappedCtx := s.Ctx
			response, err := server.SetBaseDenoms(wrappedCtx, msg)
			if testCase.pass {
				s.Require().NoError(err)
				s.Require().Equal(response, &types.MsgSetBaseDenomsResponse{})

				baseDenoms, err := s.App.AppKeepers.ProtoRevKeeper.GetAllBaseDenoms(s.Ctx)
				s.Require().NoError(err)
				s.Require().Equal(testCase.baseDenoms, baseDenoms)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
