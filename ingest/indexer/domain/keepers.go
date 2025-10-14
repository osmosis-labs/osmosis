package domain

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v31/x/poolmanager/types"
)

type PoolManagerKeeperI interface {
	GetTradingPairTakerFee(ctx sdk.Context, denom0, denom1 string) (osmomath.Dec, error)
	GetTotalPoolLiquidity(ctx sdk.Context, poolId uint64) (sdk.Coins, error)
	GetPool(ctx sdk.Context, poolId uint64) (types.PoolI, error)
	RouteCalculateSpotPrice(ctx sdk.Context, poolId uint64, quoteAssetDenom string, baseAssetDenom string) (price osmomath.BigDec, err error)
}

type Keepers struct {
	BankKeeper        bankkeeper.Keeper
	PoolManagerKeeper PoolManagerKeeperI
}
