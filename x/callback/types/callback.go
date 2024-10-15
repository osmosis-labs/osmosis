package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewCallback creates a new Callback instance.
func NewCallback(sender string, contractAddress string, height int64, jobID uint64, txFees sdk.Coin, blockReservationFees sdk.Coin, futureReservationFees sdk.Coin, surplusFees sdk.Coin) Callback {
	return Callback{
		ContractAddress: contractAddress,
		CallbackHeight:  height,
		JobId:           jobID,
		ReservedBy:      sender,
		FeeSplit: &CallbackFeesFeeSplit{
			TransactionFees:       &txFees,
			BlockReservationFees:  &blockReservationFees,
			FutureReservationFees: &futureReservationFees,
			SurplusFees:           &surplusFees,
		},
	}
}

// Validate perform object fields validation.
func (c Callback) Validate() error {
	if _, err := sdk.AccAddressFromBech32(c.GetContractAddress()); err != nil {
		return err
	}
	if _, err := sdk.AccAddressFromBech32(c.GetReservedBy()); err != nil {
		return err
	}
	if c.GetCallbackHeight() <= 0 {
		return ErrCallbackHeightNotInFuture
	}
	if err := c.GetFeeSplit().GetTransactionFees().Validate(); err != nil {
		return err
	}
	if err := c.GetFeeSplit().GetBlockReservationFees().Validate(); err != nil {
		return err
	}
	if err := c.GetFeeSplit().GetFutureReservationFees().Validate(); err != nil {
		return err
	}
	if err := c.GetFeeSplit().GetSurplusFees().Validate(); err != nil {
		return err
	}
	return nil
}
