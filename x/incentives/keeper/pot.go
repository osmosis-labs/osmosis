package keeper

import (
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreatePot lock tokens from an account for specified duration
func (k Keeper) CreatePot(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, distrTo []*types.DistrCondition, startTime time.Time, numEpochs uint64) error {
	return nil
}

// AddToPot is a utility to lock coins into module account
func (k Keeper) AddToPot(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, potID uint64) error {
	return nil
}

// Distribute is a utility to lock coins into module account
func (k Keeper) Distribute(ctx sdk.Context, pot types.Pot) error {
	return nil
}
