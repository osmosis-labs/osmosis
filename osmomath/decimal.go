package osmomath

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NOTE: never use new(BigDec) or else we will panic unmarshalling into the
// nil embedded big.Int
type BigDec struct {
	i *big.Int
}

const (
	// number of decimal places
	BigDecPrecision = 36

	// bytes required to represent the above precision
	// Ceiling[Log2[10**Precision - 1]]
	BigDecimalPrecisionBits = 120

	maxDecBitLen = maxBitLen + BigDecimalPrecisionBits

	// max number of iterations in ApproxRoot function
	maxApproxRootIterations = 100

	// max number of iterations in Log2 function
	maxLog2Iterations = 300
)

var (
	precisionReuse        = new(big.Int).Exp(big.NewInt(10), big.NewInt(BigDecPrecision), nil)
	squaredPrecisionReuse = new(big.Int).Mul(precisionReuse, precisionReuse)
	precisionReuseSDK     = new(big.Int).Exp(big.NewInt(10), big.NewInt(DecPrecision), nil)
	fivePrecision         = new(big.Int).Quo(precisionReuse, big.NewInt(2))
	precisionMultipliers  []*big.Int
	zeroInt               = big.NewInt(0)
	oneInt                = big.NewInt(1)
	tenInt                = big.NewInt(10)

	// log_2(e)
	// From: https://www.wolframalpha.com/input?i=log_2%28e%29+with+37+digits
	logOfEbase2 = MustNewBigDecFromStr("1.442695040888963407359924681001892137")

	// log_2(1.0001)
	// From: https://www.wolframalpha.com/input?i=log_2%281.0001%29+to+33+digits
	tickLogOf2 = MustNewBigDecFromStr("0.000144262291094554178391070900057480")
	// initialized in init() since requires
	// precision to be defined.
	twoBigDec BigDec = MustNewBigDecFromStr("2")

	// precisionFactors are used to adjust the scale of big.Int values to match the desired precision
	precisionFactors = make(map[uint64]*big.Int)

	zeroBigDec    BigDec = ZeroBigDec()
	oneBigDec     BigDec = OneBigDec()
	oneHalfBigDec BigDec = oneBigDec.Quo(twoBigDec)
	negOneBigDec  BigDec = oneBigDec.Neg()
)

// Decimal errors
var (
	ErrEmptyDecimalStr      = errors.New("decimal string cannot be empty")
	ErrInvalidDecimalLength = errors.New("invalid decimal length")
	ErrInvalidDecimalStr    = errors.New("invalid decimal string")
)

// Set precision multipliers
func init() {
	precisionMultipliers = make([]*big.Int, BigDecPrecision+1)
	for i := 0; i <= BigDecPrecision; i++ {
		precisionMultipliers[i] = calcPrecisionMultiplier(int64(i))
	}

	for precision := uint64(0); precision <= BigDecPrecision; precision++ {
		precisionFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(BigDecPrecision-int64(precision)), nil)
		precisionFactors[precision] = precisionFactor
	}
}

func precisionInt() *big.Int {
	return new(big.Int).Set(precisionReuse)
}

func ZeroBigDec() BigDec     { return BigDec{new(big.Int).Set(zeroInt)} }
func OneBigDec() BigDec      { return BigDec{precisionInt()} }
func SmallestBigDec() BigDec { return BigDec{new(big.Int).Set(oneInt)} }

// calculate the precision multiplier
func calcPrecisionMultiplier(prec int64) *big.Int {
	if prec > BigDecPrecision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", BigDecPrecision, prec))
	}
	zerosToAdd := BigDecPrecision - prec
	multiplier := new(big.Int).Exp(tenInt, big.NewInt(zerosToAdd), nil)
	return multiplier
}

// get the precision multiplier, do not mutate result
func precisionMultiplier(prec int64) *big.Int {
	if prec > BigDecPrecision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", BigDecPrecision, prec))
	}
	return precisionMultipliers[prec]
}

// create a new NewBigDec from integer assuming whole number
func NewBigDec(i int64) BigDec {
	return NewBigDecWithPrec(i, 0)
}

// create a new BigDec from integer with decimal place at prec
// CONTRACT: prec <= BigDecPrecision
func NewBigDecWithPrec(i, prec int64) BigDec {
	return BigDec{
		new(big.Int).Mul(big.NewInt(i), precisionMultiplier(prec)),
	}
}

// create a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= BigDecPrecision
func NewBigDecFromBigInt(i *big.Int) BigDec {
	return NewBigDecFromBigIntWithPrec(i, 0)
}

