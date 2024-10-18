package callback_test

import (
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	e2eTesting "github.com/osmosis-labs/osmosis/v26/tests/e2e/testing"
	callbackKeeper "github.com/osmosis-labs/osmosis/v26/x/callback/keeper"
	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

const (
	DECREMENT_JOBID = 0
	INCREMENT_JOBID = 1
	ERROR_JOBID     = 2
	DONOTHING_JOBID = 3
)

func TestEndBlocker(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keeper := chain.GetApp().CallbackKeeper
	msgServer := callbackKeeper.NewMsgServer(keeper)
	contractAdminAcc := chain.GetAccount(0)

	// Upload and instantiate contract
	// The test contract is based on the default counter contract and behaves the following way:
	// When job_id = 1, it increments the count value
	// When job_id = 0, it decrements the count value
	// When job_id = 2, it throws an error
	// For any other job_id, it does nothing
	codeID := chain.UploadContract(contractAdminAcc, "../../cosmwasm/contracts/callback-test/artifacts/callback_test.wasm", wasmdTypes.AllowEverybody)
	initMsg := CallbackContractInstantiateMsg{Count: 100}
	contractAddr, _ := chain.InstantiateContract(contractAdminAcc, codeID, contractAdminAcc.Address.String(), "callback_test", nil, initMsg)
	chain.NextBlock(1)

	testCases := []struct {
		testCase      string
		jobId         uint64
		expectedCount int32
	}{
		{
			testCase:      "Decrement count",
			jobId:         DECREMENT_JOBID,
			expectedCount: initMsg.Count - 1,
		},
		{
			testCase:      "Increment count",
			jobId:         INCREMENT_JOBID,
			expectedCount: initMsg.Count,
		},
		{
			testCase:      "Do nothing",
			jobId:         DONOTHING_JOBID,
			expectedCount: initMsg.Count,
		},
		{
			testCase:      "Throw error", // The contract throws error but the EndBlocker should not.
			jobId:         ERROR_JOBID,
			expectedCount: initMsg.Count,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case: %s", tc.testCase), func(t *testing.T) {
			ctx := chain.GetContext()
			feesToPay, err := getCallbackRegistrationFees(chain)
			require.NoError(t, err)

			reqMsg := &types.MsgRequestCallback{
				ContractAddress: contractAddr.String(),
				JobId:           tc.jobId,
				CallbackHeight:  ctx.BlockHeight() + 1,
				Sender:          contractAdminAcc.Address.String(),
				Fees:            feesToPay,
			}
			_, err = msgServer.RequestCallback(ctx, reqMsg)
			require.NoError(t, err)

			// Increment block height and run end blocker at the next block
			chain.NextBlock(1)
			chain.NextBlock(1)

			// Checking if the count value is as expected
			count := getCount(t, chain, contractAddr)
			require.Equal(t, tc.expectedCount, count)
		})
	}

	params, err := keeper.GetParams(chain.GetContext())
	require.NoError(t, err)

	// TEST CASE: Test CallbackGasLimit limit value reduced
	// First we set the params value to default
	// Register a callback for next block
	feesToPay, err := getCallbackRegistrationFees(chain)
	require.NoError(t, err)
	reqMsg := &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           INCREMENT_JOBID,
		CallbackHeight:  chain.GetContext().BlockHeight() + 1,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	}
	_, err = msgServer.RequestCallback(chain.GetContext(), reqMsg)
	require.NoError(t, err)

	// Setting the callbackGasLimit param to 1
	params.CallbackGasLimit = 1
	err = keeper.SetParams(chain.GetContext(), params)
	require.NoError(t, err)

	// Increment block height and run end blocker
	chain.NextBlock(1)
	chain.NextBlock(1)

	// Checking if the count value has incremented.
	// Should have incremented as the callback should have access to higher gas limit as it was registered before the gas limit was reduced
	count := getCount(t, chain, contractAddr)
	require.Equal(t, initMsg.Count+1, count)

	// TEST CASE: OUT OF GAS ERROR
	// Reserving a callback for next block
	// This callback should fail as it consumes more gas than allowed
	feesToPay, err = getCallbackRegistrationFees(chain)
	require.NoError(t, err)
	reqMsg = &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           INCREMENT_JOBID,
		CallbackHeight:  chain.GetContext().BlockHeight() + 1,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	}
	_, err = msgServer.RequestCallback(chain.GetContext(), reqMsg)
	require.NoError(t, err)

	// Increment block height and run end blocker
	chain.NextBlock(1)
	chain.NextBlock(1)

	// Checking if the count value is zero. should be as error callback rests count when out of gas error
	count = getCount(t, chain, contractAddr)
	require.Equal(t, int32(0), count)
}

func getCallbackRegistrationFees(chain *e2eTesting.TestChain) (sdk.Coin, error) {
	ctx := chain.GetContext()
	currentBlockHeight := ctx.BlockHeight()
	callbackHeight := currentBlockHeight + 1
	futureResFee, blockResFee, txFee, err := chain.GetApp().CallbackKeeper.EstimateCallbackFees(ctx, callbackHeight)
	if err != nil {
		return sdk.Coin{}, err
	}
	feesToPay := futureResFee.Add(blockResFee).Add(txFee)
	return feesToPay, nil
}

// getCount is a helper function to get the contract's count value
func getCount(t *testing.T, chain *e2eTesting.TestChain, contractAddr sdk.AccAddress) int32 {
	getCountQuery := "{\"get_count\":{}}"
	resp, err := chain.GetApp().WasmKeeper.QuerySmart(chain.GetContext(), contractAddr, []byte(getCountQuery))
	require.NoError(t, err)
	var getCountResp CallbackContractQueryMsg
	err = json.Unmarshal(resp, &getCountResp)
	require.NoError(t, err)
	return getCountResp.Count
}

type CallbackContractInstantiateMsg struct {
	Count int32 `json:"count"`
}

func (msg CallbackContractInstantiateMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Count int32 `json:"count"`
	}{
		Count: msg.Count,
	})
}

type CallbackContractQueryMsg struct {
	Count int32 `json:"count"`
}
