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
	Precision = 36

	// bytes required to represent the above precision
	// Ceiling[Log2[10**Precision - 1]]
	DecimalPrecisionBits = 120

	maxDecBitLen = maxBitLen + DecimalPrecisionBits

	// max number of iterations in ApproxRoot function
	maxApproxRootIterations = 100
)

var (
	precisionReuse       = new(big.Int).Exp(big.NewInt(10), big.NewInt(Precision), nil)
	fivePrecision        = new(big.Int).Quo(precisionReuse, big.NewInt(2))
	precisionMultipliers []*big.Int
	zeroInt              = big.NewInt(0)
	oneInt               = big.NewInt(1)
	tenInt               = big.NewInt(10)
)

// Decimal errors
var (
	ErrEmptyDecimalStr      = errors.New("decimal string cannot be empty")
	ErrInvalidDecimalLength = errors.New("invalid decimal length")
	ErrInvalidDecimalStr    = errors.New("invalid decimal string")
)

// Set precision multipliers
func init() {
	precisionMultipliers = make([]*big.Int, Precision+1)
	for i := 0; i <= Precision; i++ {
		precisionMultipliers[i] = calcPrecisionMultiplier(int64(i))
	}
}

func precisionInt() *big.Int {
	return new(big.Int).Set(precisionReuse)
}

func ZeroDec() BigDec     { return BigDec{new(big.Int).Set(zeroInt)} }
func OneDec() BigDec      { return BigDec{precisionInt()} }
func SmallestDec() BigDec { return BigDec{new(big.Int).Set(oneInt)} }

// calculate the precision multiplier
func calcPrecisionMultiplier(prec int64) *big.Int {
	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	zerosToAdd := Precision - prec
	multiplier := new(big.Int).Exp(tenInt, big.NewInt(zerosToAdd), nil)
	return multiplier
}

// get the precision multiplier, do not mutate result
func precisionMultiplier(prec int64) *big.Int {
	if prec > Precision {
		panic(fmt.Sprintf("too much precision, maximum %v, provided %v", Precision, prec))
	}
	return precisionMultipliers[prec]
}

// create a new NewBigDec from integer assuming whole number
func NewBigDec(i int64) BigDec {
	return NewDecWithPrec(i, 0)
}

// create a new BigDec from integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewDecWithPrec(i, prec int64) BigDec {
	return BigDec{
		new(big.Int).Mul(big.NewInt(i), precisionMultiplier(prec)),
	}
}

// create a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromBigInt(i *big.Int) BigDec {
	return NewDecFromBigIntWithPrec(i, 0)
}

// create a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromBigIntWithPrec(i *big.Int, prec int64) BigDec {
	return BigDec{
		new(big.Int).Mul(i, precisionMultiplier(prec)),
	}
}

// create a new BigDec from big integer assuming whole numbers
// CONTRACT: prec <= Precision
func NewDecFromInt(i BigInt) BigDec {
	return NewDecFromIntWithPrec(i, 0)
}

