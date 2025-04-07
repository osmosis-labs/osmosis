package e2e

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/address"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/iancoleman/orderedmap"

	packetforwardingtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	ibchookskeeper "github.com/osmosis-labs/osmosis/x/ibc-hooks/keeper"

	ibcratelimittypes "github.com/osmosis-labs/osmosis/v27/x/ibc-rate-limit/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/configurer/config"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/initialization"
)

var (
	// minDecTolerance minimum tolerance for osmomath.Dec, given its precision of 18.
	minDecTolerance = osmomath.MustNewDecFromStr("0.000000000000000001")
	// TODO: lower
	govPropTimeout = time.Minute
)

func (s *IntegrationTestSuite) TestPrepE2E() {
	// Reset the default taker fee to 0.15%, so we can actually run tests with it activated
	s.T().Run("SetDefaultTakerFeeChainB", func(t *testing.T) {
		t.Parallel()
		s.T().Log("resetting the default taker fee to 0.15% on chain B only")
		s.SetDefaultTakerFeeChainB()
	})

	s.T().Run("SetExpeditedVotingPeriodChainA", func(t *testing.T) {
		t.Parallel()
		s.T().Log("setting the expedited voting period to 7 seconds on chain A")
		s.SetExpeditedVotingPeriodChainA()
	})

	s.T().Run("SetExpeditedVotingPeriodChainB", func(t *testing.T) {
		t.Parallel()
		s.T().Log("setting the expedited voting period to 7 seconds on chain B")
		s.SetExpeditedVotingPeriodChainB()
	})
}

// TODO: Find more scalable way to do this
func (s *IntegrationTestSuite) TestStartE2E() {
	// Zero Dependent Tests
	s.T().Run("CreateConcentratedLiquidityPoolVoting_And_TWAP", func(t *testing.T) {
		t.Parallel()
		s.CreateConcentratedLiquidityPoolVoting_And_TWAP()
	})

	s.T().Run("ProtoRev", func(t *testing.T) {
		t.Parallel()
		s.ProtoRev()
	})

	s.T().Run("ConcentratedLiquidity", func(t *testing.T) {
		t.Parallel()
		s.ConcentratedLiquidity()
	})

	s.T().Run("SuperfluidVoting", func(t *testing.T) {
		t.Parallel()
		s.SuperfluidVoting()
	})

	s.T().Run("AddToExistingLock", func(t *testing.T) {
		t.Parallel()
		s.AddToExistingLock()
	})

	s.T().Run("ExpeditedProposals", func(t *testing.T) {
		t.Parallel()
		s.ExpeditedProposals()
	})

	s.T().Run("GeometricTWAP", func(t *testing.T) {
		t.Parallel()
		s.GeometricTWAP()
	})

	s.T().Run("LargeWasmUpload", func(t *testing.T) {
		t.Parallel()
		s.LargeWasmUpload()
	})

	s.T().Run("StableSwap", func(t *testing.T) {
		t.Parallel()
		s.StableSwap()
	})

	// Test currently disabled
	// s.T().Run("ArithmeticTWAP", func(t *testing.T) {
	// 	t.Parallel()
	// 	s.ArithmeticTWAP()
	// })

	// State Sync Dependent Tests

	if s.skipStateSync || !s.runScheduledTest {
		s.T().Skip()
	} else {
		s.T().Run("StateSync", func(t *testing.T) {
			t.Parallel()
			s.StateSync()
		})
	}

	// Upgrade Dependent Tests

	if s.skipUpgrade {
		s.T().Skip("Skipping GeometricTwapMigration test")
	} else {
		s.T().Run("GeometricTwapMigration", func(t *testing.T) {
			t.Parallel()
			s.GeometricTwapMigration()
		})
	}

	if s.skipUpgrade {
		s.T().Skip("Skipping AddToExistingLockPostUpgrade test")
	} else {
		s.T().Run("AddToExistingLockPostUpgrade", func(t *testing.T) {
			t.Parallel()
			s.AddToExistingLockPostUpgrade()
		})
	}

	// IBC Dependent Tests

	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	} else {
		s.T().Run("IBCTokenTransferRateLimiting", func(t *testing.T) {
			t.Parallel()
			s.IBCTokenTransferRateLimiting()
		})

		s.T().Run("IBCTokenTransferAndCreatePool", func(t *testing.T) {
			t.Parallel()
			s.IBCTokenTransferAndCreatePool()
		})

		s.T().Run("IBCWasmHooks", func(t *testing.T) {
			t.Parallel()
			s.IBCWasmHooks()
		})

		s.T().Run("PacketForwarding", func(t *testing.T) {
			t.Parallel()
			s.PacketForwarding()
		})
	}
}

