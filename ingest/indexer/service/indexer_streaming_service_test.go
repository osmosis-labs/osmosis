package service_test

import (
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/rand"

	"github.com/osmosis-labs/sqs/sqsdomain"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v26/app/apptesting"
	indexerdomain "github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain"
	indexermocks "github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain/mocks"
	indexerservice "github.com/osmosis-labs/osmosis/v26/ingest/indexer/service"
	sqsmocks "github.com/osmosis-labs/osmosis/v26/ingest/sqs/domain/mocks"
	concentratedliquiditytypes "github.com/osmosis-labs/osmosis/v26/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v26/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v26/x/poolmanager/types"

	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	commonmocks "github.com/osmosis-labs/osmosis/v26/ingest/common/domain/mocks"
	"github.com/osmosis-labs/osmosis/v26/ingest/common/pooltracker"
)

var (
	emptyWriteListeners      = make(map[storetypes.StoreKey][]commondomain.WriteListener)
	emptyStoreKeyMap         = make(map[string]storetypes.StoreKey)
	defaultSpreadFactor      = "0.003000000000000000"
	defaultPoolDenoms        = [][]string{{"uosmo", "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"}}
	defaultTokenInDenom      = "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"
	defaultUnfoundTokenIn    = "0"
	defaultTokenInAmount     = 1000000000
	liquidityAttributePrefix = "liquidity_"
)

type IndexerServiceTestSuite struct {
	apptesting.ConcentratedKeeperTestHelper
}

func TestIndexerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(IndexerServiceTestSuite))
}

