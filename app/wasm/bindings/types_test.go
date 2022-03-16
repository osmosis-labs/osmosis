package wasmbindings

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	swap = Swap{
		PoolId:   1,
		DenomIn:  "denom_in",
		DenomOut: "denom_out",
	}
	step = Step{
		PoolId:   2,
		DenomOut: "denom_out",
	}
	in           = sdk.NewInt(123)
	out          = sdk.NewInt(456)
	swapAmountIn = SwapAmount{
		In: &in,
	}
	swapAmountOut = SwapAmount{
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
	swapAmountExactIn = SwapAmountWithLimit{
		ExactIn: &exactIn,
	}
	swapAmountExactOut = SwapAmountWithLimit{
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
