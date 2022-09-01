package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/osmoutils"
)

// MintHooks defines an interface for mint module's hooks.
type MintHooks interface {
	AfterDistributeMintedCoin(ctx sdk.Context) error
}

var _ MintHooks = MultiMintHooks{}

// MultiMintHooks is a container for mint hooks.
// All hooks are run in sequence.
type MultiMintHooks []MintHooks

// NewMultiMintHooks returns new MultiMintHooks given hooks.
func NewMultiMintHooks(hooks ...MintHooks) MultiMintHooks {
	return hooks
}

// AfterDistributeMintedCoin is a hook that runs after minter mints and distributes coins
// at the beginning of each epoch.
func (h MultiMintHooks) AfterDistributeMintedCoin(ctx sdk.Context) error {
	for i := range h {
		wrappedHookFn := func(ctx sdk.Context) error {
			return h[i].AfterDistributeMintedCoin(ctx)
		}
		handleHooksError(ctx, osmoutils.ApplyFuncIfNoError(ctx, wrappedHookFn))
	}
	return nil
}

// handleHooksError logs the error using the ctx logger
func handleHooksError(ctx sdk.Context, err error) {
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("error in mint hook %v", err))
	}
}
