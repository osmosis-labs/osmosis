package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

type MsgsTestSuite struct {
	suite.Suite
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) TestMsgSetHotRoutes() {
	cases := []struct {
		description string
		admin       string
		hotRoutes   []*types.TokenPairArbRoutes
		pass        bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			[]*types.TokenPairArbRoutes{},
			false,
		},
		{
			"Valid message (no arb routes)",
			createAccount().String(),
			[]*types.TokenPairArbRoutes{},
			true,
		},
		{
			"Valid message (with arb routes)",
			createAccount().String(),
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
		},
		{
			"Invalid message (mismatched arb denoms)",
			createAccount().String(),
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
									TokenOut: "eth",
								},
							},
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			false,
		},
		{
			"Invalid message (with duplicate arb routes)",
			createAccount().String(),
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
		},
		{
			"Invalid message (with missing trade)",
			createAccount().String(),
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
		},
		{
			"Invalid message (with invalid route length)",
			createAccount().String(),
			[]*types.TokenPairArbRoutes{
				{
					ArbRoutes: []*types.Route{
						{
							Trades: []*types.Trade{
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
		},
		{
			"Valid message (with multiple routes)",
			createAccount().String(),
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
						{
							Trades: []*types.Trade{
								{
									Pool:     0,
									TokenIn:  "Juno",
									TokenOut: types.OsmosisDenomination,
								},
								{
									Pool:     5,
									TokenIn:  types.OsmosisDenomination,
									TokenOut: "Juno",
								},
							},
						},
					},
					TokenIn:  types.OsmosisDenomination,
					TokenOut: "Juno",
				},
			},
			true,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			msg := types.NewMsgSetHotRoutes(tc.admin, tc.hotRoutes)
			err := msg.ValidateBasic()
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *MsgsTestSuite) TestMsgSetDeveloperAccount() {
	cases := []struct {
		description string
		admin       string
		developer   string
		pass        bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			createAccount().String(),
			false,
		},
		{
			"Invalid message (invalid developer)",
			createAccount().String(),
			"developer",
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			createAccount().String(),
			true,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			msg := types.NewMsgSetDeveloperAccount(tc.admin, tc.developer)
			err := msg.ValidateBasic()
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *MsgsTestSuite) TestMsgSetMaxPoolPointsPerTx() {
	cases := []struct {
		description        string
		admin              string
		maxPoolPointsPerTx uint64
		pass               bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			1,
			false,
		},
		{
			"Invalid message (invalid max routes)",
			createAccount().String(),
			0,
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			1,
			true,
		},
		{
			"Invalid message (invalid max routes)",
			createAccount().String(),
			types.MaxPoolPointsPerTx + 1,
			false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			msg := types.NewMsgSetMaxPoolPointsPerTx(tc.admin, tc.maxPoolPointsPerTx)
			err := msg.ValidateBasic()
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *MsgsTestSuite) TestMsgSetMaxPoolPointsPerBlock() {
	cases := []struct {
		description           string
		admin                 string
		maxPoolPointsPerBlock uint64
		pass                  bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			1,
			false,
		},
		{
			"Invalid message (invalid max routes)",
			createAccount().String(),
			0,
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			10,
			true,
		},
		{
			"Invalid message (invalid max routes)",
			createAccount().String(),
			types.MaxPoolPointsPerBlock + 1,
			false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			msg := types.NewMsgSetMaxPoolPointsPerBlock(tc.admin, tc.maxPoolPointsPerBlock)
			err := msg.ValidateBasic()
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *MsgsTestSuite) TestMsgSetPoolWeights() {
	cases := []struct {
		description string
		admin       string
		poolWeights types.PoolWeights
		pass        bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			types.PoolWeights{
				BalancerWeight:     1,
				StableWeight:       1,
				ConcentratedWeight: 1,
			},
			false,
		},
		{
			"Invalid message (invalid pool weights)",
			createAccount().String(),
			types.PoolWeights{
				BalancerWeight:     0,
				StableWeight:       1,
				ConcentratedWeight: 1,
			},
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			types.PoolWeights{
				BalancerWeight:     1,
				StableWeight:       1,
				ConcentratedWeight: 1,
			},
			true,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			msg := types.NewMsgSetPoolWeights(tc.admin, tc.poolWeights)
			err := msg.ValidateBasic()
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *MsgsTestSuite) TestMsgSetBaseDenoms() {
	cases := []struct {
		description string
		admin       string
		baseDenoms  []*types.BaseDenom
		pass        bool
	}{
		{
			"Invalid message (invalid admin)",
			"admin",
			[]*types.BaseDenom{},
			false,
		},
		{
			"Invalid message (empty base denoms)",
			createAccount().String(),
			[]*types.BaseDenom{},
			false,
		},
		{
			"Invalid message (base denoms does not start with osmosis)",
			createAccount().String(),
			[]*types.BaseDenom{
				{
					Denom:    types.AtomDenomination,
					StepSize: sdk.NewInt(10),
				},
			},
			false,
		},
		{
			"Invalid message (invalid step size)",
			createAccount().String(),
			[]*types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(0),
				},
			},
			false,
		},
		{
			"Invalid message (duplicate base denoms)",
			createAccount().String(),
			[]*types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(1),
				},
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(1),
				},
			},
			false,
		},
		{
			"Valid message",
			createAccount().String(),
			[]*types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(1),
				},
			},
			true,
		},
		{
			"Valid message",
			createAccount().String(),
			[]*types.BaseDenom{
				{
					Denom:    types.OsmosisDenomination,
					StepSize: sdk.NewInt(1),
				},
				{
					Denom:    types.AtomDenomination,
					StepSize: sdk.NewInt(1),
				},
				{
					Denom:    "testDenom",
					StepSize: sdk.NewInt(1),
				},
			},
			true,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			msg := types.NewMsgSetBaseDenoms(tc.admin, tc.baseDenoms)
			err := msg.ValidateBasic()
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func createAccount() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	return sdk.AccAddress(pk.Address())
}
