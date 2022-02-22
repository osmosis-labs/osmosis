package keeper

import (
	"fmt"
	"strings"
)

func stakingSecondaryIndex(denom, valAddr string) string {
	return fmt.Sprintf("%s/superbonding/%s", denom, valAddr)
}

func unstakingSecondaryIndex(denom, valAddr string) string {
	return fmt.Sprintf("%s/superunbonding/%s", denom, valAddr)
}

// quick fix for getting the validator addresss from a synthetic denom
func ValidatorAddressFromSyntheticDenom(suffix string) (string, error) {
	if strings.Contains(suffix, "superbonding") {
		return strings.TrimLeft(suffix, "/superbonding/"), nil
	}
	if strings.Contains(suffix, "superunbonding") {
		return strings.TrimLeft(suffix, "/superunbonding/"), nil
	}
	return "", fmt.Errorf("%s is not a valid synthetic denom suffix", suffix)
}
