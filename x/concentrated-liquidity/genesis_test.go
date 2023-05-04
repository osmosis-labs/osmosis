package concentrated_liquidity_test

import (
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	osmoapp "github.com/osmosis-labs/osmosis/v15/app"
	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	clmodule "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/clmodule"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types/genesis"
)

type singlePoolGenesisEntry struct {
	pool                  model.Pool
	tick                  []genesis.FullTick
	positionData          []genesis.PositionData
	feeAccumValues        genesis.AccumObject
	incentiveAccumulators []genesis.AccumObject
	incentiveRecords      []types.IncentiveRecord
}

var (
	baseGenesis = genesis.GenesisState{
		Params: types.Params{
			AuthorizedTickSpacing:        []uint64{1, 10, 100, 1000},
			AuthorizedSwapFees:           []sdk.Dec{sdk.MustNewDecFromStr("0.0001"), sdk.MustNewDecFromStr("0.0003"), sdk.MustNewDecFromStr("0.0005")},
			AuthorizedQuoteDenoms:        []string{ETH, USDC},
			BalancerSharesRewardDiscount: types.DefaultBalancerSharesDiscount,
			AuthorizedUptimes:            types.DefaultAuthorizedUptimes,
		},
		PoolData: []genesis.PoolData{},
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
		PositionId: 1,
		PoolId:     1,
		Address:    testAddressOne.String(),
		Liquidity:  sdk.OneDec(),
		LowerTick:  -1,
		UpperTick:  100,
		JoinTime:   defaultBlockTime,
	}
	testFeeAccumRecord = accum.Record{
		NumShares:        sdk.OneDec(),
		InitAccumValue:   sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
		UnclaimedRewards: sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(5))),
		Options:          nil,
	}
)

func positionWithPoolId(position model.Position, poolId uint64) *model.Position {
	position.PoolId = poolId
	return &position
}

func withPositionId(position model.Position, positionId uint64) *model.Position {
	position.PositionId = positionId
	return &position
}

func incentiveAccumsWithPoolId(poolId uint64) []genesis.AccumObject {
	return []genesis.AccumObject{
		{
			Name: types.KeyUptimeAccumulator(poolId, uint64(0)),
			AccumContent: &accum.AccumulatorContent{
				AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(20))),
				TotalShares: sdk.NewDec(20),
			},
		},
		{
			Name: types.KeyUptimeAccumulator(poolId, uint64(1)),
			AccumContent: &accum.AccumulatorContent{
				AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("bar", sdk.NewInt(20))),
				TotalShares: sdk.NewDec(30),
			},
		},
		{
			Name: types.KeyUptimeAccumulator(poolId, uint64(2)),
			AccumContent: &accum.AccumulatorContent{
				AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("baz", sdk.NewInt(10))),
				TotalShares: sdk.NewDec(10),
			},
		},
		{
			Name: types.KeyUptimeAccumulator(poolId, uint64(3)),
			AccumContent: &accum.AccumulatorContent{
				AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("qux", sdk.NewInt(20))),
				TotalShares: sdk.NewDec(20),
			},
		},
		{
			Name: types.KeyUptimeAccumulator(poolId, uint64(4)),
			AccumContent: &accum.AccumulatorContent{
				AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("quux", sdk.NewInt(20))),
				TotalShares: sdk.NewDec(20),
			},
		},
	}
}

// setupGenesis initializes the GenesisState with the given poolGenesisEntries data.
// It returns an updated GenesisState after processing the input data.
//
// baseGenesis is the initial GenesisState.
// poolGenesisEntries is a slice of singlePoolGenesisEntry structures, each containing data
// for a single pool (the pool itself, its ticks, positions, incentives records, accumulators and the next position ID).
//
// The function iterates over the poolGenesisEntries, and for each entry, it creates a new Any type using
// the pool's data, then appends a new PoolData structure containing the pool and its corresponding
// ticks to the baseGenesis.PoolData. It also appends the corresponding positions to the
// baseGenesis.Positions, along with the incentive records and accumulator values for fees and incentives.
func setupGenesis(baseGenesis genesis.GenesisState, poolGenesisEntries []singlePoolGenesisEntry) genesis.GenesisState {
	for _, poolGenesisEntry := range poolGenesisEntries {
		poolCopy := poolGenesisEntry.pool
		poolAny, err := codectypes.NewAnyWithValue(&poolCopy)
		if err != nil {
			panic(err)
		}
		baseGenesis.PoolData = append(baseGenesis.PoolData, genesis.PoolData{
			Pool:                   poolAny,
			Ticks:                  poolGenesisEntry.tick,
			FeeAccumulator:         poolGenesisEntry.feeAccumValues,
			IncentivesAccumulators: poolGenesisEntry.incentiveAccumulators,
			IncentiveRecords:       poolGenesisEntry.incentiveRecords,
		})
		baseGenesis.PositionData = append(baseGenesis.PositionData, poolGenesisEntry.positionData...)
		baseGenesis.NextPositionId = uint64(len(poolGenesisEntry.positionData))
	}
	return baseGenesis
}

