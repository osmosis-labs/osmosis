package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
)

func (s *KeeperTestSuite) TestAdminMsgs() {
	addr0bal := int64(0)
	addr1bal := int64(0)

	bankKeeper := s.App.BankKeeper

	s.CreateDefaultDenom()
	// Make sure that the admin is set correctly
	queryRes, err := s.queryClient.DenomAuthorityMetadata(s.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: s.defaultDenom,
	})
	s.Require().NoError(err)
	s.Require().Equal(s.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

	// Test minting to admins own account
	_, err = s.msgServer.Mint(s.Ctx, types.NewMsgMint(s.TestAccs[0].String(), sdk.NewInt64Coin(s.defaultDenom, 10)))
	addr0bal += 10
	s.Require().NoError(err)
	s.Require().True(bankKeeper.GetBalance(s.Ctx, s.TestAccs[0], s.defaultDenom).Amount.Int64() == addr0bal, bankKeeper.GetBalance(s.Ctx, s.TestAccs[0], s.defaultDenom))

	// Test minting to a different account
	_, err = s.msgServer.Mint(s.Ctx, types.NewMsgMintTo(s.TestAccs[0].String(), sdk.NewInt64Coin(s.defaultDenom, 10), s.TestAccs[1].String()))
	addr1bal += 10
	s.Require().NoError(err)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[1], s.defaultDenom).Amount.Int64() == addr1bal, s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[1], s.defaultDenom))

	// Test force transferring
	_, err = s.msgServer.ForceTransfer(s.Ctx, types.NewMsgForceTransfer(s.TestAccs[0].String(), sdk.NewInt64Coin(s.defaultDenom, 5), s.TestAccs[1].String(), s.TestAccs[0].String()))
	addr1bal -= 5
	addr0bal += 5
	s.Require().NoError(err)
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], s.defaultDenom).Amount.Int64() == addr0bal, s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[0], s.defaultDenom))
	s.Require().True(s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[1], s.defaultDenom).Amount.Int64() == addr1bal, s.App.BankKeeper.GetBalance(s.Ctx, s.TestAccs[1], s.defaultDenom))

	// Test burning from own account
	_, err = s.msgServer.Burn(s.Ctx, types.NewMsgBurn(s.TestAccs[0].String(), sdk.NewInt64Coin(s.defaultDenom, 5)))
	s.Require().NoError(err)
	s.Require().True(bankKeeper.GetBalance(s.Ctx, s.TestAccs[1], s.defaultDenom).Amount.Int64() == addr1bal)

	// Test Change Admin
	_, err = s.msgServer.ChangeAdmin(s.Ctx, types.NewMsgChangeAdmin(s.TestAccs[0].String(), s.defaultDenom, s.TestAccs[1].String()))
	s.Require().NoError(err)
	queryRes, err = s.queryClient.DenomAuthorityMetadata(s.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: s.defaultDenom,
	})
	s.Require().NoError(err)
	s.Require().Equal(s.TestAccs[1].String(), queryRes.AuthorityMetadata.Admin)

	// Make sure old admin can no longer do actions
	_, err = s.msgServer.Burn(s.Ctx, types.NewMsgBurn(s.TestAccs[0].String(), sdk.NewInt64Coin(s.defaultDenom, 5)))
	s.Require().Error(err)

	// Make sure the new admin works
	_, err = s.msgServer.Mint(s.Ctx, types.NewMsgMint(s.TestAccs[1].String(), sdk.NewInt64Coin(s.defaultDenom, 5)))
	addr1bal += 5
	s.Require().NoError(err)
	s.Require().True(bankKeeper.GetBalance(s.Ctx, s.TestAccs[1], s.defaultDenom).Amount.Int64() == addr1bal)

	// Try setting admin to empty
	_, err = s.msgServer.ChangeAdmin(s.Ctx, types.NewMsgChangeAdmin(s.TestAccs[1].String(), s.defaultDenom, ""))
	s.Require().NoError(err)
	queryRes, err = s.queryClient.DenomAuthorityMetadata(s.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: s.defaultDenom,
	})
	s.Require().NoError(err)
	s.Require().Equal("", queryRes.AuthorityMetadata.Admin)
}

