package concentrated_liquidity_test

import (
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	osmoapp "github.com/osmosis-labs/osmosis/v15/app"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodule "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/clmodule"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

type singlePoolGenesisEntry struct {
	pool     model.Pool
	tick     []genesis.FullTick
	positons []model.Position
}

var (
	baseGenesis = genesis.GenesisState{
		Params: types.Params{
			AuthorizedTickSpacing: []uint64{1, 10, 50},
			AuthorizedSwapFees:    []sdk.Dec{sdk.MustNewDecFromStr("0.0001"), sdk.MustNewDecFromStr("0.0003"), sdk.MustNewDecFromStr("0.0005")}},
		PoolData: []*genesis.PoolData{},
	}
	testCoins    = sdk.NewDecCoins(cl.HundredFooCoins)
	testTickInfo = model.TickInfo{
		LiquidityGross:   sdk.OneDec(),
		LiquidityNet:     sdk.OneDec(),
		FeeGrowthOutside: testCoins,
		UptimeTrackers: []model.UptimeTracker{
			{
				UptimeGrowthOutside: testCoins,
			},
		},
	}
	defaultFullTick = genesis.FullTick{
		PoolId:    defaultPoolId,
		TickIndex: 0,
		Info:      testTickInfo,
	}
	testPositionModel = model.Position{
		PoolId:         1,
		Address:        testAddressOne.String(),
		Liquidity:      sdk.OneDec(),
		LowerTick:      -1,
		UpperTick:      100,
		JoinTime:       defaultBlockTime,
		FreezeDuration: DefaultFreezeDuration,
	}
)

func positionWithPoolId(position model.Position, poolId uint64) model.Position {
	position.PoolId = poolId
	return position
}

// setupGenesis initializes the GenesisState with the given poolGenesisEntries data.
// It returns an updated GenesisState after processing the input data.
//
// baseGenesis is the initial GenesisState.
// poolGenesisEntries is a slice of singlePoolGenesisEntry structures, each containing data
// for a single pool (the pool itself, its ticks, and positions).
//
// The function iterates over the poolGenesisEntries, and for each entry, it creates a new Any type using
// the pool's data, then appends a new PoolData structure containing the pool and its corresponding
// ticks to the baseGenesis.PoolData. It also appends the corresponding positions to the
// baseGenesis.Positions.
func setupGenesis(baseGenesis genesis.GenesisState, poolGenesisEntries []singlePoolGenesisEntry) genesis.GenesisState {
	for _, poolGenesisEntry := range poolGenesisEntries {
		poolCopy := poolGenesisEntry.pool
		poolAny, err := codectypes.NewAnyWithValue(&poolCopy)
		if err != nil {
			panic(err)
		}
		baseGenesis.PoolData = append(baseGenesis.PoolData, &genesis.PoolData{
			Pool:  poolAny,
			Ticks: poolGenesisEntry.tick,
		})
		baseGenesis.Positions = append(baseGenesis.Positions, poolGenesisEntry.positons...)

	}
	return baseGenesis
}

// TestInitGenesis tests the InitGenesis function of the ConcentratedLiquidityKeeper.
// It checks that the state is initialized correctly based on the provided genesis.
func (s *KeeperTestSuite) TestInitGenesis() {
	s.Setup()
	poolE := s.PrepareConcentratedPool()
	poolOne, ok := poolE.(*model.Pool)
	s.Require().True(ok)

	poolE = s.PrepareConcentratedPool()
	poolTwo, ok := poolE.(*model.Pool)
	s.Require().True(ok)

	testCase := []struct {
		name                   string
		genesis                genesis.GenesisState
		expectedPools          []model.Pool
		expectedTicksPerPoolId map[uint64][]genesis.FullTick
		expectedPositions      []model.Position
	}{
		{
			name: "one pool, one position, two ticks",
			genesis: setupGenesis(baseGenesis, []singlePoolGenesisEntry{
				{
					pool: *poolOne,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -10),
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 10),
					},
					positons: []model.Position{testPositionModel},
				},
			}),
			expectedPools: []model.Pool{
				*poolOne,
			},
			expectedTicksPerPoolId: map[uint64][]genesis.FullTick{
				1: {
					withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -10),
					withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 10),
				},
			},
			expectedPositions: []model.Position{testPositionModel},
		},
		{
			name: "two pools, two positions, one tick pool one, two ticks pool two",
			genesis: setupGenesis(baseGenesis, []singlePoolGenesisEntry{
				{
					pool: *poolOne,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -1234),
					},
					positons: []model.Position{testPositionModel},
				},
				{
					pool: *poolTwo,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 0),
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 999),
					},
					positons: []model.Position{positionWithPoolId(testPositionModel, 2)},
				},
			}),
			expectedPools: []model.Pool{
				*poolOne,
				*poolTwo,
			},
			expectedTicksPerPoolId: map[uint64][]genesis.FullTick{
				1: {
					withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -1234),
				},
				2: {
					withTickIndex(withPoolId(defaultFullTick, poolTwo.Id), 0),
					withTickIndex(withPoolId(defaultFullTick, poolTwo.Id), 999),
				},
			},
			expectedPositions: []model.Position{testPositionModel, positionWithPoolId(testPositionModel, 2)},
		},
	}

	for _, tc := range testCase {
		tc := tc

		s.Run(tc.name, func() {
			// This erases previously created pools.
			s.Setup()

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			clKeeper.InitGenesis(ctx, tc.genesis)

			// Check params
			clParamsAfterInitialization := clKeeper.GetParams(ctx)
			s.Require().Equal(tc.genesis.Params.String(), clParamsAfterInitialization.String())

			clPoolsAfterInitialization, err := clKeeper.GetAllPools(ctx)
			s.Require().NoError(err)

			// Check pools
			s.Require().Equal(len(clPoolsAfterInitialization), len(tc.genesis.PoolData))
			for i, actualPoolI := range clPoolsAfterInitialization {
				actualPool, ok := actualPoolI.(*model.Pool)
				s.Require().True(ok)
				s.Require().Equal(tc.expectedPools[i], *actualPool)

				expectedTicks, ok := tc.expectedTicksPerPoolId[actualPool.Id]
				s.Require().True(ok)

				actualTicks, err := clKeeper.GetAllInitializedTicksForPool(ctx, actualPool.Id)
				s.Require().NoError(err)

				// Validate ticks.
				s.validateTicks(expectedTicks, actualTicks)
			}

			// get all positions.
			positions, err := clKeeper.GetAllPositions(ctx)
			s.Require().NoError(err)

			// Validate positions
			s.Require().Equal(tc.expectedPositions, positions)
		})
	}
}

