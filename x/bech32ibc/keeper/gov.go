package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/x/bech32ibc/types"

	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
)

func (k Keeper) HandleUpdateHrpIbcChannelProposal(ctx sdk.Context, p *types.UpdateHrpIbcChannelProposal) error {
	err := types.ValidateHRP(p.Hrp)
	if err != nil {
		return err
	}

	_, found := k.channelKeeper.GetChannel(ctx, ibctransfertypes.DefaultGenesisState().GetPortId(), p.SourceChannel)

	if !found {
		return sdkerrors.Wrap(types.ErrInvalidIBCData, fmt.Sprintf("channel not found: %s", p.SourceChannel))
	}

	return k.setHrpIbcRecord(ctx, types.HrpIbcRecord{
		HRP:           p.Hrp,
		SourceChannel: p.SourceChannel,
	})
}
