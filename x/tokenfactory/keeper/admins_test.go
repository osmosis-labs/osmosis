package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestAdminMsgs() {
	suite.SetupTest()

	addr0bal := int64(0)
	addr1bal := int64(0)

	msgServer := keeper.NewMsgServerImpl(*suite.App.TokenFactoryKeeper)

	// Create a denom
	res, err := msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.Require().NoError(err)
	denom := res.GetNewTokenDenom()

	// Make sure that the admin is set correctly
	queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: res.GetNewTokenDenom(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

	// Test minting to admins own account
	_, err = msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 10)))
	addr0bal += 10
	suite.Require().NoError(err)
	suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).Amount.Int64() == addr0bal, suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom))

	// // Test force transferring
	// _, err = msgServer.ForceTransfer(sdk.WrapSDKContext(suite.Ctx), types.NewMsgForceTransfer(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 5), suite.TestAccs[1].String(), suite.TestAccs[0].String()))
	// suite.Require().NoError(err)
	// suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).IsEqual(sdk.NewInt64Coin(denom, 15)))
	// suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[1], denom).IsEqual(sdk.NewInt64Coin(denom, 5)))

	// Test burning from own account
	_, err = msgServer.Burn(sdk.WrapSDKContext(suite.Ctx), types.NewMsgBurn(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 5)))
	addr0bal -= 5
	suite.Require().NoError(err)
	suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[1], denom).Amount.Int64() == addr1bal)

	// Test Change Admin
	_, err = msgServer.ChangeAdmin(sdk.WrapSDKContext(suite.Ctx), types.NewMsgChangeAdmin(suite.TestAccs[0].String(), denom, suite.TestAccs[1].String()))
	queryRes, err = suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: res.GetNewTokenDenom(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[1].String(), queryRes.AuthorityMetadata.Admin)

	// Make sure old admin can no longer do actions
	_, err = msgServer.Burn(sdk.WrapSDKContext(suite.Ctx), types.NewMsgBurn(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 5)))
	suite.Require().Error(err)

	// Make sure the new admin works
	_, err = msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(suite.TestAccs[1].String(), sdk.NewInt64Coin(denom, 5)))
	addr1bal += 5
	suite.Require().NoError(err)
	suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[1], denom).Amount.Int64() == addr1bal)

	// Try setting admin to empty
	_, err = msgServer.ChangeAdmin(sdk.WrapSDKContext(suite.Ctx), types.NewMsgChangeAdmin(suite.TestAccs[1].String(), denom, ""))
	suite.Require().NoError(err)
	queryRes, err = suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: res.GetNewTokenDenom(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal("", queryRes.AuthorityMetadata.Admin)
}

