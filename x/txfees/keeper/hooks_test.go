package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v31/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v31/x/txfees/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const (
	preSwapDenom            = "foo"
	otherPreSwapDenom       = "bar"
	denomWithNoPool         = "baz"
	denomWithNoProtorevLink = "qux"
)

var defaultPooledAssetAmount = int64(500)

var (
	denomA = apptesting.DefaultTransmuterDenomA
	denomB = apptesting.DefaultTransmuterDenomB
	denomC = apptesting.DefaultTransmuterDenomC

	oneHundred   = osmomath.NewInt(100)
	twoHundred   = osmomath.NewInt(200)
	threeHundred = osmomath.NewInt(300)

	defaultTakerFeeShareAgreements = []poolmanagertypes.TakerFeeShareAgreement{
		{
			Denom:       denomA,
			SkimPercent: osmomath.MustNewDecFromStr("0.01"),
			SkimAddress: "osmo1785depelc44z2ezt7vf30psa9609xt0y28lrtn",
		},
		{
			Denom:       denomB,
			SkimPercent: osmomath.MustNewDecFromStr("0.02"),
			SkimAddress: "osmo1jj6t7xrevz5fhvs5zg5jtpnht2mzv539008uc2",
		},
		{
			Denom:       denomC,
			SkimPercent: osmomath.MustNewDecFromStr("0.03"),
			SkimAddress: "osmo1jermpr9yust7cyhfjme3cr08kt6n8jv6p35l39",
		},
	}

	// Test taker fee distribution with burn
	testOsmoTakerFeeDistributionWithBurn = poolmanagertypes.TakerFeeDistributionPercentage{
		StakingRewards: osmomath.MustNewDecFromStr("0.3"),
		CommunityPool:  osmomath.MustNewDecFromStr("0.0"),
		Burn:           osmomath.MustNewDecFromStr("0.7"),
	}
)

func (s *KeeperTestSuite) preparePool(denom string) (poolID uint64, pool poolmanagertypes.PoolI) {
	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	poolID = s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(baseDenom, defaultPooledAssetAmount),
		sdk.NewInt64Coin(denom, defaultPooledAssetAmount),
	)
	pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolID)
	s.Require().NoError(err)
	err = s.ExecuteUpgradeFeeTokenProposal(denom, poolID)
	s.Require().NoError(err)
	return poolID, pool
}

func (s *KeeperTestSuite) TestTxFeesAfterEpochEnd() {
	s.SetupTest(false)
	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)

	// create pools for three separate fee tokens
	uion := "uion"
	_, uionPool := s.preparePool(uion)
	atom := "atom"
	_, atomPool := s.preparePool(atom)
	ust := "ust"
	_, ustPool := s.preparePool(ust)

	tests := []struct {
		name         string
		coins        sdk.Coins
		baseDenom    string
		denoms       []string
		poolTypes    []poolmanagertypes.PoolI
		spreadFactor osmomath.Dec
		expectPass   bool
	}{
		{
			name:         "One non-osmo fee token (uion): TxFees AfterEpochEnd",
			coins:        sdk.Coins{sdk.NewInt64Coin(uion, 10)},
			baseDenom:    baseDenom,
			denoms:       []string{uion},
			poolTypes:    []poolmanagertypes.PoolI{uionPool},
			spreadFactor: osmomath.MustNewDecFromStr("0"),
		},
		{
			name:         "Multiple non-osmo fee token: TxFees AfterEpochEnd",
			coins:        sdk.Coins{sdk.NewInt64Coin(atom, 20), sdk.NewInt64Coin(ust, 30)},
			baseDenom:    baseDenom,
			denoms:       []string{atom, ust},
			poolTypes:    []poolmanagertypes.PoolI{atomPool, ustPool},
			spreadFactor: osmomath.MustNewDecFromStr("0"),
		},
	}

	finalOutputAmount := osmomath.NewInt(0)

	for _, tc := range tests {
		tc := tc

		s.Run(tc.name, func() {
			for i, coin := range tc.coins {
				// Get the output amount in osmo denom
				pool, ok := tc.poolTypes[i].(gammtypes.CFMMPoolI)
				s.Require().True(ok)

				expectedOutput, err := pool.CalcOutAmtGivenIn(s.Ctx,
					sdk.Coins{sdk.Coin{Denom: tc.denoms[i], Amount: coin.Amount}},
					tc.baseDenom,
					tc.spreadFactor)
				s.NoError(err)
				// sanity check for the expectedAmount
				s.True(coin.Amount.GTE(expectedOutput.Amount))

				finalOutputAmount = finalOutputAmount.Add(expectedOutput.Amount)

				// Deposit some fee amount (non-native-denom) to the fee module account
				_, _, addr0 := testdata.KeyTestPubAddr()
				err = testutil.FundAccount(s.Ctx, s.App.BankKeeper, addr0, sdk.Coins{coin})
				s.NoError(err)
				err = s.App.BankKeeper.SendCoinsFromAccountToModule(s.Ctx, addr0, types.NonNativeTxFeeCollectorName, sdk.Coins{coin})
				s.NoError(err)
			}

			// checks the balance of the non-native denom in module account
			moduleAddrNonNativeFee := s.App.AccountKeeper.GetModuleAddress(types.NonNativeTxFeeCollectorName)
			s.Equal(s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddrNonNativeFee), tc.coins)

			// End of epoch, so all the non-osmo fee amount should be swapped to osmo and transferred to buffer, then distributed
			// Use "day" epoch identifier to trigger smoothing buffer distribution
			futureCtx := s.Ctx.WithBlockTime(time.Now().Add(time.Minute))
			err := s.App.TxFeesKeeper.AfterEpochEnd(futureCtx, "day", int64(1))
			s.NoError(err)

			// Get smoothing factor from params
			poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
			smoothingFactor := poolManagerParams.TakerFeeParams.DailyStakingRewardsSmoothingFactor

			// check the balance of the native-basedenom in fee collector
			moduleAddrFee := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			moduleBaseDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddrFee, tc.baseDenom)

			// check the balance in the smoothing buffer
			bufferAddr := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeStakingRewardsBuffer)
			bufferBalance := s.App.BankKeeper.GetBalance(s.Ctx, bufferAddr, tc.baseDenom)

			// non-osmos module account should be empty as all the funds should be transferred to buffer or fee collector
			s.Empty(s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddrNonNativeFee))

			// With smoothing: total = (buffer + distributed) should equal finalOutputAmount
			// distributed = finalOutputAmount / smoothingFactor
			// buffer = finalOutputAmount - distributed
			expectedDistributed := finalOutputAmount.QuoRaw(int64(smoothingFactor))
			expectedBuffer := finalOutputAmount.Sub(expectedDistributed)

			s.Equal(expectedDistributed.String(), moduleBaseDenomBalance.Amount.String(), "Fee collector should receive 1/smoothing_factor of swapped amount")
			s.Equal(expectedBuffer.String(), bufferBalance.Amount.String(), "Buffer should contain remaining amount")
		})
	}
}

