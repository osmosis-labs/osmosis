package types

import (
	"errors"
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrValueParse  = errors.New("value parse error")
)

// x/concentrated-liquidity module sentinel errors.
type InvalidLowerUpperTickError struct {
	LowerTick int64
	UpperTick int64
}

func (e InvalidLowerUpperTickError) Error() string {
	return fmt.Sprintf("Lower tick must be lesser than upper. Got lower: %d, upper: %d", e.LowerTick, e.UpperTick)
}

type NotPositiveRequireAmountError struct {
	Amount string
}

func (e NotPositiveRequireAmountError) Error() string {
	return fmt.Sprintf("Required amount should be positive. Got: %s", e.Amount)
}

type PositionNotFoundError struct {
	PoolId         uint64
	LowerTick      int64
	UpperTick      int64
	JoinTime       time.Time
	FreezeDuration time.Duration
}

func (e PositionNotFoundError) Error() string {
	return fmt.Sprintf("position not found. pool id (%d), lower tick (%d), upper tick (%d), join time (%s) freeze duration (%s)", e.PoolId, e.LowerTick, e.UpperTick, e.JoinTime, e.FreezeDuration)
}

type PoolNotFoundError struct {
	PoolId uint64
}

func (e PoolNotFoundError) Error() string {
	return fmt.Sprintf("pool not found. pool id (%d)", e.PoolId)
}

type InvalidTickError struct {
	Tick    int64
	IsLower bool
	MinTick int64
	MaxTick int64
}

func (e InvalidTickError) Error() string {
	tickStr := "upper"
	if e.IsLower {
		tickStr = "lower"
	}
	return fmt.Sprintf("%s tick (%d) is invalid, Must be >= %d and <= %d", tickStr, e.Tick, e.MinTick, e.MaxTick)
}

type InsufficientLiquidityError struct {
	Actual    sdk.Dec
	Available sdk.Dec
}

func (e InsufficientLiquidityError) Error() string {
	return fmt.Sprintf("insufficient liquidity requested to withdraw. Actual: (%s). Available (%s)", e.Actual, e.Available)
}

type InsufficientLiquidityCreatedError struct {
	Actual      sdk.Int
	Minimum     sdk.Int
	IsTokenZero bool
}

func (e InsufficientLiquidityCreatedError) Error() string {
	tokenNum := uint8(0)
	if !e.IsTokenZero {
		tokenNum = 1
	}
	return fmt.Sprintf("insufficient amount of token %d created. Actual: (%s). Minimum (%s)", tokenNum, e.Actual, e.Minimum)
}

type NegativeLiquidityError struct {
	Liquidity sdk.Dec
}

func (e NegativeLiquidityError) Error() string {
	return fmt.Sprintf("liquidity cannot be negative, got (%d)", e.Liquidity)
}

type DenomDuplicatedError struct {
	TokenInDenom  string
	TokenOutDenom string
}

func (e DenomDuplicatedError) Error() string {
	return fmt.Sprintf("cannot trade same denomination in (%s) and out (%s)", e.TokenInDenom, e.TokenOutDenom)
}

type AmountLessThanMinError struct {
	TokenAmount sdk.Int
	TokenMin    sdk.Int
}

func (e AmountLessThanMinError) Error() string {
	return fmt.Sprintf("token amount calculated (%s) is lesser than min amount (%s)", e.TokenAmount, e.TokenMin)
}

type AmountGreaterThanMaxError struct {
	TokenAmount sdk.Int
	TokenMax    sdk.Int
}

func (e AmountGreaterThanMaxError) Error() string {
	return fmt.Sprintf("token amount calculated (%s) is greater than max amount (%s)", e.TokenAmount, e.TokenMax)
}

type TokenInDenomNotInPoolError struct {
	TokenInDenom string
}

func (e TokenInDenomNotInPoolError) Error() string {
	return fmt.Sprintf("tokenIn (%s) does not match any asset in pool", e.TokenInDenom)
}

type TokenOutDenomNotInPoolError struct {
	TokenOutDenom string
}

