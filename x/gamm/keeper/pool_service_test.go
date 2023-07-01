package keeper_test

import (
	"fmt"
	"time"

	_ "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	_ "github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

var (
	defaultPoolParams = balancer.PoolParams{
		SwapFee: defaultSpreadFactor,
		ExitFee: defaultZeroExitFee,
	}

	defaultScalingFactor  = []uint64{1, 1}
	defaultFutureGovernor = ""

	// pool assets
	defaultFooAsset = balancer.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("foo", sdk.NewInt(10000)),
	}
	defaultBarAsset = balancer.PoolAsset{
		Weight: sdk.NewInt(100),
		Token:  sdk.NewCoin("bar", sdk.NewInt(10000)),
	}
	defaultPoolAssets                     = []balancer.PoolAsset{defaultFooAsset, defaultBarAsset}
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
	ETH                = "eth"
	USDC               = "usdc"
	defaultTickSpacing = uint64(100)
)

func (s *KeeperTestSuite) TestCreateBalancerPool() {
	params := s.App.GAMMKeeper.GetParams(s.Ctx)
	testAccount := s.TestAccs[0]

	// get raw pool creation fee(s) as DecCoins
	poolCreationFeeDecCoins := sdk.DecCoins{}
	for _, coin := range params.PoolCreationFee {
		poolCreationFeeDecCoins = poolCreationFeeDecCoins.Add(sdk.NewDecCoin(coin.Denom, coin.Amount))
	}

	// TODO: should be moved to balancer package
	tests := []struct {
		name        string
		msg         balancer.MsgCreateBalancerPool
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
			name: "create a pool with negative spread factor",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(-1, 2),
				ExitFee: defaultZeroExitFee,
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
			name: "create a pool with non zero exit fee",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			}, defaultPoolAssets, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create the pool with empty PoolAssets",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: defaultZeroExitFee,
			}, []balancer.PoolAsset{}, defaultFutureGovernor),
			emptySender: false,
			expectPass:  false,
		}, {
			name: "create the pool with 0 weighted PoolAsset",
			msg: balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: defaultZeroExitFee,
			}, []balancer.PoolAsset{{
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
				ExitFee: defaultZeroExitFee,
			}, []balancer.PoolAsset{{
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
				ExitFee: defaultZeroExitFee,
			}, []balancer.PoolAsset{{
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
				ExitFee: defaultZeroExitFee,
			}, []balancer.PoolAsset{{
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
				ExitFee: defaultZeroExitFee,
			}, []balancer.PoolAsset{{
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
		s.SetupTest()
		gammKeeper := s.App.GAMMKeeper
		poolmanagerKeeper := s.App.PoolManagerKeeper
		distributionKeeper := s.App.DistrKeeper
		bankKeeper := s.App.BankKeeper

		// fund sender test account
		sender, err := sdk.AccAddressFromBech32(test.msg.Sender)
		s.Require().NoError(err, "test: %v", test.name)
		if !test.emptySender {
			s.FundAcc(sender, defaultAcctFunds)
		}

		// note starting balances for community fee pool and pool creator account
		feePoolBalBeforeNewPool := distributionKeeper.GetFeePoolCommunityCoins(s.Ctx)
		senderBalBeforeNewPool := bankKeeper.GetAllBalances(s.Ctx, sender)

		// attempt to create a pool with the given NewMsgCreateBalancerPool message
		poolId, err := poolmanagerKeeper.CreatePool(s.Ctx, test.msg)

		if test.expectPass {
			s.Require().NoError(err, "test: %v", test.name)

			// check to make sure new pool exists and has minted the correct number of pool shares
			pool, err := gammKeeper.GetPoolAndPoke(s.Ctx, poolId)
			s.Require().NoError(err, "test: %v", test.name)
			s.Require().Equal(types.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
				fmt.Sprintf("share token should be minted as %s initially", types.InitPoolSharesSupply.String()),
			)

			// make sure pool creation fee is correctly sent to community pool
			feePool := distributionKeeper.GetFeePoolCommunityCoins(s.Ctx)
			s.Require().Equal(feePool, feePoolBalBeforeNewPool.Add(poolCreationFeeDecCoins...))

			// get expected tokens in new pool and corresponding pool shares
			expectedPoolTokens := sdk.Coins{}
			for _, asset := range test.msg.GetPoolAssets() {
				expectedPoolTokens = expectedPoolTokens.Add(asset.Token)
			}
			expectedPoolShares := sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)

			// make sure sender's balance is updated correctly
			senderBal := bankKeeper.GetAllBalances(s.Ctx, sender)
			expectedSenderBal := senderBalBeforeNewPool.Sub(params.PoolCreationFee).Sub(expectedPoolTokens).Add(expectedPoolShares)
			s.Require().Equal(senderBal.String(), expectedSenderBal.String())

			// check pool's liquidity is correctly increased
			liquidity, err := gammKeeper.GetTotalLiquidity(s.Ctx)
			s.Require().NoError(err, "test: %v", test.name)
			s.Require().Equal(expectedPoolTokens.String(), liquidity.String())
		} else {
			s.Require().Error(err, "test: %v", test.name)
		}
	}
}

func (s *KeeperTestSuite) TestInitializePool() {
	testAccount := s.TestAccs[0]

	tests := []struct {
		name       string
		createPool func() poolmanagertypes.PoolI
		expectPass bool
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
				require.NoError(s.T(), err)
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
				require.NoError(s.T(), err)
				return &stableswapPool
			},
			expectPass: true,
		},
		{
			name: "initialize a CL pool which cause error",
			createPool: func() poolmanagertypes.PoolI {
				return s.PrepareConcentratedPool()
			},
			expectPass: false,
		},
		{
			name: "initialize pool with non-zero exit fee",
			createPool: func() poolmanagertypes.PoolI {
				balancerPool, err := balancer.NewBalancerPool(
					defaultPoolId,
					balancer.PoolParams{
						SwapFee: defaultSpreadFactor,
						ExitFee: sdk.NewDecWithPrec(5, 1),
					},
					defaultPoolAssets,
					"",
					time.Now(),
				)
				require.NoError(s.T(), err)
				return &balancerPool
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			s.SetupTest()

			gammKeeper := s.App.GAMMKeeper
			bankKeeper := s.App.BankKeeper
			poolIncentivesKeeper := s.App.PoolIncentivesKeeper

			// sender test account
			sender := testAccount
			senderBalBeforeNewPool := bankKeeper.GetAllBalances(s.Ctx, sender)

			// initializePool with a poolI
			// initializePool shoould be called by pool manager in practice.
			// We set pool route here to make sure hooks from InitializePool do not break
			s.App.PoolManagerKeeper.SetPoolRoute(s.Ctx, defaultPoolId, poolmanagertypes.Balancer)
			err := gammKeeper.InitializePool(s.Ctx, test.createPool(), sender)

			if test.expectPass {
				s.Require().NoError(err, "test: %v", test.name)

				// check to make sure new pool exists and has minted the correct number of pool shares
				pool, err := gammKeeper.GetPoolAndPoke(s.Ctx, defaultPoolId)
				s.Require().NoError(err, "test: %v", test.name)
				s.Require().Equal(types.InitPoolSharesSupply.String(), pool.GetTotalShares().String(),
					fmt.Sprintf("share token should be minted as %s initially", types.InitPoolSharesSupply),
				)

				// check to make sure user user balance increase correct number of pool shares
				s.Require().Equal(
					senderBalBeforeNewPool.Add(sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)),
					bankKeeper.GetAllBalances(s.Ctx, sender),
				)

				// get expected tokens in new pool and corresponding pool shares
				expectedPoolTokens := sdk.NewCoins()
				for _, asset := range pool.GetTotalPoolLiquidity(s.Ctx) {
					expectedPoolTokens = expectedPoolTokens.Add(asset)
				}
				expectedPoolShares := sdk.NewCoin(types.GetPoolShareDenom(pool.GetId()), types.InitPoolSharesSupply)

				// make sure expected pool tokens and expected pool shares matches the actual tokens and shares in the pool
				s.Require().Equal(expectedPoolTokens.String(), pool.GetTotalPoolLiquidity(s.Ctx).String())
				s.Require().Equal(expectedPoolShares.Amount.String(), pool.GetTotalShares().String())

				// check pool metadata
				poolShareBaseDenom := types.GetPoolShareDenom(pool.GetId())
				poolShareDisplayDenom := fmt.Sprintf("GAMM-%d", pool.GetId())
				metadata, found := bankKeeper.GetDenomMetaData(s.Ctx, poolShareBaseDenom)
				s.Require().Equal(found, true, fmt.Sprintf("Pool share denom %s is not set", poolShareDisplayDenom))
				s.Require().Equal(metadata.Base, poolShareBaseDenom, fmt.Sprintf("Pool share base denom %s is not correctly set", poolShareBaseDenom))
				s.Require().Equal(metadata.Display, poolShareDisplayDenom, fmt.Sprintf("Pool share display denom %s is not correctly set", poolShareDisplayDenom))
				s.Require().Equal(metadata.DenomUnits[0].Denom, poolShareBaseDenom)
				s.Require().Equal(metadata.DenomUnits[0].Exponent, uint32(0x0))
				s.Require().Equal(metadata.DenomUnits[0].Aliases, []string{
					"attopoolshare",
				})
				s.Require().Equal(metadata.DenomUnits[1].Denom, poolShareDisplayDenom)
				s.Require().Equal(metadata.DenomUnits[1].Exponent, uint32(types.OneShareExponent))
				s.Require().Equal(metadata.DenomUnits[1].Aliases, []string(nil))

				// check AfterPoolCreated hook
				for _, lockableDuration := range poolIncentivesKeeper.GetLockableDurations(s.Ctx) {
					gaugeId, err := poolIncentivesKeeper.GetPoolGaugeId(s.Ctx, defaultPoolId, lockableDuration)
					s.Require().NoError(err, "test: %v", test.name)

					poolIdFromPoolIncentives, err := poolIncentivesKeeper.GetPoolIdFromGaugeId(s.Ctx, gaugeId, lockableDuration)
					s.Require().NoError(err, "test: %v", test.name)
					s.Require().Equal(poolIdFromPoolIncentives, defaultPoolId)
				}
			} else {
				s.Require().Error(err, "test: %v", test.name)
			}
		})
	}
}