// create a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= BigDecPrecision
func NewBigDecFromBigIntWithPrec(i *big.Int, prec int64) BigDec {
	return BigDec{
		new(big.Int).Mul(i, precisionMultiplier(prec)),
	}
}

// create a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= BigDecPrecision
func NewBigDecFromInt(i BigInt) BigDec {
	return NewBigDecFromIntWithPrec(i, 0)
}

// create a new BigDec from big integer with decimal place at prec
// CONTRACT: prec <= BigDecPrecision
func NewBigDecFromIntWithPrec(i BigInt, prec int64) BigDec {
	return BigDec{
		new(big.Int).Mul(i.BigInt(), precisionMultiplier(prec)),
	}
}

// create a decimal from an input decimal string.
// valid must come in the form:
//
//	(-) whole integers (.) decimal integers
//
// examples of acceptable input include:
//
//	-123.456
//	456.7890
//	345
//	-456789
//
// NOTE - An error will return if more decimal places
// are provided in the string than the constant Precision.
//
// CONTRACT - This function does not mutate the input str.
func NewBigDecFromStr(str string) (BigDec, error) {
	if len(str) == 0 {
		return BigDec{}, ErrEmptyDecimalStr
	}

	// first extract any negative symbol
	neg := false
	if str[0] == '-' {
		neg = true
		str = str[1:]
	}

	if len(str) == 0 {
		return BigDec{}, ErrEmptyDecimalStr
	}

	strs := strings.Split(str, ".")
	lenDecs := 0
	combinedStr := strs[0]

	if len(strs) == 2 { // has a decimal place
		lenDecs = len(strs[1])
		if lenDecs == 0 || len(combinedStr) == 0 {
			return BigDec{}, ErrInvalidDecimalLength
		}
		combinedStr += strs[1]
	} else if len(strs) > 2 {
		return BigDec{}, ErrInvalidDecimalStr
	}

	if lenDecs > BigDecPrecision {
		return BigDec{}, fmt.Errorf("invalid precision; max: %d, got: %d", BigDecPrecision, lenDecs)
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := BigDecPrecision - lenDecs
	zeros := fmt.Sprintf(`%0`+strconv.Itoa(zerosToAdd)+`s`, "")
	combinedStr += zeros

	combined, ok := new(big.Int).SetString(combinedStr, 10) // base 10
	if !ok {
		return BigDec{}, fmt.Errorf("failed to set decimal string: %s", combinedStr)
	}
	if combined.BitLen() > maxBitLen {
		return BigDec{}, fmt.Errorf("decimal out of range; bitLen: got %d, max %d", combined.BitLen(), maxBitLen)
	}
	if neg {
		combined = new(big.Int).Neg(combined)
	}

	return BigDec{combined}, nil
}

// Decimal from string, panic on error
func MustNewBigDecFromStr(s string) BigDec {
	dec, err := NewBigDecFromStr(s)
	if err != nil {
		panic(err)
	}
	return dec
}

func (d BigDec) IsNil() bool          { return d.i == nil }                    // is decimal nil
func (d BigDec) IsZero() bool         { return (d.i).Sign() == 0 }             // is equal to zero
func (d BigDec) IsNegative() bool     { return (d.i).Sign() == -1 }            // is negative
func (d BigDec) IsPositive() bool     { return (d.i).Sign() == 1 }             // is positive
func (d BigDec) Equal(d2 BigDec) bool { return (d.i).Cmp(d2.i) == 0 }          // equal decimals
func (d BigDec) GT(d2 BigDec) bool    { return (d.i).Cmp(d2.i) > 0 }           // greater than
func (d BigDec) GTE(d2 BigDec) bool   { return (d.i).Cmp(d2.i) >= 0 }          // greater than or equal
func (d BigDec) LT(d2 BigDec) bool    { return (d.i).Cmp(d2.i) < 0 }           // less than
func (d BigDec) LTE(d2 BigDec) bool   { return (d.i).Cmp(d2.i) <= 0 }          // less than or equal
func (d BigDec) Neg() BigDec          { return BigDec{new(big.Int).Neg(d.i)} } // reverse the decimal sign
// nolint: stylecheck
func (d BigDec) Abs() BigDec { return BigDec{new(big.Int).Abs(d.i)} } // absolute value

// BigInt returns a copy of the underlying big.Int.
func (d BigDec) BigInt() *big.Int {
	if d.IsNil() {
		return nil
	}

	cp := new(big.Int)
	return cp.Set(d.i)
}

// addition
func (d BigDec) Add(d2 BigDec) BigDec {
	copy := d.Clone()
	copy.AddMut(d2)
	return copy
}

// mutative addition
func (d BigDec) AddMut(d2 BigDec) BigDec {
	d.i.Add(d.i, d2.i)

	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}

	return d
}

