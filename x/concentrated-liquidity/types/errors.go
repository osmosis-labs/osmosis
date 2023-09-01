package types

import (
	"errors"
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
)

var (
	ErrKeyNotFound                        = errors.New("key not found")
	ErrValueParse                         = errors.New("value parse error")
	ErrPositionNotFound                   = errors.New("position not found")
	ErrZeroPositionId                     = errors.New("invalid position id, cannot be 0")
	ErrPermissionlessPoolCreationDisabled = errors.New("permissionless pool creation is disabled for the concentrated liquidity module")
	ErrZeroLiquidity                      = errors.New("liquidity cannot be 0")
	ErrNextTickInfoNil                    = errors.New("next tick info cannot be nil")
	ErrPoolNil                            = errors.New("pool cannot be nil")
)

// x/concentrated-liquidity module sentinel errors.
type InvalidLowerUpperTickError struct {
	LowerTick int64
	UpperTick int64
}

func (e InvalidLowerUpperTickError) Error() string {
	return fmt.Sprintf("Lower tick must be lesser than upper. Got lower: %d, upper: %d", e.LowerTick, e.UpperTick)
}

type InvalidDirectionError struct {
	PoolTick   int64
	TargetTick int64
	ZeroForOne bool
}

func (e InvalidDirectionError) Error() string {
	return fmt.Sprintf("Given zero for one (%t) does not match swap direction. Pool tick at %d, target tick at %d", e.ZeroForOne, e.PoolTick, e.TargetTick)
}

type NotPositiveRequireAmountError struct {
	Amount string
}

func (e NotPositiveRequireAmountError) Error() string {
	return fmt.Sprintf("Required amount should be positive. Got: %s", e.Amount)
}

type QualifyingLiquidityOrTimeElapsedNotPositiveError struct {
	QualifyingLiquidity osmomath.Dec
	TimeElapsed         osmomath.Dec
}

func (e QualifyingLiquidityOrTimeElapsedNotPositiveError) Error() string {
	return fmt.Sprintf("Qualifying liquidity and time elapsed must both be positive. Got: QualifyingLiquidity (%s), timeElapsed (%s)", e.QualifyingLiquidity, e.TimeElapsed)
}

type TimeElapsedNotPositiveError struct {
	TimeElapsed osmomath.Dec
}

func (e TimeElapsedNotPositiveError) Error() string {
	return fmt.Sprintf("Time elapsed must both be positive. Got: timeElapsed (%s)", e.TimeElapsed)
}

type PositionNotFoundError struct {
	PoolId    uint64
	LowerTick int64
	UpperTick int64
	JoinTime  time.Time
}

func (e PositionNotFoundError) Error() string {
	return fmt.Sprintf("position not found. pool id (%d), lower tick (%d), upper tick (%d), join time (%s)", e.PoolId, e.LowerTick, e.UpperTick, e.JoinTime)
}

type PositionIdNotFoundError struct {
	PositionId uint64
}

func (e PositionIdNotFoundError) Error() string {
	return fmt.Sprintf("position not found. position id (%d)", e.PositionId)
}

type SpreadRewardPositionNotFoundError struct {
	PositionId uint64
}

func (e SpreadRewardPositionNotFoundError) Error() string {
	return fmt.Sprintf("position not found in spread reward accumulator. position id (%d)", e.PositionId)
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
	Actual    osmomath.Dec
	Available osmomath.Dec
}

func (e InsufficientLiquidityError) Error() string {
	return fmt.Sprintf("insufficient liquidity requested to withdraw. Actual: (%s). Available (%s)", e.Actual, e.Available)
}

type InsufficientLiquidityCreatedError struct {
	Actual      osmomath.Int
	Minimum     osmomath.Int
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
	Liquidity osmomath.Dec
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
	TokenAmount osmomath.Int
	TokenMin    osmomath.Int
}

func (e AmountLessThanMinError) Error() string {
	return fmt.Sprintf("token amount calculated (%s) is lesser than min amount (%s)", e.TokenAmount, e.TokenMin)
}

