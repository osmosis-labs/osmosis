package v8

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	superfluidtypes "github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

// MsgFilterDecorator defines an AnteHandler decorator for the v8 upgrade that
// provide height-gated message filtering acceptance.
type MsgFilterDecorator struct{}

// AnteHandle performs an AnteHandler check that returns an error if the current
// block height is less than the v8 upgrade height and contains messages that are
// not supported until the upgrade height is reached.
func (mfd MsgFilterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	currHeight := ctx.BlockHeight()

	if currHeight < UpgradeHeight && hasInvalidMsgs(tx.GetMsgs()) {
		return ctx, fmt.Errorf("tx contains unsupported message types at height %d", currHeight)
	}

	return next(ctx, tx, simulate)
}

func hasInvalidMsgs(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		switch msg.(type) {
		case *superfluidtypes.MsgUnPoolWhitelistedPool:
			return true
		}
	}

	return false
}