// subtraction
func (d BigDec) Sub(d2 BigDec) BigDec {
	res := new(big.Int).Sub(d.i, d2.i)

	if res.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{res}
}

// Clone performs a deep copy of the receiver
// and returns the new result.
func (d BigDec) Clone() BigDec {
	copy := BigDec{new(big.Int)}
	copy.i.Set(d.i)
	return copy
}

// Mut performs non-mutative multiplication.
// The receiver is not modifier but the result is.
func (d BigDec) Mul(d2 BigDec) BigDec {
	copy := d.Clone()
	copy.MulMut(d2)
	return copy
}

// Mut performs non-mutative multiplication.
// The receiver is not modifier but the result is.
func (d BigDec) MulMut(d2 BigDec) BigDec {
	d.i.Mul(d.i, d2.i)
	d.i = chopPrecisionAndRound(d.i)

	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{d.i}
}

// multiplication truncate
func (d BigDec) MulTruncate(d2 BigDec) BigDec {
	mul := new(big.Int).Mul(d.i, d2.i)
	chopped := chopPrecisionAndTruncate(mul)

	if chopped.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{chopped}
}

// multiplication round up
func (d BigDec) MulRoundUp(d2 BigDec) BigDec {
	mul := new(big.Int).Mul(d.i, d2.i)
	chopped := chopPrecisionAndRoundUpBigDec(mul)

	if chopped.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{chopped}
}

// multiplication
func (d BigDec) MulInt(i BigInt) BigDec {
	mul := new(big.Int).Mul(d.i, i.i)

	if mul.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{mul}
}

// MulInt64 - multiplication with int64
func (d BigDec) MulInt64(i int64) BigDec {
	mul := new(big.Int).Mul(d.i, big.NewInt(i))

	if mul.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{mul}
}

// quotient
func (d BigDec) Quo(d2 BigDec) BigDec {
	copy := d.Clone()
	copy.QuoMut(d2)
	return copy
}

// mutative quotient
func (d BigDec) QuoMut(d2 BigDec) BigDec {
	// multiply precision twice
	d.i.Mul(d.i, precisionReuse)
	d.i.Mul(d.i, precisionReuse)

	d.i.Quo(d.i, d2.i)
	chopPrecisionAndRound(d.i)

	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}
func (d BigDec) QuoRaw(d2 int64) BigDec {
	// multiply precision, so we can chop it later
	mul := new(big.Int).Mul(d.i, precisionReuse)

	quo := mul.Quo(mul, big.NewInt(d2))
	chopped := chopPrecisionAndRound(quo)

	if chopped.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{chopped}
}

// quotient truncate
func (d BigDec) QuoTruncate(d2 BigDec) BigDec {
	// multiply precision twice
	mul := new(big.Int).Mul(d.i, precisionReuse)
	mul.Mul(mul, precisionReuse)

	quo := mul.Quo(mul, d2.i)
	chopped := chopPrecisionAndTruncate(quo)

	if chopped.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{chopped}
}

// quotient truncate (mutative)
func (d BigDec) QuoTruncateMut(d2 BigDec) BigDec {
	// multiply precision twice
	d.i.Mul(d.i, squaredPrecisionReuse)
	d.i.Quo(d.i, d2.i)

	chopPrecisionAndTruncateMut(d.i)
	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return d
}

// quotient, round up
func (d BigDec) QuoRoundUp(d2 BigDec) BigDec {
	// multiply precision twice
	mul := new(big.Int).Mul(d.i, precisionReuse)
	mul.Mul(mul, precisionReuse)

	quo := new(big.Int).Quo(mul, d2.i)
	chopped := chopPrecisionAndRoundUpBigDec(quo)

	if chopped.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{chopped}
}

// quotient, round up (mutative)
func (d BigDec) QuoRoundUpMut(d2 BigDec) BigDec {
	// multiply precision twice
	d.i.Mul(d.i, squaredPrecisionReuse)
	d.i.Quo(d.i, d2.i)

	chopPrecisionAndRoundUpMut(d.i, precisionReuse)

	if d.i.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{d.i}
}

// quotient
func (d BigDec) QuoInt(i BigInt) BigDec {
	mul := new(big.Int).Quo(d.i, i.i)
	return BigDec{mul}
}

// QuoInt64 - quotient with int64
func (d BigDec) QuoInt64(i int64) BigDec {
	mul := new(big.Int).Quo(d.i, big.NewInt(i))
	return BigDec{mul}
}

