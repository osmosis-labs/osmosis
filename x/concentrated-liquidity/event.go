package concentrated_liquidity

import (
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	types "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

// event is the interface all event types should be implementing
type event interface {
	emit(ctx sdk.Context)
}

// guarantee that liquidityChangeEvent type implements the event interface
var _ event = &liquidityChangeEvent{}

// liquidityChangeEvent represent the fields used for event emission
// and uniquely identifying a position such as:
// - position id
// - sender
// - pool id
// - join time
// - lower tick
// - upper tick
// It also hols additional attributes for the liquidity added or removed and the actual amounts of asset0 and asset1 it translates to.
type liquidityChangeEvent struct {
	eventType      string
	positionId     uint64
	sender         sdk.AccAddress
	poolId         uint64
	lowerTick      int64
	upperTick      int64
	joinTime       time.Time
	liquidityDelta osmomath.Dec
	actualAmount0  osmomath.Int
	actualAmount1  osmomath.Int
}

// emit emits an event for a liquidity change when creating or withdrawing a position based its field.
func (l *liquidityChangeEvent) emit(ctx sdk.Context) {
	if l != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			l.eventType,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyPositionId, strconv.FormatUint(l.positionId, 10)),
			sdk.NewAttribute(sdk.AttributeKeySender, l.sender.String()),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(l.poolId, 10)),
			sdk.NewAttribute(types.AttributeLowerTick, strconv.FormatInt(l.lowerTick, 10)),
			sdk.NewAttribute(types.AttributeUpperTick, strconv.FormatInt(l.upperTick, 10)),
			sdk.NewAttribute(types.AttributeJoinTime, l.joinTime.String()),
			sdk.NewAttribute(types.AttributeLiquidity, l.liquidityDelta.String()),
			sdk.NewAttribute(types.AttributeAmount0, l.actualAmount0.String()),
			sdk.NewAttribute(types.AttributeAmount1, l.actualAmount1.String()),
		))
	}
}