func (s *KeeperTestSuite) TestSwapNonNativeFeeToDenom() {
	s.Setup()

	var (
		defaultTxFeesDenom, _  = s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
		defaultPoolCoins       = sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(100)), sdk.NewCoin(defaultTxFeesDenom, osmomath.NewInt(100)))
		balanceToSwapFoo       = sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(50)))
		balanceToSwapBaseDenom = sdk.NewCoins(sdk.NewCoin(defaultTxFeesDenom, osmomath.NewInt(50)))
	)

	tests := []struct {
		name                string
		denomToSwapTo       string
		poolCoins           sdk.Coins
		preFundCoins        sdk.Coins
		feeCollectorAddress sdk.AccAddress
		expectPass          bool
	}{
		{
			name:          "happy path",
			denomToSwapTo: balanceToSwapBaseDenom[0].Denom,
			poolCoins:     defaultPoolCoins,
			preFundCoins:  balanceToSwapFoo,
		},
		{
			name:          "same denom to swap to",
			denomToSwapTo: balanceToSwapFoo[0].Denom,
			poolCoins:     defaultPoolCoins,
			preFundCoins:  balanceToSwapFoo,
		},

		// TODO: add more test cases
		// - pool does not exist for denom pair but protorev has it set for a pair
		// - error in swap due to no liquidity
		// - same denom in balance as denomToSwapTo
		// - no pool exists for denom pair in protorev
		// - many tokens in balance, some get swapped, others don't
		// - different order of denoms in SetPoolForDenomPair()
	}

	for _, tc := range tests {
		tc := tc

		s.Run(tc.name, func() {
			s.Setup()

			// Sets up account with no balance
			testAccount := apptesting.CreateRandomAccounts(1)[0]

			// Create a pool to be swapped against
			poolId := s.PrepareConcentratedPoolWithCoins(tc.poolCoins[0].Denom, tc.poolCoins[1].Denom).GetId()

			s.FundAcc(s.TestAccs[0], tc.poolCoins)
			_, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, poolId, s.TestAccs[0], tc.poolCoins)
			s.Require().NoError(err)

			// Set the pool for the denom pair
			s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, tc.poolCoins[0].Denom, tc.poolCoins[1].Denom, poolId)

			// Fund the account with the preFundCoins
			s.FundAcc(testAccount, tc.preFundCoins)

			s.App.TxFeesKeeper.SwapNonNativeFeeToDenom(s.Ctx, tc.denomToSwapTo, testAccount)

			// Check balance
			balances := s.App.BankKeeper.GetAllBalances(s.Ctx, testAccount)
			s.Require().Len(balances, 1)

			// Check that the denomToSwapTo is the denom of the balance
			s.Require().Equal(balances[0].Denom, tc.denomToSwapTo)
		})
	}
}

func (s *KeeperTestSuite) TestSwapNonNativeFeeToDenom_SimpleCases() {
	s.Setup()

	var (
		defaultTxFeesDenom, _      = s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
		defaultPoolCoins           = sdk.NewCoins(sdk.NewCoin(preSwapDenom, osmomath.NewInt(100)), sdk.NewCoin(defaultTxFeesDenom, osmomath.NewInt(100)))
		defaultBalanceToSwap       = sdk.NewCoins(sdk.NewCoin(preSwapDenom, osmomath.NewInt(100)))
		defaultProtorevLinkDenoms  = []string{preSwapDenom, defaultTxFeesDenom}
		reversedProtorevLinkDenoms = []string{defaultTxFeesDenom, preSwapDenom}
	)

	validateFinalBalance := func(expectedEndBalanceDenoms []string, testAccount sdk.AccAddress) {
		// Check balance
		balances := s.App.BankKeeper.GetAllBalances(s.Ctx, testAccount)

		// Validate that the final denoms in the balance are as expected per test configuration.
		// On success, swapped to denomToSwapTo. On failure, kept as is.
		s.Require().Len(balances, len(expectedEndBalanceDenoms))
		for i, actualDenomInBalance := range expectedEndBalanceDenoms {
			s.Require().Contains(expectedEndBalanceDenoms[i], actualDenomInBalance)
		}
	}

	// tests SwapNonNativeFeeToDenom success and silent error cases
	// where there is only one token in the initial balance.
	s.Run("simple cases", func() {
		tests := []struct {
			name                string
			denomToSwapTo       string
			poolCoins           sdk.Coins
			preFundCoins        sdk.Coins
			protoRevLinkDenoms  []string
			feeCollectorAddress sdk.AccAddress

			doNotCreatePool   bool
			doNotAddLiquidity bool

			expectedEndBalanceDenoms []string

			expectPass bool
		}{
			{
				name:               "happy path",
				denomToSwapTo:      defaultTxFeesDenom,
				poolCoins:          defaultPoolCoins,
				preFundCoins:       defaultBalanceToSwap,
				protoRevLinkDenoms: defaultProtorevLinkDenoms,

				// Swap happened.
				expectedEndBalanceDenoms: []string{defaultTxFeesDenom},
			},
			{
				name:               "happy path with protorev link denoms reversed",
				denomToSwapTo:      defaultTxFeesDenom,
				poolCoins:          defaultPoolCoins,
				preFundCoins:       defaultBalanceToSwap,
				protoRevLinkDenoms: reversedProtorevLinkDenoms,

				// Swap happened.
				expectedEndBalanceDenoms: []string{defaultTxFeesDenom},
			},
			{
				name:               "error in swap due to pool not created but pool id to denom pair link set. No swap happens and no error/panic",
				denomToSwapTo:      defaultTxFeesDenom,
				poolCoins:          defaultPoolCoins,
				preFundCoins:       defaultBalanceToSwap,
				protoRevLinkDenoms: defaultProtorevLinkDenoms,

				doNotCreatePool: true,

				// Swap did not happen.
				expectedEndBalanceDenoms: []string{preSwapDenom},
			},
			{
				name:               "error in swap due to no liquidity. No swap happens and no error/panic",
				denomToSwapTo:      defaultTxFeesDenom,
				poolCoins:          defaultPoolCoins,
				preFundCoins:       defaultBalanceToSwap,
				protoRevLinkDenoms: defaultProtorevLinkDenoms,

				doNotAddLiquidity: true,

				// Swap did not happen.
				expectedEndBalanceDenoms: []string{preSwapDenom},
			},
			{
				name:               "same denom in balance as denomToSwapTo - no-op",
				denomToSwapTo:      defaultTxFeesDenom,
				poolCoins:          defaultPoolCoins,
				preFundCoins:       osmoutils.FilterDenoms(defaultPoolCoins, []string{defaultTxFeesDenom}),
				protoRevLinkDenoms: defaultProtorevLinkDenoms,

				// Swap did not happen but denomToSwapTo was already in balance.
				expectedEndBalanceDenoms: []string{defaultTxFeesDenom},
			},
			{
				name:          "no pool exists for denom pair in protorev - no-op",
				denomToSwapTo: defaultTxFeesDenom,
				poolCoins:     defaultPoolCoins,
				preFundCoins:  defaultBalanceToSwap,
				// Note no protorev link denoms set.

				// Swap did not happen.
				expectedEndBalanceDenoms: []string{preSwapDenom},
			},
		}

		for _, tc := range tests {
			tc := tc

			s.Run(tc.name, func() {

				// Sets up account with no balance
				testAccount := apptesting.CreateRandomAccounts(1)[0]

				poolId := uint64(1)
				if !tc.doNotCreatePool || !tc.doNotAddLiquidity {
					// Create a pool to be swapped against.
					poolId := s.PrepareConcentratedPoolWithCoins(tc.poolCoins[0].Denom, tc.poolCoins[1].Denom).GetId()

					// Add liquidity
					if !tc.doNotAddLiquidity {
						s.FundAcc(s.TestAccs[0], tc.poolCoins)
						_, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, poolId, s.TestAccs[0], tc.poolCoins)
						s.Require().NoError(err)
					}
				}

				// Set the pool for the denom pair per configuration.
				if len(tc.protoRevLinkDenoms) > 0 {
					s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, tc.poolCoins[0].Denom, tc.poolCoins[1].Denom, poolId)
				}

				// Fund the account with the preFundCoins
				s.FundAcc(testAccount, tc.preFundCoins)

				// System under test.
				s.App.TxFeesKeeper.SwapNonNativeFeeToDenom(s.Ctx, tc.denomToSwapTo, testAccount)

				// Check balance
				validateFinalBalance(tc.expectedEndBalanceDenoms, testAccount)
			})
		}
	})

	// tests SwapNonNativeFeeToDenom with multiple tokens
	// in the initial balance. Some of these tokens successfully swap, others do not and are silently skipped.
	// The denomToSwapTo in the initial balance is also silently skipped
	s.Run("multiple tokens", func() {

		denomToSwapTo := defaultTxFeesDenom

		// Prepare coins with all edge cases and success scenarios for swapping to denomToSwapTo.
		preFundCoins := prepareCoinsForSwapToDenomTest(denomToSwapTo)

		// Note: preSwapDenom and otherPreSwapDenom get swapped to denomToSwapTo.
		// Other denoms are silently skipped.
		expectedEndBalanceDenoms := []string{denomToSwapTo, denomWithNoPool, denomWithNoProtorevLink}

		// Prepare 2 test pools and link their denom pairs.
		s.preparePoolsForSwappingToDenom(preSwapDenom, otherPreSwapDenom, denomToSwapTo)

		// Sets up account with no balance
		testAccount := apptesting.CreateRandomAccounts(1)[0]

		// Fund the account with the preFundCoins
		s.FundAcc(testAccount, preFundCoins)

		// System under test.
		s.App.TxFeesKeeper.SwapNonNativeFeeToDenom(s.Ctx, denomToSwapTo, testAccount)

		// Check balance
		validateFinalBalance(expectedEndBalanceDenoms, testAccount)
	})
}

