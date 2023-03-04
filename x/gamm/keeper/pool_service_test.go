package keeper_test

import (
	"fmt"
	"time"

	_ "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	_ "github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

var (
	defaultSwapFee    = sdk.MustNewDecFromStr("0.025")
	defaultExitFee    = sdk.MustNewDecFromStr("0.025")
	defaultPoolParams = balancer.PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultStableSwapPoolParams = stableswap.PoolParams{
		SwapFee: defaultSwapFee,
		ExitFee: defaultExitFee,
	}
	defaultScalingFactor  = []uint64{1, 1}
	defaultFutureGovernor = ""

	// pool assets
	defaultFooAsset = balancertypes.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}
	defaultBarAsset = balancertypes.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}
	defaultPoolAssets                     = []balancertypes.PoolAsset{defaultFooAsset, defaultBarAsset}
	defaultStableSwapPoolAssets sdk.Coins = sdk.NewCoins(
		sdk.NewCoin("foo", sdk.NewInt(10000)),
		sdk.NewCoin("bar", sdk.NewInt(10000)),
	)
	defaultAcctFunds sdk.Coins = sdk.NewCoins(
		sdk.NewCoin("uosmo", sdk.NewInt(10000000000)),
		sdk.NewCoin("foo", sdk.NewInt(10000000)),
		sdk.NewCoin("bar", sdk.NewInt(10000000)),
		sdk.NewCoin("baz", sdk.NewInt(10000000)),
	)
	ETH                       = "eth"
	USDC                      = "usdc"
	defaultTickSpacing        = uint64(1)
	DefaultExponentAtPriceOne = sdk.NewInt(-4)
)

func (suite *KeeperTestSuite) TestCreateBalancerPool() {
	params := suite.App.GAMMKeeper.GetParams(suite.Ctx)
	testAccount := suite.TestAccs[0]

	// get raw pool creation fee(s) as DecCoins
	poolCreationFeeDecCoins := sdk.DecCoins{}
	for _, coin := range params.PoolCreationFee {
		poolCreationFeeDecCoins = poolCreationFeeDecCoins.Add(sdk.NewDecCoin(coin.Denom, coin.Amount))
	}

	// TODO: should be moved to balancer package
	tests := []struct {
		name        string
		msg         balancertypes.MsgCreateBalancerPool
		emptySender bool
		expectPass  bool
	}{
		{
			name:        "create pool with default assets",
			msg:         balancer.NewMsgCreateBalancerPool(testAccount, defaultPoolParams, defaultPoolAssets, defaultFutureGovernor),
			emptySender: false,
			expectPass:  true,
		}, {
			name:        "create pool with no assets",
			msg:         balancer.NewMsgCreateBalancerPool(testAccount, defaultPoolParams, defaultPoolAssets, defaultFutureGovernor),
			emptySender: true,
			expectPass:  false,
		}, {
			name: "create a pool with negative swap fee",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(-1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, defaultPoolAssets, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create a pool with negative exit fee",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(-1, 2),
			}, defaultPoolAssets, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create the pool with empty PoolAssets",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{}, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create the pool with 0 weighted PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(0),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create the pool with negative weighted PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(-1),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create the pool with 0 balance PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(0)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create the pool with negative balance PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token: sdk.Coin{
					Denom:  "foo",
					Amount: sdk.NewInt(-1),
				},
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create the pool with duplicated PoolAssets",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, []balancertypes.PoolAsset{{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}, {
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
			}}, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()
		gammKeeper := suite.App.GAMMKeeper
		poolmanagerKeeper := suite.App.PoolManagerKeeper
		distributionKeeper := suite.App.DistrKeeper
		bankKeeper := suite.App.BankKeeper

		// fund sender test account
		sender, err := sdk.AccAddressFromBech32(test.msg.Sender)
		suite.Require().NoError(err, "test: %v", test.name)
		if !test.emptySender {
			suite.FundAcc(sender, defaultAcctFunds)
		}

		// note starting balances for community fee pool and pool creator account
		feePoolBalBeforeNewPool := distributionKeeper.GetFeePoolCommunityCoins(suite.Ctx)
		senderBalBeforeNewPool := bankKeeper.GetAllBalances(suite.Ctx, sender)

		// attempt to create a pool with the given NewMsgCreateBalancerPool message
		poolId, err := poolmanagerKeeper.CreatePool(suite.Ctx, test.msg)

		if test.expectPass {
			suite.Require().NoError(err, "test: %v", test.name)

			// check to make sure new pool exists and has minted the correct number of pool shares
			pool, err := gammKeeper.GetPoolAndPoke(suite.Ctx, poolId)
			suite.Require().NoError(err, "test: %v", test.name)
			suite.Require().Equal(types.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
				fmt.Sprintf("share token should be minted as %s initially", types.InitPoolSharesSupply.String()),
			)

			// make sure pool creation fee is correctly sent to community pool
			feePool := distributionKeeper.GetFeePoolCommunityCoins(suite.Ctx)
			suite.Require().Equal(feePool, feePoolBalBeforeNewPool.Add(poolCreationFeeDecCoins...))

			// get expected tokens in new pool and corresponding pool shares
			expectedPoolTokens := sdk.Coins{}
			for _, asset := range test.msg.GetPoolAssets() {
				expectedPoolTokens = expectedPoolTokens.Add(asset.Token)
			}
			expectedPoolShares := sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)

			// make sure sender's balance is updated correctly
			senderBal := bankKeeper.GetAllBalances(suite.Ctx, sender)
			expectedSenderBal := senderBalBeforeNewPool.Sub(params.PoolCreationFee).Sub(expectedPoolTokens).Add(expectedPoolShares)
			suite.Require().Equal(senderBal.String(), expectedSenderBal.String())

			// check pool's liquidity is correctly increased
			liquidity := gammKeeper.GetTotalLiquidity(suite.Ctx)
			suite.Require().Equal(expectedPoolTokens.String(), liquidity.String())
		} else {
			suite.Require().Error(err, "test: %v", test.name)
		}
	}
}

