package ingester

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/pools/common"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v20/x/poolmanager/types"
)

type (
	DenomRoutingInfo = denomRoutingInfo
	PoolIngester     = poolIngester
)

func ConvertPool(ctx sdk.Context, pool poolmanagertypes.PoolI, denomToRoutingInfoMap map[string]denomRoutingInfo, bankKeeper common.BankKeeper, protorevKeeper common.ProtorevKeeper, poolManagerKeeper common.PoolManagerKeeper) (domain.PoolI, error) {
	return convertPool(ctx, pool, denomToRoutingInfoMap, bankKeeper, protorevKeeper, poolManagerKeeper)
}