type AmountGreaterThanMaxError struct {
	TokenAmount osmomath.Int
	TokenMax    osmomath.Int
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

type SqrtPriceValidationError struct {
	SqrtPriceLimit osmomath.BigDec
	LowerBound     osmomath.BigDec
	UpperBound     osmomath.BigDec
}

func (e SqrtPriceValidationError) Error() string {
	return fmt.Sprintf("invalid sqrt price given (%s), should be greater than (%s) and less than (%s)", e.SqrtPriceLimit, e.LowerBound, e.UpperBound)
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
	Amount0 osmomath.Int
	Amount1 osmomath.Int
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

type TickIndexNotWithinBoundariesError struct {
	MaxTick    int64
	MinTick    int64
	ActualTick int64
}

func (e TickIndexNotWithinBoundariesError) Error() string {
	return fmt.Sprintf("tickIndex must be within the range (%d, %d). Got (%d)", e.MinTick, e.MaxTick, e.ActualTick)
}

type TickNotFoundError struct {
	Tick int64
}

func (e TickNotFoundError) Error() string {
	return fmt.Sprintf("tick %d is not found", e.Tick)
}

type PriceBoundError struct {
	ProvidedPrice osmomath.BigDec
	MinSpotPrice  osmomath.BigDec
	MaxSpotPrice  osmomath.Dec
}

func (e PriceBoundError) Error() string {
	return fmt.Sprintf("provided price (%s) must be between %s and %s", e.ProvidedPrice, e.MinSpotPrice, e.MaxSpotPrice)
}

type SpotPriceNegativeError struct {
	ProvidedPrice osmomath.Dec
}

func (e SpotPriceNegativeError) Error() string {
	return fmt.Sprintf("provided price (%s) must be positive", e.ProvidedPrice)
}

type SqrtPriceNegativeError struct {
	ProvidedSqrtPrice osmomath.BigDec
}

func (e SqrtPriceNegativeError) Error() string {
	return fmt.Sprintf("provided sqrt price (%s) must be positive", e.ProvidedSqrtPrice)
}

type InvalidSpreadFactorError struct {
	ActualSpreadFactor osmomath.Dec
}

func (e InvalidSpreadFactorError) Error() string {
	return fmt.Sprintf("invalid spread factor(%s), must be in [0, 1) range", e.ActualSpreadFactor)
}

type PositionAlreadyExistsError struct {
	PoolId    uint64
	LowerTick int64
	UpperTick int64
	JoinTime  time.Time
}

func (e PositionAlreadyExistsError) Error() string {
	return fmt.Sprintf("position already exists with same poolId %d, lowerTick %d, upperTick %d, JoinTime %s", e.PoolId, e.LowerTick, e.UpperTick, e.JoinTime)
}

type IncentiveRecordNotFoundError struct {
	PoolId            uint64
	MinUptime         time.Duration
	IncentiveRecordId uint64
}

func (e IncentiveRecordNotFoundError) Error() string {
	return fmt.Sprintf("incentive record not found. pool id (%d), minimum uptime (%s), incentive record id (%d)", e.PoolId, e.MinUptime.String(), e.IncentiveRecordId)
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
	IncentiveAmount osmomath.Int
}

func (e IncentiveInsufficientBalanceError) Error() string {
	return fmt.Sprintf("sender has insufficient balance to create this incentive record. Pool id (%d), incentive denom (%s), incentive amount needed (%s)", e.PoolId, e.IncentiveDenom, e.IncentiveAmount)
}

type ErrInvalidBalancerPoolLiquidityError struct {
	ClPoolId              uint64
	BalancerPoolId        uint64
	BalancerPoolLiquidity sdk.Coins
}

func (e ErrInvalidBalancerPoolLiquidityError) Error() string {
	return fmt.Sprintf("canonical balancer pool for CL pool is invalid. CL pool id (%d), Balancer pool ID (%d), Balancer pool assets (%s)", e.ClPoolId, e.BalancerPoolId, e.BalancerPoolLiquidity)
}

type BalancerRecordNotFoundError struct {
	ClPoolId       uint64
	BalancerPoolId uint64
	UptimeIndex    uint64
}

func (e BalancerRecordNotFoundError) Error() string {
	return fmt.Sprintf("record not found on CL accumulators for given balancer pool. CL pool id (%d), Balancer pool ID (%d), Uptime index (%d)", e.ClPoolId, e.BalancerPoolId, e.UptimeIndex)
}

type BalancerRecordNotClearedError struct {
	ClPoolId       uint64
	BalancerPoolId uint64
	UptimeIndex    uint64
}

func (e BalancerRecordNotClearedError) Error() string {
	return fmt.Sprintf("balancer record was not cleared after reward claiming. CL pool id (%d), Balancer pool ID (%d), Uptime index (%d)", e.ClPoolId, e.BalancerPoolId, e.UptimeIndex)
}

type InvalidIncentiveCoinError struct {
	PoolId        uint64
	IncentiveCoin sdk.Coin
}

func (e InvalidIncentiveCoinError) Error() string {
	return fmt.Sprintf("incentive coin denom must be valid and have non negative amount Pool id (%d), incentive coin (%s)", e.PoolId, e.IncentiveCoin)
}

type NonPositiveEmissionRateError struct {
	PoolId       uint64
	EmissionRate osmomath.Dec
}

func (e NonPositiveEmissionRateError) Error() string {
	return fmt.Sprintf("emission rate must be position (nonzero and nonnegative). Pool id (%d), emission rate (%s)", e.PoolId, e.EmissionRate)
}

type InvalidMinUptimeError struct {
	PoolId            uint64
	MinUptime         time.Duration
	AuthorizedUptimes []time.Duration
}

func (e InvalidMinUptimeError) Error() string {
	return fmt.Sprintf("attempted to create an incentive record with an unsupported minimum uptime. Pool id (%d), specified min uptime (%s), authorized uptimes (%s)", e.PoolId, e.MinUptime, e.AuthorizedUptimes)
}

type InvalidUptimeIndexError struct {
	MinUptime        time.Duration
	SupportedUptimes []time.Duration
}

func (e InvalidUptimeIndexError) Error() string {
	return fmt.Sprintf("attempted to find index for an unsupported min uptime. Specified min uptime (%s), supported uptimes (%s)", e.MinUptime, e.SupportedUptimes)
}

type QueryRangeUnsupportedError struct {
	RequestedRange osmomath.Int
	MaxRange       osmomath.Int
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
	return fmt.Sprintf("expected tick store key to be of length (%d), was (%d)", KeyTickLengthBytes, e.Length)
}

type InsufficientPoolBalanceError struct {
	Err error
}

func (e InsufficientPoolBalanceError) Error() string {
	return fmt.Sprintf("insufficient pool balance: %s", e.Err.Error())
}

func (e *InsufficientPoolBalanceError) Unwrap() error { return e.Err }

type InsufficientUserBalanceError struct {
	Err error
}

func (e InsufficientUserBalanceError) Error() string {
	return fmt.Sprintf("insufficient user balance: %s", e.Err.Error())
}

func (e *InsufficientUserBalanceError) Unwrap() error { return e.Err }

type InvalidAmountCalculatedError struct {
	Amount osmomath.Int
}

func (e InvalidAmountCalculatedError) Error() string {
	return fmt.Sprintf("invalid amount calculated, must be >= 1, was (%s)", e.Amount)
}

type InvalidNextPositionIdError struct {
	NextPositionId uint64
}

func (e InvalidNextPositionIdError) Error() string {
	return fmt.Sprintf("invalid next incentive record id (%d), must be positive", e.NextPositionId)
}

type InvalidNextIncentiveRecordIdError struct {
	NextIncentiveRecordId uint64
}

func (e InvalidNextIncentiveRecordIdError) Error() string {
	return fmt.Sprintf("invalid next incentive record id (%d), must be positive", e.NextIncentiveRecordId)
}

type AddressPoolPositionIdNotFoundError struct {
	PositionId uint64
	Owner      string
	PoolId     uint64
}

func (e AddressPoolPositionIdNotFoundError) Error() string {
	return fmt.Sprintf("position id %d not found for address %s and pool id %d", e.PositionId, e.Owner, e.PoolId)
}

type PoolPositionIdNotFoundError struct {
	PositionId uint64
	PoolId     uint64
}

func (e PoolPositionIdNotFoundError) Error() string {
	return fmt.Sprintf("position id %d not found for pool id %d", e.PositionId, e.PoolId)
}

type NegativeDurationError struct {
	Duration time.Duration
}

func (e NegativeDurationError) Error() string {
	return fmt.Sprintf("duration cannot be negative (%s)", e.Duration)
}

type UninitializedPoolWithLiquidityError struct {
	PoolId uint64
}

func (e UninitializedPoolWithLiquidityError) Error() string {
	return fmt.Sprintf("attempted to uninitialize pool (%d) with liquidity still existing", e.PoolId)
}

type NoSpotPriceWhenNoLiquidityError struct {
	PoolId uint64
}

func (e NoSpotPriceWhenNoLiquidityError) Error() string {
	return fmt.Sprintf("error getting spot price for pool (%d), no liquidity in pool", e.PoolId)
}

type PositionQuantityTooLowError struct {
	MinNumPositions int
	NumPositions    int
}

func (e PositionQuantityTooLowError) Error() string {
	return fmt.Sprintf("position quantity must be greater than or equal to (%d), was (%d)", e.MinNumPositions, e.NumPositions)
}

type PositionOwnerMismatchError struct {
	PositionOwner string
	Sender        string
}

func (e PositionOwnerMismatchError) Error() string {
	return fmt.Sprintf("position owner mismatch, expected (%s), got (%s)", e.PositionOwner, e.Sender)
}

type PositionNotFullyChargedError struct {
	PositionId               uint64
	PositionJoinTime         time.Time
	FullyChargedMinTimestamp time.Time
}

func (e PositionNotFullyChargedError) Error() string {
	return fmt.Sprintf("position ID (%d) not fully charged, join time (%s), fully charged min timestamp (%s)", e.PositionId, e.PositionJoinTime, e.FullyChargedMinTimestamp)
}

type PositionsNotInSamePoolError struct {
	Position1PoolId uint64
	Position2PoolId uint64
}

func (e PositionsNotInSamePoolError) Error() string {
	return fmt.Sprintf("positions not in same pool, position 1 pool id (%d), position 2 pool id (%d)", e.Position1PoolId, e.Position2PoolId)
}

type PositionsNotInSameTickRangeError struct {
	Position1TickLower int64
	Position1TickUpper int64
	Position2TickLower int64
	Position2TickUpper int64
}

func (e PositionsNotInSameTickRangeError) Error() string {
	return fmt.Sprintf("positions not in same tick range, position 1 tick lower (%d), position 1 tick upper (%d), position 2 tick lower (%d), position 2 tick upper (%d)", e.Position1TickLower, e.Position1TickUpper, e.Position2TickLower, e.Position2TickUpper)
}

type InvalidDiscountRateError struct {
	DiscountRate osmomath.Dec
}

func (e InvalidDiscountRateError) Error() string {
	return fmt.Sprintf("Discount rate for Balancer shares must be in range [0, 1]. Attempted to set as %s", e.DiscountRate)
}

type UptimeNotSupportedError struct {
	Uptime time.Duration
}

func (e UptimeNotSupportedError) Error() string {
	return fmt.Sprintf("Uptime %s is not in list of supported uptimes. Full list of supported uptimes: %s", e.Uptime, SupportedUptimes)
}

type PositionIdToLockNotFoundError struct {
	PositionId uint64
}

func (e PositionIdToLockNotFoundError) Error() string {
	return fmt.Sprintf("position id (%d) does not have an underlying lock in state", e.PositionId)
}

type LockIdToPositionIdNotFoundError struct {
	LockId uint64
}

func (e LockIdToPositionIdNotFoundError) Error() string {
	return fmt.Sprintf("lock id (%d) does not have an underlying position in state", e.LockId)
}

type LockNotMatureError struct {
	PositionId uint64
	LockId     uint64
}

func (e LockNotMatureError) Error() string {
	return fmt.Sprintf("position ID %d's lock (%d) is not mature, must wait till unlocking is complete to withdraw the position", e.PositionId, e.LockId)
}

type PositionSuperfluidStakedError struct {
	PositionId uint64
}

func (e PositionSuperfluidStakedError) Error() string {
	return fmt.Sprintf("Cannot add to position ID %d as it is superfluid staked.", e.PositionId)
}

type AddToLastPositionInPoolError struct {
	PoolId     uint64
	PositionId uint64
}

func (e AddToLastPositionInPoolError) Error() string {
	return fmt.Sprintf("Cannot add to a position if it is the last position in the pool. Pool id (%d), position ID (%d).", e.PoolId, e.PositionId)
}

type NegativeAmountAddedError struct {
	PositionId   uint64
	Asset0Amount osmomath.Int
	Asset1Amount osmomath.Int
}

func (e NegativeAmountAddedError) Error() string {
	return fmt.Sprintf("Cannot add negative amounts of assets to a position. Position ID (%d), asset0 amount (%s), asset1 amount(%s).", e.PositionId, e.Asset0Amount, e.Asset1Amount)
}

type MatchingDenomError struct {
	Denom string
}

func (e MatchingDenomError) Error() string {
	return fmt.Sprintf("received matching denoms (%s), must be different", e.Denom)
}

type UnauthorizedQuoteDenomError struct {
	ProvidedQuoteDenom    string
	AuthorizedQuoteDenoms []string
}

func (e UnauthorizedQuoteDenomError) Error() string {
	return fmt.Sprintf("attempted to create pool with unauthorized quote denom (%s), must be one of the following: (%s)", e.ProvidedQuoteDenom, e.AuthorizedQuoteDenoms)
}

type UnauthorizedSpreadFactorError struct {
	ProvidedSpreadFactor    osmomath.Dec
	AuthorizedSpreadFactors []osmomath.Dec
}

func (e UnauthorizedSpreadFactorError) Error() string {
	return fmt.Sprintf("attempted to create pool with unauthorized spread factor (%s), must be one of the following: (%s)", e.ProvidedSpreadFactor, e.AuthorizedSpreadFactors)
}

type UnauthorizedTickSpacingError struct {
	ProvidedTickSpacing    uint64
	AuthorizedTickSpacings []uint64
}

func (e UnauthorizedTickSpacingError) Error() string {
	return fmt.Sprintf("attempted to create pool with unauthorized tick spacing (%d), must be one of the following: (%d)", e.ProvidedTickSpacing, e.AuthorizedTickSpacings)
}

type NonPositiveLiquidityForNewPositionError struct {
	LiquidityDelta osmomath.Dec
	PositionId     uint64
}

func (e NonPositiveLiquidityForNewPositionError) Error() string {
	return fmt.Sprintf("liquidityDelta (%s) must be positive for a new position with id (%d)", e.LiquidityDelta, e.PositionId)
}

type LiquidityWithdrawalError struct {
	PositionID       uint64
	RequestedAmount  osmomath.Dec
	CurrentLiquidity osmomath.Dec
}

func (e LiquidityWithdrawalError) Error() string {
	return fmt.Sprintf("position %d attempted to withdraw %s liquidity, but only has %s available", e.PositionID, e.RequestedAmount, e.CurrentLiquidity)
}

type LowerTickMismatchError struct {
	PositionId uint64
	Expected   int64
	Got        int64
}

func (e LowerTickMismatchError) Error() string {
	return fmt.Sprintf("position lower tick mismatch, expected (%d), got (%d), position id (%d)", e.Expected, e.Got, e.PositionId)
}

type UpperTickMismatchError struct {
	PositionId uint64
	Expected   int64
	Got        int64
}

func (e UpperTickMismatchError) Error() string {
	return fmt.Sprintf("position upper tick mismatch, expected (%d), got (%d), position id (%d)", e.Expected, e.Got, e.PositionId)
}

type JoinTimeMismatchError struct {
	PositionId uint64
	Expected   time.Time
	Got        time.Time
}

func (e JoinTimeMismatchError) Error() string {
	return fmt.Sprintf("join time does not match provided join time, expected (%s), got (%s), , position id (%d)", e.Expected.String(), e.Got.String(), e.PositionId)
}

type NotPositionOwnerError struct {
	PositionId uint64
	Address    string
}

func (e NotPositionOwnerError) Error() string {
	return fmt.Sprintf("address (%s) is not the owner of position ID (%d)", e.Address, e.PositionId)
}

type PositionNotFullRangeError struct {
	PositionId uint64
	LowerTick  int64
	UpperTick  int64
}

func (e PositionNotFullRangeError) Error() string {
	return fmt.Sprintf("position ID (%d) is not a full range position, lower tick (%d), upper tick (%d)", e.PositionId, e.LowerTick, e.UpperTick)
}

type Amount0IsNegativeError struct {
	Amount0 osmomath.Int
}

func (e Amount0IsNegativeError) Error() string {
	return fmt.Sprintf("amount0 (%s) is negative", e.Amount0)
}

type Amount1IsNegativeError struct {
	Amount1 osmomath.Int
}

func (e Amount1IsNegativeError) Error() string {
	return fmt.Sprintf("amount1 (%s) is negative", e.Amount1)
}

type ModifySamePositionAccumulatorError struct {
	PositionAccName string
}

func (e ModifySamePositionAccumulatorError) Error() string {
	return fmt.Sprintf("attempted to modify the same accumulator with name %s", e.PositionAccName)
}

type NumCoinsError struct {
	NumCoins int
}

func (e NumCoinsError) Error() string {
	return fmt.Sprintf("num coins provided (%d) must be 2 for a full range position", e.NumCoins)
}

type CoinLengthError struct {
	MaxLength int
	Length    int
}

func (e CoinLengthError) Error() string {
	return fmt.Sprintf("coin length (%d) must be less than or equal to max length (%d)", e.Length, e.MaxLength)
}

type RanOutOfTicksForPoolError struct {
	PoolId uint64
}

func (e RanOutOfTicksForPoolError) Error() string {
	return fmt.Sprintf("ran out of ticks for pool (%d) during swap", e.PoolId)
}

type SqrtRootCalculationError struct {
	SqrtPriceLimit osmomath.BigDec
}

func (e SqrtRootCalculationError) Error() string {
	return fmt.Sprintf("issue calculating square root of price limit %s", e.SqrtPriceLimit)
}

type TickToSqrtPriceConversionError struct {
	NextTick int64
}

func (e TickToSqrtPriceConversionError) Error() string {
	return fmt.Sprintf("could not convert next tick  to nextSqrtPrice (%v)", e.NextTick)
}

type SwapNoProgressError struct {
	PoolId           uint64
	UserProvidedCoin sdk.Coin
}

func (e SwapNoProgressError) Error() string {
	return fmt.Sprintf("ran out of iterations during swap. Possibly entered an infinite loop. Pool id (%d), user provided coin (%s)", e.PoolId, e.UserProvidedCoin)
}

type SwapNoProgressWithConsumptionError struct {
	ComputedSqrtPrice osmomath.BigDec
	AmountIn          osmomath.Dec
	AmountOut         osmomath.Dec
}

func (e SwapNoProgressWithConsumptionError) Error() string {
	return fmt.Sprintf("did not advance sqrt price after swap step %s, with amounts in (%s), out (%s)", e.ComputedSqrtPrice, e.AmountIn, e.AmountOut)
}

type SqrtPriceToTickError struct {
	OutOfBounds bool
}

func (e SqrtPriceToTickError) Error() string {
	return fmt.Sprintf("sqrt price to tick could not find a satisfying tick index. Hit bounds: %v", e.OutOfBounds)
}

type OverChargeSwapOutGivenInError struct {
	AmountSpecifiedRemaining osmomath.Dec
}

func (e OverChargeSwapOutGivenInError) Error() string {
	return fmt.Sprintf("over charge problem swap out given in by (%s)", e.AmountSpecifiedRemaining)
}

type ComputedSqrtPriceInequalityError struct {
	IsZeroForOne                 bool
	NextInitializedTickSqrtPrice osmomath.BigDec
	ComputedSqrtPrice            osmomath.BigDec
}

func (e ComputedSqrtPriceInequalityError) Error() string {
	return fmt.Sprintf("edge case has occurred when swapping at tick boundaries, with izZeroForOne (%t), NextInitializedTickSqrtPrice (%s), computedSqrtPrice (%s). Please try again with a different swap amount", e.IsZeroForOne, e.NextInitializedTickSqrtPrice, e.ComputedSqrtPrice)
}
