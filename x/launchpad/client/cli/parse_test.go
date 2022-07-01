package cli

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v7/x/launchpad/types"
)

func TestUnmarslahCreateSaleInputs(t *testing.T) {
	tIn := "tIn"
	tOut := "1000tOut"
	maxFee := "200tIn2"
	start := "2022-02-03T15:00:00.000Z"
	recipient := "osmo1r85gjuck87f9hw7l2c30w3zh696xrq0lus0kq6"
	creator := "osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44"
	url1 := "https://api.ipfsbrowser.com/ipfs/get.php?hash=QmcGV8fimB7aeBxnDqr7bSSLUWLeyFKUukGqDhWnvriQ3T"
	tcs := []struct {
		name     string
		errMsg   string
		input    string
		expected types.MsgCreateSale
	}{
		{
			"valid1", "",
			fmt.Sprintf(
				`{"token-in": "%s", "token-out": "%s", "start-time": "%s", "duration": "24h", "recipient": "%s", "name": "my token sale", "url": "%s"}`, tIn, tOut, start, recipient, url1),
			types.MsgCreateSale{Creator: creator, TokenIn: tIn, TokenOut: sdk.Coin{"tOut", sdk.NewInt(1000)}, StartTime: time.Date(2022, 2, 3, 15, 0, 0, 0, time.UTC), Duration: 24 * time.Hour, Recipient: recipient, MaxFee: []sdk.Coin{}},
		},
		{
			"valid1-with-max-fee", "",
			fmt.Sprintf(
				`{"token-in": "%s", "token-out": "%s", "start-time": "%s", "duration": "24h", "recipient": "%s", "name": "my token sale", "url": "%s", "max-fee": ["%s"]}`, tIn, tOut, start, recipient, url1, maxFee),
			types.MsgCreateSale{Creator: creator, TokenIn: tIn, TokenOut: sdk.Coin{"tOut", sdk.NewInt(1000)}, StartTime: time.Date(2022, 2, 3, 15, 0, 0, 0, time.UTC), Duration: 24 * time.Hour, Recipient: recipient, MaxFee: []sdk.Coin{{"tIn2", sdk.NewInt(200)}}},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)
			var i createSaleInputs
			require.NoError(i.UnmarshalJSON([]byte(tc.input)))
			m, err := i.ToMsgCreateSale(creator)
			if tc.errMsg == "" {
				require.NoError(err)
				require.Equal(tc.expected.Creator, m.Creator)
				require.Equal(tc.expected.TokenIn, m.TokenIn)
				require.Equal(tc.expected.TokenOut.String(), m.TokenOut.String())
				require.Equal(tc.expected.StartTime, m.StartTime)
				require.Equal(tc.expected.Duration, m.Duration)
				require.Equal(tc.expected.Recipient, m.Recipient)
				require.Equal(tc.expected.MaxFee, m.MaxFee)
			} else {
				require.ErrorContains(err, tc.errMsg)
			}
		})
	}
}
