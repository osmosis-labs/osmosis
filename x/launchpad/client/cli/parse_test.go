package cli

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUnmarslahCreateSaleInputs(t *testing.T) {
	require := require.New(t)
	tIn := "tIn"
	tOut := "1000tOut"
	start := "2022-02-03T15:00:00.000Z"
	recipient := "osmo1r85gjuck87f9hw7l2c30w3zh696xrq0lus0kq6"
	creator := "osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44"
	url1 := "https://api.ipfsbrowser.com/ipfs/get.php?hash=QmcGV8fimB7aeBxnDqr7bSSLUWLeyFKUukGqDhWnvriQ3T"
	input := fmt.Sprintf(
		`{"token-in": "%s", "token-out": "%s", "start-time": "%s", "duration": "24h", "recipient": "%s", "name": "my token sale", "url": "%s"}`,
		tIn, tOut, start, recipient, url1)
	var i createSaleInputs
	require.NoError(i.UnmarshalJSON([]byte(input)))
	m, err := i.ToMsgCreateSale(creator)
	require.NoError(err)
	require.Equal(m.Creator, creator)
	require.Equal(m.TokenIn, tIn)
	require.Equal(m.TokenOut.String(), tOut)
	require.Equal(m.StartTime, time.Date(2022, 2, 3, 15, 0, 0, 0, time.UTC))
	require.Equal(m.Duration, 24*time.Hour)
	require.Equal(m.Recipient, recipient)
}
