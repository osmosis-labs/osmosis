package osmoutils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

func DefaultFeeString(cfg network.Config) string {
	feeCoins := sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, osmomath.NewInt(10)))
	return fmt.Sprintf("--%s=%s", flags.FlagFees, feeCoins.String())
}

const (
	base   = 10
	bitlen = 64
)

func ParseUint64SliceFromString(s string, separator string) ([]uint64, error) {
	var parsedInts []uint64
	for _, s := range strings.Split(s, separator) {
		s = strings.TrimSpace(s)

		parsed, err := strconv.ParseUint(s, base, bitlen)
		if err != nil {
			return []uint64{}, err
		}
		parsedInts = append(parsedInts, parsed)
	}
	return parsedInts, nil
}

func ParseSdkIntFromString(s string, separator string) ([]osmomath.Int, error) {
	var parsedInts []osmomath.Int
	for _, weightStr := range strings.Split(s, separator) {
		weightStr = strings.TrimSpace(weightStr)

		parsed, err := strconv.ParseUint(weightStr, base, bitlen)
		if err != nil {
			return parsedInts, err
		}
		parsedInts = append(parsedInts, osmomath.NewIntFromUint64(parsed))
	}
	return parsedInts, nil
}

func ParseSdkDecFromString(s string, separator string) ([]osmomath.Dec, error) {
	var parsedDec []osmomath.Dec
	for _, weightStr := range strings.Split(s, separator) {
		weightStr = strings.TrimSpace(weightStr)

		parsed, err := osmomath.NewDecFromStr(weightStr)
		if err != nil {
			return parsedDec, err
		}

		parsedDec = append(parsedDec, parsed)
	}
	return parsedDec, nil
}

// CreateRandomAccounts is a function return a list of randomly generated AccAddresses
func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}