// ApproxRoot returns an approximate estimation of a Dec's positive real nth root
// using Newton's method (where n is positive). The algorithm starts with some guess and
// computes the sequence of improved guesses until an answer converges to an
// approximate answer.  It returns `|d|.ApproxRoot() * -1` if input is negative.
// A maximum number of 100 iterations is used a backup boundary condition for
// cases where the answer never converges enough to satisfy the main condition.
func (d BigDec) ApproxRoot(root uint64) (guess BigDec, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = errors.New("out of bounds")
			}
		}
	}()

	if d.IsNegative() {
		absRoot, err := d.MulInt64(-1).ApproxRoot(root)
		return absRoot.MulInt64(-1), err
	}

	if root == 1 || d.IsZero() || d.Equal(OneBigDec()) {
		return d, nil
	}

	if root == 0 {
		return OneBigDec(), nil
	}

	rootInt := NewBigIntFromUint64(root)
	guess, delta := OneBigDec(), OneBigDec()

	for iter := 0; delta.Abs().GT(SmallestBigDec()) && iter < maxApproxRootIterations; iter++ {
		prev := guess.PowerInteger(root - 1)
		if prev.IsZero() {
			prev = SmallestBigDec()
		}
		delta = d.Quo(prev)
		delta = delta.Sub(guess)
		delta = delta.QuoInt(rootInt)

		guess = guess.Add(delta)
	}

	return guess, nil
}

// ApproxSqrt is a wrapper around ApproxRoot for the common special case
// of finding the square root of a number. It returns -(sqrt(abs(d)) if input is negative.
func (d BigDec) ApproxSqrt() (BigDec, error) {
	return d.ApproxRoot(2)
}

// is integer, e.g. decimals are zero
func (d BigDec) IsInteger() bool {
	return new(big.Int).Rem(d.i, precisionReuse).Sign() == 0
}

// format decimal state
func (d BigDec) Format(s fmt.State, verb rune) {
	_, err := s.Write([]byte(d.String()))
	if err != nil {
		panic(err)
	}
}

// String returns a BigDec as a string.
func (d BigDec) String() string {
	if d.i == nil {
		return d.i.String()
	}

	isNeg := d.IsNegative()

	if isNeg {
		d = d.Neg()
	}

	bzInt, err := d.i.MarshalText()
	if err != nil {
		return ""
	}
	inputSize := len(bzInt)

	var bzStr []byte

	// TODO: Remove trailing zeros
	// case 1, purely decimal
	if inputSize <= BigDecPrecision {
		bzStr = make([]byte, BigDecPrecision+2)

		// 0. prefix
		bzStr[0] = byte('0')
		bzStr[1] = byte('.')

		// set relevant digits to 0
		for i := 0; i < BigDecPrecision-inputSize; i++ {
			bzStr[i+2] = byte('0')
		}

		// set final digits
		copy(bzStr[2+(BigDecPrecision-inputSize):], bzInt)
	} else {
		// inputSize + 1 to account for the decimal point that is being added
		bzStr = make([]byte, inputSize+1)
		decPointPlace := inputSize - BigDecPrecision

		copy(bzStr, bzInt[:decPointPlace])                   // pre-decimal digits
		bzStr[decPointPlace] = byte('.')                     // decimal point
		copy(bzStr[decPointPlace+1:], bzInt[decPointPlace:]) // post-decimal digits
	}

	if isNeg {
		return "-" + string(bzStr)
	}

	return string(bzStr)
}

// Float64 returns the float64 representation of a BigDec.
// Will return the error if the conversion failed.
func (d BigDec) Float64() (float64, error) {
	return strconv.ParseFloat(d.String(), 64)
}

// MustFloat64 returns the float64 representation of a BigDec.
// Would panic if the conversion failed.
func (d BigDec) MustFloat64() float64 {
	if value, err := strconv.ParseFloat(d.String(), 64); err != nil {
		panic(err)
	} else {
		return value
	}
}

// Dec returns the osmomath.Dec representation of a BigDec.
// Values in any additional decimal places are truncated.
func (d BigDec) Dec() Dec {
	return d.DecWithPrecision(DecPrecision)
}

// DecWithPrecision converts BigDec to Dec with desired precision
// Example:
// BigDec: 1.010100000000153000000000000000000000
// precision: 4
// Output Dec: 1.010100000000000000
// Panics if precision exceeds DecPrecision
func (d BigDec) DecWithPrecision(precision uint64) Dec {
	var precisionFactor *big.Int
	if precision > DecPrecision {
		panic(fmt.Sprintf("maximum Dec precision is (%v), provided (%v)", DecPrecision, precision))
	} else {
		precisionFactor = precisionFactors[precision]
	}

	// Truncate any additional decimal values that exist due to BigDec's additional precision
	// This relies on big.Int's Quo function doing floor division
	intRepresentation := new(big.Int).Quo(d.BigInt(), precisionFactor)

	// convert int representation back to SDK Dec precision
	truncatedDec := NewDecFromBigIntWithPrec(intRepresentation, int64(precision))

	return truncatedDec
}

