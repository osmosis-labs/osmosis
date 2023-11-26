package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

/* ----------------------------- Testing ExactIn ---------------------------- */
func (suite *KeeperTestSuite) TestDYMIsBurned_ExactIn() {
	tokenInAmt := int64(100000)
	testcases := map[string]struct {
		routes                []poolmanagertypes.SwapAmountInRoute
		tokenIn               sdk.Coin
		tokenOutMinAmount     sdk.Int
		expectError           bool
		expectedBurnEvents    bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:            []poolmanagertypes.SwapAmountInRoute{},
			tokenIn:           sdk.NewCoin("foo", sdk.NewInt(tokenInAmt)),
			tokenOutMinAmount: sdk.NewInt(1),
			expectError:       true,
		},
		"udym as tokenIn": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "foo",
				},
			},
			tokenIn:           sdk.NewCoin("udym", sdk.NewInt(tokenInAmt)),
			tokenOutMinAmount: sdk.NewInt(1),
			expectError:       false,
		},
		"udym swapped in first pool": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "udym",
				},
			},
			tokenIn:           sdk.NewCoin("foo", sdk.NewInt(tokenInAmt)),
			tokenOutMinAmount: sdk.NewInt(1),
			expectError:       false,
		},
		"usdc swapped in first pool": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        2,
					TokenOutDenom: "bar",
				},
			},
			tokenIn:           sdk.NewCoin("foo", sdk.NewInt(tokenInAmt)),
			tokenOutMinAmount: sdk.NewInt(1),
			expectError:       false,
		},
		"usdc as token in": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        2,
					TokenOutDenom: "foo",
				},
			},
			tokenIn:           sdk.NewCoin("bar", sdk.NewInt(tokenInAmt)),
			tokenOutMinAmount: sdk.NewInt(1),
			expectError:       false,
		},
		"usdc as token in - no route to dym": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        4,
					TokenOutDenom: "baz",
				},
			},
			tokenIn:           sdk.NewCoin("bar", sdk.NewInt(tokenInAmt)),
			tokenOutMinAmount: sdk.NewInt(1),
			expectError:       false,
		},
		"usdc swap with dym": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        3,
					TokenOutDenom: "udym",
				},
			},
			tokenIn:           sdk.NewCoin("bar", sdk.NewInt(tokenInAmt)),
			tokenOutMinAmount: sdk.NewInt(1),
			expectError:       false,
		},
	}

	for name, tc := range testcases {
		suite.SetupTest()
		suite.FundAcc(suite.TestAccs[0], apptesting.DefaultAcctFunds)
		params := suite.App.GAMMKeeper.GetParams(suite.Ctx)
		params.PoolCreationFee = sdk.NewCoins(
			sdk.NewCoin("udym", sdk.NewInt(100000)),
			sdk.NewCoin("bar", sdk.NewInt(100000)))
		suite.App.GAMMKeeper.SetParams(suite.Ctx, params)

		ctx := suite.Ctx
		msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

		pool1coins := []sdk.Coin{sdk.NewCoin("udym", sdk.NewInt(100000)), sdk.NewCoin("foo", sdk.NewInt(100000))}
		suite.PrepareBalancerPoolWithCoins(pool1coins...)

		//"bar" is treated as baseDenom (e.g. USDC)
		pool2coins := []sdk.Coin{sdk.NewCoin("bar", sdk.NewInt(100000)), sdk.NewCoin("foo", sdk.NewInt(100000))}
		suite.PrepareBalancerPoolWithCoins(pool2coins...)

		pool3coins := []sdk.Coin{sdk.NewCoin("bar", sdk.NewInt(100000)), sdk.NewCoin("udym", sdk.NewInt(100000))}
		suite.PrepareBalancerPoolWithCoins(pool3coins...)

		pool4coins := []sdk.Coin{sdk.NewCoin("bar", sdk.NewInt(100000)), sdk.NewCoin("baz", sdk.NewInt(100000))}
		suite.PrepareBalancerPoolWithCoins(pool4coins...)

		//check total supply before swap
		suppliesBefore := make(map[string]sdk.Int)
		suppliesBefore["udym"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "udym").Amount
		suppliesBefore["foo"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "foo").Amount
		suppliesBefore["bar"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "bar").Amount

		// check taker fee is not 0
		suite.Require().True(suite.App.GAMMKeeper.GetParams(ctx).TakerFee.GT(sdk.ZeroDec()))

		// make swap
		_, err := msgServer.SwapExactAmountIn(sdk.WrapSDKContext(ctx), &types.MsgSwapExactAmountIn{
			Sender:            suite.TestAccs[0].String(),
			Routes:            tc.routes,
			TokenIn:           tc.tokenIn,
			TokenOutMinAmount: tc.tokenOutMinAmount,
		})
		if tc.expectError {
			suite.Require().Error(err, name)
			continue
		}
		suite.Require().NoError(err, name)

		// check total supply after swap
		suppliesAfter := make(map[string]sdk.Int)
		suppliesAfter["udym"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "udym").Amount
		suppliesAfter["foo"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "foo").Amount
		suppliesAfter["bar"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "bar").Amount

		//validate total supply is reduced by taker fee
		suite.Require().True(suppliesAfter["udym"].LT(suppliesBefore["udym"]), name)
		suite.Require().True(suppliesAfter["foo"].Equal(suppliesBefore["foo"]), name)
		suite.Require().True(suppliesAfter["bar"].Equal(suppliesBefore["bar"]), name)
	}
}