// create a new BigDec from big integer with decimal place at prec
// CONTRACT: prec <= Precision
func NewDecFromIntWithPrec(i BigInt, prec int64) BigDec {
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
func NewDecFromStr(str string) (BigDec, error) {
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

	if lenDecs > Precision {
		return BigDec{}, fmt.Errorf("invalid precision; max: %d, got: %d", Precision, lenDecs)
	}

	// add some extra zero's to correct to the Precision factor
	zerosToAdd := Precision - lenDecs
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
func MustNewDecFromStr(s string) BigDec {
	dec, err := NewDecFromStr(s)
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
func (d BigDec) Abs() BigDec          { return BigDec{new(big.Int).Abs(d.i)} } // absolute value

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
	res := new(big.Int).Add(d.i, d2.i)

	if res.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{res}
}

// subtraction
func (d BigDec) Sub(d2 BigDec) BigDec {
	res := new(big.Int).Sub(d.i, d2.i)

	if res.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{res}
}

// multiplication
func (d BigDec) Mul(d2 BigDec) BigDec {
	mul := new(big.Int).Mul(d.i, d2.i)
	chopped := chopPrecisionAndRound(mul)

	if chopped.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{chopped}
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
	// multiply precision twice
	mul := new(big.Int).Mul(d.i, precisionReuse)
	mul.Mul(mul, precisionReuse)

	quo := new(big.Int).Quo(mul, d2.i)
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

// quotient, round up
func (d BigDec) QuoRoundUp(d2 BigDec) BigDec {
	// multiply precision twice
	mul := new(big.Int).Mul(d.i, precisionReuse)
	mul.Mul(mul, precisionReuse)

	quo := new(big.Int).Quo(mul, d2.i)
	chopped := chopPrecisionAndRoundUp(quo)

	if chopped.BitLen() > maxDecBitLen {
		panic("Int overflow")
	}
	return BigDec{chopped}
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

	if root == 1 || d.IsZero() || d.Equal(OneDec()) {
		return d, nil
	}

	if root == 0 {
		return OneDec(), nil
	}

	rootInt := NewIntFromUint64(root)
	guess, delta := OneDec(), OneDec()

	for iter := 0; delta.Abs().GT(SmallestDec()) && iter < maxApproxRootIterations; iter++ {
		prev := guess.Power(root - 1)
		if prev.IsZero() {
			prev = SmallestDec()
		}
		delta = d.Quo(prev)
		delta = delta.Sub(guess)
		delta = delta.QuoInt(rootInt)

		guess = guess.Add(delta)
	}

	return guess, nil
}

// Power returns a the result of raising to a positive integer power
func (d BigDec) Power(power uint64) BigDec {
	if power == 0 {
		return OneDec()
	}
	tmp := OneDec()

	for i := power; i > 1; {
		if i%2 != 0 {
			tmp = tmp.Mul(d)
		}
		i /= 2
		d = d.Mul(d)
	}

	return d.Mul(tmp)
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
	if inputSize <= Precision {
		bzStr = make([]byte, Precision+2)

		// 0. prefix
		bzStr[0] = byte('0')
		bzStr[1] = byte('.')

		// set relevant digits to 0
		for i := 0; i < Precision-inputSize; i++ {
			bzStr[i+2] = byte('0')
		}

		// set final digits
		copy(bzStr[2+(Precision-inputSize):], bzInt)
	} else {
		// inputSize + 1 to account for the decimal point that is being added
		bzStr = make([]byte, inputSize+1)
		decPointPlace := inputSize - Precision

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

// SdkDec returns the Sdk.Dec representation of a BigDec.
// Values in any additional decimal places are truncated.
func (d BigDec) SDKDec() sdk.Dec {
	precisionDiff := Precision - sdk.Precision
	precisionFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precisionDiff)), nil)

	if precisionDiff < 0 {
		panic("invalid decimal precision")
	}

	// Truncate any additional decimal values that exist due to BigDec's additional precision
	// This relies on big.Int's Quo function doing floor division
	intRepresentation := new(big.Int).Quo(d.BigInt(), precisionFactor)

	// convert int representation back to SDK Dec precision
	truncatedDec := sdk.NewDecFromBigIntWithPrec(intRepresentation, sdk.Precision)

	return truncatedDec
}

// BigDecFromSdkDec returns the BigDec representation of an SDKDec.
// Values in any additional decimal places are truncated.
func BigDecFromSDKDec(d sdk.Dec) BigDec {
	return NewDecFromBigIntWithPrec(d.BigInt(), sdk.Precision)
}

// BigDecFromSdkDecSlice returns the []BigDec representation of an []SDKDec.
// Values in any additional decimal places are truncated.
func BigDecFromSDKDecSlice(ds []sdk.Dec) []BigDec {
	result := make([]BigDec, len(ds))
	for i, d := range ds {
		result[i] = NewDecFromBigIntWithPrec(d.BigInt(), sdk.Precision)
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

// chopPrecisionAndRoundUp removes a Precision amount of rightmost digits and rounds up.
func chopPrecisionAndRoundUp(d *big.Int) *big.Int {
	// remove the negative and add it back when returning
	if d.Sign() == -1 {
		// make d positive, compute chopped value, and then un-mutate d
		d = d.Neg(d)
		// truncate since d is negative...
		d = chopPrecisionAndTruncate(d)
		d = d.Neg(d)
		return d
	}

	// get the truncated quotient and remainder
	quo, rem := d, big.NewInt(0)
	quo, rem = quo.QuoRem(d, precisionReuse, rem)

	if rem.Sign() == 0 { // remainder is zero
		return quo
	}

	return quo.Add(quo, oneInt)
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
	return NewIntFromBigInt(chopPrecisionAndRoundNonMutative(d.i))
}

// chopPrecisionAndTruncate is similar to chopPrecisionAndRound,
// but always rounds down. It does not mutate the input.
func chopPrecisionAndTruncate(d *big.Int) *big.Int {
	return new(big.Int).Quo(d, precisionReuse)
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
	return NewIntFromBigInt(chopPrecisionAndTruncate(d.i))
}

// TruncateDec truncates the decimals from the number and returns a Dec
func (d BigDec) TruncateDec() BigDec {
	return NewDecFromBigInt(chopPrecisionAndTruncate(d.i))
}

// Ceil returns the smallest interger value (as a decimal) that is greater than
// or equal to the given decimal.
func (d BigDec) Ceil() BigDec {
	tmp := new(big.Int).Set(d.i)

	quo, rem := tmp, big.NewInt(0)
	quo, rem = quo.QuoRem(tmp, precisionReuse, rem)

	// no need to round with a zero remainder regardless of sign
	if rem.Cmp(zeroInt) == 0 {
		return NewDecFromBigInt(quo)
	}

	if rem.Sign() == -1 {
		return NewDecFromBigInt(quo)
	}

	return NewDecFromBigInt(quo.Add(quo, oneInt))
}

// MaxSortableDec is the largest Dec that can be passed into SortableDecBytes()
// Its negative form is the least Dec that can be passed in.
var MaxSortableDec = OneDec().Quo(SmallestDec())

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
		return append([]byte("-"), []byte(fmt.Sprintf(fmt.Sprintf("%%0%ds", Precision*2+1), dec.Abs().String()))...)
	}
	return []byte(fmt.Sprintf(fmt.Sprintf("%%0%ds", Precision*2+1), dec.String()))
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
	newDec, err := NewDecFromStr(text)
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

// MinDec gets minimum decimal between two
func MinDec(d1, d2 BigDec) BigDec {
	if d1.LT(d2) {
		return d1
	}
	return d2
}

// MaxDec gets maximum decimal between two
func MaxDec(d1, d2 BigDec) BigDec {
	if d1.LT(d2) {
		return d2
	}
	return d1
}

// DecEq returns true if two given decimals are equal.
// Intended to be used with require/assert:  require.True(t, DecEq(...))
func DecEq(t *testing.T, exp, got BigDec) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}

// DecApproxEq returns true if the differences between two given decimals are smaller than the tolerance range.
// Intended to be used with require/assert:  require.True(t, DecEq(...))
func DecApproxEq(t *testing.T, d1 BigDec, d2 BigDec, tol BigDec) (*testing.T, bool, string, string, string) {
	diff := d1.Sub(d2).Abs()
	return t, diff.LTE(tol), "expected |d1 - d2| <:\t%v\ngot |d1 - d2| = \t\t%v", tol.String(), diff.String()
}
