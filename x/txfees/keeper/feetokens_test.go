package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (suite *KeeperTestSuite) TestBaseDenom() {
	suite.SetupTest(false)

	// Test getting basedenom (should be default from genesis)
	baseDenom, err := suite.app.TxFeesKeeper.GetBaseDenom(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.DefaultBondDenom, baseDenom)

	converted, err := suite.app.TxFeesKeeper.ConvertToBaseToken(suite.ctx, sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))
	suite.Require().True(converted.IsEqual(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)))
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestUpgradeFeeTokenProposals() {
	suite.SetupTest(false)

	uionPoolId := suite.PreparePoolWithAssets(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)

	uionPoolId2 := suite.PreparePoolWithAssets(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)

	// Make pool with fee token but no OSMO and make sure governance proposal fails
	noBasePoolId := suite.PreparePoolWithAssets(
		sdk.NewInt64Coin("uion", 500),
		sdk.NewInt64Coin("foo", 500),
	)

	// Create correct pool and governance proposal
	fooPoolId := suite.PreparePoolWithAssets(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("foo", 1000),
	)

	tests := []struct {
		name       string
		feeToken   string
		poolID     uint64
		expectPass bool
	}{
		{
			name:       "uion pool",
			feeToken:   "uion",
			poolID:     uionPoolId,
			expectPass: true,
		},
		{
			name:       "try with basedenom",
			feeToken:   sdk.DefaultBondDenom,
			poolID:     uionPoolId,
			expectPass: false,
		},
		{
			name:       "proposal with non-existent pool",
			feeToken:   "foo",
			poolID:     100000000000,
			expectPass: false,
		},
		{
			name:       "proposal with wrong pool for fee token",
			feeToken:   "foo",
			poolID:     uionPoolId,
			expectPass: false,
		},
		{
			name:       "proposal with pool with no base denom",
			feeToken:   "foo",
			poolID:     noBasePoolId,
			expectPass: false,
		},
		{
			name:       "proposal to add foo correctly",
			feeToken:   "foo",
			poolID:     fooPoolId,
			expectPass: true,
		},
		{
			name:       "proposal to replace pool for fee token",
			feeToken:   "uion",
			poolID:     uionPoolId2,
			expectPass: true,
		},
		{
			name:       "proposal to replace uion as fee denom",
			feeToken:   "uion",
			poolID:     0,
			expectPass: true,
		},
	}

	for _, tc := range tests {

		feeTokensBefore := suite.app.TxFeesKeeper.GetFeeTokens(suite.ctx)

		// Add a new whitelisted fee token via a governance proposal
		err := suite.ExecuteUpgradeFeeTokenProposal(tc.feeToken, tc.poolID)

		feeTokensAfter := suite.app.TxFeesKeeper.GetFeeTokens(suite.ctx)

		if tc.expectPass {
			// Make sure no error during setting of proposal
			suite.Require().NoError(err, "test: %s", tc.name)

			// For a proposal that adds a feetoken
			if tc.poolID != 0 {
				// Make sure the length of fee tokens is >= before
				suite.Require().GreaterOrEqual(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
				// Ensure that the fee token is convertable to base token
				_, err := suite.app.TxFeesKeeper.ConvertToBaseToken(suite.ctx, sdk.NewInt64Coin(tc.feeToken, 10))
				suite.Require().NoError(err, "test: %s", tc.name)
				// make sure the queried poolID is the same as expected
				queriedPoolId, err := suite.queryClient.DenomPoolID(suite.ctx.Context(),
					&types.QueryDenomPoolIDRequest{
						Denom: tc.feeToken,
					},
				)
				suite.Require().NoError(err, "test: %s", tc.name)
				suite.Require().Equal(tc.poolID, queriedPoolId.GetPoolID(), "test: %s", tc.name)
			} else {
				// if this proposal deleted a fee token
				// ensure that the length of fee tokens is <= to before
				suite.Require().LessOrEqual(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
				// Ensure that the fee token is not convertable to base token
				_, err := suite.app.TxFeesKeeper.ConvertToBaseToken(suite.ctx, sdk.NewInt64Coin(tc.feeToken, 10))
				suite.Require().Error(err, "test: %s", tc.name)
				// make sure the queried poolID errors
				_, err = suite.queryClient.DenomPoolID(suite.ctx.Context(),
					&types.QueryDenomPoolIDRequest{
						Denom: tc.feeToken,
					},
				)
				suite.Require().Error(err, "test: %s", tc.name)
			}
		} else {
			// Make sure errors during setting of proposal
			suite.Require().Error(err, "test: %s", tc.name)
			// fee tokens should be the same
			suite.Require().Equal(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestFeeTokenConversions() {
	suite.SetupTest(false)

	baseDenom, _ := suite.app.TxFeesKeeper.GetBaseDenom(suite.ctx)

	tests := []struct {
		name                string
		baseDenomPoolInput  sdk.Coin
		feeTokenPoolInput   sdk.Coin
		inputFee            sdk.Coin
		expectedConvertable bool
		expectedOutput      sdk.Coin
	}{
		{
			name:                "equal value",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			feeTokenPoolInput:   sdk.NewInt64Coin("uion", 100),
			inputFee:            sdk.NewInt64Coin("uion", 10),
			expectedOutput:      sdk.NewInt64Coin(baseDenom, 10),
			expectedConvertable: true,
		},
		{
			name:                "unequal value",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			feeTokenPoolInput:   sdk.NewInt64Coin("foo", 200),
			inputFee:            sdk.NewInt64Coin("foo", 10),
			expectedOutput:      sdk.NewInt64Coin(baseDenom, 20),
			expectedConvertable: true,
		},
		{
			name:                "basedenom value",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			feeTokenPoolInput:   sdk.NewInt64Coin("foo", 200),
			inputFee:            sdk.NewInt64Coin(baseDenom, 10),
			expectedOutput:      sdk.NewInt64Coin(baseDenom, 10),
			expectedConvertable: true,
		},
		{
			name:                "convert non-existent",
			baseDenomPoolInput:  sdk.NewInt64Coin(baseDenom, 100),
			feeTokenPoolInput:   sdk.NewInt64Coin("uion", 200),
			inputFee:            sdk.NewInt64Coin("foo", 10),
			expectedOutput:      sdk.Coin{},
			expectedConvertable: false,
		},
	}

	for _, tc := range tests {
		suite.SetupTest(false)

		poolID := suite.PreparePoolWithAssets(
			tc.baseDenomPoolInput,
			tc.feeTokenPoolInput,
		)

		suite.ExecuteUpgradeFeeTokenProposal(tc.feeTokenPoolInput.Denom, poolID)

		converted, err := suite.app.TxFeesKeeper.ConvertToBaseToken(suite.ctx, tc.inputFee)
		if tc.expectedConvertable {
			suite.Require().NoError(err, "test: %s", tc.name)
			suite.Require().True(converted.IsEqual(tc.expectedOutput), "test: %s", tc.name)
		} else {
			suite.Require().Error(err, "test: %s", tc.name)
		}
	}

}