// ChopPrecisionMut truncates all decimals after precision numbers after decimal point. Mutative
// CONTRACT: precision <= BigDecPrecision
// Panics if precision exceeds BigDecPrecision
func (d *BigDec) ChopPrecisionMut(precision uint64) BigDec {
	if precision > BigDecPrecision {
		panic(fmt.Sprintf("maximum BigDec precision is (%v), provided (%v)", DecPrecision, precision))
	}

	precisionFactor := precisionFactors[precision]
	// big.Quo truncates numbers that would have been after decimal point
	d.i.Quo(d.i, precisionFactor)
	d.i.Mul(d.i, precisionFactor)
	return BigDec{d.i}
}

// ChopPrecision truncates all decimals after precision numbers after decimal point
// CONTRACT: precision <= BigDecPrecision
// Panics if precision exceeds BigDecPrecision
func (d *BigDec) ChopPrecision(precision uint64) BigDec {
	copy := d.Clone()
	return copy.ChopPrecisionMut(precision)
}

// DecRoundUp returns the osmomath.Dec representation of a BigDec.
// Round up at precision end.
// Values in any additional decimal places are truncated.
func (d BigDec) DecRoundUp() Dec {
	return NewDecFromBigIntWithPrec(chopPrecisionAndRoundUpDec(d.i), DecPrecision)
}

// BigDecFromDec returns the BigDec representation of an Dec.
// Values in any additional decimal places are truncated.
func BigDecFromDec(d Dec) BigDec {
	return NewBigDecFromBigIntWithPrec(d.BigInt(), DecPrecision)
}

// BigDecFromSDKInt returns the BigDec representation of an sdkInt.
// Values in any additional decimal places are truncated.
func BigDecFromSDKInt(i Int) BigDec {
	return NewBigDecFromBigIntWithPrec(i.BigInt(), 0)
}

// BigDecFromDecSlice returns the []BigDec representation of an []Dec.
// Values in any additional decimal places are truncated.
func BigDecFromDecSlice(ds []Dec) []BigDec {
	result := make([]BigDec, len(ds))
	for i, d := range ds {
		result[i] = NewBigDecFromBigIntWithPrec(d.BigInt(), DecPrecision)
	}
	return result
}

// BigDecFromDecSlice returns the []BigDec representation of an []Dec.
// Values in any additional decimal places are truncated.
func BigDecFromDecCoinSlice(ds []sdk.DecCoin) []BigDec {
	result := make([]BigDec, len(ds))
	for i, d := range ds {
		result[i] = NewBigDecFromBigIntWithPrec(d.Amount.BigInt(), DecPrecision)
	}
	return result
}

//     ____
//  __|    |__   "chop 'em
//       ` \     round!"
// ___||  ~  _     -bankers
// |         |      __
// |       | |   __|__|__
// |_____:  /   | $$$    |
//              |________|