func (suite *KeeperTestSuite) TestInitializePool() {
	testAccount := suite.TestAccs[0]

	tests := []struct {
		name        string
		createPool  func() poolmanagertypes.PoolI
		expectPass  bool
		expectPanic bool
	}{
		{
			name: "initialize balancer pool with default assets",
			createPool: func() poolmanagertypes.PoolI {
				balancerPool, err := balancer.NewBalancerPool(
					defaultPoolId,
					defaultPoolParams,
					defaultPoolAssets,
					"",
					time.Now(),
				)
				require.NoError(suite.T(), err)
				return &balancerPool
			},
			expectPass: true,
		},
		{
			name: "initialize stableswap pool with default assets",
			createPool: func() poolmanagertypes.PoolI {
				stableswapPool, err := stableswap.NewStableswapPool(
					defaultPoolId,
					defaultPoolParamsStableSwap,
					defaultStableSwapPoolAssets,
					defaultScalingFactor,
					"",
					defaultFutureGovernor,
				)
				require.NoError(suite.T(), err)
				return &stableswapPool
			},
			expectPass: true,
		},
		{
			name: "initialize a CL pool which cause panic",
			createPool: func() poolmanagertypes.PoolI {
				return suite.PrepareConcentratedPool()
			},
			expectPanic: true,
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			osmoassert.ConditionalPanic(suite.T(), test.expectPanic, func() {
				suite.SetupTest()

				gammKeeper := suite.App.GAMMKeeper
				bankKeeper := suite.App.BankKeeper
				poolIncentivesKeeper := suite.App.PoolIncentivesKeeper

				// sender test account
				sender := testAccount
				senderBalBeforeNewPool := bankKeeper.GetAllBalances(suite.Ctx, sender)

				// initializePool with a poolI
				err := gammKeeper.InitializePool(suite.Ctx, test.createPool(), sender)

				if test.expectPass {
					suite.Require().NoError(err, "test: %v", test.name)

					// check to make sure new pool exists and has minted the correct number of pool shares
					pool, err := gammKeeper.GetPoolAndPoke(suite.Ctx, defaultPoolId)
					suite.Require().NoError(err, "test: %v", test.name)
					suite.Require().Equal(types.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
						fmt.Sprintf("share token should be minted as %s initially", types.InitPoolSharesSupply),
					)

					// check to make sure user user balance increase correct number of pool shares
					suite.Require().Equal(
						senderBalBeforeNewPool.Add(sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)),
						bankKeeper.GetAllBalances(suite.Ctx, sender),
					)

					// get expected tokens in new pool and corresponding pool shares
					expectedPoolTokens := sdk.NewCoins()
					for _, asset := range pool.GetTotalPoolLiquidity(suite.Ctx) {
						expectedPoolTokens = expectedPoolTokens.Add(asset)
					}
					expectedPoolShares := sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)

					// make sure expected pool tokens and expected pool shares matches the actual tokens and shares in the pool
					suite.Require().Equal(expectedPoolTokens.String(), pool.GetTotalPoolLiquidity(suite.Ctx).String())
					suite.Require().Equal(expectedPoolShares.Amount.String(), pool.GetTotalShares().String())

					// check pool metadata
					poolShareBaseDenom := types.GetPoolShareDenom(pool.GetId())
					poolShareDisplayDenom := fmt.Sprintf("GAMM-%d", pool.GetId())
					metadata, found := bankKeeper.GetDenomMetaData(suite.Ctx, poolShareBaseDenom)
					suite.Require().Equal(found, true, fmt.Sprintf("Pool share denom %s is not set", poolShareDisplayDenom))
					suite.Require().Equal(metadata.Base, poolShareBaseDenom, fmt.Sprintf("Pool share base denom %s is not correctly set", poolShareBaseDenom))
					suite.Require().Equal(metadata.Display, poolShareDisplayDenom, fmt.Sprintf("Pool share display denom %s is not correctly set", poolShareDisplayDenom))
					suite.Require().Equal(metadata.DenomUnits[0].Denom, poolShareBaseDenom)
					suite.Require().Equal(metadata.DenomUnits[0].Exponent, uint32(0x0))
					suite.Require().Equal(metadata.DenomUnits[0].Aliases, []string{
						"attopoolshare",
					})
					suite.Require().Equal(metadata.DenomUnits[1].Denom, poolShareDisplayDenom)
					suite.Require().Equal(metadata.DenomUnits[1].Exponent, uint32(types.OneShareExponent))
					suite.Require().Equal(metadata.DenomUnits[1].Aliases, []string(nil))

					// check AfterPoolCreated hook
					for _, lockableDuration := range poolIncentivesKeeper.GetLockableDurations(suite.Ctx) {
						gaugeId, err := poolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, defaultPoolId, lockableDuration)
						suite.Require().NoError(err, "test: %v", test.name)

						poolIdFromPoolIncentives, err := poolIncentivesKeeper.GetPoolIdFromGaugeId(suite.Ctx, gaugeId, lockableDuration)
						suite.Require().NoError(err, "test: %v", test.name)
						suite.Require().Equal(poolIdFromPoolIncentives, defaultPoolId)
					}

				} else {
					suite.Require().Error(err, "test: %v", test.name)
				}
			})
		})
	}
}

