package domain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/sqs/sqsdomain"
)

// BlockProcessor is an interface that defines an interface for processing a block.
type BlockProcessor interface {
	// Process processes the block returning the parsed pool data and taker fee map.
	// Returns error if the processor fails to extract data.
	Process(ctx sdk.Context) ([]sqsdomain.PoolI, sqsdomain.TakerFeeMap, error)
}
