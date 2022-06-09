package balancer

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	CalcPoolSharesOutGivenSingleAssetIn = calcPoolSharesOutGivenSingleAssetIn
	CalcSingleAssetInGivenPoolSharesOut = calcSingleAssetInGivenPoolSharesOut
)

func (p *Pool) CalcSingleAssetJoin(tokenIn sdk.Coin, swapFee sdk.Dec, tokenInPoolAsset PoolAsset, totalShares sdk.Int) (numShares sdk.Int, err error) {
	return p.calcSingleAssetJoin(tokenIn, swapFee, tokenInPoolAsset, totalShares)
}
