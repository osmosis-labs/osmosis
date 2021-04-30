package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	OsmoBondDenom                 = "uosmo"
	DefaultDurationUntilDecay     = time.Hour
	DefaultClaimModuleAcctBalance = sdk.NewCoins(sdk.NewCoin(OsmoBondDenom, sdk.NewInt(1e6))) // 1 OSMO = 10^6 uosmo
	DefaultDurationOfDecay        = time.Hour * 5
)
