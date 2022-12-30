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

func (suite *MsgsTestSuite) TestMsgSetMaxRoutesPerTx() {
	cases := []struct {
		description string
		admin       string
		maxRoutes   uint64
		pass        bool
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
			100,
			false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			msg := types.NewMsgSetMaxRoutesPerTx(tc.admin, tc.maxRoutes)
			err := msg.ValidateBasic()
			if tc.pass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *MsgsTestSuite) TestMsgSetMaxRoutesPerBlock() {
	cases := []struct {
		description string
		admin       string
		maxRoutes   uint64
		pass        bool
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
			300,
			false,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.description, func() {
			msg := types.NewMsgSetMaxRoutesPerBlock(tc.admin, tc.maxRoutes)
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
