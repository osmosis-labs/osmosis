package ante

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func TestSendBlockDecorator(t *testing.T) {
	testCases := []struct {
		from       sdk.AccAddress
		to         sdk.AccAddress
		expectPass bool
	}{
		{sdk.AccAddress("1234"), sdk.AccAddress("zxcv"), true},
		{sdk.AccAddress("asdf"), sdk.AccAddress("zxcv"), true},
		{sdk.AccAddress("asdf"), sdk.AccAddress("bnm,"), false},
	}

	permittedOnlySendTo := map[string]string{
		sdk.AccAddress("asdf").String(): sdk.AccAddress("zxcv").String(),
	}
	decorator := NewSendBlockDecorator(permittedOnlySendTo)

	for _, testCase := range testCases {
		err := decorator.CheckIfBlocked([]sdk.Msg{bank.NewMsgSend(testCase.from, testCase.to, sdk.NewCoins(sdk.NewInt64Coin("test", 1)))})
		if testCase.expectPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}

}
