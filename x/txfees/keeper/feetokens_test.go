package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v19/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestBaseDenom() {
	s.SetupTest(false)

	// Test getting basedenom (should be default from genesis)
	baseDenom, err := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(sdk.DefaultBondDenom, baseDenom)

	converted, err := s.App.TxFeesKeeper.ConvertToBaseToken(s.Ctx, sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))
	s.Require().True(converted.IsEqual(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)))
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestUpgradeFeeTokenProposals() {
	s.SetupTest(false)

	uionPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)

	uionPoolId2 := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)

	// Make pool with fee token but no OSMO and make sure governance proposal fails
	noBasePoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin("uion", 500),
		sdk.NewInt64Coin("foo", 500),
	)

	// Create correct pool and governance proposal
	fooPoolId := s.PrepareBalancerPoolWithCoins(
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
		s.Run(tc.name, func() {
			feeTokensBefore := s.App.TxFeesKeeper.GetFeeTokens(s.Ctx)

			// Add a new whitelisted fee token via a governance proposal
			err := s.ExecuteUpgradeFeeTokenProposal(tc.feeToken, tc.poolId)

			feeTokensAfter := s.App.TxFeesKeeper.GetFeeTokens(s.Ctx)

			if tc.expectPass {
				// Make sure no error during setting of proposal
				s.Require().NoError(err, "test: %s", tc.name)

				// For a proposal that adds a feetoken
				if tc.poolId != 0 {
					// Make sure the length of fee tokens is >= before
					s.Require().GreaterOrEqual(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
					// Ensure that the fee token is convertable to base token
					_, err := s.App.TxFeesKeeper.ConvertToBaseToken(s.Ctx, sdk.NewInt64Coin(tc.feeToken, 10))
					s.Require().NoError(err, "test: %s", tc.name)
					// make sure the queried poolId is the same as expected
					queriedPoolId, err := s.queryClient.DenomPoolId(s.Ctx.Context(),
						&types.QueryDenomPoolIdRequest{
							Denom: tc.feeToken,
						},
					)
					s.Require().NoError(err, "test: %s", tc.name)
					s.Require().Equal(tc.poolId, queriedPoolId.GetPoolID(), "test: %s", tc.name)
				} else {
					// if this proposal deleted a fee token
					// ensure that the length of fee tokens is <= to before
					s.Require().LessOrEqual(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
					// Ensure that the fee token is not convertable to base token
					_, err := s.App.TxFeesKeeper.ConvertToBaseToken(s.Ctx, sdk.NewInt64Coin(tc.feeToken, 10))
					s.Require().Error(err, "test: %s", tc.name)
					// make sure the queried poolId errors
					_, err = s.queryClient.DenomPoolId(s.Ctx.Context(),
						&types.QueryDenomPoolIdRequest{
							Denom: tc.feeToken,
						},
					)
					s.Require().Error(err, "test: %s", tc.name)
				}
			} else {
				// Make sure errors during setting of proposal
				s.Require().Error(err, "test: %s", tc.name)
				// fee tokens should be the same
				s.Require().Equal(len(feeTokensAfter), len(feeTokensBefore), "test: %s", tc.name)
			}
		})
	}
}

func (s *KeeperTestSuite) TestFeeTokenConversions() {
	s.SetupTest(false)

	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)

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
			// expected to get approximately 5 base denom
			// foo supply / stake supply =  200 / 100 = 2 foo for 1 stake
			// 10 foo in / 2 foo for 1 stake = 5 base denom
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
		s.SetupTest(false)

		s.Run(tc.name, func() {
			poolId := s.PrepareBalancerPoolWithCoins(
				tc.baseDenomPoolInput,
				tc.feeTokenPoolInput,
			)

			err := s.ExecuteUpgradeFeeTokenProposal(tc.feeTokenPoolInput.Denom, poolId)
			s.Require().NoError(err)

			converted, err := s.App.TxFeesKeeper.ConvertToBaseToken(s.Ctx, tc.inputFee)
			if tc.expectedConvertable {
				s.Require().NoError(err, "test: %s", tc.name)
				s.Require().Equal(tc.expectedOutput, converted)
			} else {
				s.Require().Error(err, "test: %s", tc.name)
			}
		})
	}
}
