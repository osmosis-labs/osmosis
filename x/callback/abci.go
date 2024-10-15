package callback

import (
	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/v26/x/callback/keeper"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
	"github.com/osmosis-labs/osmosis/v26/x/callback/utils"
)

// EndBlocker fetches all the callbacks registered for the current block height and executes them
func EndBlocker(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected) error {
	k.IterateCallbacksByHeight(ctx, ctx.BlockHeight(), callbackExec(ctx, k, wk))
	return nil
}

// callbackExec returns a function which executes the callback and deletes it from state after execution
func callbackExec(ctx sdk.Context, k keeper.Keeper, wk types.WasmKeeperExpected) func(types.Callback) bool {
	logger := k.Logger(ctx)
	return func(callback types.Callback) bool {
		// creating CallbackMsg which is encoded to json and passed as input to contract execution
		callbackMsg := types.NewCallbackMsg(callback.JobId)
		callbackMsgString := callbackMsg.String()

		logger.Debug(
			"executing callback",
			"contract_address", callback.ContractAddress,
			"job_id", callback.JobId,
			"msg", callbackMsgString,
		)

		params, err := k.GetParams(ctx)
		if err != nil {
			panic(err)
		}

		gasUsed, err := utils.ExecuteWithGasLimit(ctx, callback.MaxGasLimit, func(ctx sdk.Context) error {
			// executing the callback on the contract
			_, err := wk.Sudo(ctx, sdk.MustAccAddressFromBech32(callback.ContractAddress), callbackMsg.Bytes())
			return err
		})
		if err != nil {
			// Emit failure event
			types.EmitCallbackExecutedFailedEvent(
				ctx,
				callback.ContractAddress,
				callback.JobId,
				callbackMsgString,
				gasUsed,
				err.Error(),
			)

			errorCode := types.ModuleErrors_ERR_UNKNOWN
			// check if out of gas error
			if errorsmod.IsOf(err, sdkerrors.ErrOutOfGas) {
				errorCode = types.ModuleErrors_ERR_OUT_OF_GAS
			}
			// check if the error was due to contract execution failure
			if errorsmod.IsOf(err, wasmtypes.ErrExecuteFailed) {
				errorCode = types.ModuleErrors_ERR_CONTRACT_EXECUTION_FAILED
			}

			// log the error
			logger.Error(
				"error executing callback",
				"contract_address", callback.ContractAddress,
				"job_id", callback.JobId,
				"gas_used", gasUsed,
				"error", err,
				"error_code", errorCode,
			)

			// This is because gasUsed amount returned is greater than the gas limit. cuz ofc.
			// so we set it to callback.MaxGasLimit so when we do txFee refund, we arent trying to refund more than we should
			// e.g if callback.MaxGasLimit is 10, but gasUsed is 100, we need to use 10 to calculate txFeeRefund.
			// else the module will pay back more than it took from the user 💀
			// TLDR; this ensures in case of "out of gas error", we keep all txFees and refund nothing.
			gasUsed = callback.MaxGasLimit
		} else {
			logger.Info(
				"callback executed successfully",
				"contract_address", callback.ContractAddress,
				"job_id", callback.JobId,
				"msg", callbackMsgString,
				"gas_used", gasUsed,
			)
			// Emit success event
			types.EmitCallbackExecutedSuccessEvent(
				ctx,
				callback.ContractAddress,
				callback.JobId,
				callbackMsgString,
				gasUsed,
			)
		}

		logger.Info(
			"callback executed with pending gas",
			"contract_address", callback.ContractAddress,
			"job_id", callback.JobId,
			"used_gas", gasUsed,
		)

		// Calculate current tx fees based on gasConsumed. Refund any leftover to the address which reserved the callback
		txFeesConsumed := k.CalculateTransactionFees(ctx, gasUsed, params.GetMinPriceOfGas())
		if txFeesConsumed.IsLT(*callback.FeeSplit.TransactionFees) {
			refundAmount := callback.FeeSplit.TransactionFees.Sub(txFeesConsumed)
			err := k.RefundFromCallbackModule(ctx, callback.ReservedBy, refundAmount)
			if err != nil {
				panic(err)
			}
		} else {
			// This is to ensure that if the txFeeConsumed is higher due to rise in gas price,
			// we dont fund fee_collector more than we should
			txFeesConsumed = *callback.FeeSplit.TransactionFees
		}

		// Send fees to fee collector
		feeCollectorAmount := callback.FeeSplit.BlockReservationFees.
			Add(*callback.FeeSplit.FutureReservationFees).
			Add(*callback.FeeSplit.SurplusFees).
			Add(txFeesConsumed)
		err = k.SendToFeeCollector(ctx, feeCollectorAmount)
		if err != nil {
			panic(err)
		}

		// deleting the callback after execution
		if err := k.Callbacks.Remove(
			ctx,
			collections.Join3(
				callback.CallbackHeight,
				sdk.MustAccAddressFromBech32(callback.ContractAddress).Bytes(),
				callback.JobId,
			),
		); err != nil {
			panic(err)
		}

		return false
	}
}
