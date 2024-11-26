package keeper_test

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
)

// TestMintDenomMsg tests TypeMsgMint message is emitted on a successful mint
func (s *KeeperTestSuite) TestMintDenomMsg() {
	// Create a denom
	s.CreateDefaultDenom()

	for _, tc := range []struct {
		desc                  string
		amount                int64
		mintDenom             string
		admin                 string
		valid                 bool
		expectedMessageEvents int
	}{
		{
			desc:      "denom does not exist",
			amount:    10,
			mintDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:     s.TestAccs[0].String(),
			valid:     false,
		},
		{
			desc:                  "success case",
			amount:                10,
			mintDenom:             s.defaultDenom,
			admin:                 s.TestAccs[0].String(),
			valid:                 true,
			expectedMessageEvents: 1,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test mint message
			_, err := s.msgServer.Mint(ctx, types.NewMsgMint(tc.admin, sdk.NewInt64Coin(tc.mintDenom, 10)))
			if tc.valid {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
			// Ensure current number and type of event is emitted
			s.AssertEventEmitted(ctx, types.TypeMsgMint, tc.expectedMessageEvents)
		})
	}
}

// TestBurnDenomMsg tests TypeMsgBurn message is emitted on a successful burn
func (s *KeeperTestSuite) TestBurnDenomMsg() {
	// Create a denom.
	s.CreateDefaultDenom()
	// mint 10 default token for testAcc[0]
	_, err := s.msgServer.Mint(s.Ctx, types.NewMsgMint(s.TestAccs[0].String(), sdk.NewInt64Coin(s.defaultDenom, 10)))
	s.Require().NoError(err)

	for _, tc := range []struct {
		desc                  string
		amount                int64
		burnDenom             string
		admin                 string
		valid                 bool
		expectedMessageEvents int
	}{
		{
			desc:      "denom does not exist",
			burnDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:     s.TestAccs[0].String(),
			valid:     false,
		},
		{
			desc:                  "success case",
			burnDenom:             s.defaultDenom,
			admin:                 s.TestAccs[0].String(),
			valid:                 true,
			expectedMessageEvents: 1,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test burn message
			_, err := s.msgServer.Burn(ctx, types.NewMsgBurn(tc.admin, sdk.NewInt64Coin(tc.burnDenom, 10)))
			if tc.valid {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
			// Ensure current number and type of event is emitted
			s.AssertEventEmitted(ctx, types.TypeMsgBurn, tc.expectedMessageEvents)
		})
	}
}

func (s *KeeperTestSuite) TestForceTransferMsg() {
	// Create a denom
	s.CreateDefaultDenom()

	s.Run(fmt.Sprintf("test force transfer"), func() {
		mintAmt := sdk.NewInt64Coin(s.defaultDenom, 10)

		_, err := s.msgServer.Mint(s.Ctx, types.NewMsgMint(s.TestAccs[0].String(), mintAmt))

		govModAcc := s.App.AccountKeeper.GetModuleAccount(s.Ctx, govtypes.ModuleName)

		err = s.App.BankKeeper.SendCoins(s.Ctx, s.TestAccs[0], govModAcc.GetAddress(), sdk.NewCoins(mintAmt))
		s.Require().NoError(err)

		_, err = s.msgServer.ForceTransfer(s.Ctx, types.NewMsgForceTransfer(s.TestAccs[0].String(), mintAmt, govModAcc.GetAddress().String(), s.TestAccs[1].String()))
		s.Require().ErrorContains(err, "send from module acc not available")
	})
}

// TestCreateDenomMsg tests TypeMsgCreateDenom message is emitted on a successful denom creation
func (s *KeeperTestSuite) TestCreateDenomMsg() {
	for _, tc := range []struct {
		desc                  string
		subdenom              string
		valid                 bool
		expectedMessageEvents int
	}{
		{
			desc:     "subdenom too long",
			subdenom: "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
			valid:    false,
		},
		{
			desc:                  "success case: defaultDenomCreationFee",
			subdenom:              "evmos",
			valid:                 true,
			expectedMessageEvents: 1,
		},
	} {
		s.SetupTest()
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Set denom creation fee in params
			// Test create denom message
			_, err := s.msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(s.TestAccs[0].String(), tc.subdenom))
			if tc.valid {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
			// Ensure current number and type of event is emitted
			s.AssertEventEmitted(ctx, types.TypeMsgCreateDenom, tc.expectedMessageEvents)
		})
	}
}

