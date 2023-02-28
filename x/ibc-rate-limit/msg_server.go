package ibc_rate_limit

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/ibc-rate-limit/types"
)

type msgServer struct {
	keeper *ICS4Wrapper
}

func NewMsgServerImpl(keeper *ICS4Wrapper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

func (server msgServer) SetContractParam(goCtx context.Context, msg *types.MsgSetContractParam) (*types.MsgSetContractParamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ctx.Logger().Error("HEEEERRRRe", msg.Address)

	address, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	server.keeper.SetParams(ctx, types.Params{
		ContractAddress: address.String(),
	})
	return nil, nil
}