// This test creates several pools, and tests that:
// the condition is in a case where the balancer return value returns an overflowing value
// the SpotPrice query does not
func (suite *KeeperTestSuite) TestSpotPriceOverflow() {
	denomA := "denomA"
	denomB := "denomB"
	tests := map[string]struct {
		poolLiquidity   sdk.Coins
		poolWeights     []int64
		quoteAssetDenom string
		baseAssetDenom  string
		overflows       bool
		panics          bool
	}{
		"uniV2marginalOverflow": {
			poolLiquidity: sdk.NewCoins(sdk.NewCoin(denomA, types.MaxSpotPrice.TruncateInt().Add(sdk.OneInt())),
				sdk.NewCoin(denomB, sdk.OneInt())),
			poolWeights:     []int64{1, 1},
			quoteAssetDenom: denomA,
			baseAssetDenom:  denomB,
			overflows:       true,
		},
		"uniV2 internal error": {
			poolLiquidity: sdk.NewCoins(sdk.NewCoin(denomA, sdk.NewDec(2).Power(250).TruncateInt()),
				sdk.NewCoin(denomB, sdk.OneInt())),
			poolWeights:     []int64{1, 1 << 19},
			quoteAssetDenom: denomB,
			baseAssetDenom:  denomA,
			panics:          true,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			poolId := suite.PrepareBalancerPoolWithCoinsAndWeights(tc.poolLiquidity, tc.poolWeights)
			pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
			suite.Require().NoError(err)
			var poolSpotPrice sdk.Dec
			var poolErr error
			osmoassert.ConditionalPanic(suite.T(), tc.panics, func() {
				poolSpotPrice, poolErr = pool.SpotPrice(suite.Ctx, tc.baseAssetDenom, tc.quoteAssetDenom)
			})
			keeperSpotPrice, keeperErr := suite.App.GAMMKeeper.CalculateSpotPrice(suite.Ctx, poolId, tc.quoteAssetDenom, tc.baseAssetDenom)
			if tc.overflows {
				suite.Require().NoError(poolErr)
				suite.Require().ErrorIs(keeperErr, types.ErrSpotPriceOverflow)
				suite.Require().Error(keeperErr)
				suite.Require().Equal(types.MaxSpotPrice, keeperSpotPrice)
			} else if tc.panics {
				suite.Require().ErrorIs(keeperErr, types.ErrSpotPriceInternal)
				suite.Require().Error(keeperErr)
				suite.Require().Equal(sdk.Dec{}, keeperSpotPrice)
			} else {
				suite.Require().NoError(poolErr)
				suite.Require().NoError(keeperErr)
				suite.Require().Equal(poolSpotPrice, keeperSpotPrice)
			}
		})
	}
}

