package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/protorev/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient

	pools          []Pool
	balances       []sdk.Coin
	searcherRoutes []types.SearcherRoutes
}

type Pool struct {
	Asset1  string
	Asset2  string
	Amount1 sdk.Int
	Amount2 sdk.Int
	SwapFee sdk.Dec
	ExitFee sdk.Dec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)

	// Init balances for the accounts
	suite.balances = []sdk.Coin{
		sdk.NewCoin("akash", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("bitcoin", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("canto", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("ethereum", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin(types.AtomDenomination, sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("juno", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(9000000000000000000)),
	}
	suite.fundAllAccountsWith()
	suite.Commit()

	// Init pools
	suite.setUpPools()
	suite.Commit()

	// Init search routes
	suite.setUpSearcherRoutes()
	suite.Commit()
}

// SetUpOsmoPools sets up the Osmo pools for testing
// This creates 5 assets and pools between them all
// akash <-> types.OsmosisDenomination
// juno <-> types.OsmosisDenomination
// ethereum <-> types.OsmosisDenomination
// bitcoin <-> types.OsmosisDenomination
// canto <-> types.OsmosisDenomination
// akash <-> types.AtomDenomination
// juno <-> types.AtomDenomination
// ethereum <-> types.AtomDenomination
// bitcoin <-> types.AtomDenomination
// canto <-> types.AtomDenomination
// types.OsmosisDenomination <-> types.AtomDenomination
// akash <-> juno
// akash <-> ethereum
// akash <-> bitcoin
// akash <-> canto
// juno <-> ethereum
// juno <-> bitcoin
// juno <-> canto
// ethereum <-> bitcoin
// ethereum <-> canto
// bitcoin <-> canto
func (suite *KeeperTestSuite) setUpPools() {
	// Init pools
	suite.pools = []Pool{
		{ // Pool 1
			Asset1:  "akash",
			Asset2:  types.AtomDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 2
			Asset1:  "juno",
			Asset2:  types.AtomDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 3
			Asset1:  "ethereum",
			Asset2:  types.AtomDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 4
			Asset1:  "bitcoin",
			Asset2:  types.AtomDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 5
			Asset1:  "canto",
			Asset2:  types.AtomDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 6
			Asset1:  types.OsmosisDenomination,
			Asset2:  types.AtomDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 7
			Asset1:  "akash",
			Asset2:  types.OsmosisDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 8
			Asset1:  "juno",
			Asset2:  types.OsmosisDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 9
			Asset1:  "ethereum",
			Asset2:  types.OsmosisDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 10
			Asset1:  "bitcoin",
			Asset2:  types.OsmosisDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 11
			Asset1:  "canto",
			Asset2:  types.OsmosisDenomination,
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 12
			Asset1:  "akash",
			Asset2:  "juno",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 13
			Asset1:  "akash",
			Asset2:  "ethereum",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 14
			Asset1:  "akash",
			Asset2:  "bitcoin",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 15
			Asset1:  "akash",
			Asset2:  "canto",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 16
			Asset1:  "juno",
			Asset2:  "ethereum",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 17
			Asset1:  "juno",
			Asset2:  "bitcoin",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 18
			Asset1:  "juno",
			Asset2:  "canto",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 19
			Asset1:  "ethereum",
			Asset2:  "bitcoin",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 20
			Asset1:  "ethereum",
			Asset2:  "canto",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
		{ // Pool 21
			Asset1:  "bitcoin",
			Asset2:  "canto",
			Amount1: sdk.NewInt(1000),
			Amount2: sdk.NewInt(1000),
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
		},
	}

	for _, pool := range suite.pools {
		suite.createGAMMPool(pool.Asset1, pool.Asset2, pool.Amount1, pool.Amount2, pool.SwapFee, pool.ExitFee)
	}

	// Set all of the pool info into the stores
	suite.App.AppKeepers.ProtoRevKeeper.EpochHooks().AfterEpochEnd(suite.Ctx, "", 1)
}

func (suite *KeeperTestSuite) createGAMMPool(token1, token2 string, token1Amount, token2Amount sdk.Int, swapFee, exitFee sdk.Dec) uint64 {
	poolAssets := []balancertypes.PoolAsset{
		{
			Token: sdk.Coin{
				Denom:  token1,
				Amount: token1Amount,
			},
			Weight: sdk.NewInt(1),
		},
		{
			Token: sdk.Coin{
				Denom:  token2,
				Amount: token2Amount,
			},
			Weight: sdk.NewInt(1),
		},
	}

	poolParams := balancertypes.PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}

	return suite.prepareCustomBalancerPool(poolAssets, poolParams)
}

func (suite *KeeperTestSuite) prepareCustomBalancerPool(
	poolAssets []balancertypes.PoolAsset,
	poolParams balancer.PoolParams,
) uint64 {
	poolID, err := suite.App.GAMMKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[1], poolParams, poolAssets, ""),
	)
	suite.Require().NoError(err)

	return poolID
}

func (suite *KeeperTestSuite) fundAllAccountsWith() {
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, suite.balances)
	}
}

// SetUpRoutes sets up the searcher routes for testing
func (suite *KeeperTestSuite) setUpSearcherRoutes() {
	suite.searcherRoutes = []types.SearcherRoutes{
		{
			TokenA: "akash",
			TokenB: types.AtomDenomination,
			Routes: []*types.Route{
				{
					Pools: []uint64{0, 14, 4}, // akash/atom, akash/bitcoin, bitcoin/atom
				},
				{
					Pools: []uint64{0, 13, 3}, // akash/atom, akash/bitcoin, bitcoin/atom
				},
			},
		},
		{
			TokenA: "juno",
			TokenB: types.OsmosisDenomination,
			Routes: []*types.Route{
				{
					Pools: []uint64{7, 12, 0}, // osmo/akash, akash/juno, osmo/juno
				},
			},
		},
		{
			TokenA: "canto",
			TokenB: types.AtomDenomination,
			Routes: []*types.Route{
				{
					Pools: []uint64{11, 0, 6}, // osmo/canto, canto/atom, atom/osmo
				},
			},
		},
	}

	for _, hotRoutes := range suite.searcherRoutes {
		suite.App.AppKeepers.ProtoRevKeeper.SetSearcherRoutes(suite.Ctx, hotRoutes.TokenA, hotRoutes.TokenB, &hotRoutes)
	}
}
