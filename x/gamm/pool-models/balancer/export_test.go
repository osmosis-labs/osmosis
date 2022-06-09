package balancer

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	ErrMsgFormatFailedInterimLiquidityUpdate  = errMsgFormatFailedInterimLiquidityUpdate
	ErrMsgFormatRepeatingPoolAssetsNotAllowed = errMsgFormatRepeatingPoolAssetsNotAllowed

	CalcPoolSharesOutGivenSingleAssetIn = calcPoolSharesOutGivenSingleAssetIn
	CalcSingleAssetInGivenPoolSharesOut = calcSingleAssetInGivenPoolSharesOut
	GetPoolAssetsByDenom                = getPoolAssetsByDenom
	UpdateIntermediaryPoolAssets        = updateIntermediaryPoolAssets
)

func (p *Pool) CalcJoinSingleAssetTokensIn(tokensIn sdk.Coins, totalSharesSoFar sdk.Int, poolAssetsByDenom map[string]PoolAsset, swapFee sdk.Dec) (sdk.Int, sdk.Coins, error) {
	return p.calcJoinSingleAssetTokensIn(tokensIn, totalSharesSoFar, poolAssetsByDenom, swapFee)
}
