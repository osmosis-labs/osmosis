package cfmm_common

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func CalcExitPool(ctx sdk.Context, pool types.PoolI, exitingShares sdk.Int, exitFee sdk.Dec) (sdk.Coins, error) {
	totalShares := pool.GetTotalShares()
	if exitingShares.GTE(totalShares) {
		return sdk.Coins{}, errors.New(("too many shares out"))
	}

	// refundedShares = exitingShares * (1 - exit fee)
	// with 0 exit fee optimization
	var refundedShares sdk.Dec
	if !exitFee.IsZero() {
		// exitingShares * (1 - exit fee)
		// Todo: make a -1 constant
		oneSubExitFee := sdk.OneDec().SubMut(exitFee)
		refundedShares = oneSubExitFee.MulIntMut(exitingShares)
	} else {
		refundedShares = exitingShares.ToDec()
	}

	shareOutRatio := refundedShares.QuoInt(totalShares)
	// exitedCoins = shareOutRatio * pool liquidity
	exitedCoins := sdk.Coins{}
	poolLiquidity := pool.GetTotalPoolLiquidity(ctx)

	for _, asset := range poolLiquidity {
		// round down here, due to not wanting to over-exit
		exitAmt := shareOutRatio.MulInt(asset.Amount).TruncateInt()
		if exitAmt.LTE(sdk.ZeroInt()) {
			continue
		}
		if exitAmt.GTE(asset.Amount) {
			return sdk.Coins{}, errors.New(("too many shares out"))
		}
		exitedCoins = exitedCoins.Add(sdk.NewCoin(asset.Denom, exitAmt))
	}

	return exitedCoins, nil
}