// Invariants tested:
// Staking fee collector for staking rewards.
// - All non-native rewards that have a pool with liquidity and a link set in protorev get swapped to native denom
// - All resulting native tokens get sent to the fee collector.
// - Any non-native tokens that did not have associated pool stay in the balance of staking fee collector.
// Community pool fee collector.
// - All non-native rewards that have a pool with liquidity and a link set in protorev get swapped to a denom configured by parameter.
// - All resulting parameter denom tokens get sent to the community pool.
// - Any non-native tokens that did not have associated pool stay in the balance of community pool fee collector.
func (s *KeeperTestSuite) TestAfterEpochEnd() {
	s.Setup()

	var (
		stakingDenom, _    = s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
		communityPoolDenom = s.App.PoolManagerKeeper.GetParams(s.Ctx).TakerFeeParams.CommunityPoolDenomToSwapNonWhitelistedAssetsTo
	)

	// Prepares the initial balance of the fee collector for swapping to the given denom
	// as well as the pools and links between denoms and pool ids.
	prepareFeeCollector := func(collectorName string, denomToSwapTo string) sdk.AccAddress {
		// Prepare coins with all edge cases and success scenarios for swapping to denomToSwapTo.
		preFundCollectorCoins := prepareCoinsForSwapToDenomTest(denomToSwapTo)
		s.FundModuleAcc(collectorName, preFundCollectorCoins)

		// Prepare pools.
		s.preparePoolsForSwappingToDenom(otherPreSwapDenom, preSwapDenom, denomToSwapTo)

		return s.App.AccountKeeper.GetModuleAddress(collectorName)
	}

	// CommunityPoolDenoms that come in for stakers and not for community pool need to be swapped to stakingDenom, so we need to set the pool for the denom pair.
	stakingDenomCommunityPoolDenomPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(stakingDenom, communityPoolDenom)
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, stakingDenom, communityPoolDenom, stakingDenomCommunityPoolDenomPool.GetId())

	// Prepare the tx fee collector.
	txFeeCollectorAddress := prepareFeeCollector(types.NonNativeTxFeeCollectorName, stakingDenom)

	// Prepare the taker fee collector.
	prepareFeeCollector(types.TakerFeeCollectorName, communityPoolDenom)
	communityPoolCollectorAddress := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeCommunityPoolName)

	// Snapshot the community pool balance before the epoch end.
	communityPoolAddress := s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	communityPoolBalanceBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)

	// Set up taker fee share agreements
	for _, agreement := range defaultTakerFeeShareAgreements {
		s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, agreement)
	}

	// Set accrued values for denom pairs
	s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, "foo", oneHundred)
	s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, "bar", oneHundred)
	s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomB, "foo", twoHundred)
	s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomB, "bar", twoHundred)

	// Fund the taker fee collector
	s.FundModuleAcc(types.TakerFeeCollectorName, sdk.NewCoins(sdk.NewCoin("foo", threeHundred), sdk.NewCoin("bar", threeHundred)))

	// System under test.
	// AfterEpochEnd should not panic or error
	err := s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, "day", 1)
	s.Require().NoError(err)

	// Confirm that tx fee collector only has denomWithNoPool and denomWithNoProtorevLink left in balance.
	txFeeCollectorAddressBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, txFeeCollectorAddress)
	s.Require().Len(txFeeCollectorAddressBalance, 2)
	s.Require().Equal(txFeeCollectorAddressBalance[0].Denom, denomWithNoPool)
	s.Require().Equal(txFeeCollectorAddressBalance[1].Denom, denomWithNoProtorevLink)

	// Confirm that that all native tokens are sent to the fee collector.
	feeCollectorAddress := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	feeCollectorBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, feeCollectorAddress)
	s.Require().Len(feeCollectorBalance, 1)
	s.Require().Equal(feeCollectorBalance[0].Denom, stakingDenom)

	// Confirm that community pool fee collector only denomWithNoProtorevLink left in balance.
	// denomWithNoPool is a whitelisted asset, so it should be directly sent to the community pool.
	communityPoolCollectorBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolCollectorAddress)
	s.Require().Len(communityPoolCollectorBalance, 1)
	s.Require().Equal(communityPoolCollectorBalance[0].Denom, denomWithNoProtorevLink)

	communityPoolBalanceAfter := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)
	communityPoolBalanceDelta := communityPoolBalanceAfter.Sub(communityPoolBalanceBefore...)

	// Confirm that that all tokens that are of the configured denom parameter are sent to the community pool.
	s.Require().Len(communityPoolBalanceDelta, 4)
	s.Require().Equal(communityPoolBalanceDelta[0].Denom, otherPreSwapDenom)
	s.Require().Equal(communityPoolBalanceDelta[1].Denom, denomWithNoPool)
	s.Require().Equal(communityPoolBalanceDelta[2].Denom, preSwapDenom)
	s.Require().Equal(communityPoolBalanceDelta[3].Denom, communityPoolDenom)

	// Check the balances of the skim addresses
	for _, agreement := range defaultTakerFeeShareAgreements {
		skimAddress := sdk.MustAccAddressFromBech32(agreement.SkimAddress)
		skimAddressBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, skimAddress)
		if agreement.Denom == denomA {
			s.Require().Equal(2, skimAddressBalance.Len())
			s.Require().Equal(sdk.NewCoin("bar", oneHundred), skimAddressBalance[0])
			s.Require().Equal(sdk.NewCoin("foo", oneHundred), skimAddressBalance[1])
		} else if agreement.Denom == denomB {
			s.Require().Equal(2, skimAddressBalance.Len())
			s.Require().Equal(sdk.NewCoin("bar", twoHundred), skimAddressBalance[0])
			s.Require().Equal(sdk.NewCoin("foo", twoHundred), skimAddressBalance[1])
		}
	}

	// Confirm that all taker fee share accumulators are cleared
	allTakerFeeShareAccumulators, err := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
	s.Require().NoError(err)
	s.Require().Empty(allTakerFeeShareAccumulators)
}

