package stableswap

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ErrMsgFmtDenomDoesNotExist        = errMsgFmtNonExistentDenomGiven
	ErrMsgFmtNonPositiveTokenAmount   = errMsgFmtNonPositiveTokenAmount
	ErrMsgFmtDuplicateDenomFound      = errMsgFmtDuplicateDenomFound
	ErrMsgFmtTooLittlePoolAssetsGiven = errMsgFmtTooLittlePoolAssetsGiven
	ErrMsgFmtNonPositiveScalingFactor = errMsgFmtNonPositiveScalingFactor
	ErrMsgFmrDrainedPool              = errMsfFmtDrainedPool
	ErrMsgEmptyDenomGiven             = errMsgEmptyDenomGiven
)

func (pa Pool) GetPoolAssetAndIndex(denom string) (int, PoolAsset, error) {
	return pa.getPoolAssetAndIndex(denom)
}

func (pa Pool) GetScaledPoolAmt(denom string) (sdk.Int, error) {
	return pa.getScaledPoolAmt(denom)
}

func (pa Pool) GetDescaledPoolAmt(denom string, amtToDeScale sdk.Dec) (sdk.Dec, error) {
	return pa.getDescaledPoolAmt(denom, amtToDeScale)
}

func (pa Pool) ValidateAndSortInitialPoolAssets() error {
	return pa.validateAndSortInitialPoolAssets()
}

func (pa Pool) UpdatePoolLiquidityForSwap(tokensIn sdk.Coins, tokensOut sdk.Coins) error {
	return pa.updatePoolLiquidityForSwap(tokensIn, tokensOut)
}