// TODO: Add more edge cases around TokenInMaxs not containing every token in pool.
func (suite *KeeperTestSuite) TestJoinPoolNoSwap() {
	fiveKFooAndBar := sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(5000)), sdk.NewCoin("foo", sdk.NewInt(5000)))
	tests := []struct {
		name            string
		txSender        sdk.AccAddress
		sharesRequested sdk.Int
		tokenInMaxs     sdk.Coins
		expectPass      bool
	}{
		{
			name:            "basic join no swap",
			txSender:        suite.TestAccs[1],
			sharesRequested: types.OneShare.MulRaw(50),
			tokenInMaxs:     sdk.Coins{},
			expectPass:      true,
		},
		{
			name:            "join no swap with zero shares requested",
			txSender:        suite.TestAccs[1],
			sharesRequested: sdk.NewInt(0),
			tokenInMaxs:     sdk.Coins{},
			expectPass:      false,
		},
		{
			name:            "join no swap with negative shares requested",
			txSender:        suite.TestAccs[1],
			sharesRequested: sdk.NewInt(-1),
			tokenInMaxs:     sdk.Coins{},
			expectPass:      false,
		},
		{
			name:            "join no swap with insufficient funds",
			txSender:        suite.TestAccs[1],
			sharesRequested: sdk.NewInt(-1),
			tokenInMaxs: sdk.Coins{
				sdk.NewCoin("bar", sdk.NewInt(4999)), sdk.NewCoin("foo", sdk.NewInt(4999)),
			},
			expectPass: false,
		},
		{
			name:            "join no swap with exact tokenInMaxs",
			txSender:        suite.TestAccs[1],
			sharesRequested: types.OneShare.MulRaw(50),
			tokenInMaxs: sdk.Coins{
				fiveKFooAndBar[0], fiveKFooAndBar[1],
			},
			expectPass: true,
		},
		{
			name:            "join no swap with arbitrary extra token in tokenInMaxs",
			txSender:        suite.TestAccs[1],
			sharesRequested: types.OneShare.MulRaw(50),
			tokenInMaxs: sdk.Coins{
				fiveKFooAndBar[0], fiveKFooAndBar[1], sdk.NewCoin("baz", sdk.NewInt(5000)),
			},
			expectPass: false,
		},
		{
			name:            "join no swap with TokenInMaxs not containing every token in pool",
			txSender:        suite.TestAccs[1],
			sharesRequested: types.OneShare.MulRaw(50),
			tokenInMaxs: sdk.Coins{
				fiveKFooAndBar[0],
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		ctx := suite.Ctx
		gammKeeper := suite.App.GAMMKeeper
		poolmanagerKeeper := suite.App.PoolManagerKeeper
		bankKeeper := suite.App.BankKeeper
		testAccount := suite.TestAccs[0]

		// Mint some assets to the accounts.
		suite.FundAcc(testAccount, defaultAcctFunds)

		// Create the pool at first
		msg := balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, defaultPoolAssets, defaultFutureGovernor)
		poolId, err := poolmanagerKeeper.CreatePool(suite.Ctx, msg)
		suite.Require().NoError(err, "test: %v", test.name)

		suite.FundAcc(test.txSender, defaultAcctFunds)

		balancesBefore := bankKeeper.GetAllBalances(suite.Ctx, test.txSender)
		_, _, err = gammKeeper.JoinPoolNoSwap(suite.Ctx, test.txSender, poolId, test.sharesRequested, test.tokenInMaxs)

		if test.expectPass {
			suite.Require().NoError(err, "test: %v", test.name)
			suite.Require().Equal(test.sharesRequested.String(), bankKeeper.GetBalance(suite.Ctx, test.txSender, "gamm/pool/1").Amount.String())
			balancesAfter := bankKeeper.GetAllBalances(suite.Ctx, test.txSender)
			deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
			// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100000000gamm/pool/1.
			// Thus, to get the 50*OneShare gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be provided.
			suite.Require().Equal("5000", deltaBalances.AmountOf("foo").String())
			suite.Require().Equal("5000", deltaBalances.AmountOf("bar").String())

			liquidity := gammKeeper.GetTotalLiquidity(suite.Ctx)
			suite.Require().Equal("15000bar,15000foo", liquidity.String())

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, 1)
		} else {
			suite.Require().Error(err, "test: %v", test.name)

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, 0)
		}
	}
}

