package redis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/common"
)

// retrieveTakerFeeToMapIfNotExists retrieves the taker fee for the denom pair if it does not exist in the map
// Mutates the map with the taker fee for every uniquer denom pair in the denoms list.
// If the taker fee for a denom pair already exists in the map, it is not retrieved again.
// Note that the denoms in denomPair must always be lexicographically sorted to avoid duplicates.
// Returns error if fails to retrieve taker fee from chain. Nil otherwise
func retrieveTakerFeeToMapIfNotExists(ctx sdk.Context, denoms []string, denomPairToTakerFeeMap domain.TakerFeeMap, poolManagerKeeper common.PoolManagerKeeper) error {
	for i, denomI := range denoms {
		for j := i + 1; j < len(denoms); j++ {
			if !denomPairToTakerFeeMap.Has(denomI, denoms[j]) {
				takerFee, err := poolManagerKeeper.GetTradingPairTakerFee(ctx, denomI, denoms[j])
				if err != nil {
					// This error should not happen. As a result, we do not skip it
					return err
				}

				denomPairToTakerFeeMap.SetTakerFee(denomI, denoms[j], takerFee)
			}
		}
	}
	return nil
}