// TestProtoRev is a test that ensures that the protorev module is working as expected. In particular, this tests and ensures that:
// 1. The protorev module is correctly configured on init
// 2. The protorev module can correctly back run a swap
// 3. the protorev module correctly tracks statistics
func (s *IntegrationTestSuite) ProtoRev() {
	const (
		poolFile1 = "protorevPool1.json"
		poolFile2 = "protorevPool2.json"
		poolFile3 = "protorevPool3.json"

		walletName = "swap-that-creates-an-arb"

		denomIn      = initialization.LuncIBCDenom
		denomOut     = initialization.UstIBCDenom
		amount       = "10000"
		minAmountOut = "1"

		epochIdentifier = "day"
	)

	// NOTE: Uses chainA since IBC denoms are hard coded.
	chainA, chainANode := s.getChainACfgs()

	sender := chainANode.GetWallet(initialization.ValidatorWalletName)

	// --------------- Module init checks ---------------- //
	s.T().Logf("running protorev module init checks")

	enabled, err := chainANode.QueryProtoRevEnabled()
	s.Require().NoError(err)
	s.Require().True(enabled, "protorev enabled should be true on init")

	hotRoutes, err := chainANode.QueryProtoRevTokenPairArbRoutes()
	s.Require().NoError(err, "protorev module should have no new hot routes on init")
	s.Require().Len(hotRoutes, 0, "protorev module should have no new hot routes on init")

	_, err = chainANode.QueryProtoRevNumberOfTrades()
	s.Require().Error(err, "protorev module should have no trades on init")

	info, err := chainANode.QueryProtoRevInfoByPoolType()
	s.Require().NoError(err, "protorev module should have pool info on init")
	s.Require().NotNil(info, "protorev module should have pool info on init")

	_, err = chainANode.QueryProtoRevMaxPoolPointsPerTx()
	s.Require().NoError(err, "protorev module should have max pool points per tx on init")

	_, err = chainANode.QueryProtoRevMaxPoolPointsPerBlock()
	s.Require().NoError(err, "protorev module should have max pool points per block on init")

	supportedBaseDenoms, err := chainANode.QueryProtoRevBaseDenoms()
	s.Require().NoError(err)
	s.Require().Len(supportedBaseDenoms, 1, "protorev module should only have note as a supported base denom on init")
	s.Require().Equal(supportedBaseDenoms[0].Denom, "note", "protorev module should only have note as a supported base denom on init")

	s.T().Logf("completed protorev module init checks")

	// --------------- Set up for a calculated backrun ---------------- //
	// Create all of the pools that will be used in the test.
	swapPoolId1 := chainANode.CreateBalancerPool(poolFile1, initialization.ValidatorWalletName)
	swapPoolId2 := chainANode.CreateBalancerPool(poolFile2, initialization.ValidatorWalletName)
	swapPoolId3 := chainANode.CreateBalancerPool(poolFile3, initialization.ValidatorWalletName)

	// Wait for the creation to be propagated to the other nodes + for the protorev module to
	// correctly update the highest liquidity pools.
	s.T().Logf("waiting for the protorev module to update the highest liquidity pools (wait %.f sec) after the week epoch duration", initialization.EpochDayDuration.Seconds())
	chainA.WaitForNumEpochs(1, epochIdentifier)

	// Create a wallet to use for the swap txs.
	swapWalletAddr := chainANode.CreateWallet(walletName, chainA)
	coinIn := fmt.Sprintf("%s%s", amount, denomIn)
	chainANode.BankSend(coinIn, sender, swapWalletAddr)

	// Check supplies before swap.
	supplyBefore, err := chainANode.QuerySupply()
	s.Require().NoError(err)
	s.Require().NotNil(supplyBefore)

	// Performing the swap that creates a cyclic arbitrage opportunity.
	s.T().Logf("performing a swap that creates a cyclic arbitrage opportunity")
	chainANode.SwapExactAmountIn(coinIn, minAmountOut, fmt.Sprintf("%d", swapPoolId2), denomOut, swapWalletAddr)

	// --------------- Module checks after a calculated backrun ---------------- //

	supplyCheck := func() {
		s.T().Logf("checking that the supplies have not changed")
		supplyAfter, err := chainANode.QuerySupply()
		s.Require().NoError(err)
		s.Require().Equal(supplyBefore, supplyAfter)
	}

	// Check that the number of trades executed by the protorev module is 1.
	numTradesCheck := func() {
		numTrades, err := chainANode.QueryProtoRevNumberOfTrades()
		s.T().Logf("checking that the protorev module has executed 1 trade")
		s.Require().NoError(err)
		s.Require().Equal(uint64(1), numTrades.Uint64())
	}

	// Check that the profits of the protorev module are not nil.
	profits := func() {
		profits, err := chainANode.QueryProtoRevProfits()
		s.T().Logf("checking that the protorev module has non-nil profits: %s", profits)
		s.Require().NoError(err)
		s.Require().Len(profits, 1)

		// Check that the route statistics of the protorev module are not nil.
		routeStats, err := chainANode.QueryProtoRevAllRouteStatistics()
		s.T().Logf("checking that the protorev module has non-nil route statistics: %x", routeStats)
		s.Require().NoError(err)
		s.Require().Len(routeStats, 1)
		s.Require().Equal(osmomath.OneInt(), routeStats[0].NumberOfTrades)
		s.Require().Equal([]uint64{swapPoolId1, swapPoolId2, swapPoolId3}, routeStats[0].Route)
		s.Require().Equal(profits, routeStats[0].Profits)
	}

	runFuncsInParallelAndBlock([]func(){supplyCheck, numTradesCheck, profits})
}

func (s *IntegrationTestSuite) StableSwap() {
	chainAB, chainABNode := s.getChainCfgs()

	index := s.getChainIndex(chainAB)

	sender := chainABNode.GetWallet(initialization.ValidatorWalletName)

	const (
		denomA = "stake"
		denomB = "note"

		minAmountOut = "1"
	)

	coinAIn, coinBIn := fmt.Sprintf("20000%s", denomA), fmt.Sprintf("2%s", denomB)

	chainABNode.BankSend(initialization.WalletFeeTokens.String(), sender, config.StableswapWallet[index])
	chainABNode.BankSend(coinAIn+","+coinBIn, sender, config.StableswapWallet[index])

	s.T().Log("performing swaps")
	chainABNode.SwapExactAmountIn(coinAIn, minAmountOut, fmt.Sprintf("%d", config.PreUpgradeStableSwapPoolId[index]), denomB, config.StableswapWallet[index])
	chainABNode.SwapExactAmountIn(coinBIn, minAmountOut, fmt.Sprintf("%d", config.PreUpgradeStableSwapPoolId[index]), denomA, config.StableswapWallet[index])
}

// TestGeometricTwapMigration tests that the geometric twap record
// migration runs successfully. It does so by attempting to execute
// the swap on the pool created pre-upgrade. When a pool is created
// pre-upgrade, twap records are initialized for a pool. By running
// a swap post-upgrade, we confirm that the geometric twap was initialized
// correctly and does not cause a chain halt. This test was created
// in-response to a testnet incident when performing the geometric twap
// upgrade. Upon adding the migrations logic, the tests began to pass.
func (s *IntegrationTestSuite) GeometricTwapMigration() {
	if s.skipUpgrade {
		s.T().Skip("Skipping upgrade tests")
	}

	var (
		// Configurations for tests/e2e/scripts/pool1A.json
		// This pool gets initialized pre-upgrade.
		minAmountOut    = "1"
		otherDenom      = []string{"ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518", "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"}
		migrationWallet = "migration"
	)

	chainAB, chainABNode := s.getChainCfgs()
	index := s.getChainIndex(chainAB)

	sender := chainABNode.GetWallet(initialization.ValidatorWalletName)

	noteIn := fmt.Sprintf("1000000%s", "note")

	swapWalletAddr := chainABNode.CreateWallet(migrationWallet, chainAB)

	chainABNode.BankSend(noteIn, sender, swapWalletAddr)

	// Swap to create new twap records on the pool that was created pre-upgrade.
	chainABNode.SwapExactAmountIn(noteIn, minAmountOut, fmt.Sprintf("%d", config.PreUpgradePoolId[index]), otherDenom[index], swapWalletAddr)
}

