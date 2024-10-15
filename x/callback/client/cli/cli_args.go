package cli

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParseAccAddressArg is a helper function to parse an account address CLI argument.
func ParseAccAddressArg(argName, argValue string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(argValue)
	if err != nil {
		return sdk.AccAddress{}, fmt.Errorf("parsing %s argument: invalid address: %w", argName, err)
	}

	return addr, nil
}

// ParseUint64Arg is a helper function to parse uint64 CLI argument.
func ParseUint64Arg(argName, argValue string) (uint64, error) {
	v, err := strconv.ParseUint(argValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %s argument: invalid uint64 value: %w", argName, err)
	}

	return v, nil
}

// ParseInt64Arg is a helper function to parse int64 CLI argument.
func ParseInt64Arg(argName, argValue string) (int64, error) {
	v, err := strconv.ParseInt(argValue, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %s argument: invalid uint64 value: %w", argName, err)
	}

	return v, nil
}

// ParseCoinArg is a helper function to parse uint64 CLI argument.
func ParseCoinArg(argName, argValue string) (sdk.Coin, error) {
	deposit, err := sdk.ParseCoinNormalized(argValue)
	if err != nil {
		return deposit, fmt.Errorf("parsing %s argument: invalid sdk.Coin value: %w", argName, err)
	}
	return deposit, nil
}
