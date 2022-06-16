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
		{sdk.AccAddress("honest-sender_______"), sdk.AccAddress("honest-address"), true},
		{sdk.AccAddress("honest-sender_______"), sdk.AccAddress("recovery-address"), true},
		{sdk.AccAddress("malicious-sender____"), sdk.AccAddress("recovery-address"), true},
		{sdk.AccAddress("malicious-sender____"), sdk.AccAddress("random-address"), false},
	}

	permittedOnlySendTo := map[string]string{
		sdk.AccAddress("malicious-sender____").String(): sdk.AccAddress("recovery-address").String(),
	}
	decorator := NewSendBlockDecorator(SendBlockOptions{permittedOnlySendTo})

	for _, testCase := range testCases {
		err := decorator.CheckIfBlocked(
			[]sdk.Msg{
				bank.NewMsgSend(testCase.from, testCase.to, sdk.NewCoins(sdk.NewInt64Coin("test", 1))),
			})
		if testCase.expectPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
