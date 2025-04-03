package types_test

import (
	"fmt"
	"github.com/osmosis-labs/osmosis/osmomath"
	"math"
	"sort"
	"strconv"
	"testing"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"
)

func TestToMap(t *testing.T) {
	tests := struct {
		votes   []types.VoteForTally
		isValid []bool
	}{
		[]types.VoteForTally{
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        assets.MicroKRWDenom,
				ExchangeRate: osmomath.NewDec(1600),
				Power:        100,
			},
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        assets.MicroKRWDenom,
				ExchangeRate: osmomath.ZeroDec(),
				Power:        100,
			},
			{
				Voter:        sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address()),
				Denom:        assets.MicroKRWDenom,
				ExchangeRate: osmomath.NewDec(1500),
				Power:        100,
			},
		},
		[]bool{true, false, true},
	}

	pb := types.ExchangeRateBallot(tests.votes)
	mapData := pb.ToMap()
	for i, vote := range tests.votes {
		exchangeRate, ok := mapData[string(vote.Voter)]
		if tests.isValid[i] {
			require.True(t, ok)
			require.Equal(t, exchangeRate, vote.ExchangeRate)
		} else {
			require.False(t, ok)
		}
	}
}

func TestToCrossRate(t *testing.T) {
	data := []struct {
		base     osmomath.Dec
		quote    osmomath.Dec
		expected osmomath.Dec
	}{
		{
			base:     osmomath.NewDec(1600),
			quote:    osmomath.NewDec(100),
			expected: osmomath.NewDec(16),
		},
		{
			base:     osmomath.NewDec(0),
			quote:    osmomath.NewDec(100),
			expected: osmomath.NewDec(16),
		},
		{
			base:     osmomath.NewDec(1600),
			quote:    osmomath.NewDec(0),
			expected: osmomath.NewDec(16),
		},
	}

	pbBase := types.ExchangeRateBallot{}
	pbQuote := types.ExchangeRateBallot{}
	cb := types.ExchangeRateBallot{}
	for _, data := range data {
		valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())
		if !data.base.IsZero() {
			pbBase = append(pbBase, types.NewVoteForTally(data.base, assets.MicroKRWDenom, valAddr, 100))
		}

		pbQuote = append(pbQuote, types.NewVoteForTally(data.quote, assets.MicroKRWDenom, valAddr, 100))

		if !data.base.IsZero() && !data.quote.IsZero() {
			cb = append(cb, types.NewVoteForTally(data.base.Quo(data.quote), assets.MicroKRWDenom, valAddr, 100))
		} else {
			cb = append(cb, types.NewVoteForTally(osmomath.ZeroDec(), assets.MicroKRWDenom, valAddr, 0))
		}
	}

	baseMapBallot := pbBase.ToMap()
	require.Equal(t, cb, pbQuote.ToCrossRate(baseMapBallot))

	sort.Sort(cb)

	require.Equal(t, cb, pbQuote.ToCrossRateWithSort(baseMapBallot))
}

func TestSqrt(t *testing.T) {
	num := osmomath.NewDecWithPrec(144, 4)
	floatNum, err := strconv.ParseFloat(num.String(), 64)
	require.NoError(t, err)

	floatNum = math.Sqrt(floatNum)
	num, err = osmomath.NewDecFromStr(fmt.Sprintf("%f", floatNum))
	require.NoError(t, err)

	require.Equal(t, osmomath.NewDecWithPrec(12, 2), num)
}

func TestPBPower(t *testing.T) {
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	_, valAccAddrs, sk := types.GenerateRandomTestCase()
	pb := types.ExchangeRateBallot{}
	ballotPower := int64(0)

	for i := 0; i < len(sk.Validators()); i++ {
		val, err := sk.GetValidator(ctx, valAccAddrs[i])
		require.NoError(t, err)
		power := val.GetConsensusPower(sdk.DefaultPowerReduction)
		vote := types.NewVoteForTally(
			osmomath.ZeroDec(),
			assets.MicroSDRDenom,
			valAccAddrs[i],
			power,
		)

		pb = append(pb, vote)

		require.NotEqual(t, int64(0), vote.Power)

		ballotPower += vote.Power
	}

	require.Equal(t, ballotPower, pb.Power())

	// Mix in a fake validator, the total power should not have changed.
	pubKey := secp256k1.GenPrivKey().PubKey()
	faceValAddr := sdk.ValAddress(pubKey.Address())
	fakeVote := types.NewVoteForTally(
		osmomath.OneDec(),
		assets.MicroSDRDenom,
		faceValAddr,
		0,
	)

	pb = append(pb, fakeVote)
	require.Equal(t, ballotPower, pb.Power())
}