func (e TokenOutDenomNotInPoolError) Error() string {
	return fmt.Sprintf("tokenOut (%s) does not match any asset in pool", e.TokenOutDenom)
}

type InvalidPriceLimitError struct {
	SqrtPriceLimit sdk.Dec
	LowerBound     sdk.Dec
	UpperBound     sdk.Dec
}

func (e InvalidPriceLimitError) Error() string {
	return fmt.Sprintf("invalid sqrt price limit given (%s), should be greater than (%s) and less than (%s)", e.SqrtPriceLimit, e.LowerBound, e.UpperBound)
}

type TickSpacingError struct {
	TickSpacing uint64
	LowerTick   int64
	UpperTick   int64
}

func (e TickSpacingError) Error() string {
	return fmt.Sprintf("lowerTick (%d) and upperTick (%d) must be divisible by the pool's tickSpacing parameter (%d)", e.LowerTick, e.UpperTick, e.TickSpacing)
}

type TickSpacingBoundaryError struct {
	TickSpacing        uint64
	TickSpacingMinimum uint64
	TickSpacingMaximum uint64
}

func (e TickSpacingBoundaryError) Error() string {
	return fmt.Sprintf("requested tickSpacing (%d) is not between the minimum (%d) and maximum (%d)", e.TickSpacing, e.TickSpacingMinimum, e.TickSpacingMaximum)
}

type InitialLiquidityZeroError struct {
	Amount0 sdk.Int
	Amount1 sdk.Int
}

func (e InitialLiquidityZeroError) Error() string {
	return fmt.Sprintf("first position must contain non-zero value of both assets to determine spot price: Amount0 (%s) Amount1 (%s)", e.Amount0, e.Amount1)
}

type TickIndexMaximumError struct {
	MaxTick int64
}

func (e TickIndexMaximumError) Error() string {
	return fmt.Sprintf("tickIndex must be less than or equal to %d", e.MaxTick)
}

type TickIndexMinimumError struct {
	MinTick int64
}

func (e TickIndexMinimumError) Error() string {
	return fmt.Sprintf("tickIndex must be greater than or equal to %d", e.MinTick)
}

type ExponentAtPriceOneError struct {
	ProvidedExponentAtPriceOne  sdk.Int
	PrecisionValueAtPriceOneMin sdk.Int
	PrecisionValueAtPriceOneMax sdk.Int
}

func (e ExponentAtPriceOneError) Error() string {
	return fmt.Sprintf("exponentAtPriceOne provided (%s) must be in the range (%s, %s)", e.ProvidedExponentAtPriceOne, e.PrecisionValueAtPriceOneMin, e.PrecisionValueAtPriceOneMax)
}

type PriceBoundError struct {
	ProvidedPrice sdk.Dec
	MinSpotPrice  sdk.Dec
	MaxSpotPrice  sdk.Dec
}

func (e PriceBoundError) Error() string {
	return fmt.Sprintf("provided price (%s) must be between %s and %s", e.ProvidedPrice, e.MinSpotPrice, e.MaxSpotPrice)
}

type SpotPriceNegativeError struct {
	ProvidedPrice sdk.Dec
}

func (e SpotPriceNegativeError) Error() string {
	return fmt.Sprintf("provided price (%s) must be positive", e.ProvidedPrice)
}

type InvalidSwapFeeError struct {
	ActualFee sdk.Dec
}

func (e InvalidSwapFeeError) Error() string {
	return fmt.Sprintf("invalid swap fee(%s), must be in [0, 1) range", e.ActualFee)
}

type IncentiveRecordNotFoundError struct {
	PoolId              uint64
	IncentiveDenom      string
	MinUptime           time.Duration
	IncentiveCreatorStr string
}

func (e IncentiveRecordNotFoundError) Error() string {
	return fmt.Sprintf("incentive record not found. pool id (%d), incentive denom (%s), minimum uptime (%s), incentive creator (%s)", e.PoolId, e.IncentiveDenom, e.MinUptime.String(), e.IncentiveCreatorStr)
}

