package commondomain

import sdk "github.com/cosmos/cosmos-sdk/types"

// BlockProcessor is an interface for processing a block.
type BlockProcessor interface {
	// ProcessBlock processes a block.
	// It returns an error if the block processing fails.
	ProcessBlock(ctx sdk.Context) error
}