func (s *IndexerServiceTestSuite) TestAdjustTokenInAmountBySpreadFactor() {
	testCases := []struct {
		name                  string // Test case name
		eventType             string // Event type to be tested. Only gammtypes.TypeEvtTokenSwapped is valid
		havePoolIDAttribute   bool   // Decide whether pool_id attribute should be appended in the event attributes during test data preparation
		haveTokensInAttribute bool   // Decide whether tokens_in attribute should be appended in the event attributes during test data preparation
		useCLPoolID           bool   // Decide whether to use the pool_id from the concentrated liquidity pool or a random pool_id
		expectedError         bool   // Expected error flag
		expectedAdjustedAmt   bool   // Expected adjusted amount flag
	}{
		{
			name:                  "happy path",
			eventType:             gammtypes.TypeEvtTokenSwapped,
			havePoolIDAttribute:   true,
			haveTokensInAttribute: true,
			useCLPoolID:           true,
			expectedError:         false,
			expectedAdjustedAmt:   true,
		},
		{
			name:                  "non token_swapped event",
			eventType:             gammtypes.TypeEvtPoolJoined,
			havePoolIDAttribute:   true,
			haveTokensInAttribute: true,
			useCLPoolID:           true,
			expectedError:         false,
			expectedAdjustedAmt:   false,
		},
		{
			name:                  "use non existent pool_id",
			eventType:             gammtypes.TypeEvtTokenSwapped,
			havePoolIDAttribute:   true,
			haveTokensInAttribute: true,
			useCLPoolID:           false,
			expectedError:         true,
			expectedAdjustedAmt:   false,
		},
		{
			name:                  "no pool_id in event attributes",
			eventType:             gammtypes.TypeEvtTokenSwapped,
			havePoolIDAttribute:   false,
			haveTokensInAttribute: true,
			useCLPoolID:           true,
			expectedError:         true,
			expectedAdjustedAmt:   false,
		},
		{
			name:                  "no tokens_in in event attributes",
			eventType:             gammtypes.TypeEvtTokenSwapped,
			havePoolIDAttribute:   true,
			haveTokensInAttribute: false,
			useCLPoolID:           true,
			expectedError:         true,
			expectedAdjustedAmt:   false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// This test suite is to test the AdjustTokenInAmountBySpreadFactor method in the indexer streaming service.
			// where it applies to: token_swapped event only, i.e. gammtypes.TypeEvtTokenSwapped
			// It then looks for the pool_id, tokens_in attribute (concentratedliquiditytypes.AttributeKeyPoolId) in the
			// event attribute map.  With the pool_id, it then fetches the pool spread factor thru
			// the GetSpreadFactor function. The spread factor is then used to adjust the tokens_in amount in the attribute map.

			// Initialized chain pools
			s.PrepareAllSupportedPools()

			// Get all chain pools from state for asserting later
			concentratedPools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)

			s.Require().NoError(err)

			cfmmPools, err := s.App.GAMMKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)

			cosmWasmPools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
			s.Require().NoError(err)

			blockPools := commondomain.BlockPools{
				ConcentratedPools: concentratedPools,
				CFMMPools:         cfmmPools,
				CosmWasmPools:     cosmWasmPools,
			}

			transformedPools := []sqsdomain.PoolI{}
			for _, pool := range blockPools.GetAll() {
				// Note: balances are irrelevant for the test so we supply empty balances
				transformedPool := sqsdomain.NewPool(pool, pool.GetSpreadFactor(s.Ctx), sdk.Coins{})
				transformedPools = append(transformedPools, transformedPool)
			}

			// Initialize a mock block update process utils
			blockUpdatesProcessUtilsMock := &sqsmocks.BlockUpdateProcessUtilsMock{}

			// Initialize an empty pool tracker
			emptyPoolTracker := pooltracker.NewMemory()

			// Initialize a mock pool extractor
			poolExtractorMock := &sqsmocks.PoolsExtractorMock{
				BlockPools: commondomain.BlockPools{
					ConcentratedPools: concentratedPools,
					CFMMPools:         cfmmPools,
					CosmWasmPools:     cosmWasmPools,
				},
			}

			// Initialize a block process strategy manager
			blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

			// Initialize a concentrated pool with spread factor
			sf, _ := osmomath.NewDecFromStr(defaultSpreadFactor)
			s.CreateConcentratedPoolsAndFullRangePositionWithSpreadFactor(defaultPoolDenoms, []math.LegacyDec{sf})
			pools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			pool := pools[1]
			poolID := pool.GetId()

			// Initialize keepers
			keepers := indexerdomain.Keepers{
				PoolManagerKeeper: s.App.PoolManagerKeeper,
			}

			// Initialize tx decoder and logger
			txDecoder := s.App.GetTxConfig().TxDecoder()
			logger := s.App.Logger()

			// Initialize a mock publisher
			publisherMock := &indexermocks.PublisherMock{}

			// Initialize the node status checker mock
			nodeStatusCheckerMock := &commonmocks.NodeStatusCheckerMock{}

			indexerStreamingService := indexerservice.New(
				blockUpdatesProcessUtilsMock,
				blockProcessStrategyManager,
				publisherMock,
				emptyStoreKeyMap,
				poolExtractorMock,
				emptyPoolTracker,
				keepers,
				txDecoder,
				nodeStatusCheckerMock,
				logger)

			// Create the event based on the test cases attributes
			event := abcitypes.Event{
				Type: tc.eventType,
			}
			if tc.havePoolIDAttribute {
				event.Attributes = append(event.Attributes, abcitypes.EventAttribute{
					Key: concentratedliquiditytypes.AttributeKeyPoolId,
					Value: func() string {
						if tc.useCLPoolID {
							return strconv.Itoa(int(poolID))
						} else {
							return strconv.Itoa(rand.Intn(1000) + 1000)
						}
					}(),
					Index: false,
				})
			}
			if tc.haveTokensInAttribute {
				event.Attributes = append(event.Attributes, abcitypes.EventAttribute{
					Key:   concentratedliquiditytypes.AttributeKeyTokensIn,
					Value: strconv.Itoa(defaultTokenInAmount) + defaultTokenInDenom,
					Index: false,
				})
			}
			if !tc.havePoolIDAttribute && !tc.haveTokensInAttribute {
				event.Attributes = []abcitypes.EventAttribute{}
			}

			// Pass the event to the AddTokenLiquidity method
			err = indexerStreamingService.AdjustTokenInAmountBySpreadFactor(s.Ctx, &event)
			adjustedTokensInStr := defaultUnfoundTokenIn
			for _, attribute := range event.Attributes {
				if attribute.Key == concentratedliquiditytypes.AttributeKeyTokensIn {
					adjustedTokensInStr = attribute.Value
					break
				}
			}
			s.Require().Equal(tc.expectedError, err != nil)
			if tc.haveTokensInAttribute {
				s.Require().NotEqual(defaultUnfoundTokenIn, adjustedTokensInStr)
				// Assert the "tokens_in" event attribute
				coins, err := sdk.ParseCoinsNormalized(adjustedTokensInStr)
				s.Require().NoError(err)
				adjustedTokensInAmount := int(coins[0].Amount.Int64())
				if tc.expectedAdjustedAmt {
					s.Require().NotEqual(defaultTokenInAmount, adjustedTokensInAmount)
				} else {
					s.Require().Equal(defaultTokenInAmount, adjustedTokensInAmount)
				}
			} else {
				// No tokens_in attribute, so the adjusted amount should be 0 (not found)
				s.Require().Equal(defaultUnfoundTokenIn, adjustedTokensInStr)
			}

		})
	}
}

