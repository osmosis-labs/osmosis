package poolstransformer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/sqsdomain"

	commondomain "github.com/osmosis-labs/osmosis/v27/ingest/common/domain"
)

// retrieveTakerFeeToMapIfNotExists retrieves the taker fee for the denom pair if it does not exist in the map
// Mutates the map with the taker fee for every uniquer denom pair combination in the denoms list.
// Since bi-directional taker fee is supported, the taker fee for a denom pair is stored in both directions.
// For example, the taker fees for a pair of denoms (A, B) is stored BOTH as (A, B) and (B, A).
// If the taker fee for a denom pair already exists in the map, it is not retrieved again.
// Returns error if fails to retrieve taker fee from chain. Nil otherwise
func retrieveTakerFeeToMapIfNotExists(ctx sdk.Context, denoms []string, denomPairToTakerFeeMap sqsdomain.TakerFeeMap, poolManagerKeeper commondomain.PoolManagerKeeper) error {
	for i, denomI := range denoms {
		for j, denomJ := range denoms {
			if i != j {
				if !denomPairToTakerFeeMap.Has(denomI, denomJ) {
					takerFee, err := poolManagerKeeper.GetTradingPairTakerFee(ctx, denomI, denomJ)
					if err != nil {
						// This error should not happen. As a result, we do not skip it
						return err
					}

					denomPairToTakerFeeMap.SetTakerFee(denomI, denomJ, takerFee)
				}
			}
		}
	}

	return nil
}
