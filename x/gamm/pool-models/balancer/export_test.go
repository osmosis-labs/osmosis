package balancer

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ErrMsgFormatRepeatingPoolAssetsNotAllowed = formatRepeatingPoolAssetsNotAllowedErrFormat
	ErrMsgFormatNoPoolAssetFound              = formatNoPoolAssetFoundErrFormat
)

var (
	ErrMsgFormatFailedInterimLiquidityUpdate = failedInterimLiquidityUpdateErrFormat

	CalcPoolSharesOutGivenSingleAssetIn   = calcPoolSharesOutGivenSingleAssetIn
	CalcSingleAssetInGivenPoolSharesOut   = calcSingleAssetInGivenPoolSharesOut
	UpdateIntermediaryPoolAssetsLiquidity = updateIntermediaryPoolAssetsLiquidity

	GetPoolAssetsByDenom = getPoolAssetsByDenom
	EnsureDenomInPool    = ensureDenomInPool
)

func (p *Pool) CalcSingleAssetJoin(tokenIn sdk.Coin, spreadFactor sdk.Dec, tokenInPoolAsset PoolAsset, totalShares sdk.Int) (numShares sdk.Int, err error) {
	return p.calcSingleAssetJoin(tokenIn, spreadFactor, tokenInPoolAsset, totalShares)
}

func (p *Pool) CalcJoinSingleAssetTokensIn(tokensIn sdk.Coins, totalSharesSoFar sdk.Int, poolAssetsByDenom map[string]PoolAsset, spreadFactor sdk.Dec) (sdk.Int, sdk.Coins, error) {
	return p.calcJoinSingleAssetTokensIn(tokensIn, totalSharesSoFar, poolAssetsByDenom, spreadFactor)
}