func (s *IndexerServiceTestSuite) TestAddTokenLiquidity() {
	testCases := []struct {
		name                string // Test case name
		eventType           string // Event type to be tested. Only gammtypes.TypeEvtTokenSwapped is valid
		havePoolIDAttribute bool   // Decide whether pool_id attribute should be appended in the event attributes during test data preparation
		useCLPoolID         bool   // Decide whether to use the pool_id from the concentrated liquidity pool or a random pool_id
		expectedError       bool   // Expected error flag
		expectedLiquidity   bool   // Expected liquidity flag
	}{
		{
			name:                "happy path",
			eventType:           gammtypes.TypeEvtTokenSwapped,
			havePoolIDAttribute: true,
			useCLPoolID:         true,
			expectedError:       false,
			expectedLiquidity:   true,
		},
		{
			name:                "non token_swapped event",
			eventType:           gammtypes.TypeEvtPoolJoined,
			havePoolIDAttribute: true,
			useCLPoolID:         true,
			expectedError:       false,
			expectedLiquidity:   false,
		},
		{
			name:                "use non existent pool_id",
			eventType:           gammtypes.TypeEvtTokenSwapped,
			havePoolIDAttribute: true,
			useCLPoolID:         false,
			expectedError:       true,
			expectedLiquidity:   false,
		},
		{
			name:                "no pool_id attribute in event",
			eventType:           gammtypes.TypeEvtTokenSwapped,
			havePoolIDAttribute: false,
			useCLPoolID:         true,
			expectedError:       true,
			expectedLiquidity:   false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// This test suite is to test the AddTokenLiquidity method in the indexer streaming service.
			// where it applies to: token_swapped event only, i.e. gammtypes.TypeEvtTokenSwapped
			// It then looks for the pool_id attribute (concentratedliquiditytypes.AttributeKeyPoolId) in the
			// event attribute map.  With the pool_id, it then fetches the pool liquidity thru
			// keepers.PoolManagerKeeper.GetTotalPoolLiquidity function. The pool liquidity data is then appended
			// to the event attribute map with the key "liquidity_{denom}", value being the pool liquidity.

			// Initialized chain pools
			s.PrepareAllSupportedPools()

			// Get all chain pools from state for asserting later
			concentratedPools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)

			s.Require().NoError(err)

			cfmmPools, err := s.App.GAMMKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)

			cosmWasmPools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
			s.Require().NoError(err)

			blockPools := commondomain.BlockPools{
				ConcentratedPools: concentratedPools,
				CFMMPools:         cfmmPools,
				CosmWasmPools:     cosmWasmPools,
			}

			transformedPools := []sqsdomain.PoolI{}
			for _, pool := range blockPools.GetAll() {
				// Note: balances are irrelevant for the test so we supply empty balances
				transformedPool := sqsdomain.NewPool(pool, pool.GetSpreadFactor(s.Ctx), sdk.Coins{})
				transformedPools = append(transformedPools, transformedPool)
			}

			// Initialize a mock block update process utils
			blockUpdatesProcessUtilsMock := &sqsmocks.BlockUpdateProcessUtilsMock{}

			// Initialize an empty pool tracker
			emptyPoolTracker := pooltracker.NewMemory()

			// Initialize a mock pool extractor
			poolExtractorMock := &sqsmocks.PoolsExtractorMock{
				BlockPools: commondomain.BlockPools{
					ConcentratedPools: concentratedPools,
					CFMMPools:         cfmmPools,
					CosmWasmPools:     cosmWasmPools,
				},
			}

			// Initialize a block process strategy manager
			blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

			// Initialize a concentrated pool with spread factor
			sf, _ := math.LegacyNewDecFromStr(defaultSpreadFactor)
			s.CreateConcentratedPoolsAndFullRangePositionWithSpreadFactor(defaultPoolDenoms, []math.LegacyDec{sf})
			pools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			pool := pools[1]
			poolID := pool.GetId()

			// Initialize keepers
			keepers := indexerdomain.Keepers{
				PoolManagerKeeper: s.App.PoolManagerKeeper,
			}

			// Initialize tx decoder and logger
			txDecoder := s.App.GetTxConfig().TxDecoder()
			logger := s.App.Logger()

			// Initialize a mock publisher
			publisherMock := &indexermocks.PublisherMock{}

			// Initialize the node status checker mock
			nodeStatusCheckerMock := &commonmocks.NodeStatusCheckerMock{}

			// Initialize an indexer streaming service
			indexerStreamingService := indexerservice.New(
				blockUpdatesProcessUtilsMock,
				blockProcessStrategyManager,
				publisherMock,
				emptyStoreKeyMap,
				poolExtractorMock,
				emptyPoolTracker,
				keepers,
				txDecoder,
				nodeStatusCheckerMock,
				logger)

			// Create the event based on the test cases attributes
			event := abcitypes.Event{
				Type: tc.eventType,
				Attributes: func() []abcitypes.EventAttribute {
					if tc.havePoolIDAttribute {
						return []abcitypes.EventAttribute{
							{
								Key: concentratedliquiditytypes.AttributeKeyPoolId,
								Value: func() string {
									if tc.useCLPoolID {
										return strconv.Itoa(int(poolID))
									} else {
										return strconv.Itoa(rand.Intn(1000) + 1000)
									}
								}(),
								Index: false,
							},
						}
					} else {
						return []abcitypes.EventAttribute{}
					}
				}(),
			}

			// Pass the event to the AddTokenLiquidity method
			err = indexerStreamingService.AddTokenLiquidity(s.Ctx, &event)
			s.Require().Equal(tc.expectedError, err != nil)

			// Assert the "liquidity_{denom}"" event attribute
			denoms := pool.GetPoolDenoms(s.Ctx)
			s.Require().Equal(tc.expectedLiquidity, checkIfLiquidityAttributeExists(event, denoms))

		})
	}
}

