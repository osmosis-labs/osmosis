package cli_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

<<<<<<< HEAD
	"github.com/osmosis-labs/osmosis/v18/x/gamm/client/cli"
=======
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/gamm/client/cli"
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
)

func TestParseCoinsNoSort(t *testing.T) {
	const (
		a = "aaa"
		b = "bbb"
		c = "ccc"
		d = "ddd"
	)

	var (
		ten = osmomath.NewInt(10)

		coinA = sdk.NewCoin(a, ten)
		coinB = sdk.NewCoin(b, ten)
		coinC = sdk.NewCoin(c, ten)
		coinD = sdk.NewCoin(d, ten)
	)

	tests := map[string]struct {
		coinsStr      string
		expectedCoins sdk.Coins
	}{
		"ascending": {
			coinsStr: "10aaa,10bbb,10ccc,10ddd",
			expectedCoins: sdk.Coins{
				coinA,
				coinB,
				coinC,
				coinD,
			},
		},
		"descending": {
			coinsStr: "10ddd,10ccc,10bbb,10aaa",
			expectedCoins: sdk.Coins{
				coinD,
				coinC,
				coinB,
				coinA,
			},
		},
		"mixed with different values.": {
			coinsStr: "100ddd,20bbb,300aaa,40ccc",
			expectedCoins: sdk.Coins{
				sdk.NewCoin(d, osmomath.NewInt(100)),
				sdk.NewCoin(b, osmomath.NewInt(20)),
				sdk.NewCoin(a, osmomath.NewInt(300)),
				sdk.NewCoin(c, osmomath.NewInt(40)),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			coins, err := cli.ParseCoinsNoSort(tc.coinsStr)

			require.NoError(t, err)
			require.Equal(t, tc.expectedCoins, coins)
		})
	}
}
