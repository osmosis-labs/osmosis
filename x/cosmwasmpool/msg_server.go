package cosmwasmpool

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
)

type msgServer struct {
	keeper *Keeper
}

var (
	_ types.MsgServer        = msgServer{}
	_ model.MsgCreatorServer = msgServer{}
)

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

func NewMsgCreatorServerImpl(keeper *Keeper) model.MsgCreatorServer {
	return &msgServer{
		keeper: keeper,
	}
}

func (m msgServer) CreateCosmWasmPool(goCtx context.Context, msg *model.MsgCreateCosmWasmPool) (*model.MsgCreateCosmWasmPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	poolId, err := m.keeper.poolmanagerKeeper.CreatePool(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &model.MsgCreateCosmWasmPoolResponse{PoolID: poolId}, nil
}
