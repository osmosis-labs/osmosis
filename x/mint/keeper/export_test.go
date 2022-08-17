package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/mint/types"
)

type (
	ErrInvalidRatio                  = invalidRatioError
	ErrInsufficientDevVestingBalance = insufficientDevVestingBalanceError
)

const (
	EmptyWeightedAddressReceiver = emptyWeightedAddressReceiver
	DeveloperVestingAmount       = developerVestingAmount
)

var (
	GetProportions = getProportions
)

func (k Keeper) CreateDeveloperVestingModuleAccount(ctx sdk.Context, amount sdk.Coin) error {
	return k.createDeveloperVestingModuleAccount(ctx, amount)
}

func (k Keeper) GetDeveloperVestedAmount(ctx sdk.Context, denom string) sdk.Int {
	return k.getDeveloperVestedAmount(ctx, denom)
}

func (k Keeper) GetMintedAmount(ctx sdk.Context, denom string) sdk.Int {
	return k.getMintedAmount(ctx, denom)
}

func (k Keeper) DistributeToModule(ctx sdk.Context, recipientModule string, mintedCoin sdk.Coin, proportion sdk.Dec) (sdk.Int, error) {
	return k.distributeToModule(ctx, recipientModule, mintedCoin, proportion)
}

func (k Keeper) DistributeDeveloperRewards(ctx sdk.Context, developerRewardsCoin sdk.Coin, developerRewardsReceivers []types.WeightedAddress) (sdk.Int, error) {
	return k.distributeDeveloperRewards(ctx, developerRewardsCoin, developerRewardsReceivers)
}

func (k Keeper) DistributeTruncationDelta(ctx sdk.Context, mintedDenom string, expectedTotalMintedByCurrentEpoch sdk.Dec, expectedTotalVestedByCurrentEpoch sdk.Dec) (sdk.Int, error) {
	return k.distributeTruncationDelta(ctx, mintedDenom, expectedTotalMintedByCurrentEpoch, expectedTotalVestedByCurrentEpoch)
}

func (k Keeper) DistributeMintedCoin(ctx sdk.Context, mintedCoin sdk.Coin) error {
	return k.distributeMintedCoin(ctx, mintedCoin)
}

func (k Keeper) GetLastReductionEpochNum(ctx sdk.Context) int64 {
	return k.getLastReductionEpochNum(ctx)
}

func (k Keeper) SetLastReductionEpochNum(ctx sdk.Context, epochNum int64) {
	k.setLastReductionEpochNum(ctx, epochNum)
}

func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) error {
	return k.mintCoins(ctx, newCoins)
}

// Set the mint hooks. This is used for testing purposes only.
func (k *Keeper) SetMintHooksUnsafe(h types.MintHooks) *Keeper {
	k.hooks = h
	return k
}

// Get the mint hooks. This is used for testing purposes only.
func (k *Keeper) GetMintHooksUnsafe() types.MintHooks {
	return k.hooks
}
