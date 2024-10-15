package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

func TestSudoMsgString(t *testing.T) {
	testCases := []struct {
		testCase    string
		msg         types.SudoMsg
		expectedMsg string
	}{
		{
			"ok: callback job_id is 1",
			types.NewCallbackMsg(1),
			`{"callback":{"job_id":1}}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			res := tc.msg.String()
			require.EqualValues(t, tc.expectedMsg, res)
		})
	}
}