// Remove a Precision amount of rightmost digits and perform bankers rounding
// on the remainder (gaussian rounding) on the digits which have been removed.
//
// Mutates the input. Use the non-mutative version if that is undesired
func chopPrecisionAndRound(d *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		d = chopPrecisionAndRound(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	switch rem.Cmp(fivePrecision) {
	case -1:
		return quo
	case 1:
		return quo.Add(quo, oneInt)
	default: // bankers rounding must take place
		// always round to an even number
		if quo.Bit(0) == 0 {
			return quo
		}
		return quo.Add(quo, oneInt)
	}
}

// chopPrecisionAndRoundUpBigDec removes a Precision amount of rightmost digits and rounds up.
// Non-mutative.
func chopPrecisionAndRoundUpBigDec(d *big.Int) *big.Int {
	// make copy
	copy := new(big.Int).Set(d)
	return chopPrecisionAndRoundUpMut(copy, precisionReuse)
}

// chopPrecisionAndRoundUpDec removes  DecPrecision amount of rightmost digits and rounds up.
// Non-mutative.
func chopPrecisionAndRoundUpDec(d *big.Int) *big.Int {
	copy := new(big.Int).Set(d)
	return chopPrecisionAndRoundUpMut(copy, precisionReuseSDK)
}

// chopPrecisionAndRoundUp removes a Precision amount of rightmost digits and rounds up.
// Mutates input d.
// Mutations occur:
// - By calling chopPrecisionAndTruncateMut.
// - Using input d directly in QuoRem.
func chopPrecisionAndRoundUpMut(d *big.Int, precisionReuse *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		// truncate since d is negative...
		d = chopPrecisionAndTruncateMut(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	_, rem := d.QuoRem(d, precisionReuse, big.NewInt(0))

	if rem.Sign() == 0 { // remainder is zero
		return d
	}

	return d.Add(d, oneInt)
}

func chopPrecisionAndRoundNonMutative(d *big.Int) *big.Int {
	tmp := new(big.Int).Set(d)
	return chopPrecisionAndRound(tmp)
}

// RoundInt64 rounds the decimal using bankers rounding
func (d BigDec) RoundInt64() int64 {
	chopped := chopPrecisionAndRoundNonMutative(d.i)
	if !chopped.IsInt64() {
		panic("Int64() out of bound")
	}
	return chopped.Int64()
}

// RoundInt round the decimal using bankers rounding
func (d BigDec) RoundInt() BigInt {
	return NewBigIntFromBigInt(chopPrecisionAndRoundNonMutative(d.i))
}

// chopPrecisionAndTruncate is similar to chopPrecisionAndRound,
// but always rounds down. It does not mutate the input.
func chopPrecisionAndTruncate(d *big.Int) *big.Int {
	return new(big.Int).Quo(d, precisionReuse)
}

// chopPrecisionAndTruncate is similar to chopPrecisionAndRound,
// but always rounds down. It mutates the input.
func chopPrecisionAndTruncateMut(d *big.Int) *big.Int {
	return d.Quo(d, precisionReuse)
}

// TruncateInt64 truncates the decimals from the number and returns an int64
func (d BigDec) TruncateInt64() int64 {
	chopped := chopPrecisionAndTruncate(d.i)
	if !chopped.IsInt64() {
		panic("Int64() out of bound")
	}
	return chopped.Int64()
}

// TruncateInt truncates the decimals from the number and returns an Int
func (d BigDec) TruncateInt() BigInt {
	return NewBigIntFromBigInt(chopPrecisionAndTruncate(d.i))
}

// TruncateDec truncates the decimals from the number and returns a Dec
func (d BigDec) TruncateDec() BigDec {
	return NewBigDecFromBigInt(chopPrecisionAndTruncate(d.i))
}

// Ceil returns the smallest interger value (as a decimal) that is greater than
// or equal to the given decimal.
func (d BigDec) Ceil() BigDec {
	tmp := new(big.Int).Set(d.i)

	quo, rem := tmp, big.NewInt(0)
	quo, rem = quo.QuoRem(tmp, precisionReuse, rem)

	// no need to round with a zero remainder regardless of sign
	if rem.Cmp(zeroInt) == 0 {
		return NewBigDecFromBigInt(quo)
	}

	if rem.Sign() == -1 {
		return NewBigDecFromBigInt(quo)
	}

	return NewBigDecFromBigInt(quo.Add(quo, oneInt))
}

// MaxSortableDec is the largest Dec that can be passed into SortableDecBytes()
// Its negative form is the least Dec that can be passed in.
var MaxSortableDec = OneBigDec().Quo(SmallestBigDec())

// ValidSortableDec ensures that a Dec is within the sortable bounds,
// a BigDec can't have a precision of less than 10^-18.
// Max sortable decimal was set to the reciprocal of SmallestDec.
func ValidSortableDec(dec BigDec) bool {
	return dec.Abs().LTE(MaxSortableDec)
}

// SortableDecBytes returns a byte slice representation of a Dec that can be sorted.
// Left and right pads with 0s so there are 18 digits to left and right of the decimal point.
// For this reason, there is a maximum and minimum value for this, enforced by ValidSortableDec.
func SortableDecBytes(dec BigDec) []byte {
	if !ValidSortableDec(dec) {
		panic("dec must be within bounds")
	}
	// Instead of adding an extra byte to all sortable decs in order to handle max sortable, we just
	// makes its bytes be "max" which comes after all numbers in ASCIIbetical order
	if dec.Equal(MaxSortableDec) {
		return []byte("max")
	}
	// For the same reason, we make the bytes of minimum sortable dec be --, which comes before all numbers.
	if dec.Equal(MaxSortableDec.Neg()) {
		return []byte("--")
	}
	// We move the negative sign to the front of all the left padded 0s, to make negative numbers come before positive numbers
	if dec.IsNegative() {
		return append([]byte("-"), []byte(fmt.Sprintf(fmt.Sprintf("%%0%ds", BigDecPrecision*2+1), dec.Abs().String()))...)
	}
	return []byte(fmt.Sprintf(fmt.Sprintf("%%0%ds", BigDecPrecision*2+1), dec.String()))
}

// reuse nil values
var nilJSON []byte

func init() {
	empty := new(big.Int)
	bz, _ := empty.MarshalText()
	nilJSON, _ = json.Marshal(string(bz))
}

// MarshalJSON marshals the decimal
func (d BigDec) MarshalJSON() ([]byte, error) {
	if d.i == nil {
		return nilJSON, nil
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON defines custom decoding scheme
func (d *BigDec) UnmarshalJSON(bz []byte) error {
	if d.i == nil {
		d.i = new(big.Int)
	}

	var text string
	err := json.Unmarshal(bz, &text)
	if err != nil {
		return err
	}

	// TODO: Reuse dec allocation
	newDec, err := NewBigDecFromStr(text)
	if err != nil {
		return err
	}

	d.i = newDec.i
	return nil
}

// MarshalYAML returns the YAML representation.
func (d BigDec) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// Marshal implements the gogo proto custom type interface.
func (d BigDec) Marshal() ([]byte, error) {
	if d.i == nil {
		d.i = new(big.Int)
	}
	return d.i.MarshalText()
}

// MarshalTo implements the gogo proto custom type interface.
func (d *BigDec) MarshalTo(data []byte) (n int, err error) {
	if d.i == nil {
		d.i = new(big.Int)
	}

	if d.i.Cmp(zeroInt) == 0 {
		copy(data, []byte{0x30})
		return 1, nil
	}

	bz, err := d.Marshal()
	if err != nil {
		return 0, err
	}

	copy(data, bz)
	return len(bz), nil
}

// Unmarshal implements the gogo proto custom type interface.
func (d *BigDec) Unmarshal(data []byte) error {
	if len(data) == 0 {
		d = nil
		return nil
	}

	if d.i == nil {
		d.i = new(big.Int)
	}

	if err := d.i.UnmarshalText(data); err != nil {
		return err
	}

	if d.i.BitLen() > maxBitLen {
		return fmt.Errorf("decimal out of range; got: %d, max: %d", d.i.BitLen(), maxBitLen)
	}

	return nil
}

// Size implements the gogo proto custom type interface.
func (d *BigDec) Size() int {
	bz, _ := d.Marshal()
	return len(bz)
}

// Override Amino binary serialization by proxying to protobuf.
func (d BigDec) MarshalAmino() ([]byte, error)   { return d.Marshal() }
func (d *BigDec) UnmarshalAmino(bz []byte) error { return d.Unmarshal(bz) }

// helpers

// DecsEqual tests if two decimal arrays are equal
func DecsEqual(d1s, d2s []BigDec) bool {
	if len(d1s) != len(d2s) {
		return false
	}

	for i, d1 := range d1s {
		if !d1.Equal(d2s[i]) {
			return false
		}
	}
	return true
}

// MinBigDec gets minimum decimal between two
func MinBigDec(d1, d2 BigDec) BigDec {
	if d1.LT(d2) {
		return d1
	}
	return d2
}

// MaxBigDec gets maximum decimal between two
func MaxBigDec(d1, d2 BigDec) BigDec {
	if d1.LT(d2) {
		return d2
	}
	return d1
}

// DecEq returns true if two given decimals are equal.
// Intended to be used with require/assert:  require.True(t, DecEq(...))
//
//nolint:thelper
func DecEq(t *testing.T, exp, got BigDec) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

// DecApproxEq returns true if the differences between two given decimals are smaller than the tolerance range.
// Intended to be used with require/assert:  require.True(t, DecEq(...))
//
//nolint:thelper
func DecApproxEq(t *testing.T, d1 BigDec, d2 BigDec, tol BigDec) (*testing.T, bool, string, string, string) {
	diff := d1.Sub(d2).Abs()
	return t, diff.LTE(tol), "expected |d1 - d2| <:\t%v\ngot |d1 - d2| = \t\t%v", tol.String(), diff.String()
}

// LogBase2 returns log_2 {x}.
// Rounds down by truncations during division and right shifting.
// Accurate up to 32 precision digits.
// Implementation is based on:
// https://stm32duinoforum.com/forum/dsp/BinaryLogarithm.pdf
func (x BigDec) LogBase2() BigDec {
	// create a new decimal to avoid mutating
	// the receiver's int buffer.
	xCopy := BigDec{}
	xCopy.i = new(big.Int).Set(x.i)
	if xCopy.LTE(zeroBigDec) {
		panic(fmt.Sprintf("log is not defined at <= 0, given (%s)", xCopy))
	}

	// Normalize x to be 1 <= x < 2.

	// y is the exponent that results in a whole multiple of 2.
	y := ZeroBigDec()

	// repeat until: x >= 1.
	for xCopy.LT(oneBigDec) {
		xCopy.i.Lsh(xCopy.i, 1)
		y.AddMut(negOneBigDec)
	}

	// repeat until: x < 2.
	for xCopy.GTE(twoBigDec) {
		xCopy.i.Rsh(xCopy.i, 1)
		y.AddMut(oneBigDec)
	}

	b := oneHalfBigDec.Clone()

	// N.B. At this point x is a positive real number representing
	// mantissa of the log. We estimate it using the following
	// algorithm:
	// https://stm32duinoforum.com/forum/dsp/BinaryLogarithm.pdf
	// This has shown precision of 32 digits relative
	// to Wolfram Alpha in tests.
	for i := 0; i < maxLog2Iterations; i++ {
		xCopy.MulMut(xCopy)
		if xCopy.GTE(twoBigDec) {
			xCopy.i.Rsh(xCopy.i, 1)
			y.AddMut(b)
		}
		b.i.Rsh(b.i, 1)
	}

	return y
}

// Natural logarithm of x.
// Formula: ln(x) = log_2(x) / log_2(e)
func (x BigDec) Ln() BigDec {
	log2x := x.LogBase2()

	y := log2x.Quo(logOfEbase2)

	return y
}

// log_1.0001(x) "tick" base logarithm
// Formula: log_1.0001(b) = log_2(b) / log_2(1.0001)
func (x BigDec) TickLog() BigDec {
	log2x := x.LogBase2()

	y := log2x.Quo(tickLogOf2)

	return y
}

// log_a(x) custom base logarithm
// Formula: log_a(b) = log_2(b) / log_2(a)
func (x BigDec) CustomBaseLog(base BigDec) BigDec {
	if base.LTE(ZeroBigDec()) || base.Equal(OneBigDec()) {
		panic(fmt.Sprintf("log is not defined at base <= 0 or base == 1, base given (%s)", base))
	}

	log2x_argument := x.LogBase2()
	log2x_base := base.LogBase2()

	y := log2x_argument.Quo(log2x_base)

	return y
}

// PowerInteger takes a given decimal to an integer power
// and returns the result. Non-mutative. Uses square and multiply
// algorithm for performing the calculation.
func (d BigDec) PowerInteger(power uint64) BigDec {
	clone := d.Clone()
	return clone.PowerIntegerMut(power)
}

// PowerIntegerMut takes a given decimal to an integer power
// and returns the result. Mutative. Uses square and multiply
// algorithm for performing the calculation.
func (d BigDec) PowerIntegerMut(power uint64) BigDec {
	if power == 0 {
		return OneBigDec()
	}
	tmp := OneBigDec()

	for i := power; i > 1; {
		if i%2 != 0 {
			tmp = tmp.MulMut(d)
		}
		i /= 2
		d = d.MulMut(d)
	}

	return d.MulMut(tmp)
}

// Power returns a result of raising the given big dec to
// a positive decimal power. Panics if the power is negative.
// Panics if the base is negative. Does not mutate the receiver.
// The max supported exponent is defined by the global maxSupportedExponent.
// If a greater exponent is given, the function panics.
// The error is not bounded but expected to be around 10^-18, use with care.
// See the underlying Exp2, LogBase2 and Mul for the details of their bounds.
// WARNING: This function is broken for base < 1. The reason is that logarithm function is
// negative between zero and 1, and the Exp2(k) is undefined for negative k.
// As a result, this function panics if called for d < 1.
func (d BigDec) Power(power BigDec) BigDec {
	if d.IsNegative() {
		panic(fmt.Sprintf("negative base is not supported for Power(), base was (%s)", d))
	}
	if power.IsNegative() {
		panic(fmt.Sprintf("negative power is not supported for Power(), power was (%s)", power))
	}
	if power.Abs().GT(maxSupportedExponent) {
		panic(fmt.Sprintf("integer exponent %s is too large, max (%s)", power, maxSupportedExponent))
	}
	if power.IsInteger() {
		return d.PowerInteger(power.TruncateInt().Uint64())
	}
	if power.IsZero() {
		return OneBigDec()
	}
	if d.IsZero() {
		return ZeroBigDec()
	}
	if d.LT(OneBigDec()) {
		panic(fmt.Sprintf("Power() is not supported for base < 1, base was (%s)", d))
	}
	if d.Equal(twoBigDec) {
		return Exp2(power)
	}

	// d^power = exp2(power * log_2{base})
	result := Exp2(d.LogBase2().Mul(power))

	return result
}
