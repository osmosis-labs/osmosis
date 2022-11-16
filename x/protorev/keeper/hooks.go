package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/osmosis-labs/osmosis/v12/x/epochs/types"
)

type EpochHooks struct {
	k Keeper
}

var (
	_ epochstypes.EpochHooks = EpochHooks{}
)

func (k Keeper) EpochHooks() epochstypes.EpochHooks {
	return EpochHooks{k}
}

///////////////////////////////////////////////////////

// BeforeEpochStart is the epoch start hook.
func (hook EpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

// AfterEpochEnd is the epoch end hook.
func (hook EpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}