func (s *IndexerServiceTestSuite) TestSetSpotPrice() {
	testCases := []struct {
		name              string // Test case name
		eventType         string // Event type to be tested. Only gammtypes.TypeEvtTokenSwapped is valid
		poolID            string // pool_id attribute value
		tokenIn           string // token_in attribute value
		tokenOut          string // token_out attribute value
		expectedError     bool   // Expected error flag
		expectedSpotPrice bool   // Expected spot price flag
	}{
		{
			name:              "happy path",
			eventType:         gammtypes.TypeEvtTokenSwapped,
			poolID:            "3",
			tokenIn:           "1000bar",
			tokenOut:          "1000foo",
			expectedError:     false,
			expectedSpotPrice: true,
		},
		{
			name:              "error when no pool_id attribute",
			eventType:         gammtypes.TypeEvtTokenSwapped,
			poolID:            "",
			tokenIn:           "1000bar",
			tokenOut:          "1000foo",
			expectedError:     true,
			expectedSpotPrice: false,
		},
		{
			name:              "error when no token_in attribute",
			eventType:         gammtypes.TypeEvtTokenSwapped,
			poolID:            "3",
			tokenIn:           "",
			tokenOut:          "1000foo",
			expectedError:     true,
			expectedSpotPrice: false,
		},
		{
			name:              "error when no token_out attribute",
			eventType:         gammtypes.TypeEvtTokenSwapped,
			poolID:            "3",
			tokenIn:           "1000bar",
			tokenOut:          "",
			expectedError:     true,
			expectedSpotPrice: false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// This test suite is to test the AddTokenLiquidity method in the indexer streaming service.
			// where it applies to: token_swapped event only, i.e. gammtypes.TypeEvtTokenSwapped
			// It then looks for the pool_id attribute (concentratedliquiditytypes.AttributeKeyPoolId) in the
			// event attribute map.  With the pool_id, it then fetches the pool liquidity thru
			// keepers.PoolManagerKeeper.GetTotalPoolLiquidity function. The pool liquidity data is then appended
			// to the event attribute map with the key "liquidity_{denom}", value being the pool liquidity.

			// Initialized chain pools
			s.PrepareAllSupportedPools()

			// Get all chain pools from state for asserting later
			concentratedPools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)

			s.Require().NoError(err)

			cfmmPools, err := s.App.GAMMKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)

			cosmWasmPools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
			s.Require().NoError(err)

			// Initialize a mock block update process utils
			blockUpdatesProcessUtilsMock := &sqsmocks.BlockUpdateProcessUtilsMock{}

			// Initialize an empty pool tracker
			emptyPoolTracker := pooltracker.NewMemory()

			// Initialize a mock pool extractor
			poolExtractorMock := &sqsmocks.PoolsExtractorMock{
				BlockPools: commondomain.BlockPools{
					ConcentratedPools: concentratedPools,
					CFMMPools:         cfmmPools,
					CosmWasmPools:     cosmWasmPools,
				},
			}

			// Initialize a block process strategy manager
			blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

			// Initialize keepers
			keepers := indexerdomain.Keepers{
				PoolManagerKeeper: s.App.PoolManagerKeeper,
			}

			// Initialize tx decoder and logger
			txDecoder := s.App.GetTxConfig().TxDecoder()
			logger := s.App.Logger()

			// Initialize a mock publisher
			publisherMock := &indexermocks.PublisherMock{}

			// Initialize the node status checker mock
			nodeStatusCheckerMock := &commonmocks.NodeStatusCheckerMock{}

			// Initialize an indexer streaming service
			indexerStreamingService := indexerservice.New(
				blockUpdatesProcessUtilsMock,
				blockProcessStrategyManager,
				publisherMock,
				emptyStoreKeyMap,
				poolExtractorMock,
				emptyPoolTracker,
				keepers,
				txDecoder,
				nodeStatusCheckerMock,
				logger)

			// Create the event based on the test cases attributes
			event := abcitypes.Event{
				Type: tc.eventType,
				Attributes: func() []abcitypes.EventAttribute {
					attributes := []abcitypes.EventAttribute{
						{
							Key:   concentratedliquiditytypes.AttributeKeyPoolId,
							Value: tc.poolID,
							Index: false,
						},
						{
							Key:   concentratedliquiditytypes.AttributeKeyTokensIn,
							Value: tc.tokenIn,
							Index: false,
						},
						{
							Key:   concentratedliquiditytypes.AttributeKeyTokensOut,
							Value: tc.tokenOut,
							Index: false,
						},
					}
					return attributes
				}(),
			}

			// Pass the event to the SetSpotPrice method
			err = indexerStreamingService.SetSpotPrice(s.Ctx, &event)
			s.Require().Equal(tc.expectedError, err != nil)

			// Assert the "quote_tokenin_base_tokenout" event attribute
			if !tc.expectedError {
				s.Require().Equal(tc.expectedSpotPrice, checkIfSpotPriceAttributeExists(event))
			}

		})
	}
}

