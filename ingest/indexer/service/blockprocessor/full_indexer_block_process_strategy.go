package blockprocessor

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v26/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v26/ingest/indexer/domain"
)

type fullIndexerBlockProcessStrategy struct {
	client            domain.Publisher
	keepers           domain.Keepers
	poolExtractor     commondomain.PoolExtractor
	poolPairPublisher domain.PairPublisher
}

var _ commondomain.BlockProcessor = &fullIndexerBlockProcessStrategy{}

// IsFullBlockProcessor implements commondomain.BlockProcessor.
func (f *fullIndexerBlockProcessStrategy) IsFullBlockProcessor() bool {
	return true
}

// ProcessBlock implements commondomain.BlockProcessStrategy.
func (f *fullIndexerBlockProcessStrategy) ProcessBlock(ctx sdk.Context) (err error) {
	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()

		// Publish supplies
		f.publishAllSupplies(ctx)
	}()

	go func() {
		defer wg.Done()

		// Publish pools
		err = f.processPools(ctx)
	}()

	wg.Wait()

	if err != nil {
		return err
	}

	return nil
}

// publishAllSupplies publishes all the supplies in the block.
func (f *fullIndexerBlockProcessStrategy) publishAllSupplies(ctx sdk.Context) {
	// Ingest the initial data
	f.keepers.BankKeeper.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		// Skip LP shares
		if domain.ShouldFilterDenom(coin.Denom) {
			return false
		}

		// Publish the token supply
		err := f.client.PublishTokenSupply(ctx, domain.TokenSupply{
			Denom:  coin.Denom,
			Supply: coin.Amount,
		})

		// Skip any error silently but log it.
		if err != nil {
			// TODO: alert
			ctx.Logger().Error("failed to publish token supply", "error", err)
		}

		supplyOffset := f.keepers.BankKeeper.GetSupplyOffset(ctx, coin.Denom)

		// If supply offset is non-zero, publish it.
		if !supplyOffset.IsZero() {
			// Publish the token supply offset
			err = f.client.PublishTokenSupplyOffset(ctx, domain.TokenSupplyOffset{
				Denom:        coin.Denom,
				SupplyOffset: supplyOffset,
			})

			// Skip any error silently but log it.
			if err != nil {
				// TODO: alert
				ctx.Logger().Error("failed to publish token supply offset", "error", err)
			}
		}

		return false
	})
}

// processPools publishes all the pools in the block.
func (f *fullIndexerBlockProcessStrategy) processPools(ctx sdk.Context) error {
	blockPools, err := f.poolExtractor.ExtractAll(ctx)
	if err != nil {
		return err
	}

	// Extract pools
	pools := blockPools.GetAll()

	// Process pool pairs
	if err := f.poolPairPublisher.PublishPoolPairs(ctx, pools); err != nil {
		return err
	}

	return nil
}