// This test creates several pools, and tests that:
// the condition is in a case where the balancer return value returns an overflowing value
// the SpotPrice query does not
func (s *KeeperTestSuite) TestSpotPriceOverflow() {
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
		s.Run(name, func() {
			poolId := s.PrepareBalancerPoolWithCoinsAndWeights(tc.poolLiquidity, tc.poolWeights)
			pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
			s.Require().NoError(err)
			var poolSpotPrice sdk.Dec
			var poolErr error
			osmoassert.ConditionalPanic(s.T(), tc.panics, func() {
				poolSpotPrice, poolErr = pool.SpotPrice(s.Ctx, tc.baseAssetDenom, tc.quoteAssetDenom)
			})
			keeperSpotPrice, keeperErr := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, tc.quoteAssetDenom, tc.baseAssetDenom)
			if tc.overflows {
				s.Require().NoError(poolErr)
				s.Require().ErrorIs(keeperErr, types.ErrSpotPriceOverflow)
				s.Require().Error(keeperErr)
				s.Require().Equal(types.MaxSpotPrice, keeperSpotPrice)
			} else if tc.panics {
				s.Require().ErrorIs(keeperErr, types.ErrSpotPriceInternal)
				s.Require().Error(keeperErr)
				s.Require().Equal(sdk.Dec{}, keeperSpotPrice)
			} else {
				s.Require().NoError(poolErr)
				s.Require().NoError(keeperErr)
				s.Require().Equal(poolSpotPrice, keeperSpotPrice)
			}
		})
	}
}

