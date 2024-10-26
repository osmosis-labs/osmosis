package types

import (
	"errors"
	"fmt"

	"github.com/osmosis-labs/osmosis/osmomath"
)

const (
	amountOutPlaceholder = "out"
	amountInPlaceholder  = "in"
)

var (
	ErrEmptyRoutes                                  = errors.New("provided empty routes")
	ErrTooFewPoolAssets                             = errors.New("pool should have at least 2 assets, as they must be swapping between at least two assets")
	ErrTooManyPoolAssets                            = errors.New("pool has too many assets (currently capped at 8 assets per pool)")
	ErrDuplicateRoutesNotAllowed                    = errors.New("duplicate multihop routes are not allowed")
	ErrUnauthorizedGov                              = errors.New("only the governance module is allowed to execute this message")
	ErrSetTakerFeeShareAgreementsMapCached          = errors.New("error setting taker fee share agreements map cache")
	ErrSetTakerFeeRevenueShareUserMapCached         = errors.New("error setting taker fee revenue share user map cache")
	ErrSetTakerFeeRevenueShareSignupLookupMapCached = errors.New("error setting taker fee revenue share signup lookup map cache")
	ErrSetAllRegisteredAlloyedPoolsByDenomCached    = errors.New("error setting all registered alloyed pools by denom cache")
	ErrSetAllRegisteredAlloyedPoolsIdArrayCached    = errors.New("error setting registered alloyed pool ID array cache")
	ErrSetRegisteredAlloyedPool                     = errors.New("error setting registered alloyed pool")
	ErrInvalidKeyFormat                             = errors.New("invalid key format")
	ErrTotalAlloyedLiquidityIsZero                  = errors.New("totalAlloyedLiquidity is zero")
	ErrBadExecution                                 = errors.New("cannot execute contract: %v")
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
	Amount      osmomath.Int
}

func (e FinalAmountIsNotPositiveError) Error() string {
	amountPlaceholder := amountOutPlaceholder
	if !e.IsAmountOut {
		amountPlaceholder = amountInPlaceholder
	}
	return fmt.Sprintf("final total amount (%s) must be positive, was (%d)", amountPlaceholder, e.Amount)
}

type PriceImpactProtectionExactInError struct {
	Actual    osmomath.Int
	MinAmount osmomath.Int
}

func (e PriceImpactProtectionExactInError) Error() string {
	return fmt.Sprintf("price impact protection: expected %s to be at least %s", e.Actual, e.MinAmount)
}

type PriceImpactProtectionExactOutError struct {
	Actual    osmomath.Int
	MaxAmount osmomath.Int
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

type InactivePoolError struct {
	PoolId uint64
}

func (e InactivePoolError) Error() string {
	return fmt.Sprintf("Pool %d is not active.", e.PoolId)
}

type NotCosmWasmPoolError struct {
	PoolId uint64
}

func (e NotCosmWasmPoolError) Error() string {
	return fmt.Sprintf("pool with id %d is not a CosmWasmPool", e.PoolId)
}

type NoAccruedValueError struct {
	TakerFeeShareDenom   string
	TakerFeeChargedDenom string
}

func (e NoAccruedValueError) Error() string {
	return fmt.Sprintf("no accrued value found for takerFeeShareDenom %v and takerFeeChargedDenom %s", e.TakerFeeShareDenom, e.TakerFeeChargedDenom)
}

type NoRegisteredAlloyedPoolError struct {
	PoolId uint64
}

func (e NoRegisteredAlloyedPoolError) Error() string {
	return fmt.Sprintf("no registered alloyed pool found for poolId %d", e.PoolId)
}

type InvalidAlloyedDenomFormatError struct {
	PartsLength int
}

func (e InvalidAlloyedDenomFormatError) Error() string {
	return fmt.Sprintf("invalid format for alloyedDenom, expected 4 parts but got %d", e.PartsLength)
}

type InvalidAlloyedDenomPartError struct {
	PartIndex int
	Expected  string
	Actual    string
}

func (e InvalidAlloyedDenomPartError) Error() string {
	return fmt.Sprintf("part %d of alloyedDenom should be '%s', but got '%s'", e.PartIndex, e.Expected, e.Actual)
}

type InvalidAlloyedPoolIDError struct {
	AlloyedIDStr string
	Err          error
}

func (e InvalidAlloyedPoolIDError) Error() string {
	return fmt.Sprintf("failed to parse alloyed pool ID '%s': %v", e.AlloyedIDStr, e.Err)
}

func (e InvalidAlloyedPoolIDError) Unwrap() error {
	return e.Err
}

type InvalidTakerFeeSharePercentageError struct {
	Percentage osmomath.Dec
}

func (e InvalidTakerFeeSharePercentageError) Error() string {
	return fmt.Sprintf("invalid taker fee share percentage: %s, must be between 0 and 1", e.Percentage)
}