// preparePoolsForSwappingToDenom sets up two pools:
// 1. nonNativeDenomA and denomToSwapTo
// 2. nonNativeDenomB and denomToSwapTo
//
// For each pool creates a full range position to have some liquidity.
// For each pool, creates a link between its id and the denom pair.
func (s *KeeperTestSuite) preparePoolsForSwappingToDenom(nonNativeDenomA, nonNativeDenomB, denomToSwapTo string) {
	// Create 2 pools:
	// nonNativeDenomA and denomToSwapTo
	// nonNativeDenomB and denomToSwapTo
	poolIdOne := s.PrepareConcentratedPoolWithCoins(denomToSwapTo, nonNativeDenomA).GetId()
	poolIdTwo := s.PrepareConcentratedPoolWithCoins(nonNativeDenomB, denomToSwapTo).GetId()

	// Add liquidity to both pools
	poolOneCoins := sdk.NewCoins(sdk.NewCoin(nonNativeDenomA, osmomath.NewInt(100)), sdk.NewCoin(denomToSwapTo, osmomath.NewInt(100)))
	poolTwoCoins := sdk.NewCoins(sdk.NewCoin(nonNativeDenomB, osmomath.NewInt(100)), sdk.NewCoin(denomToSwapTo, osmomath.NewInt(100)))
	s.FundAcc(s.TestAccs[0], poolOneCoins.Add(poolTwoCoins...))

	_, err := s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, poolIdOne, s.TestAccs[0], poolOneCoins)
	s.Require().NoError(err)

	_, err = s.App.ConcentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, poolIdTwo, s.TestAccs[0], poolTwoCoins)
	s.Require().NoError(err)

	// Set the link where denoms are base - quote ordered.
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, poolOneCoins[0].Denom, poolOneCoins[1].Denom, poolIdOne)

	// Set the link where denoms are quote - base ordered.
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, poolTwoCoins[1].Denom, poolTwoCoins[0].Denom, poolIdTwo)
}

// returns a set of coins that covers all edge cases and success scenarios for swapping to denom.
func prepareCoinsForSwapToDenomTest(swapToDenom string) sdk.Coins {
	return sdk.NewCoins(
		sdk.NewCoin(preSwapDenom, osmomath.NewInt(100)),            // first pool with a link to denom pair in protorev (gets swapped)
		sdk.NewCoin(swapToDenom, osmomath.NewInt(300)),             // swapToDenom (left as is in balance)
		sdk.NewCoin(denomWithNoPool, osmomath.NewInt(400)),         // no pool exists, silently skipped
		sdk.NewCoin(denomWithNoProtorevLink, osmomath.NewInt(500)), // pool with no link to denom pair in protorev, silently skipped
		sdk.NewCoin(otherPreSwapDenom, osmomath.NewInt(600)),       // second pool with a link to denom pair in protorev (gets swapped)
	)
}

// TestClearTakerFeeShareAccumulators tests the functionality of clearing taker fee share accumulators.
// It sets up various scenarios with different taker fee share agreements and accumulators, funds the taker fee collector,
// and then calls the ClearTakerFeeShareAccumulators method to ensure that the accumulators are cleared correctly.
// The test also verifies the balances of the skim addresses to ensure that the correct amounts have been transferred.
func (s *KeeperTestSuite) TestClearTakerFeeShareAccumulators() {
	tests := []struct {
		name                              string
		setupTakerFeeShares               func()
		setupAccumulators                 func()
		fundTakerFeeCollector             func()
		expectedTakerFeeShareAccumulators []poolmanagertypes.TakerFeeSkimAccumulator
		checkSkimAddressBalance           func()
	}{
		{
			name: "one fee share accumulator set",
			setupTakerFeeShares: func() {
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, defaultTakerFeeShareAgreements[0])
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, defaultTakerFeeShareAgreements[1])
			},
			setupAccumulators: func() {
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, "foo", oneHundred)
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, "bar", oneHundred)
			},
			fundTakerFeeCollector: func() {
				s.FundModuleAcc(types.TakerFeeCollectorName, sdk.NewCoins(sdk.NewCoin("foo", oneHundred), sdk.NewCoin("bar", oneHundred)))
			},
			expectedTakerFeeShareAccumulators: []poolmanagertypes.TakerFeeSkimAccumulator{},
			checkSkimAddressBalance: func() {
				// Check balance
				skimAddress := sdk.MustAccAddressFromBech32(defaultTakerFeeShareAgreements[0].SkimAddress)
				skimAddressBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, skimAddress)
				s.Require().Equal(2, skimAddressBalance.Len())
				s.Require().Equal(sdk.NewCoin("bar", oneHundred), skimAddressBalance[0])
				s.Require().Equal(sdk.NewCoin("foo", oneHundred), skimAddressBalance[1])

				skimAddress = sdk.MustAccAddressFromBech32(defaultTakerFeeShareAgreements[1].SkimAddress)
				skimAddressBalance = s.App.BankKeeper.GetAllBalances(s.Ctx, skimAddress)
				s.Require().Equal(0, skimAddressBalance.Len())
			},
		},
		{
			name: "two fee share accumulators set",
			setupTakerFeeShares: func() {
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, defaultTakerFeeShareAgreements[0])
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, defaultTakerFeeShareAgreements[1])
			},
			setupAccumulators: func() {
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, "foo", oneHundred)
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, "bar", oneHundred)
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomB, "foo", twoHundred)
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomB, "bar", twoHundred)
			},
			fundTakerFeeCollector: func() {
				s.FundModuleAcc(types.TakerFeeCollectorName, sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(600)), sdk.NewCoin("bar", osmomath.NewInt(600))))
			},
			expectedTakerFeeShareAccumulators: []poolmanagertypes.TakerFeeSkimAccumulator{},
			checkSkimAddressBalance: func() {
				// Check balance
				skimAddress := sdk.MustAccAddressFromBech32(defaultTakerFeeShareAgreements[0].SkimAddress)
				skimAddressBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, skimAddress)
				s.Require().Equal(2, skimAddressBalance.Len())
				s.Require().Equal(sdk.NewCoin("bar", oneHundred), skimAddressBalance[0])
				s.Require().Equal(sdk.NewCoin("foo", oneHundred), skimAddressBalance[1])

				skimAddress = sdk.MustAccAddressFromBech32(defaultTakerFeeShareAgreements[1].SkimAddress)
				skimAddressBalance = s.App.BankKeeper.GetAllBalances(s.Ctx, skimAddress)
				s.Require().Equal(2, skimAddressBalance.Len())
				s.Require().Equal(sdk.NewCoin("bar", twoHundred), skimAddressBalance[0])
				s.Require().Equal(sdk.NewCoin("foo", twoHundred), skimAddressBalance[1])
			},
		},
		{
			name: "two fee share accumulators set, not enough in balance to send second loop, second loop denom not cleared but first is",
			setupTakerFeeShares: func() {
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, defaultTakerFeeShareAgreements[0])
				s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, defaultTakerFeeShareAgreements[1])
			},
			setupAccumulators: func() {
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, "foo", oneHundred)
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomA, "bar", oneHundred)
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomB, "foo", twoHundred)
				s.App.PoolManagerKeeper.SetTakerFeeShareDenomsToAccruedValue(s.Ctx, denomB, "bar", twoHundred)
			},
			fundTakerFeeCollector: func() {
				s.FundModuleAcc(types.TakerFeeCollectorName, sdk.NewCoins(sdk.NewCoin("foo", oneHundred), sdk.NewCoin("bar", oneHundred)))
			},
			expectedTakerFeeShareAccumulators: []poolmanagertypes.TakerFeeSkimAccumulator{
				{
					Denom:            denomB,
					SkimmedTakerFees: sdk.NewCoins(sdk.NewCoin("foo", twoHundred), sdk.NewCoin("bar", twoHundred)),
				},
			},
			checkSkimAddressBalance: func() {
				// Check balance
				skimAddress := sdk.MustAccAddressFromBech32(defaultTakerFeeShareAgreements[0].SkimAddress)
				skimAddressBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, skimAddress)
				s.Require().Equal(2, skimAddressBalance.Len())
				s.Require().Equal(sdk.NewCoin("bar", oneHundred), skimAddressBalance[0])
				s.Require().Equal(sdk.NewCoin("foo", oneHundred), skimAddressBalance[1])
				skimAddress = sdk.MustAccAddressFromBech32(defaultTakerFeeShareAgreements[1].SkimAddress)
				skimAddressBalance = s.App.BankKeeper.GetAllBalances(s.Ctx, skimAddress)
				s.Require().Equal(0, skimAddressBalance.Len())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			s.Setup()
			tc.setupTakerFeeShares()
			tc.setupAccumulators()
			tc.fundTakerFeeCollector()
			s.App.TxFeesKeeper.ClearTakerFeeShareAccumulators(s.Ctx)
			allTakerFeeShareAccumulators, err := s.App.PoolManagerKeeper.GetAllTakerFeeShareAccumulators(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedTakerFeeShareAccumulators, allTakerFeeShareAccumulators)
			tc.checkSkimAddressBalance()
		})
	}
}