// TestExportGenesis tests the ExportGenesis function of the ConcentratedLiquidityKeeper.
// It checks that the correct genesis state is returned.
func (s *KeeperTestSuite) TestExportGenesis() {
	s.Setup()

	poolE := s.PrepareConcentratedPool()
	poolOne, ok := poolE.(*model.Pool)
	s.Require().True(ok)

	poolE = s.PrepareConcentratedPool()
	poolTwo, ok := poolE.(*model.Pool)
	s.Require().True(ok)

	testCase := []struct {
		name    string
		genesis genesis.GenesisState
	}{
		{
			name: "one pool, one position, two ticks",
			genesis: setupGenesis(baseGenesis, []singlePoolGenesisEntry{
				{
					pool: *poolOne,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -10),
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 10),
					},
					positons: []model.Position{testPositionModel},
				},
			}),
		},
		{
			name: "two pools, two positions, one tick pool one, two ticks pool two",
			genesis: setupGenesis(baseGenesis, []singlePoolGenesisEntry{
				{
					pool: *poolOne,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -1234),
					},
					positons: []model.Position{testPositionModel},
				},
				{
					pool: *poolTwo,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolTwo.Id), 0),
						withTickIndex(withPoolId(defaultFullTick, poolTwo.Id), 999),
					},
					positons: []model.Position{positionWithPoolId(testPositionModel, 2)},
				},
			}),
		},
	}

	for _, tc := range testCase {
		tc := tc

		s.Run(tc.name, func() {
			s.Setup()

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx
			expectedGenesis := tc.genesis

			clKeeper.InitGenesis(ctx, tc.genesis)

			// Export the genesis state.
			actualExported := clKeeper.ExportGenesis(ctx)

			// Validate params.
			s.Require().Equal(expectedGenesis.Params.String(), actualExported.Params.String())

			// Validate pools and ticks.
			s.Require().Equal(len(expectedGenesis.PoolData), len(actualExported.PoolData))
			for i, actualPoolData := range actualExported.PoolData {
				expectedPoolData := expectedGenesis.PoolData[i]
				s.Require().Equal(expectedPoolData.Pool, actualPoolData.Pool)

				s.validateTicks(expectedPoolData.Ticks, actualPoolData.Ticks)
			}

			// Validate positions.
			s.Require().Equal(tc.genesis.Positions, actualExported.Positions)
		})
	}
}

// TestMarshalUnmarshalGenesis tests the MarshalUnmarshalGenesis functions of the ConcentratedLiquidityKeeper.
// It checks that the exported genesis can be marshaled and unmarshaled without panicking.
func TestMarshalUnmarshalGenesis(t *testing.T) {
	// Set up the app and context
	app := osmoapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := ctx.BlockTime()
	ctx = ctx.WithBlockTime(now.Add(time.Second))

	// Create an app module for the ConcentratedLiquidityKeeper
	encodingConfig := osmoapp.MakeEncodingConfig()
	appCodec := encodingConfig.Marshaler
	appModule := clmodule.NewAppModule(appCodec, *app.ConcentratedLiquidityKeeper)

	// Export the genesis state
	genesisExported := appModule.ExportGenesis(ctx, appCodec)

	// Test that the exported genesis can be marshaled and unmarshaled without panicking
	assert.NotPanics(t, func() {
		app := osmoapp.Setup(false)
		ctx := app.BaseApp.NewContext(false, tmproto.Header{})
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		am := clmodule.NewAppModule(appCodec, *app.ConcentratedLiquidityKeeper)
		am.InitGenesis(ctx, appCodec, genesisExported)
	})
}

func (s *KeeperTestSuite) validateTicks(expectedTicks []genesis.FullTick, actualTicks []genesis.FullTick) {
	s.Require().Equal(len(expectedTicks), len(actualTicks))
	for i, tick := range actualTicks {
		s.Require().Equal(expectedTicks[i].PoolId, tick.PoolId, "tick (%d) pool ids are not equal", i)
		s.Require().Equal(expectedTicks[i].TickIndex, tick.TickIndex, "tick (%d) pool indexes are not equal", i)
		s.Require().Equal(expectedTicks[i].Info, tick.Info, "tick (%d) infos are not equal", i)
	}
}