// TestChangeAdminDenomMsg tests TypeMsgChangeAdmin message is emitted on a successful admin change
func (s *KeeperTestSuite) TestChangeAdminDenomMsg() {
	for _, tc := range []struct {
		desc                    string
		msgChangeAdmin          func(denom string) *types.MsgChangeAdmin
		expectedChangeAdminPass bool
		expectedAdminIndex      int
		msgMint                 func(denom string) *types.MsgMint
		expectedMintPass        bool
		expectedMessageEvents   int
	}{
		{
			desc: "non-admins can't change the existing admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(s.TestAccs[1].String(), denom, s.TestAccs[2].String())
			},
			expectedChangeAdminPass: false,
			expectedAdminIndex:      0,
		},
		{
			desc: "success change admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(s.TestAccs[0].String(), denom, s.TestAccs[1].String())
			},
			expectedAdminIndex:      1,
			expectedChangeAdminPass: true,
			expectedMessageEvents:   1,
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(s.TestAccs[1].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintPass: true,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// setup test
			s.SetupTest()
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Create a denom and mint
			res, err := s.msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(s.TestAccs[0].String(), "bitcoin"))
			s.Require().NoError(err)
			testDenom := res.GetNewTokenDenom()
			_, err = s.msgServer.Mint(ctx, types.NewMsgMint(s.TestAccs[0].String(), sdk.NewInt64Coin(testDenom, 10)))
			s.Require().NoError(err)
			// Test change admin message
			_, err = s.msgServer.ChangeAdmin(ctx, tc.msgChangeAdmin(testDenom))
			if tc.expectedChangeAdminPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
			// Ensure current number and type of event is emitted
			s.AssertEventEmitted(ctx, types.TypeMsgChangeAdmin, tc.expectedMessageEvents)
		})
	}
}

// TestSetDenomMetaDataMsg tests TypeMsgSetDenomMetadata message is emitted on a successful denom metadata change
func (s *KeeperTestSuite) TestSetDenomMetaDataMsg() {
	// setup test
	s.SetupTest()
	s.CreateDefaultDenom()

	for _, tc := range []struct {
		desc                  string
		msgSetDenomMetadata   types.MsgSetDenomMetadata
		expectedPass          bool
		expectedMessageEvents int
	}{
		{
			desc: "successful set denom metadata",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(s.TestAccs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    s.defaultDenom,
						Exponent: 0,
					},
					{
						Denom:    appparams.BaseCoinUnit,
						Exponent: 6,
					},
				},
				Base:    s.defaultDenom,
				Display: appparams.BaseCoinUnit,
				Name:    "OSMO",
				Symbol:  "OSMO",
			}),
			expectedPass:          true,
			expectedMessageEvents: 1,
		},
		{
			desc: "non existent factory denom name",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(s.TestAccs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    fmt.Sprintf("factory/%s/litecoin", s.TestAccs[0].String()),
						Exponent: 0,
					},
					{
						Denom:    appparams.BaseCoinUnit,
						Exponent: 6,
					},
				},
				Base:    fmt.Sprintf("factory/%s/litecoin", s.TestAccs[0].String()),
				Display: appparams.BaseCoinUnit,
				Name:    "OSMO",
				Symbol:  "OSMO",
			}),
			expectedPass: false,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			ctx := s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))
			// Test set denom metadata message
			_, err := s.msgServer.SetDenomMetadata(ctx, &tc.msgSetDenomMetadata)
			if tc.expectedPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
			// Ensure current number and type of event is emitted
			s.AssertEventEmitted(ctx, types.TypeMsgSetDenomMetadata, tc.expectedMessageEvents)
		})
	}
}
