package redis_test

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"

	redisdb "github.com/redis/go-redis/v9"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/repository/redis"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v20/tests/mocks"
	"github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

const (
	UOSMO = "uosmo"
	UION  = "uion"
)

var (
	defaultCreationTime = time.Unix(1, 0)
	defaultAmountA      = osmomath.NewInt(1000000)
	defaultAmountB      = osmomath.NewInt(5000000)
)

var _ = Describe("RedisRepository", func() {
	const (
		numPoolsCreate           = 3
		stableSwapPoolIDOffset   = 3 + 1
		concentratedPoolIDOffset = numPoolsCreate*2 + 1
		cosmWasmPoolIDOffset     = numPoolsCreate*3 + 1
	)

	var (
		testBalancerPools     []domain.PoolI
		testStableSwapPools   []domain.PoolI
		testConcentratedPools []domain.PoolI
		testCosmWasmPools     []domain.PoolI
		clientMock            redismock.ClientMock
		ctx                   context.Context
		mockCFMMPool          *mocks.MockSQSPoolI
		mockConcentratedPool  *mocks.MockSQSPoolI
		mockCosmWasmPool      *mocks.MockSQSPoolI
	)

	BeforeEach(func() {
		// Configure test pools
		// Create balancer pools
		for i := 0; i < numPoolsCreate; i++ {
			expectedPoolID := uint64(i + 1)
			pool := withPoolID(newDefaultBalancerPool(), expectedPoolID)
			testBalancerPools = append(testBalancerPools, pool)
		}

		// Create stableswap pools
		for i := stableSwapPoolIDOffset; i < stableSwapPoolIDOffset+numPoolsCreate; i++ {
			expectedPoolID := uint64(i)
			pool := withPoolType(withPoolID(newDefaultBalancerPool(), expectedPoolID), poolmanagertypes.Balancer)
			testStableSwapPools = append(testStableSwapPools, pool)
		}

		// Create concentrated pools
		for i := concentratedPoolIDOffset; i <= concentratedPoolIDOffset+numPoolsCreate-1; i++ {
			expectedPoolID := uint64(i)
			pool := withPoolType(withPoolID(newDefaultBalancerPool(), expectedPoolID), poolmanagertypes.Concentrated)
			testConcentratedPools = append(testConcentratedPools, pool)
		}

		// Create cosmwasm pools
		for i := cosmWasmPoolIDOffset; i <= cosmWasmPoolIDOffset+numPoolsCreate-1; i++ {
			expectedPoolID := uint64(i)
			pool := withPoolType(withPoolID(newDefaultBalancerPool(), expectedPoolID), poolmanagertypes.CosmWasm)
			testCosmWasmPools = append(testCosmWasmPools, pool)
		}

		// configure redis db mock
		var client *redisdb.Client
		client, clientMock = redismock.NewClientMock()
		poolRepo = redis.NewRedisPoolsRepo(client)

		// configure CFMM pool mock
		ctrl := gomock.NewController(GinkgoT())
		mockCFMMPool = mocks.NewMockSQSPoolI(ctrl)

		// configure Concentrated pool mock
		mockConcentratedPool = mocks.NewMockSQSPoolI(ctrl)

		// configure CosmWasm pool mock
		mockCosmWasmPool = mocks.NewMockSQSPoolI(ctrl)

		// Create context
		ctx = context.Background()
	})

	AfterEach(func() {
		// clear test pools
		testBalancerPools = []domain.PoolI{}
		testStableSwapPools = []domain.PoolI{}
		testConcentratedPools = []domain.PoolI{}
		testCosmWasmPools = []domain.PoolI{}
	})

	Describe("CFMM Pools", func() {

		Context("StoreCFMM", func() {

			When("called with empty pools", func() {
				It("should succeed", func() {
					err := poolRepo.StoreCFMM(ctx, []domain.PoolI{})
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("storing balancer pools", func() {
				It("should succeed", func() {

					// Define assertions on redis client methods being called
					for i := 0; i < numPoolsCreate; i++ {
						expectedID := uint64(i + 1)

						serializedPool, err := json.Marshal(testBalancerPools[i])
						Expect(err).ToNot(HaveOccurred())
						expectedPoolKey := redis.CfmmKeyFromPoolTypeAndID(poolmanagertypes.Balancer, expectedID)
						clientMock.ExpectHSet(redis.CfmmPoolKey, expectedPoolKey, serializedPool).SetVal(1)
					}

					err := poolRepo.StoreCFMM(ctx, testBalancerPools)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("storing stableswap pools", func() {
				It("should succeed", func() {

					// Define assertions on redis client methods being called
					for i := stableSwapPoolIDOffset; i <= numPoolsCreate*2; i++ {
						expectedID := uint64(i)

						serializedPool, err := json.Marshal(testStableSwapPools[expectedID-numPoolsCreate-1])
						Expect(err).ToNot(HaveOccurred())
						expectedPoolKey := redis.CfmmKeyFromPoolTypeAndID(poolmanagertypes.Stableswap, expectedID)
						clientMock.ExpectHSet(redis.CfmmPoolKey, expectedPoolKey, serializedPool).SetVal(1)
					}

					err := poolRepo.StoreCFMM(ctx, testStableSwapPools)

					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("storing multiple balancer and stableswap pools", func() {
				It("should succeed", func() {
					// Define assertions on redis client methods being called

					// Change order by storing stableswap first - order does not matter here.
					for i := stableSwapPoolIDOffset; i <= numPoolsCreate*2; i++ {
						expectedID := uint64(i)

						serializedPool, err := json.Marshal(testStableSwapPools[expectedID-numPoolsCreate-1])
						Expect(err).ToNot(HaveOccurred())
						expectedPoolKey := redis.CfmmKeyFromPoolTypeAndID(poolmanagertypes.Stableswap, expectedID)
						clientMock.ExpectHSet(redis.CfmmPoolKey, expectedPoolKey, serializedPool).SetVal(1)
					}

					for i := 0; i < numPoolsCreate; i++ {
						expectedID := uint64(i + 1)

						serializedPool, err := json.Marshal(testBalancerPools[i])
						Expect(err).ToNot(HaveOccurred())
						expectedPoolKey := redis.CfmmKeyFromPoolTypeAndID(poolmanagertypes.Balancer, expectedID)
						clientMock.ExpectHSet(redis.CfmmPoolKey, expectedPoolKey, serializedPool).SetVal(1)
					}

					// Note that this is out of pool ID order.
					storedPools := append(testStableSwapPools, testBalancerPools...)

					err := poolRepo.StoreCFMM(ctx, storedPools)

					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("storing non-CFMM pools", func() {
				It("should fail", func() {
					mockCFMMPool.EXPECT().GetType().Return(types.Concentrated)

					err := poolRepo.StoreCFMM(ctx, []domain.PoolI{mockCFMMPool})
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(domain.InvalidPoolTypeError{PoolType: int32(types.Concentrated)}))
				})
			})
		})

		Context("GetAllCFMM", func() {
			When("no pools returned from redis", func() {
				It("should succeed", func() {
					clientMock.ExpectHGetAll(redis.CfmmPoolKey).SetVal(map[string]string{})

					result, err := poolRepo.GetAllCFMM(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeEmpty())
				})
			})

			When("stableswap & balancer pools", func() {
				It("should return them in ascending order sorted by ID", func() {

					expectedReturn := map[string]string{}

					// Mock stableswap pools in the map
					for _, stableswapPool := range testStableSwapPools {
						// cast to conrete type to be able to marshal
						stableswapModel := stableswapPool.(*domain.Pool)

						expectedSerialized, err := json.Marshal(stableswapModel)
						Expect(err).ToNot(HaveOccurred())

						expectedPoolKey := redis.CfmmKeyFromPoolTypeAndID(poolmanagertypes.Stableswap, stableswapModel.GetId())
						expectedReturn[expectedPoolKey] = string(expectedSerialized)
					}

					// Mock balancer pools in the map
					for _, balancerPool := range testBalancerPools {
						// cast to conrete type to be able to marshal
						balancerModel := balancerPool.(*domain.Pool)

						expectedSerialized, err := json.Marshal(balancerModel)
						Expect(err).ToNot(HaveOccurred())

						expectedPoolKey := redis.CfmmKeyFromPoolTypeAndID(poolmanagertypes.Balancer, balancerPool.GetId())
						expectedReturn[expectedPoolKey] = string(expectedSerialized)
					}

					clientMock.ExpectHGetAll(redis.CfmmPoolKey).SetVal(expectedReturn)

					result, err := poolRepo.GetAllCFMM(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(HaveLen(len(testStableSwapPools) + len(testBalancerPools)))

					// TODO: reduce duplication with generics
					for i := 0; i < len(result); i++ {
						currentPool := result[i]
						if i < len(result)-i {
							nextPool := result[i+1]

							// Asserts sorted order
							Expect(currentPool.GetId()).Should(Equal(nextPool.GetId() - 1))
						}

						// Assert type

						// Expecting the first 3 pools to be balancer type, and the next 3 stableswap
						expectedType := poolmanagertypes.Balancer
						if currentPool.GetId() >= stableSwapPoolIDOffset {
							expectedType = poolmanagertypes.Stableswap
						}
						Expect(currentPool.GetType()).To(Equal(expectedType))
					}
				})
			})
		})
	})

	Describe("ConcentratedPools", func() {
		Context("StoreConcentrated", func() {

			When("called with empty pools", func() {
				It("should succeed", func() {
					err := poolRepo.StoreConcentrated(ctx, []domain.PoolI{})
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("storing concentrated pools", func() {
				It("should succeed", func() {
					// Define assertions on redis client methods being called
					for expectedID := concentratedPoolIDOffset; expectedID < concentratedPoolIDOffset+numPoolsCreate; expectedID++ {

						serializedPool, err := json.Marshal(testConcentratedPools[expectedID-concentratedPoolIDOffset])
						Expect(err).ToNot(HaveOccurred())
						clientMock.ExpectHSet(redis.ConcentratedPoolKey, strconv.Itoa(expectedID), serializedPool).SetVal(1)
					}

					err := poolRepo.StoreConcentrated(ctx, testConcentratedPools)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("storing non-Concentrated pools", func() {
				It("should fail", func() {
					mockConcentratedPool.EXPECT().GetType().Return(types.Balancer)

					err := poolRepo.StoreConcentrated(ctx, []domain.PoolI{mockConcentratedPool})
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(domain.InvalidPoolTypeError{PoolType: int32(types.Balancer)}))
				})
			})
		})

		Context("GetAllConcentrated", func() {
			When("no pools returned from redis", func() {
				It("should succeed", func() {
					clientMock.ExpectHGetAll(redis.ConcentratedPoolKey).SetVal(map[string]string{})

					result, err := poolRepo.GetAllConcentrated(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeEmpty())
				})
			})

			When("non-empty", func() {
				It("should return them in ascending order sorted by ID", func() {

					expectedReturn := map[string]string{}

					// Mock concentrated pools in the map
					for _, concentratedPool := range testConcentratedPools {
						// cast to conrete type to be able to marshal
						concentratedModel := concentratedPool.(*domain.Pool)

						expectedSerialized, err := json.Marshal(concentratedModel)
						Expect(err).ToNot(HaveOccurred())

						expectedReturn[strconv.Itoa(int(concentratedPool.GetId()))] = string(expectedSerialized)
					}

					clientMock.ExpectHGetAll(redis.ConcentratedPoolKey).SetVal(expectedReturn)

					result, err := poolRepo.GetAllConcentrated(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(HaveLen(len(testConcentratedPools)))

					// TODO: reduce duplication with generics
					for i := 0; i < len(result); i++ {
						currentPool := result[i]
						if i < len(result)-i {
							nextPool := result[i+1]

							// Asserts sorted order
							Expect(currentPool.GetId()).Should(Equal(nextPool.GetId() - 1))
						}

						// Assert type
						Expect(currentPool.GetType()).To(Equal(poolmanagertypes.Concentrated))
					}
				})
			})
		})
	})

	Describe("CosmWasmPools", func() {
		Context("StoreCosmWasm", func() {

			When("called with empty pools", func() {
				It("should succeed", func() {
					err := poolRepo.StoreCosmWasm(ctx, []domain.PoolI{})
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("storing cosmwasm pools", func() {
				It("should succeed", func() {
					// Define assertions on redis client methods being called
					for expectedID := cosmWasmPoolIDOffset; expectedID < cosmWasmPoolIDOffset+numPoolsCreate; expectedID++ {

						serializedPool, err := json.Marshal(testCosmWasmPools[expectedID-cosmWasmPoolIDOffset])
						Expect(err).ToNot(HaveOccurred())
						clientMock.ExpectHSet(redis.CosmWasmPoolKey, strconv.Itoa(expectedID), serializedPool).SetVal(1)
					}

					err := poolRepo.StoreCosmWasm(ctx, testCosmWasmPools)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("storing non-Cosmwasm pools", func() {
				It("should fail", func() {
					mockCosmWasmPool.EXPECT().GetType().Return(types.Balancer)

					err := poolRepo.StoreCosmWasm(ctx, []domain.PoolI{mockCosmWasmPool})
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(domain.InvalidPoolTypeError{PoolType: int32(types.Balancer)}))
				})
			})
		})

		Context("GetAllConcentrated", func() {
			When("no pools returned from redis", func() {
				It("should succeed", func() {
					clientMock.ExpectHGetAll(redis.CosmWasmPoolKey).SetVal(map[string]string{})

					result, err := poolRepo.GetAllCosmWasm(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeEmpty())
				})
			})

			When("non-empty", func() {
				It("should return them in ascending order sorted by ID", func() {

					expectedReturn := map[string]string{}

					// Mock concentrated pools in the map
					for _, cosmWasmPool := range testCosmWasmPools {
						// cast to conrete type to be able to marshal
						cosmWasmModel := cosmWasmPool.(*domain.Pool)

						expectedSerialized, err := json.Marshal(cosmWasmModel)
						Expect(err).ToNot(HaveOccurred())

						expectedReturn[strconv.Itoa(int(cosmWasmModel.GetId()))] = string(expectedSerialized)
					}

					clientMock.ExpectHGetAll(redis.CosmWasmPoolKey).SetVal(expectedReturn)

					result, err := poolRepo.GetAllCosmWasm(ctx)

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(HaveLen(len(testCosmWasmPools)))

					// TODO: reduce duplication with generics
					for i := 0; i < len(result); i++ {
						currentPool := result[i]
						currentCosmWasmPool := currentPool.(*domain.Pool)
						if i < len(result)-i {
							nextPool := result[i+1]
							nextCosmWasmPool := nextPool.(*domain.Pool)

							// Asserts sorted order
							Expect(currentCosmWasmPool.GetId()).Should(Equal(nextCosmWasmPool.GetId() - 1))
						}
					}
				})
			})
		})
	})
})

// creates a default balancer pool to be used in tests
func newDefaultBalancerPool() *domain.Pool {
	pool := &domain.Pool{
		Id:           1,
		Type:         int(poolmanagertypes.Balancer),
		SpreadFactor: osmomath.NewDecWithPrec(1, 2).String(),
		Denoms:       []string{UOSMO, UION},
		Weights:      []string{"1", "1"},
		Balances:     sdk.NewCoins(sdk.NewCoin(UOSMO, defaultAmountA), sdk.NewCoin(UION, defaultAmountB)).String(),
		Liquidity:    sdk.NewInt(500).String(),
	}

	return pool
}

// modifies the ID of the given balancer pool to given.
func withPoolID(pool *domain.Pool, ID uint64) *domain.Pool {
	pool.Id = ID
	return pool
}

func withPoolType(pool *domain.Pool, poolType poolmanagertypes.PoolType) *domain.Pool {
	pool.Type = int(poolType)
	return pool
}