// TestMintDenom ensures the following properties of the MintMessage:
// * No one can mint tokens for a denom that doesn't exist
// * Only the admin of a denom can mint tokens for it
// * The admin of a denom can mint tokens for it
func (s *KeeperTestSuite) TestMintDenom() {
	balances := make(map[string]int64)
	for _, acc := range s.TestAccs {
		balances[acc.String()] = 0
	}

	// Create a denom
	s.CreateDefaultDenom()

	for _, tc := range []struct {
		desc       string
		mintMsg    types.MsgMint
		expectPass bool
	}{
		{
			desc: "denom does not exist",
			mintMsg: *types.NewMsgMint(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin("factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos", 10),
			),
			expectPass: false,
		},
		{
			desc: "mint is not by the admin",
			mintMsg: *types.NewMsgMintTo(
				s.TestAccs[1].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
				s.TestAccs[0].String(),
			),
			expectPass: false,
		},
		{
			desc: "success case - mint to self",
			mintMsg: *types.NewMsgMint(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
			),
			expectPass: true,
		},
		{
			desc: "success case - mint to another address",
			mintMsg: *types.NewMsgMintTo(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
				s.TestAccs[1].String(),
			),
			expectPass: true,
		},
		{
			desc: "error: try minting non-tokenfactory denom",
			mintMsg: *types.NewMsgMintTo(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(appparams.BaseCoinUnit, 10),
				s.TestAccs[1].String(),
			),
			expectPass: false,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			_, err := s.msgServer.Mint(s.Ctx, &tc.mintMsg)
			if tc.expectPass {
				s.Require().NoError(err)
				balances[tc.mintMsg.MintToAddress] += tc.mintMsg.Amount.Amount.Int64()
			} else {
				s.Require().Error(err)
			}

			mintToAddr, _ := sdk.AccAddressFromBech32(tc.mintMsg.MintToAddress)
			bal := s.App.BankKeeper.GetBalance(s.Ctx, mintToAddr, s.defaultDenom).Amount
			s.Require().Equal(bal.Int64(), balances[tc.mintMsg.MintToAddress])
		})
	}
}

func (s *KeeperTestSuite) TestBurnDenom() {
	// Create a denom.
	s.CreateDefaultDenom()

	// mint 1000 default token for all testAccs
	balances := make(map[string]int64)
	for _, acc := range s.TestAccs {
		_, err := s.msgServer.Mint(s.Ctx, types.NewMsgMintTo(s.TestAccs[0].String(), sdk.NewInt64Coin(s.defaultDenom, 1000), acc.String()))
		s.Require().NoError(err)
		balances[acc.String()] = 1000
	}

	// save sample module account address for testing
	moduleAdress := s.App.AccountKeeper.GetModuleAddress("developer_vesting_unvested")

	for _, tc := range []struct {
		desc       string
		burnMsg    types.MsgBurn
		expectPass bool
	}{
		{
			desc: "denom does not exist",
			burnMsg: *types.NewMsgBurn(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin("factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos", 10),
			),
			expectPass: false,
		},
		{
			desc: "burn is not by the admin",
			burnMsg: *types.NewMsgBurnFrom(
				s.TestAccs[1].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
				s.TestAccs[0].String(),
			),
			expectPass: false,
		},
		{
			desc: "burn more than balance",
			burnMsg: *types.NewMsgBurn(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10000),
			),
			expectPass: false,
		},
		{
			desc: "success case - burn from self",
			burnMsg: *types.NewMsgBurn(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
			),
			expectPass: true,
		},
		{
			desc: "success case - burn from another address",
			burnMsg: *types.NewMsgBurnFrom(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
				s.TestAccs[1].String(),
			),
			expectPass: true,
		},
		{
			desc: "fail case - burn from module account",
			burnMsg: *types.NewMsgBurnFrom(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
				moduleAdress.String(),
			),
			expectPass: false,
		},
		{
			desc: "fail case - burn non-tokenfactory denom",
			burnMsg: *types.NewMsgBurnFrom(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(appparams.BaseCoinUnit, 10),
				moduleAdress.String(),
			),
			expectPass: false,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			_, err := s.msgServer.Burn(s.Ctx, &tc.burnMsg)
			if tc.expectPass {
				s.Require().NoError(err)
				balances[tc.burnMsg.BurnFromAddress] -= tc.burnMsg.Amount.Amount.Int64()
			} else {
				s.Require().Error(err)
			}

			burnFromAddr, _ := sdk.AccAddressFromBech32(tc.burnMsg.BurnFromAddress)
			bal := s.App.BankKeeper.GetBalance(s.Ctx, burnFromAddr, s.defaultDenom).Amount
			s.Require().Equal(bal.Int64(), balances[tc.burnMsg.BurnFromAddress])
		})
	}
}