func (suite *KeeperTestSuite) TestMintDenom() {
	suite.SetupTest()
	var (
		denom     string
		msgServer types.MsgServer
		addr0bal  = int64(0)
	)

	// Create a denom
	msgServer = keeper.NewMsgServerImpl(*suite.App.TokenFactoryKeeper)
	res, err := msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.Require().NoError(err)
	denom = res.GetNewTokenDenom()

	for _, tc := range []struct {
		desc      string
		amount    int64
		mintDenom string
		admin     string
		valid     bool
	}{
		{
			desc:      "denom does not exist",
			amount:    10,
			mintDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:     suite.TestAccs[0].String(),
			valid:     false,
		},
		{
			desc:      "mint is not by the admin",
			amount:    10,
			mintDenom: denom,
			admin:     suite.TestAccs[1].String(),
			valid:     false,
		},
		{
			desc:      "success case",
			amount:    10,
			mintDenom: denom,
			admin:     suite.TestAccs[0].String(),
			valid:     true,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// Test minting to admins own account
			_, err := msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(tc.admin, sdk.NewInt64Coin(tc.mintDenom, 10)))
			if tc.valid {
				addr0bal += 10
				suite.Require().NoError(err)
				suite.Require().Equal(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).Amount.Int64(), addr0bal, suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBurnDenom() {
	suite.SetupTest()
	var (
		denom     string
		msgServer types.MsgServer
		addr0bal  = int64(0)
	)

	// Create a denom and mint
	// Create a denom
	msgServer = keeper.NewMsgServerImpl(*suite.App.TokenFactoryKeeper)
	res, err := msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
	suite.Require().NoError(err)
	denom = res.GetNewTokenDenom()

	// Make sure that the admin is set correctly
	queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: res.GetNewTokenDenom(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(suite.TestAccs[0].String(), queryRes.AuthorityMetadata.Admin)

	_, err = msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 10)))
	addr0bal += 10
	suite.Require().NoError(err)
	suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).Amount.Int64() == addr0bal, suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom))

	for _, tc := range []struct {
		desc      string
		amount    int64
		burnDenom string
		admin     string
		valid     bool
	}{
		{
			desc:      "denom does not exist",
			amount:    10,
			burnDenom: "factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos",
			admin:     suite.TestAccs[0].String(),
			valid:     false,
		},
		{
			desc:      "burn is not by the admin",
			amount:    10,
			burnDenom: denom,
			admin:     suite.TestAccs[1].String(),
			valid:     false,
		},
		{
			desc:      "burn amount is bigger than minted amount",
			amount:    1000,
			burnDenom: denom,
			admin:     suite.TestAccs[1].String(),
			valid:     false,
		},
		{
			desc:      "success case",
			amount:    10,
			burnDenom: denom,
			admin:     suite.TestAccs[0].String(),
			valid:     true,
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// Test minting to admins own account
			_, err := msgServer.Burn(sdk.WrapSDKContext(suite.Ctx), types.NewMsgBurn(tc.admin, sdk.NewInt64Coin(tc.burnDenom, 10)))
			if tc.valid {
				addr0bal -= 10
				suite.Require().NoError(err)
				suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).Amount.Int64() == addr0bal, suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom))
			} else {
				suite.Require().Error(err)
				suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).Amount.Int64() == addr0bal, suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestChangeAdminDenom() {
	var (
		denom     string
		msgServer types.MsgServer
		addr0bal  int64
		addr1bal  int64
	)

	for _, tc := range []struct {
		desc     string
		malleate func()
		valid    bool
	}{
		{
			desc: "creator admin can't mint after setting to '' ",
			malleate: func() {
				_, err := msgServer.ChangeAdmin(sdk.WrapSDKContext(suite.Ctx), types.NewMsgChangeAdmin(suite.TestAccs[0].String(), denom, ""))
				queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
					Denom: denom,
				})
				suite.Require().NoError(err)
				suite.Require().Equal("", queryRes.AuthorityMetadata.Admin)

				// can't mint.
				_, err = msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 5)))
				suite.Require().Error(err)
				suite.Require().Equal(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).Amount.Int64(), addr0bal)
			},
		},
		{
			desc: "non-admins can't change the existing admin",
			malleate: func() {
				// Test Change Admin
				_, err := msgServer.ChangeAdmin(sdk.WrapSDKContext(suite.Ctx), types.NewMsgChangeAdmin(suite.TestAccs[1].String(), denom, suite.TestAccs[2].String()))
				queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
					Denom: denom,
				})
				suite.Require().NoError(err)
				suite.Require().NotEqual(suite.TestAccs[2].String(), queryRes.AuthorityMetadata.Admin)
			},
		},
		{
			desc: "success change admin",
			malleate: func() {
				// Test Change Admin
				_, err := msgServer.ChangeAdmin(sdk.WrapSDKContext(suite.Ctx), types.NewMsgChangeAdmin(suite.TestAccs[0].String(), denom, suite.TestAccs[1].String()))
				queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
					Denom: denom,
				})
				suite.Require().NoError(err)
				suite.Require().Equal(suite.TestAccs[1].String(), queryRes.AuthorityMetadata.Admin)

				// New admin can mint.
				_, err = msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(suite.TestAccs[1].String(), sdk.NewInt64Coin(denom, 5)))

				addr1bal += 5
				suite.Require().NoError(err)
				suite.Require().Equal(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).Amount.Int64(), addr0bal)
			},
		},
	} {
		suite.Run(fmt.Sprintf("Case %s", tc.desc), func() {
			// setup test
			suite.SetupTest()
			addr0bal = 0
			addr1bal = 0

			// Create a denom and mint
			msgServer = keeper.NewMsgServerImpl(*suite.App.TokenFactoryKeeper)
			res, err := msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(suite.TestAccs[0].String(), "bitcoin"))
			suite.Require().NoError(err)
			denom = res.GetNewTokenDenom()

			_, err = msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(suite.TestAccs[0].String(), sdk.NewInt64Coin(denom, 10)))
			addr0bal += 10
			suite.Require().NoError(err)
			suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom).Amount.Int64() == addr0bal, suite.App.BankKeeper.GetBalance(suite.Ctx, suite.TestAccs[0], denom))

			tc.malleate()
		})
	}
}