// TestInitGenesis tests the InitGenesis function of the ConcentratedLiquidityKeeper.
// It checks that the state is initialized correctly based on the provided genesis.
func (s *KeeperTestSuite) TestInitGenesis() {
	s.SetupTest()
	poolE := s.PrepareConcentratedPool()
	poolOne, ok := poolE.(*model.Pool)
	s.Require().True(ok)

	poolE = s.PrepareConcentratedPool()
	poolTwo, ok := poolE.(*model.Pool)
	s.Require().True(ok)

	defaultTime1 := time.Unix(100, 100)
	defaultTime2 := time.Unix(300, 100)

	testCase := []struct {
		name                     string
		genesis                  genesis.GenesisState
		expectedPools            []model.Pool
		expectedTicksPerPoolId   map[uint64][]genesis.FullTick
		expectedPositionData     []genesis.PositionData
		expectedfeeAccumValues   []genesis.AccumObject
		expectedIncentiveRecords []types.IncentiveRecord
	}{
		{
			name: "one pool, one position, two ticks, one accumulator, two incentive records",
			genesis: setupGenesis(baseGenesis, []singlePoolGenesisEntry{
				{
					pool: *poolOne,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -10),
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 10),
					},
					positionData: []genesis.PositionData{
						{
							LockId:         1,
							Position:       &testPositionModel,
							FeeAccumRecord: testFeeAccumRecord,
						},
						{
							LockId:         0,
							Position:       withPositionId(testPositionModel, 2),
							FeeAccumRecord: testFeeAccumRecord,
						},
					},
					feeAccumValues: genesis.AccumObject{
						Name: types.KeyFeePoolAccumulator(1),
						AccumContent: &accum.AccumulatorContent{
							AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
							TotalShares: sdk.NewDec(10),
						},
					},
					incentiveAccumulators: incentiveAccumsWithPoolId(1),
					incentiveRecords: []types.IncentiveRecord{
						{
							PoolId:               uint64(1),
							IncentiveDenom:       "foo",
							IncentiveCreatorAddr: testAddressOne.String(),
							IncentiveRecordBody: types.IncentiveRecordBody{
								RemainingAmount: sdk.NewDec(5),
								EmissionRate:    sdk.NewDec(10),
								StartTime:       defaultTime1,
							},
							MinUptime: testUptimeOne,
						},
						{
							PoolId:               uint64(1),
							IncentiveDenom:       "bar",
							IncentiveCreatorAddr: testAddressTwo.String(),
							IncentiveRecordBody: types.IncentiveRecordBody{
								RemainingAmount: sdk.NewDec(15),
								EmissionRate:    sdk.NewDec(20),
								StartTime:       defaultTime2,
							},
							MinUptime: testUptimeOne,
						},
					},
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
			expectedPositionData: []genesis.PositionData{
				{
					LockId:         1,
					Position:       &testPositionModel,
					FeeAccumRecord: testFeeAccumRecord,
				},
				{
					LockId:         0,
					Position:       withPositionId(testPositionModel, 2),
					FeeAccumRecord: testFeeAccumRecord,
				},
			},
			expectedfeeAccumValues: []genesis.AccumObject{
				{
					Name: types.KeyFeePoolAccumulator(1),
					AccumContent: &accum.AccumulatorContent{
						AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
						TotalShares: sdk.NewDec(10),
					},
				},
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				{
					PoolId:               uint64(1),
					IncentiveDenom:       "bar",
					IncentiveCreatorAddr: testAddressTwo.String(),
					IncentiveRecordBody: types.IncentiveRecordBody{
						RemainingAmount: sdk.NewDec(15),
						EmissionRate:    sdk.NewDec(20),
						StartTime:       defaultTime2,
					},
					MinUptime: testUptimeOne,
				},
				{
					PoolId:               uint64(1),
					IncentiveDenom:       "foo",
					IncentiveCreatorAddr: testAddressOne.String(),
					IncentiveRecordBody: types.IncentiveRecordBody{
						RemainingAmount: sdk.NewDec(5),
						EmissionRate:    sdk.NewDec(10),
						StartTime:       defaultTime1,
					},
					MinUptime: testUptimeOne,
				},
			},
		},
		{
			name: "two pools, two positions, one tick pool one, two ticks pool two, two accumulators, one incentive records each",
			genesis: setupGenesis(baseGenesis, []singlePoolGenesisEntry{
				{
					pool: *poolOne,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -1234),
					},
					positionData: []genesis.PositionData{
						{
							LockId:         1,
							Position:       &testPositionModel,
							FeeAccumRecord: testFeeAccumRecord,
						},
					},
					feeAccumValues: genesis.AccumObject{
						Name: types.KeyFeePoolAccumulator(1),
						AccumContent: &accum.AccumulatorContent{
							AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
							TotalShares: sdk.NewDec(10),
						},
					},
					incentiveAccumulators: incentiveAccumsWithPoolId(1),
					incentiveRecords: []types.IncentiveRecord{
						{
							PoolId:               uint64(1),
							IncentiveDenom:       "foo",
							IncentiveCreatorAddr: testAddressOne.String(),
							IncentiveRecordBody: types.IncentiveRecordBody{
								RemainingAmount: sdk.NewDec(5),
								EmissionRate:    sdk.NewDec(10),
								StartTime:       defaultTime1,
							},
							MinUptime: testUptimeOne,
						},
					},
				},
				{
					pool: *poolTwo,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 0),
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 999),
					},
					positionData: []genesis.PositionData{
						{
							LockId:   2,
							Position: withPositionId(*positionWithPoolId(testPositionModel, 2), DefaultPositionId+1),
						},
					},

					feeAccumValues: genesis.AccumObject{
						Name: types.KeyFeePoolAccumulator(2),
						AccumContent: &accum.AccumulatorContent{
							AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("bar", sdk.NewInt(20))),
							TotalShares: sdk.NewDec(20),
						},
					},
					incentiveAccumulators: incentiveAccumsWithPoolId(2),
					incentiveRecords: []types.IncentiveRecord{
						{
							PoolId:               uint64(2),
							IncentiveDenom:       "bar",
							IncentiveCreatorAddr: testAddressOne.String(),
							IncentiveRecordBody: types.IncentiveRecordBody{
								RemainingAmount: sdk.NewDec(5),
								EmissionRate:    sdk.NewDec(10),
								StartTime:       defaultTime1,
							},
							MinUptime: testUptimeOne,
						},
					},
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
			expectedfeeAccumValues: []genesis.AccumObject{
				{
					Name: types.KeyFeePoolAccumulator(1),
					AccumContent: &accum.AccumulatorContent{
						AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
						TotalShares: sdk.NewDec(10),
					},
				},
				{
					Name: types.KeyFeePoolAccumulator(2),
					AccumContent: &accum.AccumulatorContent{
						AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("bar", sdk.NewInt(20))),
						TotalShares: sdk.NewDec(20),
					},
				},
			},
			expectedIncentiveRecords: []types.IncentiveRecord{
				{
					PoolId:               uint64(1),
					IncentiveDenom:       "foo",
					IncentiveCreatorAddr: testAddressOne.String(),
					IncentiveRecordBody: types.IncentiveRecordBody{
						RemainingAmount: sdk.NewDec(5),
						EmissionRate:    sdk.NewDec(10),
						StartTime:       defaultTime1,
					},
					MinUptime: testUptimeOne,
				},
				{
					PoolId:               uint64(2),
					IncentiveDenom:       "bar",
					IncentiveCreatorAddr: testAddressOne.String(),
					IncentiveRecordBody: types.IncentiveRecordBody{
						RemainingAmount: sdk.NewDec(5),
						EmissionRate:    sdk.NewDec(10),
						StartTime:       defaultTime1,
					},
					MinUptime: testUptimeOne,
				},
			},
			expectedPositionData: []genesis.PositionData{
				{
					LockId:         1,
					Position:       &testPositionModel,
					FeeAccumRecord: testFeeAccumRecord,
				},
				{
					LockId:         2,
					Position:       withPositionId(*positionWithPoolId(testPositionModel, 2), DefaultPositionId+1),
					FeeAccumRecord: testFeeAccumRecord,
				},
			},
		},
	}

	for _, tc := range testCase {
		tc := tc

		s.Run(tc.name, func() {
			// This erases previously created pools.
			s.SetupTest()

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx

			clKeeper.InitGenesis(ctx, tc.genesis)

			// Check params
			clParamsAfterInitialization := clKeeper.GetParams(ctx)
			s.Require().Equal(tc.genesis.Params.String(), clParamsAfterInitialization.String())

			clPoolsAfterInitialization, err := clKeeper.GetPools(ctx)
			s.Require().NoError(err)

			// Check pools
			feeAccums := []accum.AccumulatorObject{}
			incentiveRecords := []types.IncentiveRecord{}
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

				// get fee accumulator
				feeAccum, err := clKeeper.GetFeeAccumulator(s.Ctx, actualPool.GetId())
				s.Require().NoError(err)
				feeAccums = append(feeAccums, feeAccum)

				// check incentive accumulators
				acutalIncentiveAccums, err := clKeeper.GetUptimeAccumulators(ctx, actualPool.Id)
				s.Require().NoError(err)
				for j, actualIncentiveAccum := range acutalIncentiveAccums {
					expectedAccum := tc.genesis.PoolData[i].IncentivesAccumulators
					actualTotalShares, err := actualIncentiveAccum.GetTotalShares()
					s.Require().NoError(err)

					s.Require().Equal(expectedAccum[j].GetName(), actualIncentiveAccum.GetName())
					s.Require().Equal(expectedAccum[j].AccumContent.AccumValue, actualIncentiveAccum.GetValue())
					s.Require().Equal(expectedAccum[j].AccumContent.TotalShares, actualTotalShares)
				}

				// get incentive records for pool
				poolIncentiveRecords, err := clKeeper.GetAllIncentiveRecordsForPool(s.Ctx, actualPool.GetId())
				s.Require().NoError(err)
				incentiveRecords = append(incentiveRecords, poolIncentiveRecords...)
			}

			// get all positions.
			s.Require().NoError(err)
			var actualPositionData []genesis.PositionData
			for _, positionDataEntry := range tc.expectedPositionData {
				getPosition, err := clKeeper.GetPosition(ctx, positionDataEntry.Position.PositionId)
				s.Require().NoError(err)

				actualLockId := uint64(0)
				if positionDataEntry.LockId != 0 {
					actualLockId, err = clKeeper.GetLockIdFromPositionId(ctx, positionDataEntry.Position.PositionId)
					s.Require().NoError(err)
				} else {
					_, err = clKeeper.GetLockIdFromPositionId(ctx, positionDataEntry.Position.PositionId)
					s.Require().Error(err)
					s.Require().ErrorIs(err, types.PositionIdToLockNotFoundError{PositionId: positionDataEntry.Position.PositionId})
				}

				actualPositionData = append(actualPositionData, genesis.PositionData{
					LockId:         actualLockId,
					Position:       &getPosition,
					FeeAccumRecord: positionDataEntry.FeeAccumRecord,
				})
			}

			// Validate positions
			s.Require().Equal(tc.expectedPositionData, actualPositionData)

			// Validate accum objects
			s.Require().Equal(len(feeAccums), len(tc.expectedfeeAccumValues))
			for i, accumObject := range feeAccums {
				s.Require().Equal(feeAccums[i].GetValue(), tc.expectedfeeAccumValues[i].AccumContent.AccumValue)

				totalShares, err := accumObject.GetTotalShares()
				s.Require().NoError(err)
				s.Require().Equal(totalShares, tc.expectedfeeAccumValues[i].AccumContent.TotalShares)
			}

			// Validate incentive records
			s.Require().Equal(len(incentiveRecords), len(tc.expectedIncentiveRecords))
			for i, incentiveRecord := range incentiveRecords {
				s.Require().Equal(incentiveRecord.IncentiveCreatorAddr, tc.expectedIncentiveRecords[i].IncentiveCreatorAddr)
				s.Require().Equal(incentiveRecord.PoolId, tc.expectedIncentiveRecords[i].PoolId)
				s.Require().Equal(incentiveRecord.IncentiveDenom, tc.expectedIncentiveRecords[i].IncentiveDenom)
				s.Require().Equal(incentiveRecord.MinUptime, tc.expectedIncentiveRecords[i].MinUptime)
				s.Require().Equal(incentiveRecord.IncentiveRecordBody.EmissionRate.String(), tc.expectedIncentiveRecords[i].IncentiveRecordBody.EmissionRate.String())
				s.Require().Equal(incentiveRecord.IncentiveRecordBody.RemainingAmount.String(), tc.expectedIncentiveRecords[i].IncentiveRecordBody.RemainingAmount.String())
				s.Require().True(incentiveRecord.IncentiveRecordBody.StartTime.Equal(tc.expectedIncentiveRecords[i].IncentiveRecordBody.StartTime))
			}
			// Validate next position id.
			s.Require().Equal(tc.genesis.NextPositionId, clKeeper.GetNextPositionId(ctx))
		})
	}
}