func (suite *KeeperTestSuite) TestExitPool() {
	fiveKFooAndBar := sdk.NewCoins(sdk.NewCoin("bar", sdk.NewInt(5000)), sdk.NewCoin("foo", sdk.NewInt(5000)))
	tests := []struct {
		name         string
		txSender     sdk.AccAddress
		sharesIn     sdk.Int
		tokenOutMins sdk.Coins
		emptySender  bool
		expectPass   bool
	}{
		{
			name:         "attempt exit pool with no pool share balance",
			txSender:     suite.TestAccs[0],
			sharesIn:     types.OneShare.MulRaw(50),
			tokenOutMins: sdk.Coins{},
			emptySender:  true,
			expectPass:   false,
		},
		{
			name:         "exit half pool with correct pool share balance",
			txSender:     suite.TestAccs[0],
			sharesIn:     types.OneShare.MulRaw(50),
			tokenOutMins: sdk.Coins{},
			emptySender:  false,
			expectPass:   true,
		},
		{
			name:         "attempt exit pool requesting 0 share amount",
			txSender:     suite.TestAccs[0],
			sharesIn:     sdk.NewInt(0),
			tokenOutMins: sdk.Coins{},
			emptySender:  false,
			expectPass:   false,
		},
		{
			name:         "attempt exit pool requesting negative share amount",
			txSender:     suite.TestAccs[0],
			sharesIn:     sdk.NewInt(-1),
			tokenOutMins: sdk.Coins{},
			emptySender:  false,
			expectPass:   false,
		},
		{
			name:     "attempt exit pool with tokenOutMins above actual output",
			txSender: suite.TestAccs[0],
			sharesIn: types.OneShare.MulRaw(50),
			tokenOutMins: sdk.Coins{
				sdk.NewCoin("foo", sdk.NewInt(5001)),
			},
			emptySender: false,
			expectPass:  false,
		},
		{
			name:     "attempt exit pool requesting tokenOutMins at exactly the actual output",
			txSender: suite.TestAccs[0],
			sharesIn: types.OneShare.MulRaw(50),
			tokenOutMins: sdk.Coins{
				fiveKFooAndBar[1],
			},
			emptySender: false,
			expectPass:  true,
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			ctx := suite.Ctx

			gammKeeper := suite.App.GAMMKeeper
			bankKeeper := suite.App.BankKeeper
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			// Mint assets to the pool creator
			suite.FundAcc(test.txSender, defaultAcctFunds)

			// Create the pool at first
			msg := balancer.NewMsgCreateBalancerPool(test.txSender, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)
			poolId, err := poolmanagerKeeper.CreatePool(ctx, msg)

			// If we are testing insufficient pool share balances, switch tx sender from pool creator to empty account
			if test.emptySender {
				test.txSender = suite.TestAccs[1]
			}

			balancesBefore := bankKeeper.GetAllBalances(suite.Ctx, test.txSender)
			_, err = gammKeeper.ExitPool(ctx, test.txSender, poolId, test.sharesIn, test.tokenOutMins)

			if test.expectPass {
				suite.Require().NoError(err, "test: %v", test.name)
				suite.Require().Equal(test.sharesIn.String(), bankKeeper.GetBalance(suite.Ctx, test.txSender, "gamm/pool/1").Amount.String())
				balancesAfter := bankKeeper.GetAllBalances(suite.Ctx, test.txSender)
				deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
				// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100*OneShare gamm/pool/1.
				// Thus, to refund the 50*OneShare gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be refunded.
				suite.Require().Equal("-5000", deltaBalances.AmountOf("foo").String())
				suite.Require().Equal("-5000", deltaBalances.AmountOf("bar").String())

				liquidity := gammKeeper.GetTotalLiquidity(ctx)
				suite.Require().Equal("5000bar,5000foo", liquidity.String())

				suite.AssertEventEmitted(ctx, types.TypeEvtPoolExited, 1)
			} else {
				suite.Require().Error(err, "test: %v", test.name)
				suite.AssertEventEmitted(ctx, types.TypeEvtPoolExited, 0)
			}
		})
	}
}

