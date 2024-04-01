package poolsingester

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/sqs/sqsdomain"

	"github.com/osmosis-labs/osmosis/v23/ingest/sqs/domain"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v23/x/poolmanager/types"
)

const (
	OneToOnePrecision             = uosmoPrecision
	SpotPriceErrorFmtStr          = spotPriceErrorFmtStr
	RouteIngestDisablePlaceholder = routeIngestDisablePlaceholder
)

type (
	DenomRoutingInfo = denomRoutingInfo
	PoolIngester     = poolIngester
)

func (pi *poolIngester) ConvertPool(ctx sdk.Context, pool poolmanagertypes.PoolI, denomToRoutingInfoMap map[string]denomRoutingInfo, denomPairToTakerFeeMap sqsdomain.TakerFeeMap) (sqsdomain.PoolI, error) {
	return pi.convertPool(ctx, pool, denomToRoutingInfoMap, denomPairToTakerFeeMap)
}

func RetrieveTakerFeeToMapIfNotExists(ctx sdk.Context, denoms []string, denomPairToTakerFeeMap sqsdomain.TakerFeeMap, poolManagerKeeper domain.PoolManagerKeeper) error {
	return retrieveTakerFeeToMapIfNotExists(ctx, denoms, denomPairToTakerFeeMap, poolManagerKeeper)
}
