package poolstransformer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/sqsdomain"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v25/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v25/x/poolmanager/types"
)

const (
	SpotPriceErrorFmtStr          = spotPriceErrorFmtStr
	RouteIngestDisablePlaceholder = routeIngestDisablePlaceholder
	NoPoolLiquidityCapError       = noPoolLiquidityCapError
	USDC                          = usdcDenom
)

var (
	UsdcPrecisionScalingFactor = usdcPrecisionScalingFactor
)

type (
	PoolTransformer = poolTransformer
)

func (pi *poolTransformer) ConvertPool(ctx sdk.Context, pool poolmanagertypes.PoolI, denomToRoutingInfoMap map[string]osmomath.BigDec, denomPairToTakerFeeMap sqsdomain.TakerFeeMap) (sqsdomain.PoolI, error) {
	return pi.convertPool(ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap)
}

func RetrieveTakerFeeToMapIfNotExists(ctx sdk.Context, denoms []string, denomPairToTakerFeeMap sqsdomain.TakerFeeMap, poolManagerKeeper domain.PoolManagerKeeper) error {
	return retrieveTakerFeeToMapIfNotExists(ctx, denoms, denomPairToTakerFeeMap, poolManagerKeeper)
}

func (pi *poolTransformer) ComputeUOSMOPoolLiquidityCap(ctx sdk.Context, balances sdk.Coins, denomRoutingInfoMap map[string]osmomath.BigDec) (osmomath.Int, string) {
	return pi.computeUOSMOPoolLiquidityCap(ctx, balances, denomRoutingInfoMap)
}

func FilterBalances(originalBalances sdk.Coins, poolDenomsMap map[string]struct{}) sdk.Coins {
	return filterBalances(originalBalances, poolDenomsMap)
}

func GetPoolDenomsMap(poolDenoms []string) map[string]struct{} {
	return getPoolDenomsMap(poolDenoms)
}

func (pi *poolTransformer) ComputeUSDCPoolLiquidityCapFromUOSMO(ctx sdk.Context, poolLiquidityCapUOSMO osmomath.Int) (osmomath.Int, string) {
	return pi.computeUSDCPoolLiquidityCapFromUOSMO(ctx, poolLiquidityCapUOSMO)
}