// TestJoinPoolExitPool_InverseRelationship tests that joining pool and exiting pool
// guarantees same amount in and out
func (suite *KeeperTestSuite) TestJoinPoolExitPool_InverseRelationship() {
	testCases := []struct {
		name             string
		pool             balancertypes.MsgCreateBalancerPool
		joinPoolShareAmt sdk.Int
	}{
		{
			name: "pool with same token ratio",
			pool: balancer.NewMsgCreateBalancerPool(nil, balancer.PoolParams{
				SwapFee: sdk.ZeroDec(),
				ExitFee: sdk.ZeroDec(),
			}, []balancertypes.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
				},
			}, defaultFutureGovernor),
			joinPoolShareAmt: types.OneShare.MulRaw(50),
		},
		{
			name: "pool with different token ratio",
			pool: balancer.NewMsgCreateBalancerPool(nil, balancer.PoolParams{
				SwapFee: sdk.ZeroDec(),
				ExitFee: sdk.ZeroDec(),
			}, []balancertypes.PoolAsset{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("foo", sdk.NewInt(7000)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
				},
			}, defaultFutureGovernor),
			joinPoolShareAmt: types.OneShare.MulRaw(50),
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()

		suite.Run(tc.name, func() {
			ctx := suite.Ctx
			gammKeeper := suite.App.GAMMKeeper
			poolmanagerKeeper := suite.App.PoolManagerKeeper

			for _, acc := range suite.TestAccs {
				suite.FundAcc(acc, defaultAcctFunds)
			}

			createPoolAcc := suite.TestAccs[0]
			joinPoolAcc := suite.TestAccs[1]

			// test account is set on every test case iteration, we need to manually update address for pool creator
			tc.pool.Sender = createPoolAcc.String()

			poolId, err := poolmanagerKeeper.CreatePool(ctx, tc.pool)
			suite.Require().NoError(err)

			balanceBeforeJoin := suite.App.BankKeeper.GetAllBalances(ctx, joinPoolAcc)

			_, _, err = gammKeeper.JoinPoolNoSwap(ctx, joinPoolAcc, poolId, tc.joinPoolShareAmt, sdk.Coins{})
			suite.Require().NoError(err)

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, 1)

			_, err = gammKeeper.ExitPool(ctx, joinPoolAcc, poolId, tc.joinPoolShareAmt, sdk.Coins{})

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolExited, 1)

			balanceAfterExit := suite.App.BankKeeper.GetAllBalances(ctx, joinPoolAcc)
			deltaBalance, _ := balanceBeforeJoin.SafeSub(balanceAfterExit)

			// due to rounding, `balanceBeforeJoin` and `balanceAfterExit` have neglectable difference
			// coming from rounding in exitPool.Here we test if the difference is within rounding tolerance range
			roundingToleranceCoins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1)), sdk.NewCoin("bar", sdk.NewInt(1)))
			suite.Require().True(deltaBalance.AmountOf("foo").LTE(roundingToleranceCoins.AmountOf("foo")))
			suite.Require().True(deltaBalance.AmountOf("bar").LTE(roundingToleranceCoins.AmountOf("bar")))
		})
	}
}