// TestOsmoTakerFeeBurnMechanism tests the burn functionality for OSMO taker fees
func (s *KeeperTestSuite) TestOsmoTakerFeeBurnMechanism() {
	s.Setup()

	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	burnAddress := types.DefaultNullAddress

	// burn address should be the default null address
	s.Require().Equal("osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030", burnAddress.String())

	// Set taker fee distribution with burn percentage
	takerFeeParams := s.App.PoolManagerKeeper.GetParams(s.Ctx).TakerFeeParams
	takerFeeParams.OsmoTakerFeeDistribution = testOsmoTakerFeeDistributionWithBurn

	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.TakerFeeParams = takerFeeParams
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

	// Fund the taker fee collector with OSMO
	initialAmount := osmomath.NewInt(1000000) // 1 OSMO
	s.FundModuleAcc(types.TakerFeeCollectorName, sdk.NewCoins(sdk.NewCoin(baseDenom, initialAmount)))

	// Get initial balances
	stakingRewardsCollectorAddress := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)

	initialBurnBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, baseDenom)
	initialStakingRewardsBalance := s.App.BankKeeper.GetBalance(s.Ctx, stakingRewardsCollectorAddress, baseDenom)

	// System under test - trigger AfterEpochEnd to distribute taker fees
	err := s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, "day", 1)
	s.Require().NoError(err)

	// Calculate expected amounts based on distribution percentages
	expectedBurnAmount := osmomath.NewInt(700000)
	expectedStakingRewardsAmount := osmomath.NewInt(300000)

	// Verify burn address received the correct amount
	finalBurnBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, baseDenom)
	burnAmountReceived := finalBurnBalance.Amount.Sub(initialBurnBalance.Amount)
	s.Require().Equal(expectedBurnAmount, burnAmountReceived, "Burn address should receive 70% of taker fees")

	// Verify staking rewards received the correct amount
	finalStakingRewardsBalance := s.App.BankKeeper.GetBalance(s.Ctx, stakingRewardsCollectorAddress, baseDenom)
	stakingRewardsAmountReceived := finalStakingRewardsBalance.Amount.Sub(initialStakingRewardsBalance.Amount)
	s.Require().Equal(expectedStakingRewardsAmount, stakingRewardsAmountReceived, "Staking rewards should receive 30% of taker fees")

	// Verify taker fee collector is empty
	takerFeeCollectorAddress := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeCollectorName)
	finalTakerFeeCollectorBalance := s.App.BankKeeper.GetBalance(s.Ctx, takerFeeCollectorAddress, baseDenom)
	s.Require().True(finalTakerFeeCollectorBalance.IsZero(), "Taker fee collector should be empty after distribution")

	// Verify total distribution equals initial amount
	totalDistributed := burnAmountReceived.Add(stakingRewardsAmountReceived)
	s.Require().Equal(initialAmount, totalDistributed, "Total distributed amount should equal initial amount")
}

