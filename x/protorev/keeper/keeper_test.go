package keeper_test

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	"github.com/osmosis-labs/osmosis/v15/x/protorev"
	protorevkeeper "github.com/osmosis-labs/osmosis/v15/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	balancertypes "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/stableswap"

	osmosisapp "github.com/osmosis-labs/osmosis/v15/app"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	clientCtx   client.Context
	queryClient types.QueryClient

	pools              []Pool
	stableSwapPools    []StableSwapPool
	balances           sdk.Coins
	tokenPairArbRoutes []types.TokenPairArbRoutes
	adminAccount       sdk.AccAddress
}

type Pool struct {
	PoolAssets []balancertypes.PoolAsset
	Asset1     string
	Asset2     string
	Amount1    sdk.Int
	Amount2    sdk.Int
	SwapFee    sdk.Dec
	ExitFee    sdk.Dec
	PoolId     uint64
}

type StableSwapPool struct {
	initialLiquidity sdk.Coins
	poolParams       stableswap.PoolParams
	scalingFactors   []uint64
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	// Genesis on init should be the same as the default genesis
	exportDefaultGenesis := suite.App.ProtoRevKeeper.ExportGenesis(suite.Ctx)
	suite.Require().Equal(exportDefaultGenesis, types.DefaultGenesis())

	// Init module state for testing (params may differ from default params)
	suite.App.ProtoRevKeeper.SetProtoRevEnabled(suite.Ctx, true)
	suite.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(suite.Ctx, 0)
	suite.App.ProtoRevKeeper.SetLatestBlockHeight(suite.Ctx, uint64(suite.Ctx.BlockHeight()))
	suite.App.ProtoRevKeeper.SetPointCountForBlock(suite.Ctx, 0)

	// Configure max pool points per block. This roughly correlates to the ms of execution time protorev will
	// take per block
	if err := suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, 100); err != nil {
		panic(err)
	}
	// Configure max pool points per tx. This roughly correlates to the ms of execution time protorev will take
	// per tx
	if err := suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, 18); err != nil {
		panic(err)
	}

	poolWeights := types.PoolWeights{
		StableWeight:       5, // it takes around 5 ms to simulate and execute a stable swap
		BalancerWeight:     2, // it takes around 2 ms to simulate and execute a balancer swap
		ConcentratedWeight: 2, // it takes around 2 ms to simulate and execute a concentrated swap
	}
	suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, poolWeights)

	// Configure the initial base denoms used for cyclic route building
	baseDenomPriorities := []types.BaseDenom{
		{
			Denom:    types.OsmosisDenomination,
			StepSize: sdk.NewInt(1_000_000),
		},
		{
			Denom:    "Atom",
			StepSize: sdk.NewInt(1_000_000),
		},
		{
			Denom:    "test/3",
			StepSize: sdk.NewInt(1_000_000),
		},
	}
	suite.App.ProtoRevKeeper.SetBaseDenoms(suite.Ctx, baseDenomPriorities)

	encodingConfig := osmosisapp.MakeEncodingConfig()
	suite.clientCtx = client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONCodec(encodingConfig.Marshaler)

	// Set default configuration for testing
	suite.balances = sdk.NewCoins(
		sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("Atom", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("akash", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("bitcoin", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("canto", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("ethereum", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("juno", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/0E43EDE2E2A3AFA36D0CD38BDDC0B49FECA64FA426A82E102F304E430ECF46EE", sdk.NewIntFromBigInt(big.NewInt(1).Mul(big.NewInt(9000000000000000000), big.NewInt(10000)))),
		sdk.NewCoin("ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("usdc", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("usdt", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("busd", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("test/1", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("test/2", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("test/3", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("usdx", sdk.NewInt(9000000000000000000)),
		sdk.NewCoin("usdy", sdk.NewInt(9000000000000000000)),
	)
	suite.fundAllAccountsWith()
	suite.Commit()

	// Init pools
	suite.setUpPools()
	suite.Commit()

	// Init search routes
	suite.setUpTokenPairRoutes()
	suite.Commit()

	// Set the Admin Account
	suite.adminAccount = apptesting.CreateRandomAccounts(1)[0]
	err := protorev.HandleSetProtoRevAdminAccount(suite.Ctx, *suite.App.ProtoRevKeeper, &types.SetProtoRevAdminAccountProposal{Account: suite.adminAccount.String()})
	suite.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.Ctx, suite.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, protorevkeeper.NewQuerier(*suite.App.AppKeepers.ProtoRevKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)
}

// setUpPools sets up the pools needed for testing
// This creates several assets and pools between most of them (used in testing throughout the module)
// akash <-> types.OsmosisDenomination
// juno <-> types.OsmosisDenomination
// ethereum <-> types.OsmosisDenomination
// bitcoin <-> types.OsmosisDenomination
// canto <-> types.OsmosisDenomination
// and so on....
func (suite *KeeperTestSuite) setUpPools() {
	// Create any necessary sdk.Ints that require string conversion
	pool28Amount1, ok := sdk.NewIntFromString("6170367464346955818920")
	suite.Require().True(ok)

	// Init pools
	suite.pools = []Pool{
		{ // Pool 1
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  1,
		},
		{ // Pool 2
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  2,
		},
		{ // Pool 3
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ethereum", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  3,
		},
		{ // Pool 4
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("bitcoin", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  4,
		},
		{ // Pool 5
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("canto", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  5,
		},
		{ // Pool 6
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  6,
		},
		{ // Pool 7
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  7,
		},
		{ // Pool 8
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  8,
		},
		{ // Pool 9
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ethereum", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  9,
		},
		{ // Pool 10
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("bitcoin", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  10,
		},
		{ // Pool 11
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("canto", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  11,
		},
		{ // Pool 12
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("juno", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  12,
		},
		{ // Pool 13
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("ethereum", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  13,
		},
		{ // Pool 14
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("bitcoin", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  14,
		},
		{ // Pool 15
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("canto", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  15,
		},
		{ // Pool 16
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("ethereum", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  16,
		},
		{ // Pool 17
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("bitcoin", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  17,
		},
		{ // Pool 18
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("canto", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  18,
		},
		{ // Pool 19
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ethereum", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("bitcoin", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  19,
		},
		{ // Pool 20
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ethereum", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("canto", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  20,
		},
		{ // Pool 21
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("bitcoin", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("canto", sdk.NewInt(1000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(0, 2),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  21,
		},
		{ // Pool 22
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", sdk.NewInt(18986995439401)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(191801648570)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  22,
		},
		{ // Pool 23
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0", sdk.NewInt(72765460013038)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", sdk.NewInt(596032233122)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(535, 5),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  23,
		},
		{ // Pool 24
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0", sdk.NewInt(165624820984787)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(13901565323)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  24,
		},
		{ // Pool 25
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(165624820984787)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(139015653231902)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  25,
		},
		{ // Pool 26
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", sdk.NewInt(13305396712237)),
					Weight: sdk.NewInt(50),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(171274446980)),
					Weight: sdk.NewInt(50),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  26,
		},
		{ // Pool 27
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", sdk.NewInt(15766179414665)),
					Weight: sdk.NewInt(50),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(13466662920841)),
					Weight: sdk.NewInt(50),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  27,
		},
		{ // Pool 28
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0E43EDE2E2A3AFA36D0CD38BDDC0B49FECA64FA426A82E102F304E430ECF46EE", pool28Amount1),
					Weight: sdk.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4", sdk.NewInt(6073813312)),
					Weight: sdk.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", sdk.NewInt(403568175601)),
					Weight: sdk.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", sdk.NewInt(6120465766)),
					Weight: sdk.NewInt(25),
				},
			},
			SwapFee: sdk.NewDecWithPrec(4, 4),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  28,
		},
		{ // Pool 29
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("usdc", sdk.NewInt(2000000000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  29,
		},
		{ // Pool 30
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("busd", sdk.NewInt(1000000000)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  30,
		},
		{ // Pool 31
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0E43EDE2E2A3AFA36D0CD38BDDC0B49FECA64FA426A82E102F304E430ECF46EE", pool28Amount1), // Amount didn't change on mainnet
					Weight: sdk.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4", sdk.NewInt(6073813312)),
					Weight: sdk.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", sdk.NewInt(403523315860)),
					Weight: sdk.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(6121181710)),
					Weight: sdk.NewInt(25),
				},
			},
			SwapFee: sdk.NewDecWithPrec(4, 4),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  31,
		},
		{ // Pool 32
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293", sdk.NewInt(23583984695)),
					Weight: sdk.NewInt(70),
				},
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", sdk.NewInt(381295003769)),
					Weight: sdk.NewInt(30),
				},
			},
			SwapFee: sdk.NewDecWithPrec(3, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  32,
		},
		{ // Pool 33
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293", sdk.NewInt(41552173575)),
					Weight: sdk.NewInt(70),
				},
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(10285796639)),
					Weight: sdk.NewInt(30),
				},
			},
			SwapFee: sdk.NewDecWithPrec(3, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  33,
		},
		{ // Pool 34
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(364647340206)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("test/1", sdk.NewInt(1569764554938)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(3, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  34,
		},
		{ // Pool 35
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("test/1", sdk.NewInt(1026391517901)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1694086377216)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  35,
		},
		{ // Pool 36
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(2774812791932)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("test/2", sdk.NewInt(1094837653970)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(3, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  36,
		},
		{ // Pool 37
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("Atom", sdk.NewInt(406165719545)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("test/2", sdk.NewInt(1095887931673)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(3, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  37,
		},
		{ // Pool 38
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(6111815027)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("test/3", sdk.NewInt(4478366578)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  38,
		},
		{ // Pool 39
			PoolAssets: []balancertypes.PoolAsset{
				{
					Token:  sdk.NewCoin("test/3", sdk.NewInt(18631000485558)),
					Weight: sdk.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(17000185817963)),
					Weight: sdk.NewInt(1),
				},
			},
			SwapFee: sdk.NewDecWithPrec(2, 3),
			ExitFee: sdk.NewDecWithPrec(0, 2),
			PoolId:  39,
		},
	}

	for _, pool := range suite.pools {
		suite.createGAMMPool(pool.PoolAssets, pool.SwapFee, pool.ExitFee)
	}

	suite.stableSwapPools = []StableSwapPool{
		{ // Pool 40
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdc", sdk.NewInt(1000000000000000)),
				sdk.NewCoin("usdt", sdk.NewInt(1000000000000000)),
				sdk.NewCoin("busd", sdk.NewInt(1000000000000000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 4),
				ExitFee: sdk.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1, 1},
		},
		{ // Pool 41 - Used for doomsday testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdc", sdk.NewInt(1000000000000000)),
				sdk.NewCoin("usdt", sdk.NewInt(1000000000000000)),
				sdk.NewCoin("busd", sdk.NewInt(2000000000000000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 4),
				ExitFee: sdk.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1, 1},
		},
		{ // Pool 42 - Used for extended range testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdx", sdk.NewInt(1000000000000000)),
				sdk.NewCoin("usdy", sdk.NewInt(2000000000000000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 4),
				ExitFee: sdk.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 43 - Used for extended range testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdx", sdk.NewInt(2000000000000000)),
				sdk.NewCoin("usdy", sdk.NewInt(1000000000000000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 4),
				ExitFee: sdk.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 44 - Used for panic catching testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdx", sdk.NewInt(1000)),
				sdk.NewCoin("usdy", sdk.NewInt(2000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 4),
				ExitFee: sdk.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 45 - Used for panic catching testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdx", sdk.NewInt(2000)),
				sdk.NewCoin("usdy", sdk.NewInt(1000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 4),
				ExitFee: sdk.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
	}

	for _, pool := range suite.stableSwapPools {
		suite.createStableswapPool(pool.initialLiquidity, pool.poolParams, pool.scalingFactors)
	}

	// Set all of the pool info into the stores
	suite.App.ProtoRevKeeper.UpdatePools(suite.Ctx)
}

// createStableswapPool creates a stableswap pool with the given pool assets and params
func (suite *KeeperTestSuite) createStableswapPool(initialLiquidity sdk.Coins, poolParams stableswap.PoolParams, scalingFactors []uint64) {
	_, err := suite.App.PoolManagerKeeper.CreatePool(
		suite.Ctx,
		stableswap.NewMsgCreateStableswapPool(suite.TestAccs[1], poolParams, initialLiquidity, scalingFactors, ""))
	suite.Require().NoError(err)
}

// createGAMMPool creates a balancer pool with the given pool assets and params
func (suite *KeeperTestSuite) createGAMMPool(poolAssets []balancertypes.PoolAsset, swapFee, exitFee sdk.Dec) uint64 {
	poolParams := balancertypes.PoolParams{
		SwapFee: swapFee,
		ExitFee: exitFee,
	}

	return suite.prepareCustomBalancerPool(poolAssets, poolParams)
}

// prepareCustomBalancerPool creates a custom balancer pool with the given pool assets and params
func (suite *KeeperTestSuite) prepareCustomBalancerPool(
	poolAssets []balancertypes.PoolAsset,
	poolParams balancer.PoolParams,
) uint64 {
	poolID, err := suite.App.PoolManagerKeeper.CreatePool(
		suite.Ctx,
		balancer.NewMsgCreateBalancerPool(suite.TestAccs[1], poolParams, poolAssets, ""),
	)
	suite.Require().NoError(err)

	return poolID
}

// fundAllAccountsWith funds all the test accounts with the same amount of tokens
func (suite *KeeperTestSuite) fundAllAccountsWith() {
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, suite.balances)
	}
}

// setUpTokenPairRoutes sets up the searcher routes for testing
func (suite *KeeperTestSuite) setUpTokenPairRoutes() {
	// General Test Route
	atomAkash := types.NewTrade(0, "Atom", "akash")
	akashBitcoin := types.NewTrade(14, "akash", "bitcoin")
	atomBitcoin := types.NewTrade(4, "bitcoin", "Atom")

	// Stableswap Route
	uosmoUSDC := types.NewTrade(0, types.OsmosisDenomination, "usdc")
	usdcBUSD := types.NewTrade(40, "usdc", "busd")
	busdUOSMO := types.NewTrade(30, "busd", types.OsmosisDenomination)

	// Atom Route
	atomIBC1 := types.NewTrade(31, "Atom", "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC")
	ibc1IBC2 := types.NewTrade(32, "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", "ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293")
	ibc2ATOM := types.NewTrade(0, "ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293", "Atom")

	// Four-Pool Route
	fourPool0 := types.NewTrade(34, "Atom", "test/1")
	fourPool1 := types.NewTrade(35, "test/1", types.OsmosisDenomination)
	fourPool2 := types.NewTrade(36, types.OsmosisDenomination, "test/2")
	fourPool3 := types.NewTrade(0, "test/2", "Atom")

	// Two-Pool Route
	twoPool0 := types.NewTrade(0, "test/3", types.OsmosisDenomination)
	twoPool1 := types.NewTrade(39, types.OsmosisDenomination, "test/3")

	// Doomsday Route - Stableswap
	doomsdayStable0 := types.NewTrade(29, types.OsmosisDenomination, "usdc")
	doomsdayStable1 := types.NewTrade(0, "usdc", "busd")
	doomsdayStable2 := types.NewTrade(30, "busd", types.OsmosisDenomination)

	standardStepSize := sdk.NewInt(1_000_000)

	suite.tokenPairArbRoutes = []types.TokenPairArbRoutes{
		{
			TokenIn:  "akash",
			TokenOut: "Atom",
			ArbRoutes: []types.Route{
				{
					StepSize: standardStepSize,
					Trades:   []types.Trade{atomAkash, akashBitcoin, atomBitcoin},
				},
			},
		},
		{
			TokenIn:  "usdc",
			TokenOut: types.OsmosisDenomination,
			ArbRoutes: []types.Route{
				{
					StepSize: standardStepSize,
					Trades:   []types.Trade{uosmoUSDC, usdcBUSD, busdUOSMO},
				},
			},
		},
		{
			TokenIn:  "Atom",
			TokenOut: "ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293",
			ArbRoutes: []types.Route{
				{
					StepSize: standardStepSize,
					Trades:   []types.Trade{atomIBC1, ibc1IBC2, ibc2ATOM},
				},
			},
		},
		{
			TokenIn:  "Atom",
			TokenOut: "test/2",
			ArbRoutes: []types.Route{
				{
					StepSize: standardStepSize,
					Trades:   []types.Trade{fourPool0, fourPool1, fourPool2, fourPool3},
				},
			},
		},
		{
			TokenIn:  types.OsmosisDenomination,
			TokenOut: "test/3",
			ArbRoutes: []types.Route{
				{
					StepSize: standardStepSize,
					Trades:   []types.Trade{twoPool0, twoPool1},
				},
			},
		},
		{
			TokenIn:  "busd",
			TokenOut: "usdc",
			ArbRoutes: []types.Route{
				{
					StepSize: standardStepSize,
					Trades:   []types.Trade{doomsdayStable0, doomsdayStable1, doomsdayStable2},
				},
			},
		},
	}

	for _, tokenPair := range suite.tokenPairArbRoutes {
		err := tokenPair.Validate()
		suite.Require().NoError(err)
		suite.App.ProtoRevKeeper.SetTokenPairArbRoutes(suite.Ctx, tokenPair.TokenIn, tokenPair.TokenOut, tokenPair)
	}
}