func (suite *KeeperTestSuite) TestActiveBalancerPool() {
	type testCase struct {
		blockTime  time.Time
		expectPass bool
	}

	testCases := []testCase{
		{time.Unix(1000, 0), true},
	}

	for _, tc := range testCases {
		suite.Run("", func() {
			suite.SetupTest()

			ctx := suite.Ctx
			gammKeeper := suite.App.GAMMKeeper
			testAccount := suite.TestAccs[0]

			suite.FundAcc(testAccount, defaultAcctFunds)

			// Create the pool at first
			poolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})
			ctx = ctx.WithBlockTime(tc.blockTime)

			// uneffected by start time
			_, _, err := gammKeeper.JoinPoolNoSwap(ctx, testAccount, poolId, types.OneShare.MulRaw(50), sdk.Coins{})
			suite.Require().NoError(err)

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, 1)

			_, err = gammKeeper.ExitPool(ctx, testAccount, poolId, types.InitPoolSharesSupply.QuoRaw(2), sdk.Coins{})
			suite.Require().NoError(err)

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolExited, 1)

			foocoin := sdk.NewCoin("foo", sdk.NewInt(10))
			foocoins := sdk.Coins{foocoin}

			if tc.expectPass {
				_, err = gammKeeper.JoinSwapExactAmountIn(ctx, testAccount, poolId, foocoins, sdk.ZeroInt())
				suite.Require().NoError(err)
				_, err = gammKeeper.JoinSwapShareAmountOut(ctx, testAccount, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)
				_, err = gammKeeper.ExitSwapShareAmountIn(ctx, testAccount, poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				suite.Require().NoError(err)
				_, err = gammKeeper.ExitSwapExactAmountOut(ctx, testAccount, poolId, foocoin, sdk.NewInt(1000000000000000000))
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				_, err = gammKeeper.JoinSwapShareAmountOut(ctx, testAccount, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				suite.Require().Error(err)
				_, err = gammKeeper.ExitSwapShareAmountIn(ctx, testAccount, poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				suite.Require().Error(err)
				_, err = gammKeeper.ExitSwapExactAmountOut(ctx, testAccount, poolId, foocoin, sdk.NewInt(1000000000000000000))
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestJoinSwapExactAmountInConsistency() {
	testCases := []struct {
		name              string
		poolSwapFee       sdk.Dec
		poolExitFee       sdk.Dec
		tokensIn          sdk.Coins
		shareOutMinAmount sdk.Int
		expectedSharesOut sdk.Int
		tokenOutMinAmount sdk.Int
	}{
		{
			name:              "single coin with zero swap and exit fees",
			poolSwapFee:       sdk.ZeroDec(),
			poolExitFee:       sdk.ZeroDec(),
			tokensIn:          sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000000))),
			shareOutMinAmount: sdk.ZeroInt(),
			expectedSharesOut: sdk.NewInt(6265857020099440400),
			tokenOutMinAmount: sdk.ZeroInt(),
		},
		// TODO: Uncomment or remove this following test case once the referenced
		// issue is resolved.
		//
		// Ref: https://github.com/osmosis-labs/osmosis/issues/1196
		// {
		// 	name:              "single coin with positive swap fee and zero exit fee",
		// 	poolSwapFee:       sdk.NewDecWithPrec(1, 2),
		// 	poolExitFee:       sdk.ZeroDec(),
		// 	tokensIn:          sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000000))),
		// 	shareOutMinAmount: sdk.ZeroInt(),
		// 	expectedSharesOut: sdk.NewInt(6226484702880621000),
		// 	tokenOutMinAmount: sdk.ZeroInt(),
		// },
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := suite.Ctx
			gammKeeper := suite.App.GAMMKeeper
			testAccount := suite.TestAccs[0]

			poolID := suite.prepareCustomBalancerPool(
				defaultAcctFunds,
				[]balancertypes.PoolAsset{
					{
						Weight: sdk.NewInt(100),
						Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
					},
					{
						Weight: sdk.NewInt(200),
						Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
					},
				},
				balancer.PoolParams{
					SwapFee: tc.poolSwapFee,
					ExitFee: tc.poolExitFee,
				},
			)

			shares, err := gammKeeper.JoinSwapExactAmountIn(ctx, testAccount, poolID, tc.tokensIn, tc.shareOutMinAmount)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedSharesOut, shares)

			tokenOutAmt, err := gammKeeper.ExitSwapShareAmountIn(
				ctx,
				testAccount,
				poolID,
				tc.tokensIn[0].Denom,
				shares,
				tc.tokenOutMinAmount,
			)
			suite.Require().NoError(err)

			// require swapTokenOutAmt <= (tokenInAmt * (1 - tc.poolSwapFee))
			oneMinusSwapFee := sdk.OneDec().Sub(tc.poolSwapFee)
			swapFeeAdjustedAmount := oneMinusSwapFee.MulInt(tc.tokensIn[0].Amount).RoundInt()
			suite.Require().True(tokenOutAmt.LTE(swapFeeAdjustedAmount))

			// require swapTokenOutAmt + 10 > input
			suite.Require().True(
				swapFeeAdjustedAmount.Sub(tokenOutAmt).LTE(sdk.NewInt(10)),
				"expected out amount %s, actual out amount %s",
				swapFeeAdjustedAmount, tokenOutAmt,
			)
		})
	}
}

func (suite *KeeperTestSuite) TestGetPoolDenom() {
	// setup pool with denoms
	suite.FundAcc(suite.TestAccs[0], defaultAcctFunds)
	poolCreateMsg := balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], defaultPoolParams, defaultPoolAssets, defaultFutureGovernor)
	_, err := suite.App.PoolManagerKeeper.CreatePool(suite.Ctx, poolCreateMsg)
	suite.Require().NoError(err)

	for _, tc := range []struct {
		desc         string
		poolId       uint64
		expectDenoms []string
		expectErr    bool
	}{
		{
			desc:         "Valid PoolId",
			poolId:       1,
			expectDenoms: []string{"bar", "foo"},
			expectErr:    false,
		},
		{
			desc:         "Invalid PoolId",
			poolId:       2,
			expectDenoms: []string{"bar", "foo"},
			expectErr:    true,
		},
	} {
		suite.Run(tc.desc, func() {
			denoms, err := suite.App.GAMMKeeper.GetPoolDenoms(suite.Ctx, tc.poolId)

			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(denoms, tc.expectDenoms)
			}
		})
	}
}
