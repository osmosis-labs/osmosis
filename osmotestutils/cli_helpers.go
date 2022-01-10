package osmotestutils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultFeeString(cfg network.Config) string {
	feeCoins := sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, sdk.NewInt(10)))
	return fmt.Sprintf("--%s=%s", flags.FlagFees, feeCoins.String())
}

var (
	base   = 10
	bitlen = 64
)

func ParseUint64SliceFromString(s string, seperator string) ([]uint64, error) {
	var ids []uint64
	for _, s := range strings.Split(s, seperator) {
		s = strings.TrimSpace(s)

		parsed, err := strconv.ParseUint(s, base, bitlen)
		if err != nil {
			return []uint64{}, err
		}
		ids = append(ids, parsed)
	}
	return ids, nil
}

func ParseSdkIntFromString(s string, seperator string) ([]sdk.Int, error) {
	var weights []sdk.Int
	for _, weightStr := range strings.Split(s, seperator) {
		weightStr = strings.TrimSpace(weightStr)

		parsed, err := strconv.ParseUint(weightStr, base, bitlen)
		if err != nil {
			return weights, err
		}
		weights = append(weights, sdk.NewIntFromUint64(parsed))
	}
	return weights, nil
}
