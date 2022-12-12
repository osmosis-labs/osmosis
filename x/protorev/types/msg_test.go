package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

type MsgsTestSuite struct {
	suite.Suite
}

func TestMsgsTestSuite(t *testing.T) {
	suite.Run(t, new(MsgsTestSuite))
}

func (suite *MsgsTestSuite) TestMsgSetHotRoutes() {
	validArbRoutes := types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.AtomDenomination)

	notThreePoolArbRoutes := types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.AtomDenomination)
	extraTrade := types.NewTrade(100000, "a", "b")
	notThreePoolArbRoutes.ArbRoutes = append(notThreePoolArbRoutes.ArbRoutes, &types.Route{[]*types.Trade{&extraTrade}})

	invalidArbDenoms := types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", "juno", "juno")
	mismatchedDenoms := types.CreateSeacherRoutes(3, types.OsmosisDenomination, "ethereum", types.AtomDenomination, types.OsmosisDenomination)
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
			[]*types.TokenPairArbRoutes{&validArbRoutes},
			true,
		},
		{
			"Invalid message (with duplicate arb routes)",
			createAccount().String(),
			[]*types.TokenPairArbRoutes{&validArbRoutes, &validArbRoutes},
			false,
		},
		{
			"Invalid message (with invalid arb routes)",
			createAccount().String(),
			[]*types.TokenPairArbRoutes{&notThreePoolArbRoutes},
			false,
		},
		{
			"Invalid message (with invalid arb denoms)",
			createAccount().String(),
			[]*types.TokenPairArbRoutes{&invalidArbDenoms},
			false,
		},
		{
			"Invalid message (with mismatched arb denoms)",
			createAccount().String(),
			[]*types.TokenPairArbRoutes{&mismatchedDenoms},
			false,
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

func createAccount() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	return sdk.AccAddress(pk.Address())
}
