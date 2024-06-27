package poolstransformer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/sqsdomain"
	sqscosmwasmpool "github.com/osmosis-labs/sqs/sqsdomain/cosmwasmpool"

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

func (pi *poolTransformer) ConvertPool(ctx sdk.Context, pool poolmanagertypes.PoolI, priceInfoMap map[string]osmomath.BigDec, denomPairToTakerFeeMap sqsdomain.TakerFeeMap) (sqsdomain.PoolI, error) {
	return pi.convertPool(ctx, pool, priceInfoMap, denomPairToTakerFeeMap)
}

func RetrieveTakerFeeToMapIfNotExists(ctx sdk.Context, denoms []string, denomPairToTakerFeeMap sqsdomain.TakerFeeMap, poolManagerKeeper domain.PoolManagerKeeper) error {
	return retrieveTakerFeeToMapIfNotExists(ctx, denoms, denomPairToTakerFeeMap, poolManagerKeeper)
}

func (pi *poolTransformer) ComputeUOSMOPoolLiquidityCap(ctx sdk.Context, balances sdk.Coins, priceInfoMap map[string]osmomath.BigDec) (osmomath.Int, string) {
	return pi.computeUOSMOPoolLiquidityCap(ctx, balances, priceInfoMap)
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

func (pi *poolTransformer) UpdateAlloyTransmuterInfo(ctx sdk.Context, poolId uint64, contractAddress sdk.AccAddress, cosmWasmPoolModel *sqscosmwasmpool.CosmWasmPoolModel, poolDenoms *[]string) error {
	return pi.updateAlloyTransmuterInfo(ctx, poolId, contractAddress, cosmWasmPoolModel, poolDenoms)
}

func (pi *poolTransformer) UpdateOrderbookInfo(
	ctx sdk.Context,
	poolId uint64,
	contractAddress sdk.AccAddress,
	cosmWasmPoolModel *sqscosmwasmpool.CosmWasmPoolModel,
) error {
	return pi.updateOrderbookInfo(ctx, poolId, contractAddress, cosmWasmPoolModel)
}

func (pi *poolTransformer) InitCosmWasmPoolModel(
	ctx sdk.Context,
	pool poolmanagertypes.PoolI,
) sqscosmwasmpool.CosmWasmPoolModel {
	return pi.initCosmWasmPoolModel(ctx, pool)
}

func TickIndexById(ticks []sqscosmwasmpool.OrderbookTick, tickId int64) int {
	return tickIndexById(ticks, tickId)
}
