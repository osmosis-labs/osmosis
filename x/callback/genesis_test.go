package callback_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	e2eTesting "github.com/osmosis-labs/osmosis/v26/tests/e2e/testing"
	"github.com/osmosis-labs/osmosis/v26/x/callback"
	callbackKeeper "github.com/osmosis-labs/osmosis/v26/x/callback/keeper"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

func TestExportGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().CallbackKeeper
	msgServer := callbackKeeper.NewMsgServer(keeper)
	contractAdminAcc := chain.GetAccount(1)

	// Upload and instantiate contract
	codeID := chain.UploadContract(contractAdminAcc, "../../cosmwasm/contracts/callback-test/artifacts/callback_test.wasm", wasmdTypes.AllowEverybody)
	initMsg := CallbackContractInstantiateMsg{Count: 100}
	contractAddr, _ := chain.InstantiateContract(contractAdminAcc, codeID, contractAdminAcc.Address.String(), "callback_test", nil, initMsg)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	currentBlockHeight := ctx.BlockHeight()
	callbackHeight := currentBlockHeight + 1
	futureResFee, blockResFee, txFee, err := keeper.EstimateCallbackFees(ctx, callbackHeight+5)
	require.NoError(t, err)
	feesToPay := futureResFee.Add(blockResFee).Add(txFee)

	reqMsg := &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           DECREMENT_JOBID,
		CallbackHeight:  callbackHeight,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	}
	_, err = msgServer.RequestCallback(ctx, reqMsg)
	require.NoError(t, err)

	reqMsg.JobId = INCREMENT_JOBID
	_, err = msgServer.RequestCallback(ctx, reqMsg)
	require.NoError(t, err)

	reqMsg.JobId = DONOTHING_JOBID
	_, err = msgServer.RequestCallback(ctx, reqMsg)
	require.NoError(t, err)

	reqMsg.CallbackHeight = callbackHeight + 1
	_, err = msgServer.RequestCallback(ctx, reqMsg)
	require.NoError(t, err)

	params := types.Params{
		CallbackGasLimit:               1000000,
		MaxBlockReservationLimit:       1,
		MaxFutureReservationLimit:      1,
		BlockReservationFeeMultiplier:  sdkmath.LegacyZeroDec(),
		FutureReservationFeeMultiplier: sdkmath.LegacyZeroDec(),
	}
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)

	exportedState := callback.ExportGenesis(ctx, keeper)
	require.Equal(t, 4, len(exportedState.Callbacks))
	require.Equal(t, params.CallbackGasLimit, exportedState.Params.CallbackGasLimit)
	require.Equal(t, params.MaxBlockReservationLimit, exportedState.Params.MaxBlockReservationLimit)
	require.Equal(t, params.MaxFutureReservationLimit, exportedState.Params.MaxFutureReservationLimit)
	require.Equal(t, params.BlockReservationFeeMultiplier, exportedState.Params.BlockReservationFeeMultiplier)
	require.Equal(t, params.FutureReservationFeeMultiplier, exportedState.Params.FutureReservationFeeMultiplier)
}

func TestInitGenesis(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().CallbackKeeper
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	validCoin := sdk.NewInt64Coin("stake", 10)

	genParams := types.Params{
		CallbackGasLimit:               1000000,
		MaxBlockReservationLimit:       1,
		MaxFutureReservationLimit:      1,
		BlockReservationFeeMultiplier:  sdkmath.LegacyZeroDec(),
		FutureReservationFeeMultiplier: sdkmath.LegacyZeroDec(),
	}
	err := keeper.SetParams(ctx, genParams)
	require.NoError(t, err)

	genstate := types.GenesisState{
		Params: genParams,
		Callbacks: []*types.Callback{
			{
				ContractAddress: contractAddr.String(),
				JobId:           1,
				CallbackHeight:  100,
				ReservedBy:      contractAddr.String(),
				FeeSplit: &types.CallbackFeesFeeSplit{
					TransactionFees:       &validCoin,
					BlockReservationFees:  &validCoin,
					FutureReservationFees: &validCoin,
					SurplusFees:           &validCoin,
				},
			},
		},
	}

	callback.InitGenesis(ctx, keeper, genstate)

	callbacks, err := keeper.GetAllCallbacks(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(callbacks)) // Ensuring callbacks are not imported

	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, genParams.CallbackGasLimit, params.CallbackGasLimit)
	require.Equal(t, genParams.MaxBlockReservationLimit, params.MaxBlockReservationLimit)
	require.Equal(t, genParams.MaxFutureReservationLimit, params.MaxFutureReservationLimit)
	require.Equal(t, genParams.BlockReservationFeeMultiplier, params.BlockReservationFeeMultiplier)
	require.Equal(t, genParams.FutureReservationFeeMultiplier, params.FutureReservationFeeMultiplier)
}
