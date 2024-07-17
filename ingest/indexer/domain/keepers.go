package domain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/osmomath"
)

type PoolManagerKeeperI interface {
	GetTradingPairTakerFee(ctx sdk.Context, denom0, denom1 string) (osmomath.Dec, error)
}

type Keepers struct {
	BankKeeper        bankkeeper.Keeper
	PoolManagerKeeper PoolManagerKeeperI
}
