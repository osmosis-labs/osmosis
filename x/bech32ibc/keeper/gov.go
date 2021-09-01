package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/bech32ibc/types"
)

func (k Keeper) HandleUpdateHrpIbcChannelProposal(ctx sdk.Context, p *types.UpdateHrpIbcChannelProposal) error {
	err := types.ValidateHrp(p.Hrp)
	if err != nil {
		return err
	}

	_, found := k.channelKeeper.GetChannel(ctx, k.tk.GetPort(ctx), p.SourceChannel)

	if !found {
		return sdkerrors.Wrap(types.ErrInvalidIBCData, fmt.Sprintf("channel not found: %s", p.SourceChannel))
	}

	return k.setHrpIbcRecord(ctx, types.HrpIbcRecord{
		Hrp:           p.Hrp,
		SourceChannel: p.SourceChannel,
	})
}