func (suite *KeeperTestSuite) TestDYMIsBurned_ExactOut() {
	tokenInAmt := int64(100000)
	testcases := map[string]struct {
		routes      []poolmanagertypes.SwapAmountOutRoute
		tokenOut    sdk.Coin
		expectError bool
	}{
		"zero hops": {
			routes:      []poolmanagertypes.SwapAmountOutRoute{},
			tokenOut:    sdk.NewCoin("foo", sdk.NewInt(tokenInAmt)),
			expectError: true,
		},
		"udym as tokenIn": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "udym",
				},
			},
			tokenOut:    sdk.NewCoin("foo", sdk.NewInt(tokenInAmt)),
			expectError: false,
		},
		"udym swapped in first pool": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "foo",
				},
			},
			tokenOut:    sdk.NewCoin("udym", sdk.NewInt(tokenInAmt)),
			expectError: false,
		},
		"usdc swapped in first pool": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       2,
					TokenInDenom: "foo",
				},
			},
			tokenOut:    sdk.NewCoin("bar", sdk.NewInt(tokenInAmt)),
			expectError: false,
		},
		"usdc as token in": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       2,
					TokenInDenom: "bar",
				},
			},
			tokenOut:    sdk.NewCoin("foo", sdk.NewInt(tokenInAmt)),
			expectError: false,
		},
		"usdc as token in - no route to dym": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       4,
					TokenInDenom: "bar",
				},
			},
			tokenOut:    sdk.NewCoin("baz", sdk.NewInt(tokenInAmt)),
			expectError: false,
		},
		"usdc swap with dym": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       3,
					TokenInDenom: "bar",
				},
			},
			tokenOut:    sdk.NewCoin("udym", sdk.NewInt(tokenInAmt)),
			expectError: false,
		},
	}

	for name, tc := range testcases {
		suite.SetupTest()
		suite.FundAcc(suite.TestAccs[0], apptesting.DefaultAcctFunds)
		params := suite.App.GAMMKeeper.GetParams(suite.Ctx)
		params.PoolCreationFee = sdk.NewCoins(
			sdk.NewCoin("udym", sdk.NewInt(1000)),
			sdk.NewCoin("bar", sdk.NewInt(1000)))
		suite.App.GAMMKeeper.SetParams(suite.Ctx, params)

		ctx := suite.Ctx
		msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

		pool1coins := []sdk.Coin{sdk.NewCoin("udym", sdk.NewInt(100000000)), sdk.NewCoin("foo", sdk.NewInt(100000000))}
		suite.PrepareBalancerPoolWithCoins(pool1coins...)

		//"bar" is treated as baseDenom (e.g. USDC)
		pool2coins := []sdk.Coin{sdk.NewCoin("bar", sdk.NewInt(100000000)), sdk.NewCoin("foo", sdk.NewInt(100000000))}
		suite.PrepareBalancerPoolWithCoins(pool2coins...)

		pool3coins := []sdk.Coin{sdk.NewCoin("bar", sdk.NewInt(100000000)), sdk.NewCoin("udym", sdk.NewInt(100000000))}
		suite.PrepareBalancerPoolWithCoins(pool3coins...)

		pool4coins := []sdk.Coin{sdk.NewCoin("bar", sdk.NewInt(100000000)), sdk.NewCoin("baz", sdk.NewInt(100000000))}
		suite.PrepareBalancerPoolWithCoins(pool4coins...)

		//check total supply before swap
		suppliesBefore := make(map[string]sdk.Int)
		suppliesBefore["udym"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "udym").Amount
		suppliesBefore["foo"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "foo").Amount
		suppliesBefore["bar"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "bar").Amount

		// check taker fee is not 0
		suite.Require().True(suite.App.GAMMKeeper.GetParams(ctx).TakerFee.GT(sdk.ZeroDec()))

		// make swap
		_, err := msgServer.SwapExactAmountOut(sdk.WrapSDKContext(ctx), &types.MsgSwapExactAmountOut{
			Sender:           suite.TestAccs[0].String(),
			Routes:           tc.routes,
			TokenOut:         tc.tokenOut,
			TokenInMaxAmount: sdk.NewInt(1000000000000000000),
		})
		if tc.expectError {
			suite.Require().Error(err, name)
			continue
		}
		suite.Require().NoError(err, name)

		// check total supply after swap
		suppliesAfter := make(map[string]sdk.Int)
		suppliesAfter["udym"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "udym").Amount
		suppliesAfter["foo"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "foo").Amount
		suppliesAfter["bar"] = suite.App.BankKeeper.GetSupply(suite.Ctx, "bar").Amount

		//validate total supply is reduced by taker fee
		suite.Require().True(suppliesAfter["udym"].LT(suppliesBefore["udym"]), name)
		suite.Require().True(suppliesAfter["foo"].Equal(suppliesBefore["foo"]), name)
		suite.Require().True(suppliesAfter["bar"].Equal(suppliesBefore["bar"]), name)
	}
}

