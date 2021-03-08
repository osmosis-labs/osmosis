package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSdkIntMaxValue(t *testing.T) {
	require.Panics(t, func() {
		sdkIntMaxValue.Add(sdk.NewInt(1))
	})
}
