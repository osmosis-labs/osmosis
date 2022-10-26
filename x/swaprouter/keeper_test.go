package swaprouter_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	osmoapp "github.com/osmosis-labs/osmosis/v12/app"
	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

const testExpectedPoolId = 3

var (
	testPoolCreationFee = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000_000_000)}
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
}

// CreateBalancerPoolsFromCoins creates balancer pools from given sets of coins.
// Where element 1 of the input corresponds to the first pool created,
// element 2 to the second pool created, up until the last element.
func (suite *KeeperTestSuite) CreateBalancerPoolsFromCoins(poolCoins []sdk.Coins) {
	for _, curPoolCoins := range poolCoins {
		suite.FundAcc(suite.TestAccs[0], curPoolCoins)
		suite.PrepareBalancerPoolWithCoins(curPoolCoins...)
	}
}

// TODO: refactor for this to be defined on the test suite
func TestInitGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.SwapRouterKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
		},
		NextPoolId: testExpectedPoolId,
	})

	require.Equal(t, uint64(testExpectedPoolId), app.SwapRouterKeeper.GetNextPoolIdAndIncrement(ctx))
	require.Equal(t, testPoolCreationFee, app.SwapRouterKeeper.GetParams(ctx).PoolCreationFee)
}

// TODO: refactor this to be defined on the test suite.
func TestExportGenesis(t *testing.T) {
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.SwapRouterKeeper.InitGenesis(ctx, &types.GenesisState{
		Params: types.Params{
			PoolCreationFee: testPoolCreationFee,
		},
		NextPoolId: testExpectedPoolId,
	})

	genesis := app.SwapRouterKeeper.ExportGenesis(ctx)
	require.Equal(t, uint64(testExpectedPoolId), genesis.NextPoolId)
	require.Equal(t, testPoolCreationFee, genesis.Params.PoolCreationFee)
}
