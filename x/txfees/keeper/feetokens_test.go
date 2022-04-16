package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v8/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestBaseDenom() {
	suite.SetupTest(false)

	// Test getting basedenom (should be default from genesis)
	baseDenom, err := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.DefaultBondDenom, baseDenom)

	converted, err := suite.App.TxFeesKeeper.ConvertToBaseToken(suite.Ctx, sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))
	suite.Require().True(converted.IsEqual(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)))
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestUpgradeFeeTokenProposals() {
	suite.SetupTest(false)

	uionPoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)

	uionPoolId2 := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)

	// Make pool with fee token but no OSMO and make sure governance proposal fails
	noBasePoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin("uion", 500),
		sdk.NewInt64Coin("foo", 500),
	)

	// Create correct pool and governance proposal
	fooPoolId := suite.PrepareUni2PoolWithAssets(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("foo", 1000),
	)

	tests := []struct {
		name       string
		feeToken   string
		poolId     uint64
		expectPass bool
	}{
		{
			name:       "uion pool",
			feeToken:   "uion",
			poolId:     uionPoolId,
			expectPass: true,
		},
		{
			name:       "try with basedenom",
			feeToken:   sdk.DefaultBondDenom,
			poolId:     uionPoolId,
			expectPass: false,
		},
		{
			name:       "proposal with non-existent pool",
			feeToken:   "foo",
			poolId:     100000000000,
			expectPass: false,
		},
		{
			name:       "proposal with wrong pool for fee token",
			feeToken:   "foo",
			poolId:     uionPoolId,
			expectPass: false,
		},
		{
			name:       "proposal with pool with no base denom",
			feeToken:   "foo",
			poolId:     noBasePoolId,
			expectPass: false,
		},
		{
			name:       "proposal to add foo correctly",
			feeToken:   "foo",
			poolId:     fooPoolId,
			expectPass: true,
		},
		{
			name:       "proposal to replace pool for fee token",
			feeToken:   "uion",
			poolId:     uionPoolId2,
			expectPass: true,
		},
		{
			name:       "proposal to replace uion as fee denom",
			feeToken:   "uion",
			poolId:     0,
			expectPass: true,
		},
	}

	for _, tc := range tests {

		feeTokensBefore := suite.App.TxFeesKeeper.GetFeeTokens(suite.Ctx)

		// Add a new whitelisted fee token via a governance proposal
		err := suite.ExecuteUpgradeFeeTokenProposal(tc.feeToken, tc.poolId)

		feeTokensAfter := suite.App.TxFeesKeeper.GetFeeTokens(suite.Ctx)

		if tc.expectPass {
			// Make sure no error during setting of proposal
			suite.Require().NoError(err, "test: %s", tc.name)

			// For a proposal that adds a feetoken
			if tc.poolId != 0 {
				// Make sure the length of fee tokens is >= before
				suite.Require().GreaterOrEqual(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
				// Ensure that the fee token is convertable to base token
				_, err := suite.App.TxFeesKeeper.ConvertToBaseToken(suite.Ctx, sdk.NewInt64Coin(tc.feeToken, 10))
				suite.Require().NoError(err, "test: %s", tc.name)
				// make sure the queried poolId is the same as expected
				queriedPoolId, err := suite.queryClient.DenomPoolId(suite.Ctx.Context(),
					&types.QueryDenomPoolIdRequest{
						Denom: tc.feeToken,
					},
				)
				suite.Require().NoError(err, "test: %s", tc.name)
				suite.Require().Equal(tc.poolId, queriedPoolId.GetPoolID(), "test: %s", tc.name)
			} else {
				// if this proposal deleted a fee token
				// ensure that the length of fee tokens is <= to before
				suite.Require().LessOrEqual(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
				// Ensure that the fee token is not convertable to base token
				_, err := suite.App.TxFeesKeeper.ConvertToBaseToken(suite.Ctx, sdk.NewInt64Coin(tc.feeToken, 10))
				suite.Require().Error(err, "test: %s", tc.name)
				// make sure the queried poolId errors
				_, err = suite.queryClient.DenomPoolId(suite.Ctx.Context(),
					&types.QueryDenomPoolIdRequest{
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

	baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)

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
			name:               "unequal value",
			baseDenomPoolInput: sdk.NewInt64Coin(baseDenom, 100),
			feeTokenPoolInput:  sdk.NewInt64Coin("foo", 200),
			inputFee:           sdk.NewInt64Coin("foo", 10),
			// expected to get 5.000000000005368710 baseDenom without rounding
			expectedOutput:      sdk.NewInt64Coin(baseDenom, 5),
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

		poolId := suite.PrepareUni2PoolWithAssets(
			tc.baseDenomPoolInput,
			tc.feeTokenPoolInput,
		)

		suite.ExecuteUpgradeFeeTokenProposal(tc.feeTokenPoolInput.Denom, poolId)

		converted, err := suite.App.TxFeesKeeper.ConvertToBaseToken(suite.Ctx, tc.inputFee)
		if tc.expectedConvertable {
			suite.Require().NoError(err, "test: %s", tc.name)
			suite.Require().True(converted.IsEqual(tc.expectedOutput), "test: %s", tc.name)
		} else {
			suite.Require().Error(err, "test: %s", tc.name)
		}
	}
}
