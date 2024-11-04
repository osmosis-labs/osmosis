package keeper_test

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/protorev"
	protorevkeeper "github.com/osmosis-labs/osmosis/v27/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"

	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/stableswap"

	osmosisapp "github.com/osmosis-labs/osmosis/v27/app"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
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
	PoolAssets   []balancer.PoolAsset
	Asset1       string
	Asset2       string
	Amount1      osmomath.Int
	Amount2      osmomath.Int
	SpreadFactor osmomath.Dec
	ExitFee      osmomath.Dec
	PoolId       uint64
}

type StableSwapPool struct {
	initialLiquidity sdk.Coins
	poolParams       stableswap.PoolParams
	scalingFactors   []uint64
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupNoPools() {
	s.Setup()
	s.setupParams()

	queryHelper := baseapp.NewQueryServerTestHelper(s.Ctx, s.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, protorevkeeper.NewQuerier(*s.App.AppKeepers.ProtoRevKeeper))
	s.queryClient = types.NewQueryClient(queryHelper)
}

func (s *KeeperTestSuite) SetupTest() {
	s.SetupNoPools()
}

func (s *KeeperTestSuite) SetupPoolsTest() {
	s.Setup()
	s.setupParams()

	encodingConfig := osmosisapp.GetEncodingConfig()
	s.clientCtx = client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithCodec(encodingConfig.Marshaler)

	// Set default configuration for testing
	s.balances = sdk.NewCoins(
		sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("Atom", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("akash", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("bitcoin", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("canto", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("ethereum", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("juno", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/0E43EDE2E2A3AFA36D0CD38BDDC0B49FECA64FA426A82E102F304E430ECF46EE", osmomath.NewIntFromBigInt(big.NewInt(1).Mul(big.NewInt(9000000000000000000), big.NewInt(10000)))),
		sdk.NewCoin("ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("usdc", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("usdt", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("busd", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("test/1", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("test/2", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("usdx", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("usdy", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("epochOne", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("epochTwo", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("hookGamm", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("hookCL", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("hook", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("eth", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("gamm/pool/1", osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin(apptesting.DefaultTransmuterDenomA, osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin(apptesting.DefaultTransmuterDenomB, osmomath.NewInt(9000000000000000000)),
		sdk.NewCoin("stake", osmomath.NewInt(9000000000000000000)),
	)
	s.fundAllAccountsWith()

	// Init pools
	s.setUpPools()
	s.Commit()

	// Init search routes
	s.setUpTokenPairRoutes()
	s.Commit()

	queryHelper := baseapp.NewQueryServerTestHelper(s.Ctx, s.App.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, protorevkeeper.NewQuerier(*s.App.AppKeepers.ProtoRevKeeper))
	s.queryClient = types.NewQueryClient(queryHelper)
}

func (s *KeeperTestSuite) setupParams() {
	// Genesis on init should be the same as the default genesis
	exportDefaultGenesis := s.App.ProtoRevKeeper.ExportGenesis(s.Ctx)
	s.Require().Equal(exportDefaultGenesis, types.DefaultGenesis())

	// Init module state for testing (params may differ from default params)
	s.App.ProtoRevKeeper.SetProtoRevEnabled(s.Ctx, true)
	s.App.ProtoRevKeeper.SetDaysSinceModuleGenesis(s.Ctx, 0)
	s.App.ProtoRevKeeper.SetLatestBlockHeight(s.Ctx, uint64(s.Ctx.BlockHeight()))
	s.App.ProtoRevKeeper.SetPointCountForBlock(s.Ctx, 0)

	// Configure max pool points per block. This roughly correlates to the ms of execution time protorev will
	// take per block
	if err := s.App.ProtoRevKeeper.SetMaxPointsPerBlock(s.Ctx, 100); err != nil {
		panic(err)
	}
	// Configure max pool points per tx. This roughly correlates to the ms of execution time protorev will take
	// per tx
	if err := s.App.ProtoRevKeeper.SetMaxPointsPerTx(s.Ctx, 18); err != nil {
		panic(err)
	}

	// Set the Admin Account
	s.adminAccount = apptesting.CreateRandomAccounts(1)[0]
	err := protorev.HandleSetProtoRevAdminAccount(s.Ctx, *s.App.ProtoRevKeeper, &types.SetProtoRevAdminAccountProposal{Account: s.adminAccount.String()})
	s.Require().NoError(err)

	// Configure the initial base denoms used for cyclic route building
	baseDenomPriorities := []types.BaseDenom{
		{
			Denom:    types.OsmosisDenomination,
			StepSize: osmomath.NewInt(1_000_000),
		},
		{
			Denom:    "Atom",
			StepSize: osmomath.NewInt(1_000_000),
		},
		{
			Denom:    "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7",
			StepSize: osmomath.NewInt(1_000_000),
		},
	}
	err = s.App.ProtoRevKeeper.SetBaseDenoms(s.Ctx, baseDenomPriorities)
	s.Require().NoError(err)
}

// setUpPools sets up the pools needed for testing
// This creates several assets and pools between most of them (used in testing throughout the module)
// akash <-> types.OsmosisDenomination
// juno <-> types.OsmosisDenomination
// ethereum <-> types.OsmosisDenomination
// bitcoin <-> types.OsmosisDenomination
// canto <-> types.OsmosisDenomination
// and so on....
func (s *KeeperTestSuite) setUpPools() {
	// Create any necessary osmomath.Ints that require string conversion
	pool28Amount1, ok := osmomath.NewIntFromString("6170367464346955818920")
	s.Require().True(ok)

	// Init pools
	s.pools = []Pool{
		{ // Pool 1
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       1,
		},
		{ // Pool 2
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       2,
		},
		{ // Pool 3
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ethereum", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       3,
		},
		{ // Pool 4
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("bitcoin", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       4,
		},
		{ // Pool 5
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("canto", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       5,
		},
		{ // Pool 6
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       6,
		},
		{ // Pool 7
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       7,
		},
		{ // Pool 8
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       8,
		},
		{ // Pool 9
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ethereum", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       9,
		},
		{ // Pool 10
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("bitcoin", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       10,
		},
		{ // Pool 11
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("canto", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       11,
		},
		{ // Pool 12
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("juno", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       12,
		},
		{ // Pool 13
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("ethereum", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       13,
		},
		{ // Pool 14
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("bitcoin", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       14,
		},
		{ // Pool 15
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("akash", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("canto", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       15,
		},
		{ // Pool 16
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("ethereum", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       16,
		},
		{ // Pool 17
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("bitcoin", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       17,
		},
		{ // Pool 18
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("juno", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("canto", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       18,
		},
		{ // Pool 19
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ethereum", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("bitcoin", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       19,
		},
		{ // Pool 20
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ethereum", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("canto", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       20,
		},
		{ // Pool 21
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("bitcoin", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("canto", osmomath.NewInt(1000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(0, 2),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       21,
		},
		{ // Pool 22
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", osmomath.NewInt(18986995439401)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(191801648570)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       22,
		},
		{ // Pool 23
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0", osmomath.NewInt(72765460013038)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", osmomath.NewInt(596032233122)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(535, 5),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       23,
		},
		{ // Pool 24
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0", osmomath.NewInt(165624820984787)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(13901565323)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       24,
		},
		{ // Pool 25
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(165624820984787)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(139015653231902)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       25,
		},
		{ // Pool 26
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", osmomath.NewInt(13305396712237)),
					Weight: osmomath.NewInt(50),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(171274446980)),
					Weight: osmomath.NewInt(50),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       26,
		},
		{ // Pool 27
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", osmomath.NewInt(15766179414665)),
					Weight: osmomath.NewInt(50),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(13466662920841)),
					Weight: osmomath.NewInt(50),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       27,
		},
		{ // Pool 28
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0E43EDE2E2A3AFA36D0CD38BDDC0B49FECA64FA426A82E102F304E430ECF46EE", pool28Amount1),
					Weight: osmomath.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4", osmomath.NewInt(6073813312)),
					Weight: osmomath.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", osmomath.NewInt(403568175601)),
					Weight: osmomath.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", osmomath.NewInt(6120465766)),
					Weight: osmomath.NewInt(25),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(4, 4),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       28,
		},
		{ // Pool 29
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("usdc", osmomath.NewInt(2000000000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       29,
		},
		{ // Pool 30
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("busd", osmomath.NewInt(1000000000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       30,
		},
		{ // Pool 31
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0E43EDE2E2A3AFA36D0CD38BDDC0B49FECA64FA426A82E102F304E430ECF46EE", pool28Amount1), // Amount didn't change on mainnet
					Weight: osmomath.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/8242AD24008032E457D2E12D46588FD39FB54FB29680C6C7663D296B383C37C4", osmomath.NewInt(6073813312)),
					Weight: osmomath.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", osmomath.NewInt(403523315860)),
					Weight: osmomath.NewInt(25),
				},
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(6121181710)),
					Weight: osmomath.NewInt(25),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(4, 4),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       31,
		},
		{ // Pool 32
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293", osmomath.NewInt(23583984695)),
					Weight: osmomath.NewInt(70),
				},
				{
					Token:  sdk.NewCoin("ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC", osmomath.NewInt(381295003769)),
					Weight: osmomath.NewInt(30),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(3, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       32,
		},
		{ // Pool 33
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293", osmomath.NewInt(41552173575)),
					Weight: osmomath.NewInt(70),
				},
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(10285796639)),
					Weight: osmomath.NewInt(30),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(3, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       33,
		},
		{ // Pool 34
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(364647340206)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("test/1", osmomath.NewInt(1569764554938)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(3, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       34,
		},
		{ // Pool 35
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("test/1", osmomath.NewInt(1026391517901)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1694086377216)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       35,
		},
		{ // Pool 36
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(2774812791932)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("test/2", osmomath.NewInt(1094837653970)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(3, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       36,
		},
		{ // Pool 37
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("Atom", osmomath.NewInt(40616571954500000)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("test/2", osmomath.NewInt(109588793167300000)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(3, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       37,
		},
		{ // Pool 38
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(6111815027)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", osmomath.NewInt(4478366578)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       38,
		},
		{ // Pool 39
			PoolAssets: []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", osmomath.NewInt(18631000485558)),
					Weight: osmomath.NewInt(1),
				},
				{
					Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(17000185817963)),
					Weight: osmomath.NewInt(1),
				},
			},
			SpreadFactor: osmomath.NewDecWithPrec(2, 3),
			ExitFee:      osmomath.NewDecWithPrec(0, 2),
			PoolId:       39,
		},
	}

	for _, pool := range s.pools {
		s.createGAMMPool(pool.PoolAssets, pool.SpreadFactor, pool.ExitFee)
	}

	s.stableSwapPools = []StableSwapPool{
		{ // Pool 40
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdc", osmomath.NewInt(1000000000000000)),
				sdk.NewCoin("usdt", osmomath.NewInt(1000000000000000)),
				sdk.NewCoin("busd", osmomath.NewInt(1000000000000000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1, 1},
		},
		{ // Pool 41 - Used for doomsday testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdc", osmomath.NewInt(1000000000000000)),
				sdk.NewCoin("usdt", osmomath.NewInt(1000000000000000)),
				sdk.NewCoin("busd", osmomath.NewInt(2000000000000000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1, 1},
		},
		{ // Pool 42 - Used for extended range testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdx", osmomath.NewInt(1000000000000000)),
				sdk.NewCoin("usdy", osmomath.NewInt(2000000000000000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 43 - Used for extended range testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdx", osmomath.NewInt(2000000000000000)),
				sdk.NewCoin("usdy", osmomath.NewInt(1000000000000000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 44 - Used for panic catching testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdx", osmomath.NewInt(1000)),
				sdk.NewCoin("usdy", osmomath.NewInt(2000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 45 - Used for panic catching testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("usdx", osmomath.NewInt(2000)),
				sdk.NewCoin("usdy", osmomath.NewInt(1000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 46 - Used for epoch_hook UpdateHighestLiquidityPool testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("epochOne", osmomath.NewInt(1000)),
				sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 47 - Used for epoch_hook UpdateHighestLiquidityPool testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("epochOne", osmomath.NewInt(1000)),
				sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(2000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 48 - Used for epoch_hook UpdateHighestLiquidityPool testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin("epochTwo", osmomath.NewInt(1000)),
				sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(1, 4),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
		{ // Pool 49 - Used for CL testing
			initialLiquidity: sdk.NewCoins(
				sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10_000_000_000_000)),
				sdk.NewCoin("epochTwo", osmomath.NewInt(8_000_000_000_000)),
			),
			poolParams: stableswap.PoolParams{
				SwapFee: osmomath.NewDecWithPrec(0, 2),
				ExitFee: osmomath.NewDecWithPrec(0, 2),
			},
			scalingFactors: []uint64{1, 1},
		},
	}

	for _, pool := range s.stableSwapPools {
		s.createStableswapPool(pool.initialLiquidity, pool.poolParams, pool.scalingFactors)
	}

	// Create a concentrated liquidity pool for epoch_hook testing
	// Pool 50
	s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("epochTwo", appparams.BaseCoinUnit)

	// Create a cosmwasm pool for testing
	// Pool 51
	cwPool := s.PrepareCustomTransmuterPool(s.TestAccs[0], []string{"Atom", "test/2"})
	cwFunds := sdk.NewCoins(sdk.NewCoin("Atom", osmomath.NewInt(1000000000000)), sdk.NewCoin("test/2", osmomath.NewInt(1000000000000)))
	s.JoinTransmuterPool(s.TestAccs[0], cwPool.GetId(), cwFunds)

	// Add the new cosmwasm pool to the pool info
	poolInfo := types.DefaultPoolTypeInfo
	poolInfo.Cosmwasm.WeightMaps = []types.WeightMap{
		{
			ContractAddress: cwPool.GetContractAddress(),
			Weight:          4,
		},
	}
	s.App.ProtoRevKeeper.SetInfoByPoolType(s.Ctx, poolInfo)

	// Create a duplicate pool for testing
	// Pool 52
	s.createGAMMPool(
		[]balancer.PoolAsset{
			{
				Token:  sdk.NewCoin("Atom", osmomath.NewInt(10000)),
				Weight: osmomath.NewInt(1),
			},
			{
				Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(10000)),
				Weight: osmomath.NewInt(1),
			},
		},
		osmomath.NewDecWithPrec(2, 3),
		osmomath.NewDecWithPrec(0, 2),
	)

	// Create a duplicate pool for testing
	// Pool 53
	s.createGAMMPool(
		[]balancer.PoolAsset{
			{
				Token:  sdk.NewCoin("usdc", osmomath.NewInt(10000)),
				Weight: osmomath.NewInt(1),
			},
			{
				Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(10000)),
				Weight: osmomath.NewInt(1),
			},
		},
		osmomath.NewDecWithPrec(2, 3),
		osmomath.NewDecWithPrec(0, 2),
	)

	// Create a duplicate pool for testing
	// Pool 54
	s.createGAMMPool(
		[]balancer.PoolAsset{
			{
				Token:  sdk.NewCoin("stake", osmomath.NewInt(10000)),
				Weight: osmomath.NewInt(1),
			},
			{
				Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(10000)),
				Weight: osmomath.NewInt(1),
			},
		},
		osmomath.NewDecWithPrec(2, 3),
		osmomath.NewDecWithPrec(0, 2),
	)

	// Create a duplicate pool for testing
	// Pool 55
	s.createGAMMPool(
		[]balancer.PoolAsset{
			{
				Token:  sdk.NewCoin("stake", osmomath.NewInt(100000000)),
				Weight: osmomath.NewInt(1),
			},
			{
				Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000000000)),
				Weight: osmomath.NewInt(1),
			},
		},
		osmomath.NewDecWithPrec(2, 3),
		osmomath.NewDecWithPrec(0, 2),
	)

	// Create a duplicate pool for testing
	// Pool 56
	s.createGAMMPool(
		[]balancer.PoolAsset{
			{
				Token:  sdk.NewCoin("bitcoin", osmomath.NewInt(100)),
				Weight: osmomath.NewInt(1),
			},
			{
				Token:  sdk.NewCoin("Atom", osmomath.NewInt(100)),
				Weight: osmomath.NewInt(1),
			},
		},
		osmomath.NewDecWithPrec(2, 3),
		osmomath.NewDecWithPrec(0, 2),
	)

	// Create a concentrated liquidity pool for range testing
	// Pool 58
	// Create the CL pool
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], "epochTwo", appparams.BaseCoinUnit, apptesting.DefaultTickSpacing, osmomath.ZeroDec())
	fundCoins := sdk.NewCoins(sdk.NewCoin("epochTwo", osmomath.NewInt(10_000_000_000_000)), sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10_000_000_000_000)))
	s.FundAcc(s.TestAccs[0], fundCoins)
	s.CreateFullRangePosition(clPool, fundCoins)

	// Create a concentrated liquidity pool for range testing
	// Pool 59
	// Create the CL pool
	clPool = s.PrepareCustomConcentratedPool(s.TestAccs[0], "epochTwo", appparams.BaseCoinUnit, apptesting.DefaultTickSpacing, osmomath.ZeroDec())
	fundCoins = sdk.NewCoins(sdk.NewCoin("epochTwo", osmomath.NewInt(2_000_000_000)), sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1_000_000_000)))
	s.FundAcc(s.TestAccs[0], fundCoins)
	s.CreateFullRangePosition(clPool, fundCoins)

	// Set all of the pool info into the stores
	err := s.App.ProtoRevKeeper.UpdatePools(s.Ctx)
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) CreateCLPoolAndArbRouteWith_28000_Ticks() {
	// Create the CL pool
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[2], "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", appparams.BaseCoinUnit, 100, osmomath.NewDecWithPrec(2, 3))
	fundCoins := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1000000000000000000)), sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", osmomath.NewInt(1000000000000000000)))
	s.FundAcc(s.TestAccs[2], fundCoins)

	// Create 28000 ticks in the CL pool, 14000 on each side
	tokensProvided := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100000)), sdk.NewCoin("ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", osmomath.NewInt(100000)))
	amount0Min := osmomath.NewInt(0)
	amount1Min := osmomath.NewInt(0)
	lowerTick := int64(0)
	upperTick := int64(100)

	for i := int64(0); i < 14000; i++ {
		s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPool.GetId(), s.TestAccs[2], tokensProvided, amount0Min, amount1Min, lowerTick-(100*i), upperTick-(100*i))
		s.App.ConcentratedLiquidityKeeper.CreatePosition(s.Ctx, clPool.GetId(), s.TestAccs[2], tokensProvided, amount0Min, amount1Min, lowerTick+(100*i), upperTick+(100*i))
	}

	// Set 2-pool hot route between new CL pool and respective Balancer
	s.App.ProtoRevKeeper.SetTokenPairArbRoutes(
		s.Ctx,
		appparams.BaseCoinUnit,
		"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7",
		types.NewTokenPairArbRoutes(
			[]types.Route{
				{
					Trades: []types.Trade{
						{
							Pool:     38,
							TokenIn:  appparams.BaseCoinUnit,
							TokenOut: "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7",
						},
						{
							Pool:     0,
							TokenIn:  "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7",
							TokenOut: appparams.BaseCoinUnit,
						},
					},
					StepSize: osmomath.NewInt(100000),
				},
			},
			appparams.BaseCoinUnit,
			"ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7",
		),
	)
}

// createStableswapPool creates a stableswap pool with the given pool assets and params
func (s *KeeperTestSuite) createStableswapPool(initialLiquidity sdk.Coins, poolParams stableswap.PoolParams, scalingFactors []uint64) uint64 {
	poolId, err := s.App.PoolManagerKeeper.CreatePool(
		s.Ctx,
		stableswap.NewMsgCreateStableswapPool(s.TestAccs[1], poolParams, initialLiquidity, scalingFactors, ""))
	s.Require().NoError(err)
	return poolId
}

// createGAMMPool creates a balancer pool with the given pool assets and params
func (s *KeeperTestSuite) createGAMMPool(poolAssets []balancer.PoolAsset, spreadFactor, exitFee osmomath.Dec) uint64 {
	poolParams := balancer.PoolParams{
		SwapFee: spreadFactor,
		ExitFee: exitFee,
	}

	return s.prepareCustomBalancerPool(poolAssets, poolParams)
}

// prepareCustomBalancerPool creates a custom balancer pool with the given pool assets and params
func (s *KeeperTestSuite) prepareCustomBalancerPool(
	poolAssets []balancer.PoolAsset,
	poolParams balancer.PoolParams,
) uint64 {
	poolID, err := s.App.PoolManagerKeeper.CreatePool(
		s.Ctx,
		balancer.NewMsgCreateBalancerPool(s.TestAccs[1], poolParams, poolAssets, ""),
	)
	s.Require().NoError(err)

	return poolID
}

// fundAllAccountsWith funds all the test accounts with the same amount of tokens
func (s *KeeperTestSuite) fundAllAccountsWith() {
	for _, acc := range s.TestAccs {
		s.FundAcc(acc, s.balances)
	}
}

// setUpTokenPairRoutes sets up the searcher routes for testing
func (s *KeeperTestSuite) setUpTokenPairRoutes() {
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
	twoPool0 := types.NewTrade(0, "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7", types.OsmosisDenomination)
	twoPool1 := types.NewTrade(39, types.OsmosisDenomination, "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7")

	// Doomsday Route - Stableswap
	doomsdayStable0 := types.NewTrade(29, types.OsmosisDenomination, "usdc")
	doomsdayStable1 := types.NewTrade(0, "usdc", "busd")
	doomsdayStable2 := types.NewTrade(30, "busd", types.OsmosisDenomination)

	standardStepSize := osmomath.NewInt(1_000_000)

	s.tokenPairArbRoutes = []types.TokenPairArbRoutes{
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
			TokenOut: "ibc/0CD3A0285E1341859B5E86B6AB7682F023D03E97607CCC1DC95706411D866DF7",
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

	for _, tokenPair := range s.tokenPairArbRoutes {
		err := tokenPair.Validate()
		s.Require().NoError(err)
		err = s.App.ProtoRevKeeper.SetTokenPairArbRoutes(s.Ctx, tokenPair.TokenIn, tokenPair.TokenOut, tokenPair)
		s.Require().NoError(err)
	}
}
