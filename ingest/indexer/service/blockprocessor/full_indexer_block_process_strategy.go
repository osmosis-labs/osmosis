package blockprocessor

import (
	"strings"
	"sync"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commondomain "github.com/osmosis-labs/osmosis/v25/ingest/common/domain"
	"github.com/osmosis-labs/osmosis/v25/ingest/indexer/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

type fullIndexerBlockProcessStrategy struct {
	client        domain.Publisher
	keepers       domain.Keepers
	poolExtracter commondomain.PoolExtracter
}

var _ commondomain.BlockProcessor = &fullIndexerBlockProcessStrategy{}

// ProcessBlock implements commondomain.BlockProcessStrategy.
func (f *fullIndexerBlockProcessStrategy) ProcessBlock(ctx types.Context) (err error) {

	wg := sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()

		// Publish supplies
		err = f.publishAllSupplies(ctx)
	}()

	go func() {
		defer wg.Done()

		// Publish pools
		err = f.publishAllPools(ctx)
	}()

	wg.Wait()

	if err != nil {
		return err
	}

	return nil
}

func (f *fullIndexerBlockProcessStrategy) publishAllSupplies(ctx sdk.Context) error {

	// Ingest the initial data
	f.keepers.BankKeeper.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		// Skip CL pool shares
		if strings.Contains(coin.Denom, "cl/pool") {
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

	return nil
}

// publishAllPools publishes all the pools in the block.
func (f *fullIndexerBlockProcessStrategy) publishAllPools(ctx sdk.Context) error {
	blockPools, err := f.poolExtracter.ExtractAll(ctx)
	if err != nil {
		return err
	}

	// TODO: consider worker pool

	pools := blockPools.GetAll()

	result := make(chan error, len(pools))

	// Publish all the pools
	for _, pool := range pools {

		go func(pool poolmanagertypes.PoolI) {
			// Publish the pool
			err := f.client.PublishPool(ctx, domain.Pool{
				ChainModel: pool,
			})

			result <- err
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
