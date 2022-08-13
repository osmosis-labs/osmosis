package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MintHooks interface {
	AfterDistributeMintedCoin(ctx sdk.Context)
}

var _ MintHooks = MultiMintHooks{}

// combine multiple mint hooks, all hook functions are run in array sequence.
type MultiMintHooks []MintHooks

func NewMultiMintHooks(hooks ...MintHooks) MultiMintHooks {
	return hooks
}

// AfterDistributeMintedCoin is a hook that runs after minter mints and distributes coins
// at the beginning of each epoch.
func (h MultiMintHooks) AfterDistributeMintedCoin(ctx sdk.Context) {
	for i := range h {
		h[i].AfterDistributeMintedCoin(ctx)
	}
}
