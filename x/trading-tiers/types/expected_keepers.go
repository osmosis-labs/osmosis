package types

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
)

type StakingKeeper interface {
	BondDenom(ctx context.Context) (string, error)
	GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (osmomath.Int, error)
}

type TxFeesKeeperI interface {
	IsFeeToken(ctx sdk.Context, denom string) (bool, error)
}

type TwapKeeperI interface {
	GetArithmeticTwapToNow(ctx sdk.Context, poolId uint64, baseAssetDenom string, quoteAssetDenom string, startTime time.Time) (osmomath.Dec, error)
}

type ProtorevKeeper interface {
	GetPoolForDenomPair(ctx sdk.Context, baseDenom, denomToMatch string) (uint64, error)
}

type EpochsKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochtypes.EpochInfo
}