// TODO: Add more edge cases around TokenInMaxs not containing every token in pool.
func (s *KeeperTestSuite) TestJoinPoolNoSwap() {
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
			txSender:        s.TestAccs[1],
			sharesRequested: types.OneShare.MulRaw(50),
			tokenInMaxs:     sdk.Coins{},
			expectPass:      true,
		},
		{
			name:            "join no swap with zero shares requested",
			txSender:        s.TestAccs[1],
			sharesRequested: sdk.NewInt(0),
			tokenInMaxs:     sdk.Coins{},
			expectPass:      false,
		},
		{
			name:            "join no swap with negative shares requested",
			txSender:        s.TestAccs[1],
			sharesRequested: sdk.NewInt(-1),
			tokenInMaxs:     sdk.Coins{},
			expectPass:      false,
		},
		{
			name:            "join no swap with insufficient funds",
			txSender:        s.TestAccs[1],
			sharesRequested: sdk.NewInt(-1),
			tokenInMaxs: sdk.Coins{
				sdk.NewCoin("bar", sdk.NewInt(4999)), sdk.NewCoin("foo", sdk.NewInt(4999)),
			},
			expectPass: false,
		},
		{
			name:            "join no swap with exact tokenInMaxs",
			txSender:        s.TestAccs[1],
			sharesRequested: types.OneShare.MulRaw(50),
			tokenInMaxs: sdk.Coins{
				fiveKFooAndBar[0], fiveKFooAndBar[1],
			},
			expectPass: true,
		},
		{
			name:            "join no swap with arbitrary extra token in tokenInMaxs",
			txSender:        s.TestAccs[1],
			sharesRequested: types.OneShare.MulRaw(50),
			tokenInMaxs: sdk.Coins{
				fiveKFooAndBar[0], fiveKFooAndBar[1], sdk.NewCoin("baz", sdk.NewInt(5000)),
			},
			expectPass: false,
		},
		{
			name:            "join no swap with TokenInMaxs not containing every token in pool",
			txSender:        s.TestAccs[1],
			sharesRequested: types.OneShare.MulRaw(50),
			tokenInMaxs: sdk.Coins{
				fiveKFooAndBar[0],
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		ctx := s.Ctx
		gammKeeper := s.App.GAMMKeeper
		poolmanagerKeeper := s.App.PoolManagerKeeper
		bankKeeper := s.App.BankKeeper
		testAccount := s.TestAccs[0]

		// Mint some assets to the accounts.
		s.FundAcc(testAccount, defaultAcctFunds)

		// Create the pool at first
		msg := balancer.NewMsgCreateBalancerPool(testAccount, balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: defaultZeroExitFee,
		}, defaultPoolAssets, defaultFutureGovernor)
		poolId, err := poolmanagerKeeper.CreatePool(s.Ctx, msg)
		s.Require().NoError(err, "test: %v", test.name)

		s.FundAcc(test.txSender, defaultAcctFunds)

		balancesBefore := bankKeeper.GetAllBalances(s.Ctx, test.txSender)
		_, _, err = gammKeeper.JoinPoolNoSwap(s.Ctx, test.txSender, poolId, test.sharesRequested, test.tokenInMaxs)

		if test.expectPass {
			s.Require().NoError(err, "test: %v", test.name)
			s.Require().Equal(test.sharesRequested.String(), bankKeeper.GetBalance(s.Ctx, test.txSender, "gamm/pool/1").Amount.String())
			balancesAfter := bankKeeper.GetAllBalances(s.Ctx, test.txSender)
			deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
			// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100000000gamm/pool/1.
			// Thus, to get the 50*OneShare gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be provided.
			s.Require().Equal("5000", deltaBalances.AmountOf("foo").String())
			s.Require().Equal("5000", deltaBalances.AmountOf("bar").String())

			liquidity, err := gammKeeper.GetTotalLiquidity(s.Ctx)
			s.Require().NoError(err, "test: %v", test.name)
			s.Require().Equal("15000bar,15000foo", liquidity.String())

			s.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, 1)
		} else {
			s.Require().Error(err, "test: %v", test.name)

			s.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, 0)
		}
	}
}