// TestIBCTokenTransfer tests that IBC token transfers work as expected.
// Additionally, it attempts to create a pool with IBC denoms.
func (s *IntegrationTestSuite) IBCTokenTransferAndCreatePool() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA, chainANode := s.getChainACfgs()
	chainB, chainBNode := s.getChainBCfgs()

	ibcSendConfigs := []struct {
		srcCfg    *chain.Config
		destCfg   *chain.Config
		srcNode   *chain.NodeConfig
		recipient string
	}{{chainA, chainB, chainANode, chainBNode.PublicAddress}, {chainB, chainA, chainBNode, chainANode.PublicAddress}}
	tokens := []sdk.Coin{initialization.MelodyToken, initialization.StakeToken}

	unlockFn := chain.IbcLockAddrs([]string{chainANode.PublicAddress, chainBNode.PublicAddress, initialization.ValidatorWalletName})
	defer unlockFn()
	var wg sync.WaitGroup
	wg.Add(4)
	for i := range ibcSendConfigs {
		for j := range tokens {
			cfg := ibcSendConfigs[i]
			token := tokens[j]
			go func() {
				cfg.srcNode.SendIBCNoMutex(cfg.srcCfg, cfg.destCfg, cfg.recipient, token)
				wg.Done()
			}()
		}
	}
	wg.Wait() // Wait for all goroutines to finish

	chainANode.CreateBalancerPool("ibcDenomPool.json", initialization.ValidatorWalletName)
}

// TestSuperfluidVoting tests that superfluid voting is functioning as expected.
// It does so by doing the following:
// - creating a pool
// - attempting to submit a proposal to enable superfluid voting in that pool
// - voting yes on the proposal from the validator wallet
// - voting no on the proposal from the delegator wallet
// - ensuring that delegator's wallet overwrites the validator's vote
func (s *IntegrationTestSuite) SuperfluidVoting() {
	chainAB, chainABNode := s.getChainCfgs()

	poolId := chainABNode.CreateBalancerPool("nativeDenomPool.json", initialization.ValidatorWalletName)

	// enable superfluid assets
	chainABNode.EnableSuperfluidAsset(chainAB, fmt.Sprintf("gamm/pool/%d", poolId), true)

	// setup wallets and send gamm tokens to these wallets (both chains)
	superfluidVotingWallet := chainABNode.CreateWallet("TestSuperfluidVoting", chainAB)
	chainABNode.BankSend(fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId), initialization.ValidatorWalletName, superfluidVotingWallet)
	lockId := chainABNode.LockTokens(fmt.Sprintf("%v%s", osmomath.NewInt(1000000000000000000), fmt.Sprintf("gamm/pool/%d", poolId)), "240s", superfluidVotingWallet)

	chainABNode.SuperfluidDelegate(lockId, chainABNode.OperatorAddress, superfluidVotingWallet)

	// create a text prop and vote yes
	propNumber := chainABNode.SubmitTextProposal("superfluid vote overwrite test", false, true)

	chain.AllValsVoteOnProposal(chainAB, propNumber)

	// set delegator vote to no
	chainABNode.VoteNoProposal(superfluidVotingWallet, propNumber)

	s.Eventually(
		func() bool {
			propTally, err := chainABNode.QueryPropTally(propNumber)
			if err != nil {
				return false
			}
			if propTally.Abstain.Int64()+propTally.No.Int64()+propTally.NoWithVeto.Int64()+propTally.Yes.Int64() <= 0 {
				return false
			}
			return true
		},
		govPropTimeout,
		10*time.Millisecond,
		"Symphony node failed to retrieve prop tally",
	)
	propTally, err := chainABNode.QueryPropTally(propNumber)
	s.Require().NoError(err)
	noTotalFinal, err := strconv.Atoi(propTally.No.String())
	s.NoError(err)

	s.Eventually(
		func() bool {
			intAccountBalance, err := chainABNode.QueryIntermediaryAccount(fmt.Sprintf("gamm/pool/%d", poolId), chainABNode.OperatorAddress)
			s.Require().NoError(err)
			if err != nil {
				return false
			}
			if noTotalFinal != intAccountBalance {
				fmt.Printf("noTotalFinal %v does not match intAccountBalance %v", noTotalFinal, intAccountBalance)
				return false
			}
			return true
		},
		1*time.Minute,
		10*time.Millisecond,
		"superfluid delegation vote overwrite not working as expected",
	)
}

