package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gammtypes "github.com/osmosis-labs/osmosis/x/gamm/types"
	"github.com/osmosis-labs/osmosis/x/txfees/types"
)

func (suite *KeeperTestSuite) TestFeeTokens() { // test for all unlockable coins
	suite.SetupTest()

	uionPoolId := suite.preparePool(
		[]gammtypes.PoolAsset{
			{
				Weight: sdk.NewInt(1),
				Token:  sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
			},
			{
				Weight: sdk.NewInt(1),
				Token:  sdk.NewInt64Coin("uion", 500),
			},
		},
	)

	// Test getting basedenom (should be default from genesis)
	baseDenom, err := suite.app.TxFeesKeeper.GetBaseDenom(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.DefaultBondDenom, baseDenom)

	converted, err := suite.app.TxFeesKeeper.ConvertToBaseToken(suite.ctx, sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))
	suite.Require().True(converted.IsEqual(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)))
	suite.Require().NoError(err)

	// Make sure there's no external whitelisted fee tokens at launch
	feeTokens := suite.app.TxFeesKeeper.GetFeeTokens(suite.ctx)
	suite.Require().Len(feeTokens, 0)

	// Add a new whitelisted fee token via a governance proposal
	upgradeProp := types.NewUpdateFeeTokenProposal(
		"Test Proposal",
		"test",
		types.FeeToken{
			Denom:  "uion",
			PoolID: uionPoolId,
		},
	)
	err = suite.app.TxFeesKeeper.HandleUpdateFeeTokenProposal(suite.ctx, &upgradeProp)
	suite.Require().NoError(err)

	// Check to make sure length of whitelisted fee tokens increased
	feeTokens = suite.app.TxFeesKeeper.GetFeeTokens(suite.ctx)
	suite.Require().Len(feeTokens, 1)

	// Make sure new fee token was set correct and is convertable
	suite.Require().Equal("uion", feeTokens[0].Denom)
	suite.Require().NoError(suite.app.TxFeesKeeper.ValidateFeeToken(suite.ctx, feeTokens[0]))
	converted, err = suite.app.TxFeesKeeper.ConvertToBaseToken(suite.ctx, sdk.NewInt64Coin("uion", 10))
	suite.Require().NoError(err)
	suite.Require().True(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10).IsEqual(converted))
	queriedPoolId, err := suite.queryClient.DenomPoolId(suite.ctx.Context(),
		&types.QueryDenomPoolIdRequest{
			Denom: "uion",
		},
	)
	suite.Require().Equal(uionPoolId, queriedPoolId.GetPoolID())

	// Upgrade proposal for non-existent pool
	upgradeProp = types.NewUpdateFeeTokenProposal(
		"Test Proposal 2",
		"test",
		types.FeeToken{
			Denom:  "foo",
			PoolID: 5,
		},
	)
	err = suite.app.TxFeesKeeper.HandleUpdateFeeTokenProposal(suite.ctx, &upgradeProp)
	suite.Require().Error(err)

	// Upgrade proposal with wrong pool
	upgradeProp = types.NewUpdateFeeTokenProposal(
		"Test Proposal 3",
		"test",
		types.FeeToken{
			Denom:  "foo",
			PoolID: uionPoolId,
		},
	)
	err = suite.app.TxFeesKeeper.HandleUpdateFeeTokenProposal(suite.ctx, &upgradeProp)
	suite.Require().Error(err)

	// Make pool with fee token but no OSMO and make sure governance proposal fails
	badPoolId := suite.preparePool(
		[]gammtypes.PoolAsset{
			{
				Weight: sdk.NewInt(1),
				Token:  sdk.NewInt64Coin("uion", 500),
			},
			{
				Weight: sdk.NewInt(1),
				Token:  sdk.NewInt64Coin("foo", 500),
			},
		},
	)
	upgradeProp = types.NewUpdateFeeTokenProposal(
		"Test Proposal 4",
		"test",
		types.FeeToken{
			Denom:  "foo",
			PoolID: badPoolId,
		},
	)
	err = suite.app.TxFeesKeeper.HandleUpdateFeeTokenProposal(suite.ctx, &upgradeProp)
	suite.Require().Error(err)

	// Create correct pool and governance proposal
	fooPoolId := suite.preparePool(
		[]gammtypes.PoolAsset{
			{
				Weight: sdk.NewInt(1),
				Token:  sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
			},
			{
				Weight: sdk.NewInt(1),
				Token:  sdk.NewInt64Coin("foo", 500),
			},
		},
	)
	upgradeProp = types.NewUpdateFeeTokenProposal(
		"Test Proposal 5",
		"test",
		types.FeeToken{
			Denom:  "foo",
			PoolID: fooPoolId,
		},
	)
	err = suite.app.TxFeesKeeper.HandleUpdateFeeTokenProposal(suite.ctx, &upgradeProp)
	suite.Require().NoError(err)

	// Make sure there's two whitelisted fee tokens
	responseFeeTokens, err := suite.queryClient.FeeTokens(suite.ctx.Context(), &types.QueryFeeTokensRequest{})
	suite.Require().Len(responseFeeTokens.FeeTokens, 2)
	suite.Require().NoError(err)
}