func (s *KeeperTestSuite) TestExitPool() {
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
			txSender:     s.TestAccs[0],
			sharesIn:     types.OneShare.MulRaw(50),
			tokenOutMins: sdk.Coins{},
			emptySender:  true,
			expectPass:   false,
		},
		{
			name:         "exit half pool with correct pool share balance",
			txSender:     s.TestAccs[0],
			sharesIn:     types.OneShare.MulRaw(50),
			tokenOutMins: sdk.Coins{},
			emptySender:  false,
			expectPass:   true,
		},
		{
			name:         "attempt exit pool requesting 0 share amount",
			txSender:     s.TestAccs[0],
			sharesIn:     sdk.NewInt(0),
			tokenOutMins: sdk.Coins{},
			emptySender:  false,
			expectPass:   false,
		},
		{
			name:         "attempt exit pool requesting negative share amount",
			txSender:     s.TestAccs[0],
			sharesIn:     sdk.NewInt(-1),
			tokenOutMins: sdk.Coins{},
			emptySender:  false,
			expectPass:   false,
		},
		{
			name:     "attempt exit pool with tokenOutMins above actual output",
			txSender: s.TestAccs[0],
			sharesIn: types.OneShare.MulRaw(50),
			tokenOutMins: sdk.Coins{
				sdk.NewCoin("foo", sdk.NewInt(5001)),
			},
			emptySender: false,
			expectPass:  false,
		},
		{
			name:     "attempt exit pool requesting tokenOutMins at exactly the actual output",
			txSender: s.TestAccs[0],
			sharesIn: types.OneShare.MulRaw(50),
			tokenOutMins: sdk.Coins{
				fiveKFooAndBar[1],
			},
			emptySender: false,
			expectPass:  true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			ctx := s.Ctx

			gammKeeper := s.App.GAMMKeeper
			bankKeeper := s.App.BankKeeper
			poolmanagerKeeper := s.App.PoolManagerKeeper

			// Mint assets to the pool creator
			s.FundAcc(test.txSender, defaultAcctFunds)

			// Create the pool at first
			msg := balancer.NewMsgCreateBalancerPool(test.txSender, balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			}, defaultPoolAssets, defaultFutureGovernor)
			poolId, err := poolmanagerKeeper.CreatePool(ctx, msg)
			s.Require().NoError(err)

			// If we are testing insufficient pool share balances, switch tx sender from pool creator to empty account
			if test.emptySender {
				test.txSender = s.TestAccs[1]
			}

			balancesBefore := bankKeeper.GetAllBalances(s.Ctx, test.txSender)
			_, err = gammKeeper.ExitPool(ctx, test.txSender, poolId, test.sharesIn, test.tokenOutMins)

			if test.expectPass {
				s.Require().NoError(err, "test: %v", test.name)
				s.Require().Equal(test.sharesIn.String(), bankKeeper.GetBalance(s.Ctx, test.txSender, "gamm/pool/1").Amount.String())
				balancesAfter := bankKeeper.GetAllBalances(s.Ctx, test.txSender)
				deltaBalances, _ := balancesBefore.SafeSub(balancesAfter)
				// The pool was created with the 10000foo, 10000bar, and the pool share was minted as 100*OneShare gamm/pool/1.
				// Thus, to refund the 50*OneShare gamm/pool/1, (10000foo, 10000bar) * (1 / 2) balances should be refunded.
				s.Require().Equal("-5000", deltaBalances.AmountOf("foo").String())
				s.Require().Equal("-5000", deltaBalances.AmountOf("bar").String())

				liquidity, err := gammKeeper.GetTotalLiquidity(ctx)
				s.Require().NoError(err)
				s.Require().Equal("5000bar,5000foo", liquidity.String())

				s.AssertEventEmitted(ctx, types.TypeEvtPoolExited, 1)
			} else {
				s.Require().Error(err, "test: %v", test.name)
				s.AssertEventEmitted(ctx, types.TypeEvtPoolExited, 0)
			}
		})
	}
}

