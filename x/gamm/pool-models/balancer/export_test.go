package balancer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

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

func (p *Pool) CalcSingleAssetJoin(tokenIn sdk.Coin, spreadFactor osmomath.Dec, tokenInPoolAsset PoolAsset, totalShares osmomath.Int) (numShares osmomath.Int, err error) {
	return p.calcSingleAssetJoin(tokenIn, spreadFactor, tokenInPoolAsset, totalShares)
}

func (p *Pool) CalcJoinSingleAssetTokensIn(tokensIn sdk.Coins, totalSharesSoFar osmomath.Int, poolAssetsByDenom map[string]PoolAsset, spreadFactor osmomath.Dec) (osmomath.Int, sdk.Coins, error) {
	return p.calcJoinSingleAssetTokensIn(tokensIn, totalSharesSoFar, poolAssetsByDenom, spreadFactor)
}
