package keeper

import (
	"fmt"

	epochstypes "github.com/osmosis-labs/osmosis/v11/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v11/x/mint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeforeEpochStart is a hook which is executed before the start of an epoch. It is a no-op for mint module.
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// no-op
}

// AfterEpochEnd is a hook which is executed after the end of an epoch.
// This hook should attempt to mint and distribute coins according to
// the configuration set via parameters. In addition, it handles the logic
// for reducing minted coins according to the parameters.
// For an attempt to mint to occur:
// - given epochIdentifier must be equal to the mint epoch identifier set via parameters.
// - given epochNumber must be greater than or equal to the mint start epoch set via parameters.
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

		// Check if we have hit an epoch where we update the inflation parameter.
		// We measure time between reductions in number of epochs.
		// This avoids issues with measuring in block numbers, as epochs have fixed intervals, with very
		// low variance at the relevant sizes. As a result, it is safe to store the epoch number
		// of the last reduction to be later retrieved for comparison.
		if epochNumber >= params.ReductionPeriodInEpochs+k.GetLastReductionEpochNum(ctx) {
			// Reduce the reward per reduction period
			minter.EpochProvisions = minter.NextEpochProvisions(params)
			k.SetMinter(ctx, minter)
			k.SetLastReductionEpochNum(ctx, epochNumber)
		}

		// mint coins, update supply
		mintedCoin := minter.EpochProvision(params)
		mintedCoins := sdk.NewCoins(mintedCoin)

		err := k.bankKeeper.MintCoins(ctx, types.ModuleName, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to the fee collector account
		err = k.DistributeMintedCoin(ctx, mintedCoin)
		if err != nil {
			panic(err)
		}

		developerVestingCoin := minter.DeveloperVestingEpochProvision(params)
		developerRewardsDistributionProportion := params.DistributionProportions.DeveloperRewards
		// allocate dev rewards to respective accounts from developer vesting module account.
		developerVestingAmount, err := k.distributeDeveloperRewards(ctx, developerVestingCoin, params.WeightedDeveloperRewardsReceivers)
		if err != nil {
			panic(err)
		}

		devRewardsProportion := minter.EpochProvisions.Mul(params.DistributionProportions.DeveloperRewards)
		minter.LastTotalVestedAmount = minter.LastTotalVestedAmount.Add(devRewardsProportion)
		minter.LastTotalMintedAmount = minter.LastTotalMintedAmount.Add(minter.EpochProvisions.Sub(devRewardsProportion))
		k.SetMinter(ctx, minter)

		distributedTruncationDelta, err := k.distributeTruncationDelta(ctx, mintedCoin.Denom, minter.LastTotalMintedAmount, minter.LastTotalVestedAmount, developerRewardsDistributionProportion)
		if err != nil {
			panic(err)
		}

		totalDistributed := mintedCoin.Amount.Add(developerVestingAmount).Add(distributedTruncationDelta)

		// call a hook after the minting and distribution of new coins
		k.hooks.AfterDistributeMintedCoin(ctx)

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(totalDistributed.Int64()), "minted_tokens")
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.ModuleName,
				sdk.NewAttribute(types.AttributeEpochNumber, fmt.Sprintf("%d", epochNumber)),
				sdk.NewAttribute(types.AttributeKeyEpochProvisions, minter.EpochProvisions.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, totalDistributed.String()),
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