func (s *IntegrationTestSuite) IBCTokenTransferRateLimiting() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA, chainANode := s.getChainACfgs()
	chainB, chainBNode := s.getChainBCfgs()

	receiver := chainBNode.GetWallet(initialization.ValidatorWalletName)

	// If the RL param is already set. Remember it to set it back at the end
	param := chainANode.QueryParams(ibcratelimittypes.ModuleName, string(ibcratelimittypes.KeyContractAddress))
	fmt.Println("param", param)

	melodySupply, err := chainANode.QuerySupplyOf("note")
	s.Require().NoError(err)

	f, err := melodySupply.ToLegacyDec().Float64()
	s.Require().NoError(err)

	over := f * 0.02

	paths := fmt.Sprintf(`{"channel_id": "channel-0", "denom": "%s", "quotas": [{"name":"testQuota", "duration": 86400, "send_recv": [1, 1]}] }`, initialization.MelodyToken.Denom)

	// Sending >1%
	fmt.Println("Sending >1%")
	chainANode.SendIBC(chainA, chainB, receiver, sdk.NewInt64Coin(initialization.MelodyDenom, int64(over)))

	contract, err := chainANode.SetupRateLimiting(paths, chainANode.PublicAddress, chainA, true)
	s.Require().NoError(err)

	s.Eventually(
		func() bool {
			val := chainANode.QueryParams(ibcratelimittypes.ModuleName, string(ibcratelimittypes.KeyContractAddress))
			return strings.Contains(val, contract)
		},
		govPropTimeout,
		10*time.Millisecond,
		"Symphony node failed to retrieve params",
	)

	// Sending <1%. Should work
	fmt.Println("Sending <1%. Should work")
	chainANode.SendIBC(chainA, chainB, receiver, sdk.NewInt64Coin(initialization.MelodyDenom, 1))
	// Sending >1%. Should fail
	fmt.Println("Sending >1%. Should fail")
	chainANode.FailIBCTransfer(initialization.ValidatorWalletName, receiver, fmt.Sprintf("%dnote", int(over)))

	// Removing the rate limit so it doesn't affect other tests
	chainANode.WasmExecute(contract, `{"remove_path": {"channel_id": "channel-0", "denom": "note"}}`, initialization.ValidatorWalletName)
	// reset the param to the original contract if it existed
	if param != "" {
		err = chainANode.ParamChangeProposal(
			ibcratelimittypes.ModuleName,
			string(ibcratelimittypes.KeyContractAddress),
			[]byte(param),
			chainA,
			true,
		)
		s.Require().NoError(err)
		s.Eventually(func() bool {
			val := chainANode.QueryParams(ibcratelimittypes.ModuleName, string(ibcratelimittypes.KeyContractAddress))
			return strings.Contains(val, param)
		}, time.Second*30, 10*time.Millisecond)
	}
}

func (s *IntegrationTestSuite) LargeWasmUpload() {
	_, chainNode := s.getChainCfgs()
	validatorAddr := chainNode.GetWallet(initialization.ValidatorWalletName)
	chainNode.StoreWasmCode("bytecode/large.wasm", validatorAddr)
}

func (s *IntegrationTestSuite) IBCWasmHooks() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA, chainANode := s.getChainACfgs()
	_, chainBNode := s.getChainBCfgs()

	contractAddr := s.UploadAndInstantiateCounter(chainA)

	transferAmount := int64(10)
	validatorAddr := chainBNode.GetWallet(initialization.ValidatorWalletName)
	fmt.Println("Sending IBC transfer IBCWasmHooks")
	coin := sdk.NewCoin("note", osmomath.NewInt(transferAmount))
	chainBNode.SendIBCTransfer(chainA, validatorAddr, contractAddr,
		fmt.Sprintf(`{"wasm":{"contract":"%s","msg": {"increment": {}} }}`, contractAddr), coin)

	// check the balance of the contract
	denomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "note"))
	ibcDenom := denomTrace.IBCDenom()
	s.CallCheckBalance(chainANode, contractAddr, ibcDenom, transferAmount)

	// sender wasm addr
	senderBech32, err := ibchookskeeper.DeriveIntermediateSender("channel-0", validatorAddr, "symphony")

	var response map[string]interface{}
	s.Require().Eventually(func() bool {
		response, err = chainANode.QueryWasmSmartObject(contractAddr, fmt.Sprintf(`{"get_total_funds": {"addr": "%s"}}`, senderBech32))
		if err != nil {
			return false
		}

		totalFundsIface, ok := response["total_funds"].([]interface{})
		if !ok || len(totalFundsIface) == 0 {
			return false
		}

		totalFunds, ok := totalFundsIface[0].(map[string]interface{})
		if !ok {
			return false
		}

		amount, ok := totalFunds["amount"].(string)
		if !ok {
			return false
		}

		denom, ok := totalFunds["denom"].(string)
		if !ok {
			return false
		}

		// check if denom contains "note"
		return amount == strconv.FormatInt(transferAmount, 10) && strings.Contains(denom, "ibc")
	},

		15*time.Second,
		10*time.Millisecond,
	)
}

// TestPacketForwarding sends a packet from chainA to chainB, and forwards it
// back to chainA with a custom memo to execute the counter contract on chain A
func (s *IntegrationTestSuite) PacketForwarding() {
	if s.skipIBC {
		s.T().Skip("Skipping IBC tests")
	}
	chainA, chainANode := s.getChainACfgs()
	chainB, _ := s.getChainBCfgs()

	// Instantiate the counter contract on chain A
	contractAddr := s.UploadAndInstantiateCounter(chainA)

	transferAmount := int64(10)
	validatorAddr := chainANode.GetWallet(initialization.ValidatorWalletName)
	// Specify that the counter contract should be called on chain A when the packet is received
	contractCallMemo := []byte(fmt.Sprintf(`{"wasm":{"contract":"%s","msg": {"increment": {}} }}`, contractAddr))
	// Generate the forward metadata
	forwardMetadata := packetforwardingtypes.ForwardMetadata{
		Receiver: contractAddr,
		Port:     "transfer",
		Channel:  "channel-0",
		Next:     packetforwardingtypes.NewJSONObject(false, contractCallMemo, orderedmap.OrderedMap{}), // The packet sent to chainA will have this memo
	}
	memoData := packetforwardingtypes.PacketMetadata{Forward: &forwardMetadata}
	forwardMemo, err := json.Marshal(memoData)
	s.NoError(err)
	// Send the transfer from chainA to chainB. ChainB will parse the memo and forward the packet back to chainA
	coin := sdk.NewCoin("note", osmomath.NewInt(transferAmount))
	chainANode.SendIBCTransfer(chainB, validatorAddr, validatorAddr, string(forwardMemo), coin)

	// check the balance of the contract
	s.CallCheckBalance(chainANode, contractAddr, "note", transferAmount)

	// Getting the sender as set by PFM
	senderStr := fmt.Sprintf("channel-0/%s", validatorAddr)
	senderHash32 := address.Hash(packetforwardingtypes.ModuleName, []byte(senderStr)) // typo intended
	sender := sdk.AccAddress(senderHash32[:20])
	bech32Prefix := "melody"
	pfmSender, err := sdk.Bech32ifyAddressBytes(bech32Prefix, sender)
	s.Require().NoError(err)

	// sender wasm addr
	senderBech32, err := ibchookskeeper.DeriveIntermediateSender("channel-0", pfmSender, "melody")
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		response, err := chainANode.QueryWasmSmartObject(contractAddr, fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, senderBech32))
		if err != nil {
			return false
		}
		countValue, ok := response["count"].(float64)
		if !ok {
			return false
		}
		return countValue == 0
	},
		15*time.Second,
		10*time.Millisecond,
	)
}