//TODO: test estimation when taker fee is 0

// TestEstimateMultihopSwapExactAmountIn tests that the estimation done via `EstimateSwapExactAmountIn`
func (suite *KeeperTestSuite) TestEstimateMultihopSwapExactAmountIn() {
	type param struct {
		routes            []poolmanagertypes.SwapAmountInRoute
		tokenIn           sdk.Coin
		tokenOutMinAmount sdk.Int
	}

	tests := []struct {
		name     string
		param    param
		poolType poolmanagertypes.PoolType
	}{
		{
			name: "Proper swap - foo -> bar(pool 1) - bar(pool 2) -> baz",
			param: param{
				routes: []poolmanagertypes.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: "bar",
					},
					{
						PoolId:        2,
						TokenOutDenom: "baz",
					},
				},
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
		},
		{
			name: "Swap - foo -> udym(pool 1) - udym(pool 2) -> baz ",
			param: param{
				routes: []poolmanagertypes.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: "udym",
					},
					{
						PoolId:        2,
						TokenOutDenom: "baz",
					},
				},
				tokenIn:           sdk.NewCoin("udym", sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			firstEstimatePoolId := suite.PrepareBalancerPool()
			secondEstimatePoolId := suite.PrepareBalancerPool()

			firstEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			// calculate token out amount using `EstimateMultihopSwapExactAmountIn`
			queryClient := suite.queryClient
			estimateMultihopTokenOutAmountWithTakerFee, errEstimate := queryClient.EstimateSwapExactAmountIn(
				suite.Ctx,
				&types.QuerySwapExactAmountInRequest{
					TokenIn: test.param.tokenIn.String(),
					Routes:  test.param.routes,
				},
			)
			suite.Require().NoError(errEstimate, "test: %v", test.name)

			// ensure that pool state has not been altered after estimation
			firstEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)
			suite.Require().Equal(firstEstimatePool, firstEstimatePoolAfterSwap)
			suite.Require().Equal(secondEstimatePool, secondEstimatePoolAfterSwap)

			// calculate token out amount using `MultihopSwapExactAmountIn`
			poolmanagerKeeper := suite.App.PoolManagerKeeper
			multihopTokenOutAmount, errMultihop := poolmanagerKeeper.MultihopEstimateOutGivenExactAmountIn(
				suite.Ctx,
				test.param.routes,
				test.param.tokenIn,
			)
			suite.Require().NoError(errMultihop, "test: %v", test.name)
			// the pool manager estimation is without taker fee, so it should be higher
			suite.Require().True(multihopTokenOutAmount.GT(estimateMultihopTokenOutAmountWithTakerFee.TokenOutAmount))

			// Now reducing taker fee from the input, we expect the estimation to be the same
			reducedTokenIn := sdk.NewDecFromInt(test.param.tokenIn.Amount).MulTruncate(sdk.OneDec().Sub(suite.App.GAMMKeeper.GetParams(suite.Ctx).TakerFee))
			reducedTokenInCoin := sdk.NewCoin(test.param.tokenIn.Denom, reducedTokenIn.TruncateInt())

			multihopTokenOutAmountTakerFeeReduced, errMultihop := poolmanagerKeeper.MultihopEstimateOutGivenExactAmountIn(
				suite.Ctx,
				test.param.routes,
				reducedTokenInCoin,
			)
			suite.Require().Equal(estimateMultihopTokenOutAmountWithTakerFee.TokenOutAmount, multihopTokenOutAmountTakerFeeReduced)
		})
	}
}

