package cosmwasmpool

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v25/x/cosmwasmpool/types"
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

// Gov messages

func (server msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	govAddr := server.keeper.accountKeeper.GetModuleAddress(govtypes.ModuleName)
	if msg.Sender != govAddr.String() {
		return nil, fmt.Errorf("unauthorized: expected sender to be %s, got %s", govAddr, msg.Sender)
	}

	server.keeper.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