func (s *IndexerServiceTestSuite) TestTrackCreatedPoolID() {
	testCases := []struct {
		name                     string    // Test case name
		eventType                string    // Event type to be tested. Only poolmanagertypes.TypeEvtPoolCreated is valid type to be tracked
		blockHeight              int64     // Block height to be used in the test
		blockTime                time.Time // Block time to be used in the test
		txnHash                  string    // Transaction hash to be used in the test
		havePoolIDAttribute      bool      // Decide whether pool_id attribute should be appended in the event attributes during test data preparation
		expectedError            bool      // Expected error flag
		expectedPoolBeingTracked bool      // Expected whether the pool is being tracked
	}{
		{
			name:                     "happy path",
			eventType:                poolmanagertypes.TypeEvtPoolCreated,
			blockHeight:              1000,
			blockTime:                time.Now().UTC(),
			txnHash:                  "txnhash",
			havePoolIDAttribute:      true,
			expectedError:            false,
			expectedPoolBeingTracked: true,
		},
		{
			name:                     "Should not track with non pool_created event",
			eventType:                gammtypes.TypeEvtTokenSwapped,
			blockHeight:              1000,
			blockTime:                time.Now().UTC(),
			txnHash:                  "txnhash",
			havePoolIDAttribute:      true,
			expectedError:            true,
			expectedPoolBeingTracked: false,
		},
		{
			name:                     "Should not track with no pool_id attribute",
			eventType:                poolmanagertypes.TypeEvtPoolCreated,
			blockHeight:              1000,
			blockTime:                time.Now().UTC(),
			txnHash:                  "txnhash",
			havePoolIDAttribute:      false,
			expectedError:            true,
			expectedPoolBeingTracked: false,
		},
		{
			name:                     "Should not track with no block height",
			eventType:                poolmanagertypes.TypeEvtPoolCreated,
			blockHeight:              0,
			blockTime:                time.Now().UTC(),
			txnHash:                  "txnhash",
			havePoolIDAttribute:      true,
			expectedError:            true,
			expectedPoolBeingTracked: false,
		},
		{
			name:                     "Should not track with no block time (zero time)",
			eventType:                poolmanagertypes.TypeEvtPoolCreated,
			blockHeight:              1000,
			blockTime:                time.Unix(0, 0),
			txnHash:                  "txnhash",
			havePoolIDAttribute:      true,
			expectedError:            true,
			expectedPoolBeingTracked: false,
		},
		{
			name:                     "Should not track with no txn hash",
			eventType:                poolmanagertypes.TypeEvtPoolCreated,
			blockHeight:              1000,
			blockTime:                time.Now().UTC(),
			txnHash:                  "",
			havePoolIDAttribute:      true,
			expectedError:            true,
			expectedPoolBeingTracked: false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Setup()

			// This test suite is to test the trackCreatedPoolID method in the indexer streaming service.
			// where it applies to: pool_created event only, i.e. poolmanagertypes.TypeEvtPoolCreated
			// It then looks for the pool_id attribute (concentratedliquiditytypes.AttributeKeyPoolId) in the
			// event attribute map.  With the pool_id, it then passes the pool_id to the TrackCreatedPoolID method
			// and in return the underlying pool tracker will track the pool_id. The tests will assert
			// whether the pool is being tracked or not.

			// Initialized chain pools
			s.PrepareAllSupportedPools()

			// Get all chain pools from state for asserting later
			concentratedPools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)

			s.Require().NoError(err)

			cfmmPools, err := s.App.GAMMKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)

			cosmWasmPools, err := s.App.CosmwasmPoolKeeper.GetPoolsWithWasmKeeper(s.Ctx)
			s.Require().NoError(err)

			blockPools := commondomain.BlockPools{
				ConcentratedPools: concentratedPools,
				CFMMPools:         cfmmPools,
				CosmWasmPools:     cosmWasmPools,
			}

			transformedPools := []sqsdomain.PoolI{}
			for _, pool := range blockPools.GetAll() {
				// Note: balances are irrelevant for the test so we supply empty balances
				transformedPool := sqsdomain.NewPool(pool, pool.GetSpreadFactor(s.Ctx), sdk.Coins{})
				transformedPools = append(transformedPools, transformedPool)
			}

			// Initialize a mock block update process utils
			blockUpdatesProcessUtilsMock := &sqsmocks.BlockUpdateProcessUtilsMock{}

			// Initialize an empty pool tracker
			poolTracker := pooltracker.NewMemory()

			// Initialize a mock pool extractor
			poolExtractorMock := &sqsmocks.PoolsExtractorMock{
				BlockPools: commondomain.BlockPools{
					ConcentratedPools: concentratedPools,
					CFMMPools:         cfmmPools,
					CosmWasmPools:     cosmWasmPools,
				},
			}

			// Initialize a block process strategy manager
			blockProcessStrategyManager := commondomain.NewBlockProcessStrategyManager()

			// Initialize a concentrated pool with spread factor
			sf, _ := math.LegacyNewDecFromStr(defaultSpreadFactor)
			s.CreateConcentratedPoolsAndFullRangePositionWithSpreadFactor(defaultPoolDenoms, []math.LegacyDec{sf})
			pools, err := s.App.ConcentratedLiquidityKeeper.GetPools(s.Ctx)
			s.Require().NoError(err)
			pool := pools[1]
			poolID := pool.GetId()

			// Initialize keepers
			keepers := indexerdomain.Keepers{
				PoolManagerKeeper: s.App.PoolManagerKeeper,
			}

			// Initialize tx decoder and logger
			txDecoder := s.App.GetTxConfig().TxDecoder()
			logger := s.App.Logger()

			// Initialize a mock publisher
			publisherMock := &indexermocks.PublisherMock{}

			// Initialize the node status checker mock
			nodeStatusCheckerMock := &commonmocks.NodeStatusCheckerMock{}

			// Initialize an indexer streaming service
			indexerStreamingService := indexerservice.New(
				blockUpdatesProcessUtilsMock,
				blockProcessStrategyManager,
				publisherMock,
				emptyStoreKeyMap,
				poolExtractorMock,
				poolTracker,
				keepers,
				txDecoder,
				nodeStatusCheckerMock,
				logger)

			// Create the event based on the test cases attributes
			event := abcitypes.Event{
				Type: tc.eventType,
				Attributes: func() []abcitypes.EventAttribute {
					if tc.havePoolIDAttribute {
						return []abcitypes.EventAttribute{
							{
								Key: concentratedliquiditytypes.AttributeKeyPoolId,
								Value: func() string {
									return strconv.Itoa(int(poolID))
								}(),
								Index: false,
							},
						}
					} else {
						return []abcitypes.EventAttribute{}
					}
				}(),
			}

			// Pass the event to the trackCreatedPoolID method
			err = indexerStreamingService.TrackCreatedPoolID(event, tc.blockHeight, tc.blockTime, tc.txnHash)
			s.Require().Equal(tc.expectedError, err != nil)

			createdPoolIDs := poolTracker.GetCreatedPoolIDs()
			s.Require().NotNil(createdPoolIDs)
			if tc.expectedPoolBeingTracked {
				s.Require().NotNil(createdPoolIDs[poolID])
				poolCreation := createdPoolIDs[poolID]
				s.Require().Equal(tc.expectedPoolBeingTracked, poolCreation.PoolId == poolID)
				s.Require().Equal(tc.blockHeight, poolCreation.BlockHeight)
				s.Require().Equal(tc.blockTime, poolCreation.BlockTime)
				s.Require().Equal(tc.txnHash, poolCreation.TxnHash)
			} else {
				s.Require().Empty(createdPoolIDs)
			}

		})
	}
}

