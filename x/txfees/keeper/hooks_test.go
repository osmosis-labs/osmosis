package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/txfees/types"

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

			// End of epoch, so all the non-osmo fee amount should be swapped to osmo and transfer to fee module account
			params := s.App.IncentivesKeeper.GetParams(s.Ctx)
			futureCtx := s.Ctx.WithBlockTime(time.Now().Add(time.Minute))
			err := s.App.TxFeesKeeper.AfterEpochEnd(futureCtx, params.DistrEpochIdentifier, int64(1))
			s.NoError(err)

			// check the balance of the native-basedenom in module
			moduleAddrFee := s.App.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			moduleBaseDenomBalance := s.App.BankKeeper.GetBalance(s.Ctx, moduleAddrFee, tc.baseDenom)

			// non-osmos module account should be empty as all the funds should be transferred to osmo module
			s.Empty(s.App.BankKeeper.GetAllBalances(s.Ctx, moduleAddrNonNativeFee))
			// check that the total osmo amount has been transferred to module account
			s.Equal(moduleBaseDenomBalance.Amount.String(), finalOutputAmount.String())
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
		stakingDenom, _ = s.App.TxFeesKeeper.GetBaseDenom(s.Ctx)
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

	// Prepare the tx fee collector.
	txFeeCollectorAddress := prepareFeeCollector(types.NonNativeTxFeeCollectorName, stakingDenom)

	// Snapshot the community pool balance before the epoch end.
	communityPoolAddress := s.App.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	communityPoolBalanceBefore := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)

	// Set up taker fee share agreements
	for _, agreement := range defaultTakerFeeShareAgreements {
		s.App.PoolManagerKeeper.SetTakerFeeShareAgreementForDenom(s.Ctx, agreement)
	}

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

	communityPoolBalanceAfter := s.App.BankKeeper.GetAllBalances(s.Ctx, communityPoolAddress)
	communityPoolBalanceDelta := communityPoolBalanceAfter.Sub(communityPoolBalanceBefore...)

	// Confirm that that all tokens that are of the configured denom parameter are sent to the community pool.
	s.Require().Len(communityPoolBalanceDelta, 4)
	s.Require().Equal(communityPoolBalanceDelta[0].Denom, otherPreSwapDenom)
	s.Require().Equal(communityPoolBalanceDelta[1].Denom, denomWithNoPool)
	s.Require().Equal(communityPoolBalanceDelta[2].Denom, preSwapDenom)

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
