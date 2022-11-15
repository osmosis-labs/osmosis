package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

type poolsCase struct {
	poolName string
	poolId   uint64
}

type testCaseI interface {
	getName() string
	initializeNew(string, uint64) testCaseI
}

type shareInTestCase struct {
	name          string
	poolId        uint64
	shareInAmount sdk.Int
	expectedErr   error
}

var _ testCaseI = &shareInTestCase{}

func (tc shareInTestCase) getName() string {
	return tc.name
}

func (tc *shareInTestCase) initializeNew(name string, poolId uint64) testCaseI {
	newTc := *tc
	newTc.name = name
	newTc.poolId = poolId
	return &newTc
}

type tokensInTestCase struct {
	name        string
	poolId      uint64
	tokensIn    sdk.Coins
	expectedErr error
}

var _ testCaseI = &shareInTestCase{}

func (tc tokensInTestCase) getName() string {
	return tc.name
}

func (tc *tokensInTestCase) initializeNew(name string, poolId uint64) testCaseI {
	newTc := *tc
	newTc.name = name
	newTc.poolId = poolId
	return &newTc
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func (suite *KeeperTestSuite) prepareCustomBalancerPool(
	balances sdk.Coins,
	poolAssets []balancertypes.PoolAsset,
	poolParams balancer.PoolParams,
) uint64 {
	suite.fundAllAccountsWith(balances)

	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[0], poolParams, poolAssets, ""),
	)
	suite.Require().NoError(err)

	return poolID
}

func (suite *KeeperTestSuite) prepareCustomStableswapPool(
	balances sdk.Coins,
	poolParams stableswap.PoolParams,
	initialLiquidity sdk.Coins,
	scalingFactors []uint64,
) uint64 {
	suite.fundAllAccountsWith(balances)

	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		stableswap.NewMsgCreateStableswapPool(suite.TestAccs[0], poolParams, initialLiquidity, scalingFactors, ""),
	)
	suite.Require().NoError(err)

	return poolID
}

func (suite *KeeperTestSuite) fundAllAccountsWith(balances sdk.Coins) {
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, balances)
	}
}

// initializeValidTestPools returns a list of all supported pool models.
// Each pool model has a pool with id created and a name assigned.
func (suite *KeeperTestSuite) initializeValidTestPools() []poolsCase {
	return []poolsCase{
		{
			poolName: "balancer",
			poolId:   suite.PrepareBalancerPool(),
		},
		{
			poolName: "stableswap",
			poolId:   suite.PrepareBasicStableswapPool(),
		},
	}
}

// createTestCasesForGivenPools creates a test case in sharedTestCases for each pool in validPools
// and returns the final merged list of all test cases.
// This helps to reduce code duplciation when testing the same logic for multiple pool models.
func createTestCasesForGivenPools[T testCaseI](validPools []poolsCase, sharedTestCases []T) []T {
	allTestCases := make([]T, 0)
	for _, pool := range validPools {
		curPool := pool
		for _, sharedTestCase := range sharedTestCases {
			// Make a copy of a test case shared by all pools under test.
			// Update its name and set pool id.
			poolTestCase := sharedTestCase.initializeNew(fmt.Sprintf("%s - %s", curPool.poolName, sharedTestCase.getName()), curPool.poolId)
			allTestCases = append(allTestCases, poolTestCase.(T))
		}
	}
	return allTestCases
}
