package stableswap

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ErrMsgFmtDenomDoesNotExist        = errMsgFmtNonExistentDenomGiven
	ErrMsgFmtNonPositiveTokenAmount   = errMsgFmtNonPositiveTokenAmount
	ErrMsgFmtDuplicateDenomFound      = errMsgFmtDuplicateDenomFound
	ErrMsgFmtTooLittlePoolAssetsGiven = errMsgFmtTooLittlePoolAssetsGiven
	ErrMsgFmtNonPositiveScalingFactor = errMsgFmtNonPositiveScalingFactor
	ErrMsgEmptyDenomGiven             = errMsgEmptyDenomGiven
)

func (pa Pool) GetScaledPoolAmt(denom string) (sdk.Int, error) {
	return pa.getScaledPoolAmt(denom)
}

func (pa Pool) GetDescaledPoolAmt(denom string, amtToDeScale sdk.Dec) (sdk.Dec, error) {
	return pa.getDescaledPoolAmt(denom, amtToDeScale)
}

func (pa Pool) ValidateAndSortInitialPoolAssets() error {
	return pa.validateAndSortInitialPoolAssets()
}
