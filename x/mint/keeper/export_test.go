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
	GetProportion = getProportion
)

func (k Keeper) CreateDeveloperVestingModuleAccount(ctx sdk.Context, amount sdk.Coin) error {
	return k.createDeveloperVestingModuleAccount(ctx, amount)
}

func (k Keeper) DistributeToModule(ctx sdk.Context, recipientModule string, mintedAmount sdk.Dec, proportion sdk.Dec) (sdk.Int, error) {
	return k.distributeToModule(ctx, recipientModule, mintedAmount, proportion)
}

func (k Keeper) DistributeDeveloperRewards(ctx sdk.Context, totalMintedAmount sdk.Dec, developerRewardsProportion sdk.Dec, developerRewardsReceivers []types.WeightedAddress) (sdk.Int, error) {
	return k.distributeDeveloperRewards(ctx, totalMintedAmount, developerRewardsProportion, developerRewardsReceivers)
}

func (k Keeper) GetLastReductionEpochNum(ctx sdk.Context) int64 {
	return k.getLastReductionEpochNum(ctx)
}

func (k Keeper) SetLastReductionEpochNum(ctx sdk.Context, epochNum int64) {
	k.setLastReductionEpochNum(ctx, epochNum)
}

func (k Keeper) MintAmount(ctx sdk.Context, amount sdk.Int) error {
	return k.mintAmount(ctx, amount)
}

func (k Keeper) GetDeveloperVestedAmount(ctx sdk.Context, denom string) sdk.Int {
	return k.getDeveloperVestedAmount(ctx, denom)
}

func (k Keeper) GetInflationAmount(ctx sdk.Context, denom string) sdk.Int {
	return k.getInflationAmount(ctx, denom)
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
