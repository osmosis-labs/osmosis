package keeper

import (
	"fmt"

	epochstypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v7/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	params := k.GetParams(ctx)

	if epochIdentifier == params.EpochIdentifier {
		// not distribute rewards if it's not time yet for rewards distribution
		if epochNumber < params.MintingRewardsDistributionStartEpoch {
			return
		} else if epochNumber == params.MintingRewardsDistributionStartEpoch {
			k.SetLastReductionEpochNum(ctx, epochNumber)
		}
		// fetch stored minter & params
		minter := k.GetMinter(ctx)
		params := k.GetParams(ctx)

		// Check if we have hit an epoch where we update the inflation parameter.
		// We measure time between reductions in number of epochs.
		// This avoids issues with measuring in block numbers, as epochs have fixed intervals, with very
		// low variance at the relevant sizes. As a result, it is safe to store the epoch number
		// of the last reduction to be later retrieved for comparison.
		if epochNumber >= k.GetParams(ctx).ReductionPeriodInEpochs+k.GetLastReductionEpochNum(ctx) {
			// Reduce the reward per reduction period
			minter.EpochProvisions = minter.NextEpochProvisions(params)
			k.SetMinter(ctx, minter)
			k.SetLastReductionEpochNum(ctx, epochNumber)
		}

		// mint coins, update supply
		mintedCoin := minter.EpochProvision(params)
		mintedCoins := sdk.NewCoins(mintedCoin)

		// We over-allocate by the developer vesting portion, and burn this later
		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to the fee collector account
		err = k.DistributeMintedCoin(ctx, mintedCoin)
		if err != nil {
			panic(err)
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.ModuleName,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
				sdk.NewAttribute(types.AttributeKeyEpochProvisions, minter.EpochProvisions.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			),
		)
	}
}

// ___________________________________________________________________________________________________

// Hooks wrapper struct for incentives keeper.
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

// Return the wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks.
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