// TestAddToExistingLockPostUpgrade ensures addToExistingLock works for locks created preupgrade.
func (s *IntegrationTestSuite) AddToExistingLockPostUpgrade() {
	if s.skipUpgrade {
		s.T().Skip("Skipping AddToExistingLockPostUpgrade test")
	}

	chainAB, chainABNode := s.getChainCfgs()
	index := s.getChainIndex(chainAB)

	// ensure we can add to existing locks and superfluid locks that existed pre upgrade on chainA
	// we use the hardcoded gamm/pool/1 and these specific wallet names to match what was created pre upgrade
	preUpgradePoolShareDenom := fmt.Sprintf("gamm/pool/%d", config.PreUpgradePoolId[index])

	lockupWalletAddr, lockupWalletSuperfluidAddr := chainABNode.GetWallet("lockup-wallet"), chainABNode.GetWallet("lockup-wallet-superfluid")
	chainABNode.AddToExistingLock(osmomath.NewInt(1000000000000000000), preUpgradePoolShareDenom, "240s", lockupWalletAddr, 1)
	chainABNode.AddToExistingLock(osmomath.NewInt(1000000000000000000), preUpgradePoolShareDenom, "240s", lockupWalletSuperfluidAddr, 2)
}

// TestAddToExistingLock tests lockups to both regular and superfluid locks.
func (s *IntegrationTestSuite) AddToExistingLock() {
	chainAB, chainABNode := s.getChainCfgs()

	funder := chainABNode.GetWallet(initialization.ValidatorWalletName)
	// ensure we can add to new locks and superfluid locks
	// create pool and enable superfluid assets
	poolId := chainABNode.CreateBalancerPool("nativeDenomPool.json", funder)
	chainABNode.EnableSuperfluidAsset(chainAB, fmt.Sprintf("gamm/pool/%d", poolId), true)

	// setup wallets and send gamm tokens to these wallets on chainA
	gammShares := fmt.Sprintf("10000000000000000000gamm/pool/%d", poolId)
	fundTokens := []string{gammShares, initialization.WalletFeeTokens.String()}
	lockupWalletAddr := chainABNode.CreateWalletAndFundFrom("TestAddToExistingLock", funder, fundTokens, chainAB)
	lockupWalletSuperfluidAddr := chainABNode.CreateWalletAndFundFrom("TestAddToExistingLockSuperfluid", funder, fundTokens, chainAB)

	// ensure we can add to new locks and superfluid locks on chainA
	chainABNode.LockAndAddToExistingLock(chainAB, osmomath.NewInt(1000000000000000000), fmt.Sprintf("gamm/pool/%d", poolId), lockupWalletAddr, lockupWalletSuperfluidAddr)
}

