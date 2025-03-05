package types

import (
	errorsmod "cosmossdk.io/errors"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewCronJob(cronId uint64, name, description string, msgs []MsgContractCron) CronJob {
	return CronJob{
		Id:              cronId,
		Name:            name,
		Description:     description,
		MsgContractCron: msgs,
	}
}

func (m CronJob) Validate() error {
	if m.Id == 0 {
		return errorsmod.Wrap(errors.ErrInvalidRequest, "id must not be 0")
	}
	if m.Name == "" {
		return errorsmod.Wrap(errors.ErrInvalidRequest, "name must not be empty")
	}
	if m.Description == "" {
		return errorsmod.Wrap(errors.ErrInvalidRequest, "description must not be empty")
	}
	if len(m.Name) > 20 {
		return errorsmod.Wrap(errors.ErrInvalidRequest, "name must not exceed 20 characters")
	}
	if len(m.Description) > 1000 {
		return errorsmod.Wrap(errors.ErrInvalidRequest, "description must not exceed 1000 characters")
	}
	// loop through the contract cron messages
	for _, msg := range m.MsgContractCron {
		if _, err := sdk.AccAddressFromBech32(msg.ContractAddress); err != nil {
			return errorsmod.Wrapf(errors.ErrInvalidAddress, "invalid contract address: %v", err)
		}
		if !json.Valid([]byte(msg.JsonMsg)) {
			return errorsmod.Wrap(errors.ErrInvalidRequest, "json msg is invalid")
		}
	}
	return nil
}