// checkIfLiquidityAttributeExists checks if the liquidity attribute exists in the event attributes
// as they should be appended by the AddTokenLiquidity method in the indexer streaming service.
// i.e. "liquidity_{denom}" must exist in the event.Attributes where {denom} is the pool denoms
func checkIfLiquidityAttributeExists(event abcitypes.Event, denoms []string) bool {
	liquidityKey0 := liquidityAttributePrefix + denoms[0]
	liquidityKey1 := liquidityAttributePrefix + denoms[1]
	var foundKey0, foundKey1 bool
	for _, attribute := range event.Attributes {
		if attribute.Key == liquidityKey0 {
			foundKey0 = true
			continue
		}
		if attribute.Key == liquidityKey1 {
			foundKey1 = true
			continue
		}
	}
	return foundKey0 && foundKey1
}

// checkIfSpotPriceAttributeExists checks if the spot price attribute exists in the event attributes
// as they should be appended by the SetSpotPrice method in the indexer streaming service.
// i.e. "quote_tokenin_base_tokenout" must exist in the event.Attributes
func checkIfSpotPriceAttributeExists(event abcitypes.Event) bool {
	spotPriceKey := "quote_tokenin_base_tokenout"
	for _, attribute := range event.Attributes {
		if attribute.Key == spotPriceKey {
			return true
		}
	}
	return false
}
