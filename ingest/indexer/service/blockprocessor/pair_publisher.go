package blockprocessor

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
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
// TODO: unit test
func (p PairPublisher) PublishPoolPairs(ctx sdk.Context, pools []poolmanagertypes.PoolI) error {
	result := make(chan error, len(pools))

	// Use map to cache the taker fee for each denom pair
	// and avoid querying the chain for the same pair multiple times.
	mu := sync.RWMutex{}
	denomPairToTakerFeeMap := map[string]osmomath.Dec{}

	// Publish all the pools
	for _, pool := range pools {
		go func(pool poolmanagertypes.PoolI) {
			denoms := pool.GetPoolDenoms(ctx)

			// Sort for order consistency
			sort.Strings(denoms)

			spreadFactor := pool.GetSpreadFactor(ctx)
			poolID := pool.GetId()

			resultChan := make(chan error, len(denoms)*(len(denoms)-1)/2)

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
						takerFee, err = p.poolManagerKeeper.GetTradingPairTakerFee(ctx, denomI, denomJ)
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

					// Create pair
					pair := domain.Pair{
						PoolID:    poolID,
						Denom0:    denomI,
						IdxDenom0: uint8(i),
						Denom1:    denoms[j],
						IdxDenom1: uint8(j),
						FeeBps:    takerFee.Add(spreadFactor).MulInt64(10000).TruncateInt().Uint64(),
					}

					// Publish the pair in a goroutine
					// to avoid blocking the loop
					go func(pair domain.Pair) {
						if pair.IdxDenom0 == pair.IdxDenom1 {
							resultChan <- fmt.Errorf("denom0 and denom1 index are the same for pair %v", pair)
							return
						}

						// Publish the pool
						err := p.client.PublishPair(ctx, pair)

						resultChan <- err
					}(pair)
				}
			}

			// Accumulate publishing errors in a string
			errStr := ""

			// Wait for all the publish results
			for i := 0; i < len(denoms)*(len(denoms)-1)/2; i++ {
				err := <-resultChan
				if err != nil {
					errStr += err.Error() + ", "
				}
			}

			// Return the accumulated errors
			if errStr != "" {
				result <- errors.New(errStr)
			}

			result <- nil
		}(pool)
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