// TestArithmeticTWAP tests TWAP by creating a pool, performing a swap.
// These two operations should create TWAP records.
// Then, we wait until the epoch for the records to be pruned.
// The records are guaranteed to be pruned at the next epoch
// because twap keep time = epoch time / 4 and we use a timer
// to wait for at least the twap keep time.
func (s *IntegrationTestSuite) ArithmeticTWAP() {
	s.T().Skip("TODO: investigate further: https://github.com/osmosis-labs/osmosis/issues/4342")

	const (
		poolFile   = "nativeDenomThreeAssetPool.json"
		walletName = "arithmetic-twap-wallet"

		denomA = "stake"
		denomB = "uion"
		denomC = "note"

		minAmountOut = "1"

		epochIdentifier = "day"
	)

	coinAIn, coinBIn, coinCIn := fmt.Sprintf("2000000%s", denomA), fmt.Sprintf("2000000%s", denomB), fmt.Sprintf("2000000%s", denomC)

	chainAB, chainABNode := s.getChainCfgs()
	sender := chainABNode.GetWallet(initialization.ValidatorWalletName)

	// Triggers the creation of TWAP records.
	poolId := chainABNode.CreateBalancerPool(poolFile, initialization.ValidatorWalletName)
	swapWalletAddr := chainABNode.CreateWalletAndFund(walletName, []string{initialization.WalletFeeTokens.String()}, chainAB)

	timeBeforeSwap := chainABNode.QueryLatestBlockTime()
	// Wait for the next height so that the requested twap
	// start time (timeBeforeSwap) is not equal to the block time.
	chainAB.WaitForNumHeights(2)

	s.T().Log("querying for the first TWAP to now before swap")
	twapFromBeforeSwapToBeforeSwapOneAB, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapOneBC, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapOneCA, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	s.Require().NoError(err)

	swapAmt := coinAIn + "," + coinBIn + "," + coinCIn
	chainABNode.BankSend(swapAmt, sender, swapWalletAddr)

	s.T().Log("querying for the second TWAP to now before swap, must equal to first")
	twapFromBeforeSwapToBeforeSwapTwoAB, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapTwoBC, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)
	twapFromBeforeSwapToBeforeSwapTwoCA, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap.Add(50*time.Millisecond))
	s.Require().NoError(err)

	// Since there were no swaps between the two queries, the TWAPs should be the same.
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneAB, twapFromBeforeSwapToBeforeSwapTwoAB, osmomath.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneBC, twapFromBeforeSwapToBeforeSwapTwoBC, osmomath.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapFromBeforeSwapToBeforeSwapOneCA, twapFromBeforeSwapToBeforeSwapTwoCA, osmomath.NewDecWithPrec(1, 3))

	s.T().Log("performing swaps")
	chainABNode.SwapExactAmountIn(coinAIn, minAmountOut, fmt.Sprintf("%d", poolId), denomB, swapWalletAddr)
	chainABNode.SwapExactAmountIn(coinBIn, minAmountOut, fmt.Sprintf("%d", poolId), denomC, swapWalletAddr)
	chainABNode.SwapExactAmountIn(coinCIn, minAmountOut, fmt.Sprintf("%d", poolId), denomA, swapWalletAddr)

	keepPeriodCountDown := time.NewTimer(initialization.TWAPPruningKeepPeriod)

	// Make sure that we are still producing blocks and move far enough for the swap TWAP record to be created
	// so that we can measure start time post-swap (timeAfterSwap).
	chainAB.WaitForNumHeights(2)

	// Measure time after swap and wait for a few blocks to be produced.
	// This is needed to ensure that start time is before the block time
	// when we query for TWAP.
	timeAfterSwap := chainABNode.QueryLatestBlockTime()
	chainAB.WaitForNumHeights(2)

	// TWAP "from before to after swap" should be different from "from before to before swap"
	// because swap introduces a new record with a different spot price.
	s.T().Log("querying for the TWAP from before swap to now after swap")
	twapFromBeforeSwapToAfterSwapAB, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToAfterSwapBC, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().NoError(err)
	twapFromBeforeSwapToAfterSwapCA, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	s.Require().NoError(err)
	// We had a swap of 2000000stake for some amount of uion,
	// 2000000uion for some amount of note, and
	// 2000000note for some amount of stake
	// Because we traded the same amount of all three assets, we expect the asset with the greatest
	// initial value (B, or uion) to have a largest negative price impact,
	// to the benefit (positive price impact) of the other two assets (A&C, or stake and note)
	s.Require().True(twapFromBeforeSwapToAfterSwapAB.GT(twapFromBeforeSwapToBeforeSwapOneAB))
	s.Require().True(twapFromBeforeSwapToAfterSwapBC.LT(twapFromBeforeSwapToBeforeSwapOneBC))
	s.Require().True(twapFromBeforeSwapToAfterSwapCA.GT(twapFromBeforeSwapToBeforeSwapOneCA))

	s.T().Log("querying for the TWAP from after swap to now")
	twapFromAfterToNowAB, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeAfterSwap)
	s.Require().NoError(err)
	twapFromAfterToNowBC, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeAfterSwap)
	s.Require().NoError(err)
	twapFromAfterToNowCA, err := chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeAfterSwap)
	s.Require().NoError(err)
	// Because twapFromAfterToNow has a higher time weight for the after swap period,
	// we expect the results to be flipped from the previous comparison to twapFromBeforeSwapToBeforeSwapOne
	s.Require().True(twapFromBeforeSwapToAfterSwapAB.LT(twapFromAfterToNowAB))
	s.Require().True(twapFromBeforeSwapToAfterSwapBC.GT(twapFromAfterToNowBC))
	s.Require().True(twapFromBeforeSwapToAfterSwapCA.LT(twapFromAfterToNowCA))

	s.T().Log("querying for the TWAP from after swap to after swap + 10ms")
	twapAfterSwapBeforePruning10MsAB, err := chainABNode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterSwapBeforePruning10MsBC, err := chainABNode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterSwapBeforePruning10MsCA, err := chainABNode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	// Again, because twapAfterSwapBeforePruning10Ms has a higher time weight for the after swap period between the two,
	// we expect no change in the inequality
	s.Require().True(twapFromBeforeSwapToAfterSwapAB.LT(twapAfterSwapBeforePruning10MsAB))
	s.Require().True(twapFromBeforeSwapToAfterSwapBC.GT(twapAfterSwapBeforePruning10MsBC))
	s.Require().True(twapFromBeforeSwapToAfterSwapCA.LT(twapAfterSwapBeforePruning10MsCA))

	// These must be equal because they are calculated over time ranges with the stable and equal spot price.
	// There are potential rounding errors requiring us to approximate the comparison.
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsAB, twapFromAfterToNowAB, osmomath.NewDecWithPrec(2, 3))
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsBC, twapFromAfterToNowBC, osmomath.NewDecWithPrec(2, 3))
	osmoassert.DecApproxEq(s.T(), twapAfterSwapBeforePruning10MsCA, twapFromAfterToNowCA, osmomath.NewDecWithPrec(2, 3))

	// Make sure that the pruning keep period has passed.
	s.T().Logf("waiting for pruning keep period of (%.f) seconds to pass", initialization.TWAPPruningKeepPeriod.Seconds())
	<-keepPeriodCountDown.C

	// Epoch end triggers the prunning of TWAP records.
	// Records before swap should be pruned.
	chainAB.WaitForNumEpochs(1, epochIdentifier)

	// We should not be able to get TWAP before swap since it should have been pruned.
	s.T().Log("pruning is now complete, querying TWAP for period that should be pruned")
	_, err = chainABNode.QueryArithmeticTwapToNow(poolId, denomA, denomB, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")
	_, err = chainABNode.QueryArithmeticTwapToNow(poolId, denomB, denomC, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")
	_, err = chainABNode.QueryArithmeticTwapToNow(poolId, denomC, denomA, timeBeforeSwap)
	s.Require().ErrorContains(err, "too old")

	// TWAPs for the same time range should be the same when we query for them before and after pruning.
	s.T().Log("querying for TWAP for period before pruning took place but should not have been pruned")
	twapAfterPruning10msAB, err := chainABNode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterPruning10msBC, err := chainABNode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	twapAfterPruning10msCA, err := chainABNode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterSwap.Add(10*time.Millisecond))
	s.Require().NoError(err)
	s.Require().Equal(twapAfterSwapBeforePruning10MsAB, twapAfterPruning10msAB)
	s.Require().Equal(twapAfterSwapBeforePruning10MsBC, twapAfterPruning10msBC)
	s.Require().Equal(twapAfterSwapBeforePruning10MsCA, twapAfterPruning10msCA)

	// TWAP "from after to after swap" should equal to "from after swap to after pruning"
	// These must be equal because they are calculated over time ranges with the stable and equal spot price.
	timeAfterPruning := chainABNode.QueryLatestBlockTime()
	s.T().Log("querying for TWAP from after swap to after pruning")
	twapToNowPostPruningAB, err := chainABNode.QueryArithmeticTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	twapToNowPostPruningBC, err := chainABNode.QueryArithmeticTwap(poolId, denomB, denomC, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	twapToNowPostPruningCA, err := chainABNode.QueryArithmeticTwap(poolId, denomC, denomA, timeAfterSwap, timeAfterPruning)
	s.Require().NoError(err)
	// There are potential rounding errors requiring us to approximate the comparison.
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningAB, twapAfterSwapBeforePruning10MsAB, osmomath.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningBC, twapAfterSwapBeforePruning10MsBC, osmomath.NewDecWithPrec(1, 3))
	osmoassert.DecApproxEq(s.T(), twapToNowPostPruningCA, twapAfterSwapBeforePruning10MsCA, osmomath.NewDecWithPrec(1, 3))
}

