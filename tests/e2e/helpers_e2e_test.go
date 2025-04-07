package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/configurer/chain"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/initialization"
	"github.com/osmosis-labs/osmosis/v27/tests/e2e/util"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
)

var defaultFeePerTx = osmomath.NewInt(1000)

// Get balances for address
func (s *IntegrationTestSuite) addrBalance(node *chain.NodeConfig, address string) sdk.Coins {
	addrBalances, err := node.QueryBalances(address)
	s.Require().NoError(err)
	return addrBalances
}

var currentNodeIndexA int

func (s *IntegrationTestSuite) getChainACfgs() (*chain.Config, *chain.NodeConfig) {
	chainA := s.configurer.GetChainConfig(0)
	chainANodes := chainA.GetAllChainNodes()
	chosenNode := chainANodes[currentNodeIndexA]
	currentNodeIndexA = (currentNodeIndexA + 1) % len(chainANodes)
	return chainA, chosenNode
}

var currentNodeIndexB int

func (s *IntegrationTestSuite) getChainBCfgs() (*chain.Config, *chain.NodeConfig) {
	chainB := s.configurer.GetChainConfig(1)
	chainBNodes := chainB.GetAllChainNodes()
	chosenNode := chainBNodes[currentNodeIndexB]
	currentNodeIndexB = (currentNodeIndexB + 1) % len(chainBNodes)
	return chainB, chosenNode
}

var useChainA bool

func (s *IntegrationTestSuite) getChainCfgs() (*chain.Config, *chain.NodeConfig) {
	if useChainA {
		useChainA = false
		return s.getChainACfgs()
	} else {
		useChainA = true
		return s.getChainBCfgs()
	}
}

// Helper function for calculating uncollected spread rewards since the time that spreadRewardGrowthInsideLast corresponds to
// positionLiquidity - current position liquidity
// spreadRewardGrowthBelow - spread reward growth below lower tick
// spreadRewardGrowthAbove - spread reward growth above upper tick
// spreadRewardGrowthInsideLast - amount of spread reward growth inside range at the time from which we want to calculate the amount of uncollected spread rewards
// spreadRewardGrowthGlobal - variable for tracking global spread reward growth
func calculateUncollectedSpreadRewards(positionLiquidity, spreadRewardGrowthBelow, spreadRewardGrowthAbove, spreadRewardGrowthInsideLast osmomath.Dec, spreadRewardGrowthGlobal osmomath.Dec) osmomath.Dec {
	// Calculating spread reward growth inside range [-1200; 400]
	spreadRewardGrowthInside := calculateSpreadRewardGrowthInside(spreadRewardGrowthGlobal, spreadRewardGrowthBelow, spreadRewardGrowthAbove)

	// Calculating uncollected spread rewards
	// Formula for finding uncollected spread rewards in time range [t1; t2]:
	// F_u = position_liquidity * (spread_rewards_growth_inside_t2 - spread_rewards_growth_inside_t1).
	spreadRewardsUncollected := positionLiquidity.Mul(spreadRewardGrowthInside.Sub(spreadRewardGrowthInsideLast))

	return spreadRewardsUncollected
}

func (s *IntegrationTestSuite) updatedCFMMPool(node *chain.NodeConfig, poolId uint64) gammtypes.CFMMPoolI {
	cfmmPool, err := node.QueryCFMMPool(poolId)
	s.Require().NoError(err)
	return cfmmPool
}

func formatCLIInt(i int) string {
	if i < 0 {
		return fmt.Sprintf("[%d]", i)
	}
	return strconv.Itoa(i)
}

func (s *IntegrationTestSuite) CallCheckBalance(node *chain.NodeConfig, addr, denom string, amount int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.CheckBalance(node, addr, denom, amount)
}

// CheckBalance Checks the balance of an address
func (s *IntegrationTestSuite) CheckBalance(node *chain.NodeConfig, addr, denom string, amount int64) {
	// check the balance of the contract
	s.Require().Eventually(func() bool {
		// TODO: Change to QueryBalance(addr, denom)
		balance, err := node.QueryBalances(addr)
		s.Require().NoError(err)
		if len(balance) == 0 {
			return false
		}
		// check that the amount is in one of the balances inside the balance list
		for _, b := range balance {
			if b.Denom == denom && b.Amount.Int64() == amount {
				return true
			}
		}
		return false
	},
		1*time.Minute,
		10*time.Millisecond,
	)
}

func (s *IntegrationTestSuite) UploadAndInstantiateCounter(chain *chain.Config) string {
	// copy the contract from tests/ibc-hooks/bytecode
	wd, err := os.Getwd()
	s.NoError(err)
	// co up two levels
	projectDir := filepath.Dir(filepath.Dir(wd))
	_, err = util.CopyFile(projectDir+"/tests/ibc-hooks/bytecode/counter.wasm", wd+"/scripts/counter.wasm")
	s.NoError(err)
	node, err := s.configurer.GetChainConfig(0).GetNodeAtIndex(0)
	s.NoError(err)

	codeId := node.StoreWasmCode("counter.wasm", initialization.ValidatorWalletName)
	node.InstantiateWasmContract(
		strconv.Itoa(codeId),
		`{"count": 0}`,
		initialization.ValidatorWalletName)

	contracts, err := node.QueryContractsFromId(codeId)
	s.NoError(err)
	s.Require().Len(contracts, 1, "Wrong number of contracts for the counter")
	contractAddr := contracts[0]
	return contractAddr
}

func (s *IntegrationTestSuite) getChainIndex(chain *chain.Config) int {
	if chain.Id == "melody-test-a" {
		return 0
	} else {
		return 1
	}
}

func runFuncsInParallelAndBlock(funcs []func()) {
	var wg sync.WaitGroup
	wg.Add(len(funcs))
	for _, f := range funcs {
		go func(g func()) {
			defer wg.Done()
			g()
		}(f)
	}
	wg.Wait()
}
