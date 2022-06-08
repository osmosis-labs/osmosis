package balancer

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	ErrMsgFormatFailedInterimLiquidityUpdate  = errMsgFormatFailedInterimLiquidityUpdate
	ErrMsgFormatRepeatingPoolAssetsNotAllowed = errMsgFormatRepeatingPoolAssetsNotAllowed

	CalcPoolSharesOutGivenSingleAssetIn = calcPoolSharesOutGivenSingleAssetIn
	CalcSingleAssetInGivenPoolSharesOut = calcSingleAssetInGivenPoolSharesOut
	GetPoolAssetsByDenom                = getPoolAssetsByDenom
	UpdateIntermediaryLiquidity         = updateIntermediaryLiquidity
)

func (p *Pool) CalcJoinMultipleSingleAssetTokensIn(tokensIn sdk.Coins, totalSharesSoFar sdk.Int, poolAssetsByDenom map[string]PoolAsset, swapFee sdk.Dec) (sdk.Int, sdk.Coins, error) {
	return p.calcJoinMultipleSingleAssetTokensIn(tokensIn, totalSharesSoFar, poolAssetsByDenom, swapFee)
}