// TestNonOsmoTakerFeeBurnMechanism tests the burn functionality for non-OSMO taker fees
// Non-OSMO tokens should be swapped to OSMO before being sent to the burn address
// Tests multiple denoms and failed swap recovery scenarios
func (s *KeeperTestSuite) TestNonOsmoTakerFeeBurnMechanism() {
	s.Setup()

	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
	burnAddress := types.DefaultNullAddress

	// Use real IBC denoms
	daiDenom := "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"  // DAI
	usdcDenom := "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4" // USDC
	failedSwapDenom := "ibc/FAILEDSWAP"                                                 // This denom will not have a pool, simulating failed swap

	// burn address should be the default null address
	s.Require().Equal("osmo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqmcn030", burnAddress.String())

	var poolAssetAmount = int64(500000000)
	// Set up pools for DAI and USDC to OSMO swapping
	daiPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(baseDenom, poolAssetAmount),
		sdk.NewInt64Coin(daiDenom, poolAssetAmount),
	)

	usdcPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(baseDenom, poolAssetAmount),
		sdk.NewInt64Coin(usdcDenom, poolAssetAmount),
	)

	// Set the pools for the denom pairs in protorev
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, daiDenom, baseDenom, daiPoolId)
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, usdcDenom, baseDenom, usdcPoolId)
	// Note: failedSwapDenom is intentionally not set up to test failed swap scenario

	// Verify pools exist (needed for swapping)
	_, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, daiPoolId)
	s.Require().NoError(err)
	_, err = s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, usdcPoolId)
	s.Require().NoError(err)

	// Set non-OSMO taker fee distribution: staking_rewards=22.5%, burn=52.5%, community_pool=25%
	takerFeeParams := s.App.PoolManagerKeeper.GetParams(s.Ctx).TakerFeeParams
	takerFeeParams.NonOsmoTakerFeeDistribution = poolmanagertypes.TakerFeeDistributionPercentage{
		StakingRewards: osmomath.MustNewDecFromStr("0.225"), // 22.5%
		CommunityPool:  osmomath.MustNewDecFromStr("0.25"),  // 25%
		Burn:           osmomath.MustNewDecFromStr("0.525"), // 52.5%
	}

	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.TakerFeeParams = takerFeeParams
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

	// Fund the taker fee collector with multiple non-OSMO tokens
	initialDaiAmount := osmomath.NewInt(1000000)   // 1 DAI
	initialUsdcAmount := osmomath.NewInt(1000000)  // 1 USDC
	initialFailedAmount := osmomath.NewInt(500000) // 0.5 of failed swap token

	s.FundModuleAcc(types.TakerFeeCollectorName, sdk.NewCoins(
		sdk.NewCoin(daiDenom, initialDaiAmount),
		sdk.NewCoin(usdcDenom, initialUsdcAmount),
		sdk.NewCoin(failedSwapDenom, initialFailedAmount),
	))

	communityPoolPercentage := osmomath.MustNewDecFromStr("0.25")                                      // 25%
	daiForCommunityPool := initialDaiAmount.ToLegacyDec().Mul(communityPoolPercentage).TruncateInt()   // 25% of 1,000,000 = 250,000
	usdcForCommunityPool := initialUsdcAmount.ToLegacyDec().Mul(communityPoolPercentage).TruncateInt() // 25% of 1,000,000 = 250,000

	burnPercentage := osmomath.MustNewDecFromStr("0.525")                            // 52.5%
	daiForBurn := initialDaiAmount.ToLegacyDec().Mul(burnPercentage).TruncateInt()   // 52.5% of 1,000,000 = 525,000
	usdcForBurn := initialUsdcAmount.ToLegacyDec().Mul(burnPercentage).TruncateInt() // 52.5% of 1,000,000 = 525,000

	daiForStaking := initialDaiAmount.Sub(daiForCommunityPool).Sub(daiForBurn)
	usdcForStaking := initialUsdcAmount.Sub(usdcForCommunityPool).Sub(usdcForBurn)

	// Get pools to calculate expected swap outputs
	daiCfmmPool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, daiPoolId)
	s.Require().NoError(err)
	usdcCfmmPool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, usdcPoolId)
	s.Require().NoError(err)

	// Calculate expected OSMO outputs from swapping
	expectedDaiBurnOsmo, err := daiCfmmPool.CalcOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(daiDenom, daiForBurn)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)

	expectedUsdcBurnOsmo, err := usdcCfmmPool.CalcOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(usdcDenom, usdcForBurn)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)

	// update state
	_, err = daiCfmmPool.SwapOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(daiDenom, daiForBurn)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)
	_, err = usdcCfmmPool.SwapOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(usdcDenom, usdcForBurn)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)

	expectedDaiStakingOsmo, err := daiCfmmPool.CalcOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(daiDenom, daiForStaking)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)
	expectedUsdcStakingOsmo, err := usdcCfmmPool.CalcOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(usdcDenom, usdcForStaking)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)

	expectedTotalBurnOsmo := expectedDaiBurnOsmo.Amount.Add(expectedUsdcBurnOsmo.Amount)
	expectedTotalStakingOsmo := expectedDaiStakingOsmo.Amount.Add(expectedUsdcStakingOsmo.Amount)

	// Get initial balances
	stakingRewardsCollectorAddress := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	takerFeeBurnModuleAddress := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeBurnName)
	takerFeeStakersModuleAddress := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeStakersName)

	initialBurnBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, baseDenom)
	initialStakingRewardsBalance := s.App.BankKeeper.GetBalance(s.Ctx, stakingRewardsCollectorAddress, baseDenom)

	// Trigger AfterEpochEnd to distribute taker fees (first time)
	err = s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, "day", 1)
	s.Require().NoError(err)

	// Verify burn address received the exact expected OSMO amount (key test for burn mechanism)
	finalBurnBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, baseDenom)
	burnAmountReceived := finalBurnBalance.Amount.Sub(initialBurnBalance.Amount)
	s.Require().Equal(expectedTotalBurnOsmo, burnAmountReceived,
		"Burn address should receive exact expected OSMO amount: expected=%s, actual=%s", expectedTotalBurnOsmo.String(), burnAmountReceived.String())

	// Verify staking rewards received exact expected OSMO amount (22.5% of total)
	finalStakingRewardsBalance := s.App.BankKeeper.GetBalance(s.Ctx, stakingRewardsCollectorAddress, baseDenom)
	stakingRewardsAmountReceived := finalStakingRewardsBalance.Amount.Sub(initialStakingRewardsBalance.Amount)

	s.Require().Equal(expectedTotalStakingOsmo, stakingRewardsAmountReceived,
		"Staking rewards should receive exact expected OSMO amount: expected=%s, actual=%s", expectedTotalStakingOsmo.String(), stakingRewardsAmountReceived.String())

	// Verify taker fee collector is empty
	takerFeeCollectorAddress := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeCollectorName)
	balancesAfterEpochEnd := s.App.BankKeeper.GetAllBalances(s.Ctx, takerFeeCollectorAddress)
	s.Require().True(balancesAfterEpochEnd.IsZero(), "Taker fee collector should be empty after distribution")

	// Verify that failed swap tokens remain in their respective module accounts with exact amounts
	// (since swap failed, they should accumulate in the module accounts until a pool is available)
	burnModuleFailedBalance := s.App.BankKeeper.GetBalance(s.Ctx, takerFeeBurnModuleAddress, failedSwapDenom)
	stakersModuleFailedBalance := s.App.BankKeeper.GetBalance(s.Ctx, takerFeeStakersModuleAddress, failedSwapDenom)

	expectedBurnModuleFailedAmount := takerFeeParams.NonOsmoTakerFeeDistribution.Burn.MulInt(initialFailedAmount).TruncateInt()
	expectedStakersModuleFailedAmount := takerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards.MulInt(initialFailedAmount).TruncateInt()

	s.Require().Equal(expectedBurnModuleFailedAmount, burnModuleFailedBalance.Amount,
		"Burn module should contain exact expected failed tokens: expected=%s, actual=%s", expectedBurnModuleFailedAmount.String(), burnModuleFailedBalance.Amount.String())
	s.Require().Equal(expectedStakersModuleFailedAmount, stakersModuleFailedBalance.Amount,
		"Stakers module should contain exact expected failed tokens: expected=%s, actual=%s", expectedStakersModuleFailedAmount.String(), stakersModuleFailedBalance.Amount.String())

	// Now simulate second epoch where we create a pool for the failed token
	// This tests the recovery mechanism where previously failed swaps get picked up

	// Create pool for the previously failed token
	failedTokenPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(baseDenom, poolAssetAmount),
		sdk.NewInt64Coin(failedSwapDenom, poolAssetAmount),
	)

	// Set the pool for the denom pair in protorev
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, failedSwapDenom, baseDenom, failedTokenPoolId)

	// Verify new pool exists
	_, err = s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, failedTokenPoolId)
	s.Require().NoError(err)

	// Calculate expected OSMO outputs for the failed token recovery
	failedTokenCfmmPool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, failedTokenPoolId)
	s.Require().NoError(err)

	// Use actual amounts from the first epoch to calculate expected swap outputs
	actualBurnTokenAmount := burnModuleFailedBalance.Amount
	expectedFailedBurnOsmo, err := failedTokenCfmmPool.CalcOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(failedSwapDenom, actualBurnTokenAmount)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)

	// update state
	_, err = failedTokenCfmmPool.SwapOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(failedSwapDenom, actualBurnTokenAmount)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)

	actualStakersTokenAmount := stakersModuleFailedBalance.Amount
	expectedFailedStakersOsmo, err := failedTokenCfmmPool.CalcOutAmtGivenIn(s.Ctx, sdk.Coins{sdk.NewCoin(failedSwapDenom, actualStakersTokenAmount)}, baseDenom, osmomath.ZeroDec())
	s.Require().NoError(err)

	// Record balances before second epoch end
	burnBalanceBeforeSecond := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, baseDenom)
	stakingBalanceBeforeSecond := s.App.BankKeeper.GetBalance(s.Ctx, stakingRewardsCollectorAddress, baseDenom)

	// System under test - trigger AfterEpochEnd again (second time with pool available)
	err = s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, "day", 2)
	s.Require().NoError(err)

	// Verify that the previously failed tokens are now successfully swapped and burned/distributed with exact amounts
	finalBurnBalanceSecond := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, baseDenom)
	finalStakingBalanceSecond := s.App.BankKeeper.GetBalance(s.Ctx, stakingRewardsCollectorAddress, baseDenom)

	secondBurnIncrease := finalBurnBalanceSecond.Amount.Sub(burnBalanceBeforeSecond.Amount)
	secondStakingIncrease := finalStakingBalanceSecond.Amount.Sub(stakingBalanceBeforeSecond.Amount)

	// Verify exact amounts for the second epoch swaps
	s.Require().Equal(expectedFailedBurnOsmo.Amount, secondBurnIncrease,
		"Second epoch burn should match exact expected amount: expected=%s, actual=%s", expectedFailedBurnOsmo.Amount.String(), secondBurnIncrease.String())
	s.Require().Equal(expectedFailedStakersOsmo.Amount, secondStakingIncrease,
		"Second epoch staking should match exact expected amount: expected=%s, actual=%s", expectedFailedStakersOsmo.Amount.String(), secondStakingIncrease.String())

	// Verify no non-OSMO tokens remain in burn address (they should have been swapped to OSMO first)
	burnAddressDaiBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, daiDenom)
	burnAddressUsdcBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, usdcDenom)
	burnAddressFailedBalance := s.App.BankKeeper.GetBalance(s.Ctx, burnAddress, failedSwapDenom)

	s.Require().True(burnAddressDaiBalance.IsZero(), "Burn address should not contain DAI tokens")
	s.Require().True(burnAddressUsdcBalance.IsZero(), "Burn address should not contain USDC tokens")
	s.Require().True(burnAddressFailedBalance.IsZero(), "Burn address should not contain failed swap tokens")
}