func (s *KeeperTestSuite) TestForceTransferDenom() {
	// Create a denom.
	s.CreateDefaultDenom()

	// mint 1000 default token for all testAccs
	balances := make(map[string]int64)
	for _, acc := range s.TestAccs {
		_, err := s.msgServer.Mint(s.Ctx, types.NewMsgMintTo(s.TestAccs[0].String(), sdk.NewInt64Coin(s.defaultDenom, 1000), acc.String()))
		s.Require().NoError(err)
		balances[acc.String()] = 1000
	}

	for _, tc := range []struct {
		desc             string
		forceTransferMsg types.MsgForceTransfer
		expectPass       bool
	}{
		{
			desc: "valid force transfer",
			forceTransferMsg: *types.NewMsgForceTransfer(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
				s.TestAccs[1].String(),
				s.TestAccs[2].String(),
			),
			expectPass: true,
		},
		{
			desc: "denom does not exist",
			forceTransferMsg: *types.NewMsgForceTransfer(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin("factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos", 10),
				s.TestAccs[1].String(),
				s.TestAccs[2].String(),
			),
			expectPass: false,
		},
		{
			desc: "forceTransfer is not by the admin",
			forceTransferMsg: *types.NewMsgForceTransfer(
				s.TestAccs[1].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10),
				s.TestAccs[1].String(),
				s.TestAccs[2].String(),
			),
			expectPass: false,
		},
		{
			desc: "forceTransfer is greater than the balance of",
			forceTransferMsg: *types.NewMsgForceTransfer(
				s.TestAccs[0].String(),
				sdk.NewInt64Coin(s.defaultDenom, 10000),
				s.TestAccs[1].String(),
				s.TestAccs[2].String(),
			),
			expectPass: false,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			_, err := s.msgServer.ForceTransfer(s.Ctx, &tc.forceTransferMsg)
			if tc.expectPass {
				s.Require().NoError(err)

				balances[tc.forceTransferMsg.TransferFromAddress] -= tc.forceTransferMsg.Amount.Amount.Int64()
				balances[tc.forceTransferMsg.TransferToAddress] += tc.forceTransferMsg.Amount.Amount.Int64()
			} else {
				s.Require().Error(err)
			}

			fromAddr, err := sdk.AccAddressFromBech32(tc.forceTransferMsg.TransferFromAddress)
			s.Require().NoError(err)
			fromBal := s.App.BankKeeper.GetBalance(s.Ctx, fromAddr, s.defaultDenom).Amount
			s.Require().True(fromBal.Int64() == balances[tc.forceTransferMsg.TransferFromAddress])

			toAddr, err := sdk.AccAddressFromBech32(tc.forceTransferMsg.TransferToAddress)
			s.Require().NoError(err)
			toBal := s.App.BankKeeper.GetBalance(s.Ctx, toAddr, s.defaultDenom).Amount
			s.Require().True(toBal.Int64() == balances[tc.forceTransferMsg.TransferToAddress])
		})
	}
}

