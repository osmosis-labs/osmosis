package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

func newSale(treasury string, id uint64, tokenIn, tokenOut string, start, end time.Time, totalOut sdk.Int) types.Sale {
	zero := sdk.ZeroInt()
	return types.Sale{
		Treasury:  treasury,
		Id:        id,
		TokenOut:  tokenOut,
		TokenIn:   tokenIn,
		StartTime: start,
		EndTime:   end,

		OutRemaining: totalOut,
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