// TestDistributeSmoothingBufferToStakers tests the distributeSmoothingBufferToStakers function
func (s *KeeperTestSuite) TestDistributeSmoothingBufferToStakers() {
	s.SetupTest(false)

	tests := []struct {
		name                    string
		bufferBalance           osmomath.Int
		smoothingFactor         uint64
		expectedDistribution    osmomath.Int
		expectedRemainingBuffer osmomath.Int
	}{
		{
			name:                    "Normal smoothing with factor 7",
			bufferBalance:           osmomath.NewInt(7000000),
			smoothingFactor:         7,
			expectedDistribution:    osmomath.NewInt(1000000), // 7000000 / 7
			expectedRemainingBuffer: osmomath.NewInt(6000000),
		},
		{
			name:                    "No smoothing (factor 1)",
			bufferBalance:           osmomath.NewInt(5000000),
			smoothingFactor:         1,
			expectedDistribution:    osmomath.NewInt(5000000), // All distributed
			expectedRemainingBuffer: osmomath.ZeroInt(),
		},
		{
			name:                    "Large smoothing factor",
			bufferBalance:           osmomath.NewInt(30000000),
			smoothingFactor:         30,
			expectedDistribution:    osmomath.NewInt(1000000), // 30000000 / 30
			expectedRemainingBuffer: osmomath.NewInt(29000000),
		},
		{
			name:                    "Empty buffer",
			bufferBalance:           osmomath.ZeroInt(),
			smoothingFactor:         7,
			expectedDistribution:    osmomath.ZeroInt(),
			expectedRemainingBuffer: osmomath.ZeroInt(),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest(false)
			baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)

			// Set smoothing factor in params
			poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
			poolManagerParams.TakerFeeParams.DailyStakingRewardsSmoothingFactor = tc.smoothingFactor
			s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

			// Fund the smoothing buffer by sending from a test account
			if tc.bufferBalance.GT(osmomath.ZeroInt()) {
				bufferAddr := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeStakingRewardsBuffer)
				err := s.App.BankKeeper.SendCoins(s.Ctx, s.TestAccs[0], bufferAddr, sdk.NewCoins(sdk.NewCoin(baseDenom, tc.bufferBalance)))
				s.Require().NoError(err)
			}

			// Get initial fee collector balance
			feeCollectorAddr := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			initialFeeCollectorBalance := s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddr, baseDenom)

			// Execute distribution
			s.App.TxFeesKeeper.DistributeSmoothingBufferToStakers(s.Ctx, baseDenom)

			// Check fee collector received the expected amount
			finalFeeCollectorBalance := s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddr, baseDenom)
			distributed := finalFeeCollectorBalance.Amount.Sub(initialFeeCollectorBalance.Amount)
			s.Require().Equal(tc.expectedDistribution, distributed, "Fee collector should receive expected distribution")

			// Check buffer has expected remaining balance
			bufferAddr := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeStakingRewardsBuffer)
			finalBufferBalance := s.App.BankKeeper.GetBalance(s.Ctx, bufferAddr, baseDenom)
			s.Require().Equal(tc.expectedRemainingBuffer, finalBufferBalance.Amount, "Buffer should have expected remaining balance")
		})
	}
}

