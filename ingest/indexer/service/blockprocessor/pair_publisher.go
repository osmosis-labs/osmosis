package blockprocessor

import (
	"errors"
	"fmt"
	"sync"

	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	commondomain "github.com/osmosis-labs/osmosis/v27/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v27/ingest/indexer/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

type PairPublisher struct {
	client            domain.Publisher
	poolManagerKeeper domain.PoolManagerKeeperI
}

const (
	keySeparator = "|"
)

var _ domain.PairPublisher = &PairPublisher{}

// NewPairPublisher creates a new pair publisher.
func NewPairPublisher(client domain.Publisher, poolManagerKeeper domain.PoolManagerKeeperI) domain.PairPublisher {
	return PairPublisher{
		client:            client,
		poolManagerKeeper: poolManagerKeeper,
	}
}

// PublishPoolPairs publishes the denom pairs contained in the pools.
// Each pool pair is published with the taker fee and spread factor.
// Invalid denoms are skipped as per domain.ShouldFilterDenom function.
// Returns error if at least one of the pairs failed to be published.
// Nil otherwise.
func (p PairPublisher) PublishPoolPairs(ctx sdk.Context, pools []poolmanagertypes.PoolI, createdPoolIDs map[uint64]commondomain.PoolCreation) error {
	result := make(chan error, len(pools))

	// Use map to cache the taker fee for each denom pair
	// and avoid querying the chain for the same pair multiple times.
	mu := sync.RWMutex{}
	denomPairToTakerFeeMap := map[string]osmomath.Dec{}

	// Publish all the pools
	for _, pool := range pools {
		go func(pool poolmanagertypes.PoolI, ctx sdk.Context) {
			// This is to make each go routine have its own gas meter
			// to avoid race conditions.
			ctx = ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())

			denoms := pool.GetPoolDenoms(ctx)
			// Get spread factor for the pool, note cosmossdk isn't thread safe
			// so using mutex to make it thread safe.
			mu.Lock()
			spreadFactor := pool.GetSpreadFactor(ctx)
			mu.Unlock()
			poolID := pool.GetId()

			// Wait for all the pairs to be published
			publishPairWg := sync.WaitGroup{}
			// Initial empty error string
			// We accumulate errors from concurrent publish
			// goroutines.
			errStr := atomic.Value{}
			errStr.Store("")

			for i, denomI := range denoms {
				// Skip unsupported denoms.
				if domain.ShouldFilterDenom(denomI) {
					continue
				}

				for j := i + 1; j < len(denoms); j++ {
					denomJ := denoms[j]
					// Skip unsupported denoms.
					if domain.ShouldFilterDenom(denomJ) {
						continue
					}

					// Retrieve the taker fee for the denom pair if it does not exist in the map
					takerFeeKey := denomI + keySeparator + denomJ
					mu.RLock()
					takerFee, ok := denomPairToTakerFeeMap[takerFeeKey]
					mu.RUnlock()
					if !ok {
						var err error
						// Get taker fee for the denom pair, note cosmossdk isn't thread safe
						// so using mutex to make it thread safe.
						mu.Lock()
						takerFee, err = p.poolManagerKeeper.GetTradingPairTakerFee(ctx, denomI, denomJ)
						mu.Unlock()
						if err != nil {
							// This error should not happen. As a result, we do not skip it
							result <- err

							// Continue to the next pair, if any
							continue
						}

						mu.Lock()
						denomPairToTakerFeeMap[takerFeeKey] = takerFee
						mu.Unlock()
					}

					// Create pair struct and publish it
					pair := domain.Pair{
						PoolID:     poolID,
						MultiAsset: domain.IsMultiDenom(denoms),
						Denom0:     denomI,
						IdxDenom0:  uint8(i),
						Denom1:     denoms[j],
						IdxDenom1:  uint8(j),
						FeeBps:     takerFee.Add(spreadFactor).MulInt64(10000).TruncateInt().Uint64(),
					}
					if poolCreation, ok := createdPoolIDs[poolID]; ok {
						pair.PairCreatedAt = poolCreation.BlockTime
						pair.PairCreatedAtHeight = uint64(poolCreation.BlockHeight)
						pair.PairCreatedAtTxnHash = poolCreation.TxnHash
					}

					publishPairWg.Add(1)

					// Publish the pair in a goroutine
					// to avoid blocking the loop
					go func(pair domain.Pair) {
						defer publishPairWg.Done()

						if pair.IdxDenom0 == pair.IdxDenom1 {
							curErrStr := errStr.Load()
							errStr.Store(fmt.Sprintf("%s, denom0 and denom1 index are the same for pair %v", curErrStr, pair))
							return
						}

						// Publish the pool pair.
						if err := p.client.PublishPair(ctx, pair); err != nil {
							curErrStr := errStr.Load()
							errStr.Store(fmt.Sprintf("%s, %s", curErrStr, err.Error()))
						}
					}(pair)
				}
			}

			// Wait for all the pairs to be published
			publishPairWg.Wait()

			// Load the final error string
			finalErrorStr, ok := errStr.Load().(string)
			if !ok {
				result <- errors.New("failed to parse error when processing pairs")
				return
			}

			// Return the accumulated errors
			if finalErrorStr != "" {
				result <- errors.New(finalErrorStr)
				return
			}

			result <- nil
		}(pool, ctx)
	}

	// Wait for all the results
	for i := 0; i < len(pools); i++ {
		err := <-result
		if err != nil {
			return err
		}
	}

	return nil
}