type StartTimeTooEarlyError struct {
	PoolId           uint64
	CurrentBlockTime time.Time
	StartTime        time.Time
}

func (e StartTimeTooEarlyError) Error() string {
	return fmt.Sprintf("start time cannot be before current blocktime. Pool id (%d), current blocktime (%s), start time (%s)", e.PoolId, e.CurrentBlockTime.String(), e.StartTime.String())
}

type IncentiveInsufficientBalanceError struct {
	PoolId          uint64
	IncentiveDenom  string
	IncentiveAmount sdk.Int
}

func (e IncentiveInsufficientBalanceError) Error() string {
	return fmt.Sprintf("sender has insufficient balance to create this incentive record. Pool id (%d), incentive denom (%s), incentive amount needed (%s)", e.PoolId, e.IncentiveDenom, e.IncentiveAmount)
}

type NonPositiveIncentiveAmountError struct {
	PoolId          uint64
	IncentiveAmount sdk.Dec
}

func (e NonPositiveIncentiveAmountError) Error() string {
	return fmt.Sprintf("incentive amount must be position (nonzero and nonnegative). Pool id (%d), incentive amount (%s)", e.PoolId, e.IncentiveAmount)
}

type NonPositiveEmissionRateError struct {
	PoolId       uint64
	EmissionRate sdk.Dec
}

func (e NonPositiveEmissionRateError) Error() string {
	return fmt.Sprintf("emission rate must be position (nonzero and nonnegative). Pool id (%d), emission rate (%s)", e.PoolId, e.EmissionRate)
}

type InvalidMinUptimeError struct {
	PoolId           uint64
	MinUptime        time.Duration
	SupportedUptimes []time.Duration
}

func (e InvalidMinUptimeError) Error() string {
	return fmt.Sprintf("attempted to create an incentive record with an unsupported minimum uptime. Pool id (%d), specified min uptime (%s), supported uptimes (%s)", e.PoolId, e.MinUptime, e.SupportedUptimes)
}

type QueryRangeUnsupportedError struct {
	RequestedRange sdk.Int
	MaxRange       sdk.Int
}

func (e QueryRangeUnsupportedError) Error() string {
	return fmt.Sprintf("tick range given (%s) is greater than max range supported(%s)", e.RequestedRange, e.MaxRange)
}

type ValueNotFoundForKeyError struct {
	Key []byte
}

func (e ValueNotFoundForKeyError) Error() string {
	return fmt.Sprintf("value not found for key (%x)", e.Key)
}

type InvalidKeyComponentError struct {
	KeyStr                string
	KeySeparator          string
	NumComponentsExpected int
	ComponentsExpectedStr string
}

func (e InvalidKeyComponentError) Error() string {
	return fmt.Sprintf(`invalid key (%s), must have at least (%d) components:
	(%s),
	all separated by (%s)`, e.KeyStr, e.NumComponentsExpected, e.ComponentsExpectedStr, e.KeySeparator)
}

type InvalidPrefixError struct {
	Actual   string
	Expected string
}

func (e InvalidPrefixError) Error() string {
	return fmt.Sprintf("invalid prefix (%s), expected (%s)", e.Actual, e.Expected)
}

type ValueParseError struct {
	Wrapped error
}

func (e ValueParseError) Error() string {
	return e.Wrapped.Error()
}

func (e ValueParseError) Unwrap() error {
	return ErrValueParse
}

type InvalidTickIndexEncodingError struct {
	Length int
}

func (e InvalidTickIndexEncodingError) Error() string {
	return fmt.Sprintf("invalid encoded tick index length; expected: 9, got: %d", e.Length)
}

type InvalidTickKeyByteLengthError struct {
	Length int
}

func (e InvalidTickKeyByteLengthError) Error() string {
	return fmt.Sprintf("expected tick store key to be of length (%d), was (%d)", TickKeyLengthBytes, e.Length)
}

type InvalidNextPositionIdError struct {
	NextPositionId uint64
}

func (e InvalidNextPositionIdError) Error() string {
	return fmt.Sprintf("invalid next position id (%d), must be positive", e.NextPositionId)
}