func (s *KeeperTestSuite) TestChangeAdminDenom() {
	for _, tc := range []struct {
		desc                    string
		msgChangeAdmin          func(denom string) *types.MsgChangeAdmin
		expectedChangeAdminPass bool
		expectedAdminIndex      int
		msgMint                 func(denom string) *types.MsgMint
		expectedMintPass        bool
	}{
		{
			desc: "creator admin can't mint after setting to '' ",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(s.TestAccs[0].String(), denom, "")
			},
			expectedChangeAdminPass: true,
			expectedAdminIndex:      -1,
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(s.TestAccs[0].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintPass: false,
		},
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
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(s.TestAccs[1].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintPass: true,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// setup test
			s.SetupTest()

			// Create a denom and mint
			res, err := s.msgServer.CreateDenom(s.Ctx, types.NewMsgCreateDenom(s.TestAccs[0].String(), "bitcoin"))
			s.Require().NoError(err)

			testDenom := res.GetNewTokenDenom()

			_, err = s.msgServer.Mint(s.Ctx, types.NewMsgMint(s.TestAccs[0].String(), sdk.NewInt64Coin(testDenom, 10)))
			s.Require().NoError(err)

			_, err = s.msgServer.ChangeAdmin(s.Ctx, tc.msgChangeAdmin(testDenom))
			if tc.expectedChangeAdminPass {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}

			queryRes, err := s.queryClient.DenomAuthorityMetadata(s.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
				Denom: testDenom,
			})
			s.Require().NoError(err)

			// expectedAdminIndex with negative value is assumed as admin with value of ""
			const emptyStringAdminIndexFlag = -1
			if tc.expectedAdminIndex == emptyStringAdminIndexFlag {
				s.Require().Equal("", queryRes.AuthorityMetadata.Admin)
			} else {
				s.Require().Equal(s.TestAccs[tc.expectedAdminIndex].String(), queryRes.AuthorityMetadata.Admin)
			}

			// we test mint to test if admin authority is performed properly after admin change.
			if tc.msgMint != nil {
				_, err := s.msgServer.Mint(s.Ctx, tc.msgMint(testDenom))
				if tc.expectedMintPass {
					s.Require().NoError(err)
				} else {
					s.Require().Error(err)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetDenomMetaData() {
	// setup test
	s.SetupTest()
	s.CreateDefaultDenom()

	for _, tc := range []struct {
		desc                string
		msgSetDenomMetadata types.MsgSetDenomMetadata
		expectedPass        bool
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
			expectedPass: true,
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
		{
			desc: "non-factory denom",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(s.TestAccs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    appparams.BaseCoinUnit,
						Exponent: 0,
					},
					{
						Denom:    "uosmoo",
						Exponent: 6,
					},
				},
				Base:    appparams.BaseCoinUnit,
				Display: "uosmoo",
				Name:    "OSMO",
				Symbol:  "OSMO",
			}),
			expectedPass: false,
		},
		{
			desc: "wrong admin",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(s.TestAccs[1].String(), banktypes.Metadata{
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
			expectedPass: false,
		},
		{
			desc: "invalid metadata (missing display denom unit)",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(s.TestAccs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    s.defaultDenom,
						Exponent: 0,
					},
				},
				Base:    s.defaultDenom,
				Display: appparams.BaseCoinUnit,
				Name:    "OSMO",
				Symbol:  "OSMO",
			}),
			expectedPass: false,
		},
	} {
		s.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			bankKeeper := s.App.BankKeeper
			res, err := s.msgServer.SetDenomMetadata(s.Ctx, &tc.msgSetDenomMetadata)
			if tc.expectedPass {
				s.Require().NoError(err)
				s.Require().NotNil(res)

				md, found := bankKeeper.GetDenomMetaData(s.Ctx, s.defaultDenom)
				s.Require().True(found)
				s.Require().Equal(tc.msgSetDenomMetadata.Metadata.Name, md.Name)
			} else {
				s.Require().Error(err)
			}
		})
	}
}
