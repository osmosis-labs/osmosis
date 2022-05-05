package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestAdminMsgs() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	addr1bal := int64(0)
	addr2bal := int64(0)

	msgServer := keeper.NewMsgServerImpl(*suite.App.TokenFactoryKeeper)

	// Create a denom
	res, err := msgServer.CreateDenom(sdk.WrapSDKContext(suite.Ctx), types.NewMsgCreateDenom(addr1.String(), "bitcoin"))
	suite.Require().NoError(err)
	denom := res.GetNewTokenDenom()

	// Make sure that the admin is set correctly
	queryRes, err := suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: res.GetNewTokenDenom(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(addr1.String(), queryRes.AuthorityMetadata.Admin)

	// Test minting to admins own account
	_, err = msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(addr1.String(), sdk.NewInt64Coin(denom, 10)))
	addr1bal += 10
	suite.Require().NoError(err)
	suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, denom).Amount.Int64() == addr1bal, suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, denom))

	// // Test force transferring
	// _, err = msgServer.ForceTransfer(sdk.WrapSDKContext(suite.Ctx), types.NewMsgForceTransfer(addr1.String(), sdk.NewInt64Coin(denom, 5), addr2.String(), addr1.String()))
	// suite.Require().NoError(err)
	// suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, addr1, denom).IsEqual(sdk.NewInt64Coin(denom, 15)))
	// suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, addr2, denom).IsEqual(sdk.NewInt64Coin(denom, 5)))

	// Test burning from own account
	_, err = msgServer.Burn(sdk.WrapSDKContext(suite.Ctx), types.NewMsgBurn(addr1.String(), sdk.NewInt64Coin(denom, 5)))
	addr1bal -= 5
	suite.Require().NoError(err)
	suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, addr2, denom).Amount.Int64() == addr2bal)

	// Test Change Admin
	_, err = msgServer.ChangeAdmin(sdk.WrapSDKContext(suite.Ctx), types.NewMsgChangeAdmin(addr1.String(), denom, addr2.String()))
	queryRes, err = suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: res.GetNewTokenDenom(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(addr2.String(), queryRes.AuthorityMetadata.Admin)

	// Make sure old admin can no longer do actions
	_, err = msgServer.Burn(sdk.WrapSDKContext(suite.Ctx), types.NewMsgBurn(addr1.String(), sdk.NewInt64Coin(denom, 5)))
	suite.Require().Error(err)

	// Make sure the new admin works
	_, err = msgServer.Mint(sdk.WrapSDKContext(suite.Ctx), types.NewMsgMint(addr2.String(), sdk.NewInt64Coin(denom, 5)))
	addr2bal += 5
	suite.Require().NoError(err)
	suite.Require().True(suite.App.BankKeeper.GetBalance(suite.Ctx, addr2, denom).Amount.Int64() == addr2bal)

	// Try setting admin to empty
	_, err = msgServer.ChangeAdmin(sdk.WrapSDKContext(suite.Ctx), types.NewMsgChangeAdmin(addr2.String(), denom, ""))
	suite.Require().NoError(err)
	queryRes, err = suite.queryClient.DenomAuthorityMetadata(suite.Ctx.Context(), &types.QueryDenomAuthorityMetadataRequest{
		Denom: res.GetNewTokenDenom(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal("", queryRes.AuthorityMetadata.Admin)
}