// TestExportGenesis tests the ExportGenesis function of the ConcentratedLiquidityKeeper.
// It checks that the correct genesis state is returned.
func (s *KeeperTestSuite) TestExportGenesis() {
	s.SetupTest()

	poolE := s.PrepareConcentratedPool()
	poolOne, ok := poolE.(*model.Pool)
	s.Require().True(ok)

	poolE = s.PrepareConcentratedPool()
	poolTwo, ok := poolE.(*model.Pool)
	s.Require().True(ok)

	defaultTime1 := time.Unix(100, 100)
	defaultTime2 := time.Unix(300, 100)

	testCase := []struct {
		name    string
		genesis genesis.GenesisState
	}{
		{
			name: "one pool, one position, two ticks, one accumulator, two incentive records",
			genesis: setupGenesis(baseGenesis, []singlePoolGenesisEntry{
				{
					pool: *poolOne,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -10),
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), 10),
					},
					positionData: []genesis.PositionData{
						{
							LockId:         1,
							Position:       &testPositionModel,
							FeeAccumRecord: testFeeAccumRecord,
						},
					},
					feeAccumValues: genesis.AccumObject{
						Name: types.KeyFeePoolAccumulator(poolOne.Id),
						AccumContent: &accum.AccumulatorContent{
							AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
							TotalShares: sdk.NewDec(10),
						},
					},
					incentiveAccumulators: incentiveAccumsWithPoolId(1),
					incentiveRecords: []types.IncentiveRecord{
						{
							PoolId:               uint64(1),
							IncentiveDenom:       "bar",
							IncentiveCreatorAddr: testAddressTwo.String(),
							IncentiveRecordBody: types.IncentiveRecordBody{
								RemainingAmount: sdk.NewDec(15),
								EmissionRate:    sdk.NewDec(20),
								StartTime:       defaultTime2,
							},
							MinUptime: testUptimeOne,
						},
						{
							PoolId:               uint64(1),
							IncentiveDenom:       "foo",
							IncentiveCreatorAddr: testAddressOne.String(),
							IncentiveRecordBody: types.IncentiveRecordBody{
								RemainingAmount: sdk.NewDec(5),
								EmissionRate:    sdk.NewDec(10),
								StartTime:       defaultTime1,
							},
							MinUptime: testUptimeOne,
						},
					},
				},
			}),
		},
		{
			name: "two pools, two positions, one tick pool one, two ticks pool two, two accumulators, one incentive records each",
			genesis: setupGenesis(baseGenesis, []singlePoolGenesisEntry{
				{
					pool: *poolOne,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolOne.Id), -1234),
					},
					positionData: []genesis.PositionData{
						{
							LockId:         1,
							Position:       &testPositionModel,
							FeeAccumRecord: testFeeAccumRecord,
						},
						{
							LockId:         0,
							Position:       withPositionId(testPositionModel, DefaultPositionId+1),
							FeeAccumRecord: testFeeAccumRecord,
						},
					},
					feeAccumValues: genesis.AccumObject{
						Name: types.KeyFeePoolAccumulator(poolOne.Id),
						AccumContent: &accum.AccumulatorContent{
							AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("foo", sdk.NewInt(10))),
							TotalShares: sdk.NewDec(10),
						},
					},
					incentiveAccumulators: incentiveAccumsWithPoolId(1),
					incentiveRecords: []types.IncentiveRecord{
						{
							PoolId:               uint64(1),
							IncentiveDenom:       "foo",
							IncentiveCreatorAddr: testAddressOne.String(),
							IncentiveRecordBody: types.IncentiveRecordBody{
								RemainingAmount: sdk.NewDec(5),
								EmissionRate:    sdk.NewDec(10),
								StartTime:       defaultTime1,
							},
							MinUptime: testUptimeOne,
						},
					},
				},
				{
					pool: *poolTwo,
					tick: []genesis.FullTick{
						withTickIndex(withPoolId(defaultFullTick, poolTwo.Id), 0),
						withTickIndex(withPoolId(defaultFullTick, poolTwo.Id), 999),
					},
					feeAccumValues: genesis.AccumObject{
						Name: types.KeyFeePoolAccumulator(poolTwo.Id),
						AccumContent: &accum.AccumulatorContent{
							AccumValue:  sdk.NewDecCoins(sdk.NewDecCoin("bar", sdk.NewInt(20))),
							TotalShares: sdk.NewDec(20),
						},
					},
					incentiveAccumulators: incentiveAccumsWithPoolId(2),
					incentiveRecords: []types.IncentiveRecord{
						{
							PoolId:               uint64(2),
							IncentiveDenom:       "bar",
							IncentiveCreatorAddr: testAddressOne.String(),
							IncentiveRecordBody: types.IncentiveRecordBody{
								RemainingAmount: sdk.NewDec(5),
								EmissionRate:    sdk.NewDec(10),
								StartTime:       defaultTime1,
							},
							MinUptime: testUptimeOne,
						},
					},
					positionData: []genesis.PositionData{
						{
							LockId:         2,
							Position:       withPositionId(*positionWithPoolId(testPositionModel, 2), DefaultPositionId+2),
							FeeAccumRecord: testFeeAccumRecord,
						},
					},
				},
			}),
		},
	}

	for _, tc := range testCase {
		tc := tc

		s.Run(tc.name, func() {
			s.SetupTest()

			clKeeper := s.App.ConcentratedLiquidityKeeper
			ctx := s.Ctx
			expectedGenesis := tc.genesis

			// System Under test
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

				// validate fee accumulators
				s.Require().Equal(expectedPoolData.FeeAccumulator, actualPoolData.FeeAccumulator)

				// validate incentive accumulator
				for i, incentiveAccumulator := range actualPoolData.IncentivesAccumulators {
					s.Require().Equal(expectedPoolData.IncentivesAccumulators[i], incentiveAccumulator)
				}

				// Validate Incentive Records
				s.Require().Equal(len(expectedPoolData.IncentiveRecords), len(actualPoolData.IncentiveRecords))
				for i, incentiveRecord := range actualPoolData.IncentiveRecords {
					s.Require().Equal(incentiveRecord.IncentiveCreatorAddr, expectedPoolData.IncentiveRecords[i].IncentiveCreatorAddr)
					s.Require().Equal(incentiveRecord.IncentiveDenom, expectedPoolData.IncentiveRecords[i].IncentiveDenom)
					s.Require().Equal(incentiveRecord.PoolId, expectedPoolData.IncentiveRecords[i].PoolId)
					s.Require().Equal(incentiveRecord.MinUptime, expectedPoolData.IncentiveRecords[i].MinUptime)
					s.Require().Equal(incentiveRecord.IncentiveRecordBody.EmissionRate.String(), expectedPoolData.IncentiveRecords[i].IncentiveRecordBody.EmissionRate.String())
					s.Require().Equal(incentiveRecord.IncentiveRecordBody.RemainingAmount.String(), expectedPoolData.IncentiveRecords[i].IncentiveRecordBody.RemainingAmount.String())
					s.Require().True(incentiveRecord.IncentiveRecordBody.StartTime.Equal(expectedPoolData.IncentiveRecords[i].IncentiveRecordBody.StartTime))
				}

			}

			// Validate positions.
			s.Require().Equal(tc.genesis.PositionData, actualExported.PositionData)

			// Validate next position id.
			s.Require().Equal(tc.genesis.NextPositionId, actualExported.NextPositionId)
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
