package redis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/pools/common"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v21/x/poolmanager/types"
)

const (
	OneToOnePrecision             = uosmoPrecision
	SpotPriceErrorFmtStr          = spotPriceErrorFmtStr
	NoTokenPrecisionErrorFmtStr   = noTokenPrecisionErrorFmtStr
	RouteIngestDisablePlaceholder = routeIngestDisablePlaceholder
)

type (
	DenomRoutingInfo = denomRoutingInfo
	PoolIngester     = poolIngester
)

func (pi *poolIngester) ConvertPool(ctx sdk.Context, pool poolmanagertypes.PoolI, denomToRoutingInfoMap map[string]denomRoutingInfo, denomPairToTakerFeeMap domain.TakerFeeMap, tokenPrecisionMap map[string]int) (domain.PoolI, error) {
	return pi.convertPool(ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap, tokenPrecisionMap)
}

func RetrieveTakerFeeToMapIfNotExists(ctx sdk.Context, denoms []string, denomPairToTakerFeeMap domain.TakerFeeMap, poolManagerKeeper common.PoolManagerKeeper) error {
	return retrieveTakerFeeToMapIfNotExists(ctx, denoms, denomPairToTakerFeeMap, poolManagerKeeper)
}
