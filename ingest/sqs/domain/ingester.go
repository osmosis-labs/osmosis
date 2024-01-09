package domain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/sqsdomain/repository"
)

// AtomicIngester is an interface that defines the methods for the atomic ingester.
// It processes a block by writing data into a transaction.
// The caller must call Exec on the transaction to flush data to sink.
type AtomicIngester interface {
	// ProcessBlock processes the block by writing data into a transaction.
	// Returns error if fails to process.
	// It does not flush data to sink. The caller must call Exec on the transaction
	ProcessBlock(ctx sdk.Context, tx repository.Tx) error
}