// TestJoinPoolExitPool_InverseRelationship tests that joining pool and exiting pool
// guarantees same amount in and out
func (s *KeeperTestSuite) TestJoinPoolExitPool_InverseRelationship() {
	testCases := []struct {
		name             string
		pool             balancer.MsgCreateBalancerPool
		joinPoolShareAmt sdk.Int
	}{
		{
			name: "pool with same token ratio",
			pool: balancer.NewMsgCreateBalancerPool(nil, balancer.PoolParams{
				SwapFee: sdk.ZeroDec(),
				ExitFee: sdk.ZeroDec(),
			}, []balancer.PoolAsset{
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
			}, []balancer.PoolAsset{
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
		s.SetupTest()

		s.Run(tc.name, func() {
			ctx := s.Ctx
			gammKeeper := s.App.GAMMKeeper
			poolmanagerKeeper := s.App.PoolManagerKeeper

			for _, acc := range s.TestAccs {
				s.FundAcc(acc, defaultAcctFunds)
			}

			createPoolAcc := s.TestAccs[0]
			joinPoolAcc := s.TestAccs[1]

			// test account is set on every test case iteration, we need to manually update address for pool creator
			tc.pool.Sender = createPoolAcc.String()

			poolId, err := poolmanagerKeeper.CreatePool(ctx, tc.pool)
			s.Require().NoError(err)

			balanceBeforeJoin := s.App.BankKeeper.GetAllBalances(ctx, joinPoolAcc)

			_, _, err = gammKeeper.JoinPoolNoSwap(ctx, joinPoolAcc, poolId, tc.joinPoolShareAmt, sdk.Coins{})
			s.Require().NoError(err)

			s.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, 1)

			_, err = gammKeeper.ExitPool(ctx, joinPoolAcc, poolId, tc.joinPoolShareAmt, sdk.Coins{})
			s.Require().NoError(err)

			s.AssertEventEmitted(ctx, types.TypeEvtPoolExited, 1)

			balanceAfterExit := s.App.BankKeeper.GetAllBalances(ctx, joinPoolAcc)
			deltaBalance, _ := balanceBeforeJoin.SafeSub(balanceAfterExit)

			// due to rounding, `balanceBeforeJoin` and `balanceAfterExit` have neglectable difference
			// coming from rounding in exitPool.Here we test if the difference is within rounding tolerance range
			roundingToleranceCoins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1)), sdk.NewCoin("bar", sdk.NewInt(1)))
			s.Require().True(deltaBalance.AmountOf("foo").LTE(roundingToleranceCoins.AmountOf("foo")))
			s.Require().True(deltaBalance.AmountOf("bar").LTE(roundingToleranceCoins.AmountOf("bar")))
		})
	}
}