func (s *IntegrationTestSuite) ExpeditedProposals() {
	chainAB, chainABNode := s.getChainCfgs()

	sender := chainABNode.GetWallet(initialization.ValidatorWalletName)
	govModuleAccount := chainABNode.QueryGovModuleAccount()
	propMetadata := []byte{42}
	validProp := fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.gov.v1.MsgExecLegacyContent",
			"authority": "%s",
			"content": {
				"@type": "/cosmos.gov.v1beta1.TextProposal",
				"title": "My awesome title",
				"description": "My awesome description"
			}
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"metadata": "%s",
	"deposit": "%s",
	"expedited": true
}`, govModuleAccount, base64.StdEncoding.EncodeToString(propMetadata), sdk.NewCoin("note", math.NewInt(5000000000)))

	propNumber := chainABNode.SubmitNewV1ProposalType(validProp, sender)

	totalTimeChan := make(chan time.Duration, 1)
	go chainABNode.QueryPropStatusTimed(propNumber, "PROPOSAL_STATUS_PASSED", totalTimeChan)

	chain.AllValsVoteOnProposal(chainAB, propNumber)

	// if querying proposal takes longer than timeoutPeriod, stop the goroutine and error
	var elapsed time.Duration
	timeoutPeriod := 2 * govPropTimeout
	select {
	case elapsed = <-totalTimeChan:
	case <-time.After(timeoutPeriod):
		err := fmt.Errorf("go routine took longer than %s", timeoutPeriod)
		s.Require().NoError(err)
	}

	// compare the time it took to reach pass status to expected expedited voting period
	expeditedVotingPeriodDuration := time.Duration(chainAB.ExpeditedVotingPeriod * float32(time.Second))
	timeDelta := elapsed - expeditedVotingPeriodDuration
	// ensure delta is within two seconds of expected time
	s.Require().Less(timeDelta, 2*time.Second)
	s.T().Logf("expeditedVotingPeriodDuration within two seconds of expected time: %v", timeDelta)
	close(totalTimeChan)
}

// TestGeometricTWAP tests geometric twap.
// It does the following:  creates a pool, queries twap, performs a swap , and queries twap again.
// Twap is expected to change after the swap.
// The pool is created with 1_000_000 note and 2_000_000 stake and equal weights.
// Assuming base asset is note, the initial twap is 2
// Upon swapping 1_000_000 note for stake, supply changes, making note less expensive.
// As a result of the swap, twap changes to 0.5.
// Note: do not use chain B in this test as it has taker fee set.
// This TWAP test depends on specific values that might be affected
// by the taker fee.
func (s *IntegrationTestSuite) GeometricTWAP() {
	const (
		// This pool contains 1_000_000 note and 2_000_000 stake.
		// Equals weights.
		poolFile   = "geometricPool.json"
		walletName = "geometric-twap-wallet"

		denomA = "note"  // 1_000_000 note
		denomB = "stake" // 2_000_000 stake

		minAmountOut = "1"
	)

	// Note: use chain A specifically as this is the chain where we do not
	// set taker fee.
	chainA, chainANode := s.getChainACfgs()

	sender := chainANode.GetWallet(initialization.ValidatorWalletName)

	// Triggers the creation of TWAP records.
	poolId := chainANode.CreateBalancerPool(poolFile, initialization.ValidatorWalletName)
	swapWalletAddr := chainANode.CreateWalletAndFund(walletName, []string{initialization.WalletFeeTokens.String()}, chainA)

	// We add 5 ms to avoid landing directly on block time in twap. If block time
	// is provided as start time, the latest spot price is used. Otherwise
	// interpolation is done.
	timeBeforeSwapPlus5ms := chainANode.QueryLatestBlockTime().Add(5 * time.Millisecond)
	s.T().Log("geometric twap, start time ", timeBeforeSwapPlus5ms)

	// Wait for the next height so that the requested twap
	// start time (timeBeforeSwap) is not equal to the block time.
	chainA.WaitUntilBlockTime(timeBeforeSwapPlus5ms.Add(time.Second * 3))

	s.T().Log("querying for the first geometric TWAP to now (before swap)")
	// Assume base = note, quote = stake
	// At pool creation time, the twap should be:
	// quote asset supply / base asset supply = 2_000_000 / 1_000_000 = 2
	curBlockTime := chainANode.QueryLatestBlockTime().Unix()
	s.T().Log("geometric twap, end time ", curBlockTime)

	initialTwapBOverA, err := chainANode.QueryGeometricTwapToNow(poolId, denomA, denomB, timeBeforeSwapPlus5ms)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewDec(2), initialTwapBOverA)

	// Assume base = stake, quote = note
	// At pool creation time, the twap should be:
	// quote asset supply / base asset supply = 1_000_000 / 2_000_000 = 0.5
	initialTwapAOverB, err := chainANode.QueryGeometricTwapToNow(poolId, denomB, denomA, timeBeforeSwapPlus5ms)
	s.Require().NoError(err)
	s.Require().Equal(osmomath.NewDecWithPrec(5, 1), initialTwapAOverB)

	coinAIn := fmt.Sprintf("1000000%s", denomA)
	chainANode.BankSend(coinAIn, sender, swapWalletAddr)

	s.T().Logf("performing swap of %s for %s", coinAIn, denomB)

	// stake out = stake supply * (1 - (note supply before / note supply after)^(note weight / stake weight))
	//           = 2_000_000 * (1 - (1_000_000 / 2_000_000)^1)
	//           = 2_000_000 * 0.5
	//           = 1_000_000
	chainANode.SwapExactAmountIn(coinAIn, minAmountOut, fmt.Sprintf("%d", poolId), denomB, swapWalletAddr)

	// New supply post swap:
	// stake = 2_000_000 - 1_000_000 - 1_000_000
	// note = 1_000_000 + 1_000_000 = 2_000_000

	timeAfterSwap := chainANode.QueryLatestBlockTime()
	chainA.WaitForNumHeights(1)
	timeAfterSwapPlus1Height := chainANode.QueryLatestBlockTime()
	chainA.WaitForNumHeights(1)
	s.T().Log("querying for the TWAP from after swap to now")
	afterSwapTwapBOverA, err := chainANode.QueryGeometricTwap(poolId, denomA, denomB, timeAfterSwap, timeAfterSwapPlus1Height)
	s.Require().NoError(err)

	// We swap note so note's supply will increase and stake will decrease.
	// The the price after will be smaller than the previous one.
	s.Require().True(initialTwapBOverA.GT(afterSwapTwapBOverA))

	// Assume base = note, quote = stake
	// At pool creation, we had:
	// quote asset supply / base asset supply = 2_000_000 / 1_000_000 = 2
	// Next, we swapped 1_000_000 note for stake.
	// Now, we roughly have
	// uatom = 1_000_000
	// note = 2_000_000
	// quote asset supply / base asset supply = 1_000_000 / 2_000_000 = 0.5
	osmoassert.DecApproxEq(s.T(), osmomath.NewDecWithPrec(5, 1), afterSwapTwapBOverA, osmomath.NewDecWithPrec(1, 2))
}

// Only set taker fee on chain B as some tests depend on the exact swap values.
// For example, Geometric twap. As a result, we use chain A for these tests.
//
// Similarly, CL tests depend on taker fee being set.
// As a result, we deterministically configure chain B's taker fee prior to running CL tests.
func (s *IntegrationTestSuite) SetDefaultTakerFeeChainB() {
	chainB, chainBNode := s.getChainBCfgs()
	err := chainBNode.ParamChangeProposal("poolmanager", string(poolmanagertypes.KeyDefaultTakerFee), json.RawMessage(`"0.001500000000000000"`), chainB, true)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) SetExpeditedVotingPeriodChainA() {
	chainA, chainANode := s.getChainACfgs()

	sender := chainANode.GetWallet(initialization.ValidatorWalletName)
	govModuleAccount := chainANode.QueryGovModuleAccount()
	propMetadata := []byte{42}
	validProp := fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.gov.v1.MsgUpdateParams",
			"authority": "%s",
			"params": {
				"min_deposit": [
					{
					"denom": "note",
					"amount": "10000000"
					}
				],
				"max_deposit_period": "172800s",
				"voting_period": "11s",
				"quorum": "0.334000000000000000",
				"threshold": "0.500000000000000000",
				"veto_threshold": "0.334000000000000000",
				"min_initial_deposit_ratio": "0.000000000000000000",
				"expedited_voting_period": "7s",
				"expedited_threshold": "0.667000000000000000",
				"expedited_min_deposit": [
				{
					"denom": "note",
					"amount": "50000000"
				}
				],
				"burn_vote_quorum": false,
				"burn_proposal_deposit_prevote": false,
				"burn_vote_veto": true
			}
		}
	],
	"title": "Gov params update",
	"summary": "Gov params update description",
	"metadata": "%s",
	"deposit": "%s",
	"expedited": false
}`, govModuleAccount, base64.StdEncoding.EncodeToString(propMetadata), sdk.NewCoin("note", math.NewInt(10000000)))

	proposalID := chainANode.SubmitNewV1ProposalType(validProp, sender)

	chain.AllValsVoteOnProposal(chainA, proposalID)

	s.Eventually(func() bool {
		status, err := chainANode.QueryPropStatus(proposalID)
		if err != nil {
			return false
		}
		return status == "PROPOSAL_STATUS_PASSED"
	}, time.Minute*2, 10*time.Millisecond)
}

