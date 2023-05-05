package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	amountOutPlaceholder = "out"
	amountInPlaceholder  = "in"
)

var (
	ErrEmptyRoutes               = errors.New("provided empty routes")
	ErrTooFewPoolAssets          = errors.New("pool should have at least 2 assets, as they must be swapping between at least two assets")
	ErrTooManyPoolAssets         = errors.New("pool has too many assets (currently capped at 8 assets per pool)")
	ErrDuplicateRoutesNotAllowed = errors.New("duplicate multihop routes are not allowed")
)

type nonPositiveAmountError struct {
	Amount string
}

func (e nonPositiveAmountError) Error() string {
	return fmt.Sprintf("min out amount or max in amount should be positive, was (%s)", e.Amount)
}

type FailedToFindRouteError struct {
	PoolId uint64
}

func (e FailedToFindRouteError) Error() string {
	return fmt.Sprintf("failed to find route for pool id (%d)", e.PoolId)
}

type UndefinedRouteError struct {
	PoolType PoolType
	PoolId   uint64
}

func (e UndefinedRouteError) Error() string {
	return fmt.Sprintf("route is not defined for the given pool type (%s) and pool id (%d)", e.PoolType, e.PoolId)
}

type FinalAmountIsNotPositiveError struct {
	IsAmountOut bool
	Amount      sdk.Int
}

func (e FinalAmountIsNotPositiveError) Error() string {
	amountPlaceholder := amountOutPlaceholder
	if !e.IsAmountOut {
		amountPlaceholder = amountInPlaceholder
	}
	return fmt.Sprintf("final total amount (%s) must be positive, was (%d)", amountPlaceholder, e.Amount)
}

type PriceImpactProtectionExactInError struct {
	Actual    sdk.Int
	MinAmount sdk.Int
}

func (e PriceImpactProtectionExactInError) Error() string {
	return fmt.Sprintf("price impact protection: expected %s to be at least %s", e.Actual, e.MinAmount)
}

type PriceImpactProtectionExactOutError struct {
	Actual    sdk.Int
	MaxAmount sdk.Int
}

func (e PriceImpactProtectionExactOutError) Error() string {
	return fmt.Sprintf("price impact protection: expected %s to be at most %s", e.Actual, e.MaxAmount)
}

type InvalidFinalTokenOutError struct {
	TokenOutGivenA string
	TokenOutGivenB string
}

func (e InvalidFinalTokenOutError) Error() string {
	return fmt.Sprintf("invalid final token out, each path must end on the same token out, had (%s) and (%s)  mismatch", e.TokenOutGivenA, e.TokenOutGivenB)
}

type InvalidSenderError struct {
	Sender string
}

func (e InvalidSenderError) Error() string {
	return fmt.Sprintf("Invalid sender address (%s)", e.Sender)
}

type InvalidPoolCreatorError struct {
	CreatorAddresss string
}

func (e InvalidPoolCreatorError) Error() string {
	return fmt.Sprintf("invalid pool creator (%s), only poolmanager can create pools with no fee", e.CreatorAddresss)
}

type InvalidPoolTypeError struct {
	PoolType PoolType
}

func (e InvalidPoolTypeError) Error() string {
	return fmt.Sprintf("invalid pool type (%s)", PoolType_name[int32(e.PoolType)])
}

type IncorrectPoolIdError struct {
	ExpectedPoolId uint64
	ActualPoolId   uint64
}

func (e IncorrectPoolIdError) Error() string {
	return fmt.Sprintf("Pool was attempted to be created with incorrect pool ID. Expected (%d), actual (%d)", e.ExpectedPoolId, e.ActualPoolId)
}

type IncorrectPoolAddressError struct {
	ExpectedPoolAddress string
	ActualPoolAddress   string
}

func (e IncorrectPoolAddressError) Error() string {
	return fmt.Sprintf("Pool was attempted to be created with incorrect pool address. Expected (%s), actual (%s)", e.ExpectedPoolAddress, e.ActualPoolAddress)
}
