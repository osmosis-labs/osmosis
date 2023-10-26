package types

import (
	"errors"
	fmt "fmt"

	"cosmossdk.io/math"
)

var (
	ErrNoDelegation = errors.New("No existing delegation")
)

type UndelegateMoreThanDelegatedError struct {
	TotalDelegatedAmt math.LegacyDec
	UndelegationAmt   math.Int
}

func (e UndelegateMoreThanDelegatedError) Error() string {
	return fmt.Sprintf("total tokenAmountToUndelegate more than delegated amount have %s got %s\n", e.TotalDelegatedAmt, e.UndelegationAmt)
}

type NoValidatorSetOrExistingDelegationsError struct {
	DelegatorAddr string
}

func (e NoValidatorSetOrExistingDelegationsError) Error() string {
	return fmt.Sprintf("user %s doesn't have validator set or existing delegations", e.DelegatorAddr)
}

type ValsetRatioGreaterThanOneError struct {
	ValsetRatio math.LegacyDec
}

func (e ValsetRatioGreaterThanOneError) Error() string {
	return fmt.Sprintf("valset ratio %s greater than one", e.ValsetRatio)
}

type ValidatorNotFoundError struct {
	ValidatorAddr string
}

func (e ValidatorNotFoundError) Error() string {
	return fmt.Sprintf("validator %s not found", e.ValidatorAddr)
}
