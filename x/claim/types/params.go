package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DefaultOsmoBondDenom          = "uosmo"
	DefaultDurationUntilDecay     = time.Hour
	DefaultClaimModuleAcctBalance = sdk.NewCoins(sdk.NewCoin(DefaultOsmoBondDenom, sdk.NewInt(1e6))) // 1 OSMO = 10^6 uosmo
	DefaultDurationOfDecay        = time.Hour * 5
)
