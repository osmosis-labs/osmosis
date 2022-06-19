package keeper_test

import (
	"time"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	balancertypes "github.com/osmosis-labs/osmosis/v7/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"

	appParams "github.com/osmosis-labs/osmosis/v7/app/params"
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

var (
	defaultExitFee            = sdk.MustNewDecFromStr("0.025")
	defaultSwapFee            = sdk.MustNewDecFromStr("0.01")
	defaultPoolId             = uint64(10)
	defaultFutureGovernor     = ""
	defaultCurBlockTime       = time.Unix(1618700000, 0)
	defaultPoolAssetAmount    = sdk.NewInt(10000000000)
	defaultPoolWeight         = sdk.NewInt(100)
	defaultBalancerPoolParams = balancertypes.PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	dummyPoolAssets = []balancertypes.PoolAsset{}
)

// In order to test the GetTotalSwapFee method for exact swaps in, we need to create both a
// MsgSwapExactAmountIn and a pool for each denom pair in the list of Routes.  (still need to consider how to handle 3+ asset pools)
// When the pool(s) are created, we cannot change their pool Id nor create a pool with a given pool ID
// Therefore, we must reassign the values in the msg routes to reflect the pool id's that the method
// we are testing checks the correct, newly created pools. Once the pools have been created, the check
// should reflect that the number of hops on the route * defaultSwapFee, is the returned total swap fee
func (suite *KeeperTestSuite) TestGetTotalSwapFeeForSwapInBalancerPool() {
	// check tx must be true however SetupTest does not actually use its parameter
	suite.SetupTest(true)
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()

	//suite.SetupTokenFactory()

	// define a create msg function
	createMsg := func(after func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
		properMsg := gammtypes.MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []gammtypes.SwapAmountInRoute{{
				PoolId:        1,
				TokenOutDenom: "test1",
			}},
			TokenIn:           sdk.NewCoin("test0", defaultPoolAssetAmount),
			TokenOutMinAmount: defaultPoolAssetAmount,
		}

		return after(properMsg)
	}

	// create a msg
	msg := createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
		// do nothing
		return msg
	})

	// verify msg was created correctly
	suite.Require().Equal(msg.Route(), gammtypes.RouterKey)
	suite.Require().Equal(msg.Type(), "swap_exact_amount_in")
	signers := msg.GetSigners()
	suite.Require().Equal(len(signers), 1)
	suite.Require().Equal(signers[0].String(), addr1)

	// // create balancer pool msg
	// createPoolMsg := func(after func(msg balancertypes.MsgCreateBalancerPool) balancertypes.MsgCreateBalancerPool) balancertypes.MsgCreateBalancerPool {
	// 	properMsg := balancertypes.MsgCreateBalancerPool{
	// 		Sender:             addr1,
	// 		PoolParams:         &defaultBalancerPoolParams,
	// 		PoolAssets:         dummyPoolAssets,
	// 		FuturePoolGovernor: defaultFutureGovernor,
	// 	}

	// 	return after(properMsg)
	// }

	// // create a create balancer pool msg
	// msgPool := createPoolMsg(func(msg balancertypes.MsgCreateBalancerPool) balancertypes.MsgCreateBalancerPool {
	// 	// do nothing
	// 	return msg
	// })

	// // verify msgPool was created correctly
	// suite.Require().Equal(msgPool.Type(), "create_balancer_pool")
	// msgSigners := msgPool.GetSigners()
	// suite.Require().Equal(len(msgSigners), 1)
	// suite.Require().Equal(msgSigners[0].String(), addr1)

	// define test cases
	tests := []struct {
		name               string
		msg                gammtypes.MsgSwapExactAmountIn
		expectPass         bool
		expectTotalSwapFee sdk.Dec
	}{
		{
			name: "Single denom pair",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				// do nothing
				return msg
			}),
			expectPass: true,
			// expect zero because test denoms are not fee tokens
			expectTotalSwapFee: sdk.ZeroDec(),
		},
		{
			name: "Two denom pairs",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				// create 3 denoms on the route
				msg.Routes = []gammtypes.SwapAmountInRoute{{
					PoolId:        1,
					TokenOutDenom: "test1",
				}, {
					PoolId:        2,
					TokenOutDenom: "test2",
				}, {
					PoolId:        3,
					TokenOutDenom: "test3",
				},
				}

				return msg
			}),
			expectPass: true,
			// expect zero because test denoms are not fee tokens
			expectTotalSwapFee: sdk.ZeroDec(),
		},
	}

	for _, test := range tests {
		// check test cases
		if test.expectPass {
			// get pool ids from test msg
			poolIds := test.msg.GetPoolIdOnPath()
			suite.Require().NotEmpty(poolIds)

			// get denomPath from test msg
			denomPath := test.msg.GetTokenDenomsOnPath()
			suite.Require().NotEmpty(denomPath)

			// check denom path is longer than pool ids by 1
			// *** if join/exit with swap msgs are added, denomPath length must be 2
			suite.Require().Equal(len(poolIds)+1, len(denomPath), "test: %v", test.name)

			// create pools for denom pairs and update pool ids to reflect returned pool ids from creation
			for i := range poolIds {
				// create pool assets
				poolAssets := make([]balancertypes.PoolAsset, 2)

				poolAssets[0] = balancertypes.PoolAsset{
					Token:  sdk.NewCoin(denomPath[i], defaultPoolAssetAmount),
					Weight: defaultPoolWeight,
				}

				poolAssets[1] = balancertypes.PoolAsset{
					Token:  sdk.NewCoin(denomPath[i+1], defaultPoolAssetAmount),
					Weight: defaultPoolWeight,
				}

				// Add coins for pool creation
				fundCoins := sdk.Coins{sdk.NewCoin(denomPath[i], defaultPoolAssetAmount)}
				fundCoins = fundCoins.Add(poolAssets[1].Token)
				suite.FundAcc(suite.TestAccs[0], fundCoins)

				// get a msg to create a new balancer pool
				msg := balancertypes.NewMsgCreateBalancerPool(suite.TestAccs[0], balancertypes.PoolParams{
					SwapFee: defaultSwapFee,
					ExitFee: sdk.ZeroDec(),
				}, poolAssets, "")

				// use msg to create balancer pool for pool assets with a swap fee
				poolId, err := suite.App.GAMMKeeper.CreatePool(suite.Ctx, msg)
				suite.NoError(err)

				test.msg.Routes[i].PoolId = poolId

				// // Add extra coin at start, and only add poolAsset[1] after
				// if i == 0 {
				// 	baseDenom, err := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
				// 	suite.Require().NoError(err, "test: %s", test.name)
				// 	if baseDenom != poolAssets[0].Token.Denom {
				// 		err := suite.ExecuteUpgradeFeeTokenProposal(poolAssets[0].Token.Denom, poolId)
				// 		suite.Require().NoError(err, "test: %s", test.name)
				// 	}
				// }

				// baseDenom, err := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
				// suite.Require().NoError(err, "test: %s", test.name)
				// if baseDenom != poolAssets[1].Token.Denom {
				// 	err := suite.ExecuteUpgradeFeeTokenProposal(poolAssets[1].Token.Denom, poolId)
				// 	suite.Require().NoError(err, "test: %s", test.name)
				// }

				// make the denoms fee tokens
				// feeTokens := make([]types.FeeToken, 0, len(poolAssets))
				// for j, poolAssets := range poolAssets {
				// 	feeToken := types.FeeToken{
				// 		Denom:  poolAssets[j].Token.Denom,
				// 		PoolID: poolId,
				// 	}

				// 	feeTokens = append(feeTokens, feeToken)
				// }
				// feeTokens := make([]types.FeeToken, 0, len(poolAssets))

				// if i == 0 {
				// 	suite.App.TxFeesKeeper.SetFeeTokens(suite.Ctx, []types.FeeToken{
				// 		{poolAssets[0].Token.Denom, poolId},
				// 		{poolAssets[1].Token.Denom, poolId},
				// 	})
				// } else {
				// 	suite.App.TxFeesKeeper.SetFeeTokens(suite.Ctx, )
				// }
			}
			// now that all the pools on the route have been created, we can test getTotalSwapFee
			swapFees, err := suite.App.TxFeesKeeper.GetTotalSwapFee(suite.Ctx, poolIds, denomPath)
			suite.Require().NoError(err, "test: %v", test.name)
			suite.Require().Equal(test.expectTotalSwapFee, swapFees, "test: %v", test.name)
			suite.Require().NoError(msg.ValidateBasic(), "test: %v", test)
		} else {
			suite.Require().Error(msg.ValidateBasic(), "test: %v", test)
		}

	}
}
