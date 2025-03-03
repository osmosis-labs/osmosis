package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v29/x/cron/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) RegisterCron(goCtx context.Context, msg *types.MsgRegisterCron) (*types.MsgRegisterCronResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error(fmt.Sprintf("request invalid: %s", err))
		return &types.MsgRegisterCronResponse{}, err
	}
	// Validation such that only the user who instantiated the contract can register contract
	contractAddr, err := sdk.AccAddressFromBech32(msg.ContractAddress)
	if err != nil {
		return &types.MsgRegisterCronResponse{}, sdkerrors.ErrInvalidAddress
	}
	contractInfo := k.conOps.GetContractInfo(ctx, contractAddr)
	if contractInfo == nil {
		return &types.MsgRegisterCronResponse{}, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract not found")
	}
	// check if sender is authorized
	exists := k.CheckSecurityAddress(ctx, msg.SecurityAddress)
	if !exists {
		return &types.MsgRegisterCronResponse{}, sdkerrors.ErrUnauthorized
	}
	// create a struct of type MsgContractCron
	msgContractCron := types.MsgContractCron{
		ContractAddress: msg.ContractAddress,
		JsonMsg:         msg.JsonMsg,
	}
	cronId := k.GetCronID(ctx)
	cron := types.CronJob{
		Id:              cronId + 1,
		Name:            msg.Name,
		Description:     msg.Description,
		MsgContractCron: []types.MsgContractCron{msgContractCron},
		EnableCron:      true,
	}
	k.SetCronJob(ctx, cron)
	k.SetCronID(ctx, cronId+1)
	return &types.MsgRegisterCronResponse{}, nil
}

func (k msgServer) UpdateCronJob(goCtx context.Context, msg *types.MsgUpdateCronJob) (*types.MsgUpdateCronJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error(fmt.Sprintf("request invalid: %s", err))
		return &types.MsgUpdateCronJobResponse{}, err
	}
	// check if sender is authorized
	exists := k.CheckSecurityAddress(ctx, msg.SecurityAddress)
	if !exists {
		return &types.MsgUpdateCronJobResponse{}, sdkerrors.ErrUnauthorized
	}
	// Get the cron job
	cronJob, found := k.GetCronJob(ctx, msg.Id)
	if !found {
		return &types.MsgUpdateCronJobResponse{}, errorsmod.Wrapf(sdkerrors.ErrNotFound, "cron job not found")
	}
	for _, cron := range cronJob.MsgContractCron {
		if cron.ContractAddress == msg.ContractAddress {
			return &types.MsgUpdateCronJobResponse{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "contract address already exists")
		}
	}
	cronJob.MsgContractCron = append(cronJob.MsgContractCron, types.MsgContractCron{
		ContractAddress: msg.ContractAddress,
		JsonMsg:         msg.JsonMsg,
	})
	k.SetCronJob(ctx, cronJob)
	return &types.MsgUpdateCronJobResponse{}, nil
}

func (k msgServer) DeleteCronJob(goCtx context.Context, msg *types.MsgDeleteCronJob) (*types.MsgDeleteCronJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error(fmt.Sprintf("request invalid: %s", err))
		return &types.MsgDeleteCronJobResponse{}, err
	}
	// check if sender is authorized
	exists := k.CheckSecurityAddress(ctx, msg.SecurityAddress)
	if !exists {
		return &types.MsgDeleteCronJobResponse{}, sdkerrors.ErrUnauthorized
	}
	// Get the cron job
	cronJob, found := k.GetCronJob(ctx, msg.Id)
	if !found {
		return &types.MsgDeleteCronJobResponse{}, errorsmod.Wrapf(sdkerrors.ErrNotFound, "cron job not found")
	}
	// check if contract address exists in the cron job
	var foundContract bool
	for i, cron := range cronJob.MsgContractCron {
		if cron.ContractAddress == msg.ContractAddress {
			cronJob.MsgContractCron = append(cronJob.MsgContractCron[:i], cronJob.MsgContractCron[i+1:]...)
			foundContract = true
			break
		}
	}
	if !foundContract {
		return &types.MsgDeleteCronJobResponse{}, errorsmod.Wrapf(sdkerrors.ErrNotFound, "contract address not found")
	}
	k.SetCronJob(ctx, cronJob)
	return &types.MsgDeleteCronJobResponse{}, nil
}

func (k msgServer) ToggleCronJob(goCtx context.Context, msg *types.MsgToggleCronJob) (*types.MsgToggleCronJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error(fmt.Sprintf("request invalid: %s", err))
		return &types.MsgToggleCronJobResponse{}, err
	}
	// check if sender is authorized
	exists := k.CheckSecurityAddress(ctx, msg.SecurityAddress)
	if !exists {
		return &types.MsgToggleCronJobResponse{}, sdkerrors.ErrUnauthorized
	}
	// Get the cron job
	cronJob, found := k.GetCronJob(ctx, msg.Id)
	if !found {
		return &types.MsgToggleCronJobResponse{}, errorsmod.Wrapf(sdkerrors.ErrNotFound, "cron job not found")
	}
	cronJob.EnableCron = !cronJob.EnableCron
	k.SetCronJob(ctx, cronJob)
	return &types.MsgToggleCronJobResponse{}, nil
}
