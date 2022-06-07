package bindings

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	swapJson = []byte("{ \"pool_id\": 1, \"denom_in\": \"denomIn\", \"denom_out\": \"denomOut\" }")
	swap     = Swap{
		PoolId:   1,
		DenomIn:  "denomIn",
		DenomOut: "denomOut",
	}
	stepJson = []byte("{ \"pool_id\": 2, \"denom_out\": \"denomOut\" }")
	step     = Step{
		PoolId:   2,
		DenomOut: "denomOut",
	}
	swapAmountInJson = []byte("{ \"in\": \"123\", \"out\": null }")
	in               = sdk.NewInt(123)
	swapAmountIn     = SwapAmount{
		In: &in,
	}
	swapAmountOutJson = []byte("{ \"out\": \"456\" }")
	out               = sdk.NewInt(456)
	swapAmountOut     = SwapAmount{
		Out: &out,
	}
	exactIn = ExactIn{
		Input:     sdk.NewInt(789),
		MinOutput: sdk.NewInt(101112),
	}
	exactOut = ExactOut{
		MaxInput: sdk.NewInt(131415),
		Output:   sdk.NewInt(161718),
	}
	swapAmountExactInJson = []byte("{ \"exact_in\": { \"input\": \"789\", \"min_output\": \"101112\" } }")
	swapAmountExactIn     = SwapAmountWithLimit{
		ExactIn: &exactIn,
	}
	swapAmountExactOutJson = []byte("{ \"exact_in\": null, \"exact_out\": { \"max_input\": \"131415\", \"output\": \"161718\" } }")
	swapAmountExactOut     = SwapAmountWithLimit{
		ExactOut: &exactOut,
	}
)

func TestTypesEncodeDecode(t *testing.T) {
	// Swap
	// Marshal
	bzSwap, err := json.Marshal(swap)
	require.NoError(t, err)
	// Unmarshal
	var swap1 Swap
	err = json.Unmarshal(bzSwap, &swap1)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swap, swap1)

	// Step
	// Marshal
	bzStep, err := json.Marshal(step)
	require.NoError(t, err)
	// Unmarshal
	var step1 Step
	err = json.Unmarshal(bzStep, &step1)
	require.NoError(t, err)
	// Check
	assert.Equal(t, step, step1)

	// SwapAmount
	// Marshal
	bzSwapAmount, err := json.Marshal(swapAmountOut)
	require.NoError(t, err)
	// Unmarshal
	var swapAmount1 SwapAmount
	err = json.Unmarshal(bzSwapAmount, &swapAmount1)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountOut, swapAmount1)

	// SwapAmount in
	// Marshal
	bzSwapAmount2, err := json.Marshal(swapAmountIn)
	require.NoError(t, err)
	// Unmarshal
	var swapAmount2 SwapAmount
	err = json.Unmarshal(bzSwapAmount2, &swapAmount2)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountIn, swapAmount2)

	// SwapAmount out
	// Marshal
	bzSwapAmount3, err := json.Marshal(swapAmountOut)
	require.NoError(t, err)
	// Unmarshal
	var swapAmount3 SwapAmount
	err = json.Unmarshal(bzSwapAmount3, &swapAmount3)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountOut, swapAmount3)

	// SwapAmount exact in
	// Marshal
	bzSwapAmountWithLimit1, err := json.Marshal(swapAmountExactIn)
	require.NoError(t, err)
	// Unmarshal
	var swapAmountWithLimit1 SwapAmountWithLimit
	err = json.Unmarshal(bzSwapAmountWithLimit1, &swapAmountWithLimit1)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountExactIn, swapAmountWithLimit1)

	// SwapAmount exact out
	// Marshal
	bzSwapAmountWithLimit2, err := json.Marshal(swapAmountExactOut)
	require.NoError(t, err)
	// Unmarshal
	var swapAmountWithLimit2 SwapAmountWithLimit
	err = json.Unmarshal(bzSwapAmountWithLimit2, &swapAmountWithLimit2)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountExactOut, swapAmountWithLimit2)
}

func TestTypesDecode(t *testing.T) {
	// Swap
	// Unmarshal
	var swap1 Swap
	err := json.Unmarshal([]byte(swapJson), &swap1)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swap, swap1)

	// Step
	// Unmarshal
	var step1 Step
	err = json.Unmarshal(stepJson, &step1)
	require.NoError(t, err)
	// Check
	assert.Equal(t, step, step1)

	// SwapAmount in
	// Unmarshal
	var swapAmount1 SwapAmount
	err = json.Unmarshal(swapAmountInJson, &swapAmount1)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountIn, swapAmount1)

	// SwapAmount out
	// Unmarshal
	var swapAmount2 SwapAmount
	err = json.Unmarshal(swapAmountOutJson, &swapAmount2)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountOut, swapAmount2)

	// SwapAmount exact in
	// Unmarshal
	var swapAmountWithLimit1 SwapAmountWithLimit
	err = json.Unmarshal(swapAmountExactInJson, &swapAmountWithLimit1)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountExactIn, swapAmountWithLimit1)

	// SwapAmount exact out
	// Unmarshal
	var swapAmountWithLimit2 SwapAmountWithLimit
	err = json.Unmarshal(swapAmountExactOutJson, &swapAmountWithLimit2)
	require.NoError(t, err)
	// Check
	assert.Equal(t, swapAmountExactOut, swapAmountWithLimit2)
}
