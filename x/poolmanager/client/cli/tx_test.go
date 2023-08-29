package cli_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/client/cli"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

func TestParseCoinsNoSort(t *testing.T) {
	const (
		a = "aaa"
		b = "bbb"
		c = "ccc"
		d = "ddd"
	)

	var (
		ten = sdk.NewInt(10)

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
				sdk.NewCoin(d, sdk.NewInt(100)),
				sdk.NewCoin(b, sdk.NewInt(20)),
				sdk.NewCoin(a, sdk.NewInt(300)),
				sdk.NewCoin(c, sdk.NewInt(40)),
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

func TestParseDenomPairTakerFee(t *testing.T) {
	tests := map[string]struct {
		denomPairTakerFeeStr      string
		expectedDenomPairTakerFee []types.DenomPairTakerFee
		expectedError             error
	}{
		"one set": {
			denomPairTakerFeeStr: "uion,uosmo,0.0016",
			expectedDenomPairTakerFee: []types.DenomPairTakerFee{
				{
					Denom0:   "uion",
					Denom1:   "uosmo",
					TakerFee: sdk.MustNewDecFromStr("0.0016"),
				}},
		},
		"two sets": {
			denomPairTakerFeeStr: "uion,uosmo,0.0016,stake,uosmo,0.005",
			expectedDenomPairTakerFee: []types.DenomPairTakerFee{
				{
					Denom0:   "uion",
					Denom1:   "uosmo",
					TakerFee: sdk.MustNewDecFromStr("0.0016"),
				},
				{
					Denom0:   "stake",
					Denom1:   "uosmo",
					TakerFee: sdk.MustNewDecFromStr("0.005"),
				},
			},
		},
		"error: wrong format": {
			denomPairTakerFeeStr: "uion,uosmo,0.0016,stake",
			expectedError:        fmt.Errorf("denomPairTakerFeeRecords must be a list of denom0, denom1, and takerFee separated by commas"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			denomPairTakerFee, err := cli.ParseDenomPairTakerFee(tc.denomPairTakerFeeStr)
			if tc.expectedError != nil {
				require.ErrorAs(t, err, &tc.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedDenomPairTakerFee, denomPairTakerFee)
		})
	}
}