// TestStakingRewardSmoothing tests the complete staking reward smoothing flow:
// - Starts with empty buffer and smoothing factor of 7
// - First day epoch: collects OSMO and non-OSMO taker fees, swaps non-OSMO, accumulates in buffer, distributes 1/7
// - Week epoch: does nothing to staking rewards
// - Second day epoch: distributes 1/7 of remaining buffer
func (s *KeeperTestSuite) TestStakingRewardSmoothing() {
	s.Setup()

	baseDenom, _ := s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)

	// Use real IBC denoms
	daiDenom := "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7"  // DAI
	usdcDenom := "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4" // USDC

	// Set smoothing factor to 7
	poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolManagerParams.TakerFeeParams.DailyStakingRewardsSmoothingFactor = 7
	// Set staking rewards distribution to 30% for OSMO (rest to burn/community pool)
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.3")
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.7")
	poolManagerParams.TakerFeeParams.OsmoTakerFeeDistribution.CommunityPool = osmomath.ZeroDec()
	// Set staking rewards distribution to 22.5% for non-OSMO
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.StakingRewards = osmomath.MustNewDecFromStr("0.225")
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.Burn = osmomath.MustNewDecFromStr("0.525")
	poolManagerParams.TakerFeeParams.NonOsmoTakerFeeDistribution.CommunityPool = osmomath.MustNewDecFromStr("0.25")
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

	// Create pools for DAI and USDC with liquidity
	var poolAssetAmount = int64(1000000000) // 1000 tokens
	daiPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(baseDenom, poolAssetAmount),
		sdk.NewInt64Coin(daiDenom, poolAssetAmount),
	)
	usdcPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(baseDenom, poolAssetAmount),
		sdk.NewInt64Coin(usdcDenom, poolAssetAmount),
	)

	// Set protorev links for swapping
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, daiDenom, baseDenom, daiPoolId)
	s.App.ProtoRevKeeper.SetPoolForDenomPair(s.Ctx, usdcDenom, baseDenom, usdcPoolId)

	// Get module addresses
	bufferAddr := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeStakingRewardsBuffer)
	feeCollectorAddr := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	takerFeeCollectorAddr := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeCollectorName)
	nonNativeFeeCollectorAddr := s.App.AccountKeeper.GetModuleAddress(types.NonNativeTxFeeCollectorName)
	takerFeeStakersAddr := s.App.AccountKeeper.GetModuleAddress(types.TakerFeeStakersName)

	// Verify buffer starts empty
	initialBufferBalance := s.App.BankKeeper.GetBalance(s.Ctx, bufferAddr, baseDenom)
	s.Require().True(initialBufferBalance.IsZero(), "Buffer should start empty")

	// Fund taker fee collector with OSMO and non-OSMO tokens
	osmoTakerFees := osmomath.NewInt(7000000) // 7 OSMO
	daiTakerFees := osmomath.NewInt(4000000)  // 4 DAI
	usdcTakerFees := osmomath.NewInt(6000000) // 6 USDC
	s.FundModuleAcc(types.TakerFeeCollectorName, sdk.NewCoins(
		sdk.NewCoin(baseDenom, osmoTakerFees),
		sdk.NewCoin(daiDenom, daiTakerFees),
		sdk.NewCoin(usdcDenom, usdcTakerFees),
	))

	// Also fund NonNativeTxFeeCollectorName with some non-OSMO tokens
	s.FundModuleAcc(types.NonNativeTxFeeCollectorName, sdk.NewCoins(
		sdk.NewCoin(daiDenom, osmomath.NewInt(1000000)),  // 1 DAI
		sdk.NewCoin(usdcDenom, osmomath.NewInt(2000000)), // 2 USDC
	))

	// Record initial fee collector balance
	initialFeeCollectorBalance := s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddr, baseDenom)

	// ===== FIRST DAY EPOCH =====
	err := s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, "day", 1)
	s.Require().NoError(err)

	// Verify taker fee collector is empty
	takerFeeCollectorBalance := s.App.BankKeeper.GetBalance(s.Ctx, takerFeeCollectorAddr, baseDenom)
	s.Require().True(takerFeeCollectorBalance.IsZero())

	// Verify non-native collectors are empty
	nonNativeFeeCollectorBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, nonNativeFeeCollectorAddr)
	s.Require().True(nonNativeFeeCollectorBalance.IsZero())

	takerFeeStakersBalance := s.App.BankKeeper.GetAllBalances(s.Ctx, takerFeeStakersAddr)
	s.Require().True(takerFeeStakersBalance.IsZero())

	// Get actual buffer balance and distribution after first day
	bufferBalanceAfterFirstDay := s.App.BankKeeper.GetBalance(s.Ctx, bufferAddr, baseDenom)
	feeCollectorBalanceAfterFirstDay := s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddr, baseDenom)
	firstDayDistribution := feeCollectorBalanceAfterFirstDay.Amount.Sub(initialFeeCollectorBalance.Amount)

	// Calculate total that went into buffer before distribution (buffer + what was distributed)
	totalInBufferBeforeDistribution := bufferBalanceAfterFirstDay.Amount.Add(firstDayDistribution)

	// Verify the 1/7 distribution: firstDayDistribution should be exactly 1/7 of total
	expectedFirstDistribution := totalInBufferBeforeDistribution.QuoRaw(7)
	s.Require().Equal(expectedFirstDistribution, firstDayDistribution)

	// Verify buffer has exactly 6/7 remaining
	expectedRemainingInBuffer := totalInBufferBeforeDistribution.Sub(expectedFirstDistribution)
	s.Require().Equal(expectedRemainingInBuffer, bufferBalanceAfterFirstDay.Amount)

	// ===== WEEK EPOCH (should do nothing to staking rewards) =====
	bufferBalanceBeforeWeekEpoch := s.App.BankKeeper.GetBalance(s.Ctx, bufferAddr, baseDenom)
	feeCollectorBalanceBeforeWeekEpoch := s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddr, baseDenom)

	err = s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, "week", 1)
	s.Require().NoError(err)

	// Verify buffer unchanged
	bufferBalanceAfterWeekEpoch := s.App.BankKeeper.GetBalance(s.Ctx, bufferAddr, baseDenom)
	s.Require().Equal(bufferBalanceBeforeWeekEpoch.Amount, bufferBalanceAfterWeekEpoch.Amount)

	// Verify fee collector unchanged
	feeCollectorBalanceAfterWeekEpoch := s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddr, baseDenom)
	s.Require().Equal(feeCollectorBalanceBeforeWeekEpoch.Amount, feeCollectorBalanceAfterWeekEpoch.Amount)

	// ===== SECOND DAY EPOCH (should distribute 1/7 of remaining buffer) =====
	bufferBeforeSecondDay := bufferBalanceAfterWeekEpoch.Amount
	expectedSecondDistribution := bufferBeforeSecondDay.QuoRaw(7)
	expectedRemainingAfterSecondDay := bufferBeforeSecondDay.Sub(expectedSecondDistribution)

	err = s.App.TxFeesKeeper.AfterEpochEnd(s.Ctx, "day", 2)
	s.Require().NoError(err)

	// Verify buffer has correct remaining amount
	bufferBalanceAfterSecondDay := s.App.BankKeeper.GetBalance(s.Ctx, bufferAddr, baseDenom)
	s.Require().Equal(expectedRemainingAfterSecondDay, bufferBalanceAfterSecondDay.Amount)

	// Verify fee collector received 1/7 of what was in buffer
	feeCollectorBalanceAfterSecondDay := s.App.BankKeeper.GetBalance(s.Ctx, feeCollectorAddr, baseDenom)
	secondDayDistribution := feeCollectorBalanceAfterSecondDay.Amount.Sub(feeCollectorBalanceAfterWeekEpoch.Amount)
	s.Require().Equal(expectedSecondDistribution, secondDayDistribution)

	// Verify total distributions add up correctly
	totalDistributed := firstDayDistribution.Add(secondDayDistribution)
	totalRemaining := bufferBalanceAfterSecondDay.Amount
	s.Require().Equal(totalInBufferBeforeDistribution, totalDistributed.Add(totalRemaining))
}
