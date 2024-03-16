package osmoutils

import (
	"testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func decCoin(denom string, amount int64) sdk.DecCoin {
	mInt := math.NewInt(amount)
	return sdk.NewDecCoin(denom, mInt)
}

var subDecValues = []struct {
	name string
	a, b []sdk.DecCoins
	want []sdk.DecCoins
}{
	{
		name: "empty coins",
		a:    []sdk.DecCoins{},
		b:    []sdk.DecCoins{},
		want: []sdk.DecCoins{},
	},
	{
		name: "1 coin per set",
		a:    []sdk.DecCoins{sdk.NewDecCoins(decCoin("osmo", 100))},
		b:    []sdk.DecCoins{sdk.NewDecCoins(decCoin("osmo", 10))},
		want: []sdk.DecCoins{sdk.NewDecCoins(decCoin("osmo", 90))},
	},
	{
		name: "2 uniq coins",
		a:    []sdk.DecCoins{sdk.NewDecCoins(decCoin("osmo", 100), decCoin("uosmo", 200))},
		b:    []sdk.DecCoins{sdk.NewDecCoins(decCoin("osmo", 10), decCoin("uosmo", 20))},
		want: []sdk.DecCoins{sdk.NewDecCoins(decCoin("osmo", 90), decCoin("uosmo", 180))},
	},
	{
		name: "10 uniq coins",
		a: []sdk.DecCoins{
			sdk.NewDecCoins(
				decCoin("osmo", 100), decCoin("uosmo", 200), decCoin("qck", 271), decCoin("tia", 2280),
				decCoin("mosmo", 100), decCoin("posmo", 200), decCoin("uqck", 271), decCoin("utia", 2280),
				decCoin("atom", 100), decCoin("uatom", 200),
			),
		},
		b: []sdk.DecCoins{
			sdk.NewDecCoins(
				decCoin("osmo", 80), decCoin("uosmo", 200), decCoin("qck", 270), decCoin("tia", 1000),
				decCoin("mosmo", 100), decCoin("posmo", 200), decCoin("uqck", 271), decCoin("utia", 2200),
				decCoin("atom", 99), decCoin("uatom", 192),
			),
		},
		want: []sdk.DecCoins{
			sdk.NewDecCoins(
				decCoin("osmo", 20), decCoin("uosmo", 0), decCoin("qck", 1), decCoin("tia", 1280),
				decCoin("mosmo", 0), decCoin("posmo", 0), decCoin("uqck", 0), decCoin("utia", 80),
				decCoin("atom", 1), decCoin("uatom", 8),
			),
		},
	},
}

var sink any = nil

func BenchmarkSafeSubDecCoinArrays(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, val := range subDecValues {
			got, err := SafeSubDecCoinArrays(val.a, val.b)
			if err != nil {
				b.Fatal(err)
			}
			if g, w := len(got), len(val.want); g != w {
				b.Fatalf("Unequal lengths\n\tGot:  %d\n\twant: %d", g, w)
			}
			for j := range got {
				gj := got[j]
				wj := val.want[j]
				if !gj.IsEqual(wj) {
					b.Fatalf("#%d: unexpected result\n\tGot:  %v\n\tWant: %v", j, gj, wj)
				}
			}
			sink = got
		}
	}
	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}
	sink = nil
}

func BenchmarkSubDecCoinArrays(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, val := range subDecValues {
			got, err := SubDecCoinArrays(val.a, val.b)
			if err != nil {
				b.Fatal(err)
			}
			if g, w := len(got), len(val.want); g != w {
				b.Fatalf("Unequal lengths\n\tGot:  %d\n\twant: %d", g, w)
			}
			for j := range got {
				gj := got[j]
				wj := val.want[j]
				if !gj.IsEqual(wj) {
					b.Fatalf("#%d: unexpected result\n\tGot:  %v\n\tWant: %v", j, gj, wj)
				}
			}
			sink = got
		}
	}
	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}
	sink = nil
}