func (s *KeeperTestSuite) TestActiveBalancerPool() {
	type testCase struct {
		blockTime  time.Time
		expectPass bool
	}

	testCases := []testCase{
		{time.Unix(1000, 0), true},
	}

	for _, tc := range testCases {
		s.Run("", func() {
			s.SetupTest()

			ctx := s.Ctx
			gammKeeper := s.App.GAMMKeeper
			testAccount := s.TestAccs[0]

			s.FundAcc(testAccount, defaultAcctFunds)

			// Create the pool at first
			poolId := s.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})
			ctx = ctx.WithBlockTime(tc.blockTime)

			// uneffected by start time
			_, _, err := gammKeeper.JoinPoolNoSwap(ctx, testAccount, poolId, types.OneShare.MulRaw(50), sdk.Coins{})
			s.Require().NoError(err)

			s.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, 1)

			_, err = gammKeeper.ExitPool(ctx, testAccount, poolId, types.InitPoolSharesSupply.QuoRaw(2), sdk.Coins{})
			s.Require().NoError(err)

			s.AssertEventEmitted(ctx, types.TypeEvtPoolExited, 1)

			foocoin := sdk.NewCoin("foo", sdk.NewInt(10))
			foocoins := sdk.Coins{foocoin}

			if tc.expectPass {
				_, err = gammKeeper.JoinSwapExactAmountIn(ctx, testAccount, poolId, foocoins, sdk.ZeroInt())
				s.Require().NoError(err)
				_, err = gammKeeper.JoinSwapShareAmountOut(ctx, testAccount, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				s.Require().NoError(err)
				_, err = gammKeeper.ExitSwapShareAmountIn(ctx, testAccount, poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				s.Require().NoError(err)
				_, err = gammKeeper.ExitSwapExactAmountOut(ctx, testAccount, poolId, foocoin, sdk.NewInt(1000000000000000000))
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				_, err = gammKeeper.JoinSwapShareAmountOut(ctx, testAccount, poolId, "foo", types.OneShare.MulRaw(10), sdk.NewInt(1000000000000000000))
				s.Require().Error(err)
				_, err = gammKeeper.ExitSwapShareAmountIn(ctx, testAccount, poolId, "foo", types.OneShare.MulRaw(10), sdk.ZeroInt())
				s.Require().Error(err)
				_, err = gammKeeper.ExitSwapExactAmountOut(ctx, testAccount, poolId, foocoin, sdk.NewInt(1000000000000000000))
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestJoinSwapExactAmountInConsistency() {
	testCases := []struct {
		name              string
		poolSpreadFactor  sdk.Dec
		poolExitFee       sdk.Dec
		tokensIn          sdk.Coins
		shareOutMinAmount sdk.Int
		expectedSharesOut sdk.Int
		tokenOutMinAmount sdk.Int
	}{
		{
			name:              "single coin with zero swap and exit fees",
			poolSpreadFactor:  sdk.ZeroDec(),
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
		// 	name:              "single coin with positive spread factor and zero exit fee",
		// 	poolSpreadFactor:       sdk.NewDecWithPrec(1, 2),
		// 	poolExitFee:       sdk.ZeroDec(),
		// 	tokensIn:          sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(1000000))),
		// 	shareOutMinAmount: sdk.ZeroInt(),
		// 	expectedSharesOut: sdk.NewInt(6226484702880621000),
		// 	tokenOutMinAmount: sdk.ZeroInt(),
		// },
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()
			ctx := s.Ctx
			gammKeeper := s.App.GAMMKeeper
			testAccount := s.TestAccs[0]

			poolID := s.prepareCustomBalancerPool(
				defaultAcctFunds,
				[]balancer.PoolAsset{
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
					SwapFee: tc.poolSpreadFactor,
					ExitFee: tc.poolExitFee,
				},
			)

			shares, err := gammKeeper.JoinSwapExactAmountIn(ctx, testAccount, poolID, tc.tokensIn, tc.shareOutMinAmount)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedSharesOut, shares)

			tokenOutAmt, err := gammKeeper.ExitSwapShareAmountIn(
				ctx,
				testAccount,
				poolID,
				tc.tokensIn[0].Denom,
				shares,
				tc.tokenOutMinAmount,
			)
			s.Require().NoError(err)

			// require swapTokenOutAmt <= (tokenInAmt * (1 - tc.poolSpreadFactor))
			oneMinusSwapFee := sdk.OneDec().Sub(tc.poolSpreadFactor)
			spreadFactorAdjustedAmount := oneMinusSwapFee.MulInt(tc.tokensIn[0].Amount).RoundInt()
			s.Require().True(tokenOutAmt.LTE(spreadFactorAdjustedAmount))

			// require swapTokenOutAmt + 10 > input
			s.Require().True(
				spreadFactorAdjustedAmount.Sub(tokenOutAmt).LTE(sdk.NewInt(10)),
				"expected out amount %s, actual out amount %s",
				spreadFactorAdjustedAmount, tokenOutAmt,
			)
		})
	}
}

func (s *KeeperTestSuite) TestGetPoolDenom() {
	// setup pool with denoms
	s.FundAcc(s.TestAccs[0], defaultAcctFunds)
	poolCreateMsg := balancer.NewMsgCreateBalancerPool(s.TestAccs[0], defaultPoolParams, defaultPoolAssets, defaultFutureGovernor)
	_, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, poolCreateMsg)
	s.Require().NoError(err)

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
		s.Run(tc.desc, func() {
			denoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, tc.poolId)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(denoms, tc.expectDenoms)
			}
		})
	}
}
