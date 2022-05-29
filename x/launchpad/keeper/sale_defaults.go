package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/launchpad/api"
)

func newSale(treasury string, id uint64, tokenIn, tokenOut string, start, end time.Time, totalOut sdk.Int) api.Sale {
	zero := sdk.ZeroInt()
	return api.Sale{
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

func newUserPosition() api.UserPosition {
	zero := sdk.ZeroInt()
	return api.UserPosition{
		Shares:      zero,
		Staked:      zero,
		OutPerShare: zero,
		Spent:       zero,
		Purchased:   zero,
	}
}
