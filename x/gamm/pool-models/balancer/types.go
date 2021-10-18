package balancer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BalancerPoolParamsI interface {
	GetPoolSwapFee() sdk.Dec
	GetPoolExitFee() sdk.Dec
}