func (s *IntegrationTestSuite) SetExpeditedVotingPeriodChainB() {
	chainB, chainBNode := s.getChainBCfgs()

	sender := chainBNode.GetWallet(initialization.ValidatorWalletName)
	govModuleAccount := chainBNode.QueryGovModuleAccount()
	propMetadata := []byte{42}
	validProp := fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.gov.v1.MsgUpdateParams",
			"authority": "%s",
			"params": {
				"min_deposit": [
					{
					"denom": "note",
					"amount": "10000000"
					}
				],
				"max_deposit_period": "172800s",
				"voting_period": "11s",
				"quorum": "0.334000000000000000",
				"threshold": "0.500000000000000000",
				"veto_threshold": "0.334000000000000000",
				"min_initial_deposit_ratio": "0.000000000000000000",
				"expedited_voting_period": "7s",
				"expedited_threshold": "0.667000000000000000",
				"expedited_min_deposit": [
				{
					"denom": "note",
					"amount": "50000000"
				}
				],
				"burn_vote_quorum": false,
				"burn_proposal_deposit_prevote": false,
				"burn_vote_veto": true
			}
		}
	],
	"title": "Gov params update",
	"summary": "Gov params update description",
	"metadata": "%s",
	"deposit": "%s",
	"expedited": false
}`, govModuleAccount, base64.StdEncoding.EncodeToString(propMetadata), sdk.NewCoin("note", math.NewInt(10000000)))

	proposalID := chainBNode.SubmitNewV1ProposalType(validProp, sender)

	chain.AllValsVoteOnProposal(chainB, proposalID)

	s.Eventually(func() bool {
		status, err := chainBNode.QueryPropStatus(proposalID)
		if err != nil {
			return false
		}
		return status == "PROPOSAL_STATUS_PASSED"
	}, time.Minute*2, 10*time.Millisecond)
}