func (suite *KeeperTestSuite) TestEstimateMultihopSwapExactAmountOut() {
	type param struct {
		routes           []poolmanagertypes.SwapAmountOutRoute
		tokenOut         sdk.Coin
		tokenInMinAmount sdk.Int
	}

	tests := []struct {
		name     string
		param    param
		poolType poolmanagertypes.PoolType
	}{
		{
			name: "Proper swap - foo -> bar(pool 1) - bar(pool 2) -> baz",
			param: param{
				routes: []poolmanagertypes.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: "foo",
					},
					{
						PoolId:       2,
						TokenInDenom: "bar",
					},
				},
				tokenInMinAmount: sdk.NewInt(1),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
		},
		{
			name: "Swap - foo -> udym(pool 1) - udym(pool 2) -> baz ",
			param: param{
				routes: []poolmanagertypes.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: "foo",
					},
					{
						PoolId:       2,
						TokenInDenom: "udym",
					},
				},
				tokenInMinAmount: sdk.NewInt(1),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			firstEstimatePoolId := suite.PrepareBalancerPool()
			secondEstimatePoolId := suite.PrepareBalancerPool()

			firstEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			// calculate token out amount using `EstimateMultihopSwapExactAmountIn`
			queryClient := suite.queryClient
			estimateMultihopTokenInAmountWithTakerFee, errEstimate := queryClient.EstimateSwapExactAmountOut(
				suite.Ctx,
				&types.QuerySwapExactAmountOutRequest{
					TokenOut: test.param.tokenOut.String(),
					Routes:   test.param.routes,
				},
			)
			suite.Require().NoError(errEstimate, "test: %v", test.name)

			// ensure that pool state has not been altered after estimation
			firstEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)
			suite.Require().Equal(firstEstimatePool, firstEstimatePoolAfterSwap)
			suite.Require().Equal(secondEstimatePool, secondEstimatePoolAfterSwap)

			// calculate token out amount using `MultihopSwapExactAmountIn`
			poolmanagerKeeper := suite.App.PoolManagerKeeper
			multihopTokenInAmount, errMultihop := poolmanagerKeeper.MultihopEstimateInGivenExactAmountOut(
				suite.Ctx,
				test.param.routes,
				test.param.tokenOut,
			)
			suite.Require().NoError(errMultihop, "test: %v", test.name)
			// the pool manager estimation is without taker fee, so it should be lower (less tokens in for same amount out)
			suite.Require().True(multihopTokenInAmount.LT(estimateMultihopTokenInAmountWithTakerFee.TokenInAmount))

			takerFee := suite.App.GAMMKeeper.GetParams(suite.Ctx).TakerFee
			tokensAfterTakerFeeReduction := sdk.NewDecFromInt(estimateMultihopTokenInAmountWithTakerFee.TokenInAmount).MulTruncate(sdk.OneDec().Sub(takerFee))

			// Now reducing taker fee from the input, we expect the estimation to be the same
			suite.Require().Equal(tokensAfterTakerFeeReduction.TruncateInt(), multihopTokenInAmount)
		})
	}
}
