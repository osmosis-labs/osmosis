package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

func newSale(treasury string, id uint64, tokenIn string, tokenOut sdk.Coin, start, end time.Time) types.Sale {
	zero := sdk.ZeroInt()
	return types.Sale{
		Treasury:       treasury,
		Id:             id,
		TokenOut:       tokenOut.Denom,
		TokenOutSupply: tokenOut.Amount,
		TokenIn:        tokenIn,
		StartTime:      start,
		EndTime:        end,

		OutRemaining: tokenOut.Amount,
		OutSold:      zero,
		OutPerShare:  zero,

		Staked: zero,
		Income: zero,
		Shares: zero,

		Round:    0,
		EndRound: currentRound(start, end, end),
	}
}

func newUserPosition() types.UserPosition {
	zero := sdk.ZeroInt()
	return types.UserPosition{
		Shares:      zero,
		Staked:      zero,
		OutPerShare: zero,
		Spent:       zero,
		Purchased:   zero,
	}
}