func TestPBWeightedMedian(t *testing.T) {
	tests := []struct {
		inputs      []int64
		weights     []int64
		isValidator []bool
		median      osmomath.Dec
		panic       bool
	}{
		{
			// Supermajority one number
			[]int64{1, 2, 10, 100000},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			osmomath.NewDec(10),
			false,
		},
		{
			// Adding fake validator doesn't change outcome
			[]int64{1, 2, 10, 100000, 10000000000},
			[]int64{1, 1, 100, 1, 10000},
			[]bool{true, true, true, true, false},
			osmomath.NewDec(10),
			false,
		},
		{
			// Tie votes
			[]int64{1, 2, 3, 4},
			[]int64{1, 100, 100, 1},
			[]bool{true, true, true, true},
			osmomath.NewDec(2),
			false,
		},
		{
			// No votes
			[]int64{},
			[]int64{},
			[]bool{true, true, true, true},
			osmomath.NewDec(0),
			false,
		},
		{
			// not sorted panic
			[]int64{2, 1, 10, 100000},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			osmomath.NewDec(10),
			true,
		},
	}

	for _, tc := range tests {
		pb := types.ExchangeRateBallot{}
		for i, input := range tc.inputs {
			valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())

			power := tc.weights[i]
			if !tc.isValidator[i] {
				power = 0
			}

			vote := types.NewVoteForTally(
				osmomath.NewDec(input),
				assets.MicroSDRDenom,
				valAddr,
				power,
			)

			pb = append(pb, vote)
		}

		if tc.panic {
			require.Panics(t, func() {
				if !sort.IsSorted(pb) {
					panic("ballot must be sorted")
				}
			})
		} else {
			require.Equal(t, tc.median, pb.WeightedMedian())
		}
	}
}

func TestPBStandardDeviation(t *testing.T) {
	tests := []struct {
		inputs            []float64
		weights           []int64
		isValidator       []bool
		standardDeviation osmomath.Dec
	}{
		{
			// Supermajority one number
			[]float64{1.0, 2.0, 10.0, 100000.0},
			[]int64{1, 1, 100, 1},
			[]bool{true, true, true, true},
			osmomath.NewDecWithPrec(4999500036300, types.OracleDecPrecision),
		},
		{
			// Adding fake validator doesn't change outcome
			[]float64{1.0, 2.0, 10.0, 100000.0, 10000000000},
			[]int64{1, 1, 100, 1, 10000},
			[]bool{true, true, true, true, false},
			osmomath.NewDecWithPrec(447213595075100600, types.OracleDecPrecision),
		},
		{
			// Tie votes
			[]float64{1.0, 2.0, 3.0, 4.0},
			[]int64{1, 100, 100, 1},
			[]bool{true, true, true, true},
			osmomath.NewDecWithPrec(122474500, types.OracleDecPrecision),
		},
		{
			// No votes
			[]float64{},
			[]int64{},
			[]bool{true, true, true, true},
			osmomath.NewDecWithPrec(0, 0),
		},
	}

	base := math.Pow10(types.OracleDecPrecision)
	for _, tc := range tests {
		pb := types.ExchangeRateBallot{}
		for i, input := range tc.inputs {
			valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())

			power := tc.weights[i]
			if !tc.isValidator[i] {
				power = 0
			}

			vote := types.NewVoteForTally(
				osmomath.NewDecWithPrec(int64(input*base), int64(types.OracleDecPrecision)),
				assets.MicroSDRDenom,
				valAddr,
				power,
			)

			pb = append(pb, vote)
		}

		require.Equal(t, tc.standardDeviation, pb.StandardDeviation(pb.WeightedMedian()))
	}
}

func TestPBStandardDeviationOverflow(t *testing.T) {
	valAddr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address())
	exchangeRate, err := osmomath.NewDecFromStr("100000000000000000000000000000000000000000000000000000000.0")
	require.NoError(t, err)

	pb := types.ExchangeRateBallot{types.NewVoteForTally(
		osmomath.ZeroDec(),
		assets.MicroSDRDenom,
		valAddr,
		2,
	), types.NewVoteForTally(
		exchangeRate,
		assets.MicroSDRDenom,
		valAddr,
		1,
	)}

	require.Equal(t, osmomath.ZeroDec(), pb.StandardDeviation(pb.WeightedMedian()))
}

func TestNewClaim(t *testing.T) {
	power := int64(10)
	weight := int64(11)
	winCount := int64(1)
	addr := sdk.ValAddress(secp256k1.GenPrivKey().PubKey().Address().Bytes())
	claim := types.NewClaim(power, weight, winCount, addr)
	require.Equal(t, types.Claim{
		Power:     power,
		Weight:    weight,
		WinCount:  winCount,
		Recipient: addr,
	}, claim)
}
