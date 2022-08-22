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
	GetProportions                        = getProportions
	GetTruncationStoreKeyForModuleAccount = getTruncationStoreKeyForModuleAccount
)

func (k Keeper) CalculateTotalTruncationDelta(ctx sdk.Context, key []byte, provisions sdk.Dec, amountDistributed sdk.Int) sdk.Dec {
	return k.calculateTotalTruncationDelta(ctx, key, provisions, amountDistributed)
}

func (k Keeper) CreateDeveloperVestingModuleAccount(ctx sdk.Context, amount sdk.Coin) error {
	return k.createDeveloperVestingModuleAccount(ctx, amount)
}

func (k Keeper) DistributeToModule(ctx sdk.Context, recipientModule string, provisionsCoin sdk.DecCoin, proportion sdk.Dec) (sdk.Int, error) {
	return k.distributeToModule(ctx, recipientModule, provisionsCoin, proportion)
}

func (k Keeper) DistributeDeveloperRewards(ctx sdk.Context, developerRewardsCoin sdk.DecCoin, developerRewardsReceivers []types.WeightedAddress) (sdk.Int, error) {
	return k.distributeDeveloperVestingProvisions(ctx, developerRewardsCoin, developerRewardsReceivers)
}

func (k Keeper) DistributeInflationProvisions(ctx sdk.Context, provisionsCoin sdk.DecCoin, proportions types.DistributionProportions) (sdk.Int, error) {
	return k.distributeInflationProvisions(ctx, provisionsCoin, proportions)
}

func (k Keeper) HandleTruncationDelta(ctx sdk.Context, moduleAccountName string, provisions sdk.DecCoin, amountDistributed sdk.Int) (sdk.Int, error) {
	return k.handleTruncationDelta(ctx, moduleAccountName, provisions, amountDistributed)
}

func (k Keeper) MintInflationCoin(ctx sdk.Context, inflationCoins sdk.Coins) error {
	return k.mintInflationCoins(ctx, inflationCoins)
}

func (k Keeper) GetLastReductionEpochNum(ctx sdk.Context) int64 {
	return k.getLastReductionEpochNum(ctx)
}

func (k Keeper) SetLastReductionEpochNum(ctx sdk.Context, epochNum int64) {
	k.setLastReductionEpochNum(ctx, epochNum)
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

// GetDeveloperVestedAmount returns the vestes amount from the developer vesting module account.
func (k Keeper) GetDeveloperVestedAmount(ctx sdk.Context, denom string) sdk.Int {
	unvestedAmount := k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(types.DeveloperVestingModuleAcctName), denom).Amount
	vestedAmount := sdk.NewInt(developerVestingAmount).Sub(unvestedAmount)
	return vestedAmount
}

// GetInflationAmount returns the amount minted by the mint module account
// without considering the developer rewards module account.
// The developer rewards were pre-minted to its own module account at genesis.
// Therefore, the developer rewards can be distributed separately.
// As a result, we should not consider the original developer
// vesting amount when calculating the minted amount.
func (k Keeper) GetInflationAmount(ctx sdk.Context, denom string) sdk.Int {
	totalSupply := k.bankKeeper.GetSupply(ctx, denom).Amount
	return totalSupply.Sub(sdk.NewInt(developerVestingAmount))
}
