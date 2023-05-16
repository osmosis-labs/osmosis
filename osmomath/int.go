package osmomath

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
)

const maxBitLen = 1024

func newIntegerFromString(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 0)
}

func equal(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == 0 }

// Greater than
func gt(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == 1 }

// Greater than or equal to
func gte(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) >= 0 }

// Less than
func lt(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) == -1 }

// Less than or equal to
func lte(i *big.Int, i2 *big.Int) bool { return i.Cmp(i2) <= 0 }

func add(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Add(i, i2) }

func sub(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Sub(i, i2) }

func mul(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mul(i, i2) }

func div(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Quo(i, i2) }

func mod(i *big.Int, i2 *big.Int) *big.Int { return new(big.Int).Mod(i, i2) }

func neg(i *big.Int) *big.Int { return new(big.Int).Neg(i) }

func abs(i *big.Int) *big.Int { return new(big.Int).Abs(i) }

func min(i *big.Int, i2 *big.Int) *big.Int {
	if i.Cmp(i2) == 1 {
		return new(big.Int).Set(i2)
	}

	return new(big.Int).Set(i)
}

func max(i *big.Int, i2 *big.Int) *big.Int {
	if i.Cmp(i2) == -1 {
		return new(big.Int).Set(i2)
	}

	return new(big.Int).Set(i)
}

func unmarshalText(i *big.Int, text string) error {
	if err := i.UnmarshalText([]byte(text)); err != nil {
		return err
	}

	if i.BitLen() > maxBitLen {
		return fmt.Errorf("integer out of range: %s", text)
	}

	return nil
}

// Int wraps big.Int with a 257 bit range bound
// Checks overflow, underflow and division by zero
// Exists in range from -(2^256 - 1) to 2^256 - 1
type BigInt struct {
	i *big.Int
}

// BigInt converts Int to big.Int
func (i BigInt) BigInt() *big.Int {
	if i.IsNil() {
		return nil
	}
	return new(big.Int).Set(i.i)
}

// IsNil returns true if Int is uninitialized
func (i BigInt) IsNil() bool {
	return i.i == nil
}

// NewInt constructs Int from int64
func NewInt(n int64) BigInt {
	return BigInt{big.NewInt(n)}
}

// NewIntFromUint64 constructs an Int from a uint64.
func NewIntFromUint64(n uint64) BigInt {
	b := big.NewInt(0)
	b.SetUint64(n)
	return BigInt{b}
}

// NewIntFromBigInt constructs Int from big.Int. If the provided big.Int is nil,
// it returns an empty instance. This function panics if the bit length is > 256.
func NewIntFromBigInt(i *big.Int) BigInt {
	if i == nil {
		return BigInt{}
	}

	if i.BitLen() > maxBitLen {
		panic("NewIntFromBigInt() out of bound")
	}
	return BigInt{i}
}

// NewIntFromString constructs Int from string
func NewIntFromString(s string) (res BigInt, ok bool) {
	i, ok := newIntegerFromString(s)
	if !ok {
		return
	}
	// Check overflow
	if i.BitLen() > maxBitLen {
		ok = false
		return
	}
	return BigInt{i}, true
}

// NewIntWithDecimal constructs Int with decimal
// Result value is n*10^dec
func NewIntWithDecimal(n int64, dec int) BigInt {
	if dec < 0 {
		panic("NewIntWithDecimal() decimal is negative")
	}
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
	i := new(big.Int)
	i.Mul(big.NewInt(n), exp)

	// Check overflow
	if i.BitLen() > maxBitLen {
		panic("NewIntWithDecimal() out of bound")
	}
	return BigInt{i}
}

// ZeroInt returns Int value with zero
func ZeroInt() BigInt { return BigInt{big.NewInt(0)} }

// OneInt returns Int value with one
func OneInt() BigInt { return BigInt{big.NewInt(1)} }

// ToDec converts Int to Dec
func (i BigInt) ToDec() BigDec {
	return NewDecFromInt(i)
}

// Int64 converts Int to int64
// Panics if the value is out of range
func (i BigInt) Int64() int64 {
	if !i.i.IsInt64() {
		panic("Int64() out of bound")
	}
	return i.i.Int64()
}

// IsInt64 returns true if Int64() not panics
func (i BigInt) IsInt64() bool {
	return i.i.IsInt64()
}

// Uint64 converts Int to uint64
// Panics if the value is out of range
func (i BigInt) Uint64() uint64 {
	if !i.i.IsUint64() {
		panic("Uint64() out of bounds")
	}
	return i.i.Uint64()
}

// IsUint64 returns true if Uint64() not panics
func (i BigInt) IsUint64() bool {
	return i.i.IsUint64()
}

// IsZero returns true if Int is zero
func (i BigInt) IsZero() bool {
	return i.i.Sign() == 0
}

// IsNegative returns true if Int is negative
func (i BigInt) IsNegative() bool {
	return i.i.Sign() == -1
}

// IsPositive returns true if Int is positive
func (i BigInt) IsPositive() bool {
	return i.i.Sign() == 1
}

// Sign returns sign of Int
func (i BigInt) Sign() int {
	return i.i.Sign()
}

// Equal compares two Ints
func (i BigInt) Equal(i2 BigInt) bool {
	return equal(i.i, i2.i)
}

// GT returns true if first Int is greater than second
func (i BigInt) GT(i2 BigInt) bool {
	return gt(i.i, i2.i)
}

// GTE returns true if receiver Int is greater than or equal to the parameter
// Int.
func (i BigInt) GTE(i2 BigInt) bool {
	return gte(i.i, i2.i)
}

// LT returns true if first Int is lesser than second
func (i BigInt) LT(i2 BigInt) bool {
	return lt(i.i, i2.i)
}

// LTE returns true if first Int is less than or equal to second
func (i BigInt) LTE(i2 BigInt) bool {
	return lte(i.i, i2.i)
}

// Add adds Int from another
func (i BigInt) Add(i2 BigInt) (res BigInt) {
	res = BigInt{add(i.i, i2.i)}
	// Check overflow
	if res.i.BitLen() > maxBitLen {
		panic("Int overflow")
	}
	return
}

// AddRaw adds int64 to Int
func (i BigInt) AddRaw(i2 int64) BigInt {
	return i.Add(NewInt(i2))
}

// Sub subtracts Int from another
func (i BigInt) Sub(i2 BigInt) (res BigInt) {
	res = BigInt{sub(i.i, i2.i)}
	// Check overflow
	if res.i.BitLen() > maxBitLen {
		panic("Int overflow")
	}
	return
}

// SubRaw subtracts int64 from Int
func (i BigInt) SubRaw(i2 int64) BigInt {
	return i.Sub(NewInt(i2))
}

// Mul multiples two Ints
func (i BigInt) Mul(i2 BigInt) (res BigInt) {
	// Check overflow
	if i.i.BitLen()+i2.i.BitLen()-1 > maxBitLen {
		panic("Int overflow")
	}
	res = BigInt{mul(i.i, i2.i)}
	// Check overflow if sign of both are same
	if res.i.BitLen() > maxBitLen {
		panic("Int overflow")
	}
	return
}

// MulRaw multipies Int and int64
func (i BigInt) MulRaw(i2 int64) BigInt {
	return i.Mul(NewInt(i2))
}

// Quo divides Int with Int
func (i BigInt) Quo(i2 BigInt) (res BigInt) {
	// Check division-by-zero
	if i2.i.Sign() == 0 {
		panic("Division by zero")
	}
	return BigInt{div(i.i, i2.i)}
}

// QuoRaw divides Int with int64
func (i BigInt) QuoRaw(i2 int64) BigInt {
	return i.Quo(NewInt(i2))
}

// Mod returns remainder after dividing with Int
func (i BigInt) Mod(i2 BigInt) BigInt {
	if i2.Sign() == 0 {
		panic("division-by-zero")
	}
	return BigInt{mod(i.i, i2.i)}
}

// ModRaw returns remainder after dividing with int64
func (i BigInt) ModRaw(i2 int64) BigInt {
	return i.Mod(NewInt(i2))
}

// Neg negates Int
func (i BigInt) Neg() (res BigInt) {
	return BigInt{neg(i.i)}
}

// Abs returns the absolute value of Int.
func (i BigInt) Abs() BigInt {
	return BigInt{abs(i.i)}
}

// return the minimum of the ints
func MinInt(i1, i2 BigInt) BigInt {
	return BigInt{min(i1.BigInt(), i2.BigInt())}
}

// MaxInt returns the maximum between two integers.
func MaxInt(i, i2 BigInt) BigInt {
	return BigInt{max(i.BigInt(), i2.BigInt())}
}

// Human readable string
func (i BigInt) String() string {
	return i.i.String()
}

// MarshalJSON defines custom encoding scheme
func (i BigInt) MarshalJSON() ([]byte, error) {
	if i.i == nil { // Necessary since default Uint initialization has i.i as nil
		i.i = new(big.Int)
	}
	return marshalJSON(i.i)
}

// UnmarshalJSON defines custom decoding scheme
func (i *BigInt) UnmarshalJSON(bz []byte) error {
	if i.i == nil { // Necessary since default Int initialization has i.i as nil
		i.i = new(big.Int)
	}
	return unmarshalJSON(i.i, bz)
}

// MarshalJSON for custom encoding scheme
// Must be encoded as a string for JSON precision
func marshalJSON(i encoding.TextMarshaler) ([]byte, error) {
	text, err := i.MarshalText()
	if err != nil {
		return nil, err
	}

	return json.Marshal(string(text))
}

// UnmarshalJSON for custom decoding scheme
// Must be encoded as a string for JSON precision
func unmarshalJSON(i *big.Int, bz []byte) error {
	var text string
	if err := json.Unmarshal(bz, &text); err != nil {
		return err
	}

	return unmarshalText(i, text)
}

// MarshalYAML returns the YAML representation.
func (i BigInt) MarshalYAML() (interface{}, error) {
	return i.String(), nil
}

// Marshal implements the gogo proto custom type interface.
func (i BigInt) Marshal() ([]byte, error) {
	if i.i == nil {
		i.i = new(big.Int)
	}
	return i.i.MarshalText()
}

// MarshalTo implements the gogo proto custom type interface.
func (i *BigInt) MarshalTo(data []byte) (n int, err error) {
	if i.i == nil {
		i.i = new(big.Int)
	}
	if i.i.BitLen() == 0 { // The value 0
		copy(data, []byte{0x30})
		return 1, nil
	}

	bz, err := i.Marshal()
	if err != nil {
		return 0, err
	}

	copy(data, bz)
	return len(bz), nil
}

// Unmarshal implements the gogo proto custom type interface.
func (i *BigInt) Unmarshal(data []byte) error {
	if len(data) == 0 {
		i = nil
		return nil
	}

	if i.i == nil {
		i.i = new(big.Int)
	}

	if err := i.i.UnmarshalText(data); err != nil {
		return err
	}

	if i.i.BitLen() > maxBitLen {
		return fmt.Errorf("integer out of range; got: %d, max: %d", i.i.BitLen(), maxBitLen)
	}

	return nil
}

// Size implements the gogo proto custom type interface.
func (i *BigInt) Size() int {
	bz, _ := i.Marshal()
	return len(bz)
}

// Override Amino binary serialization by proxying to protobuf.
func (i BigInt) MarshalAmino() ([]byte, error)   { return i.Marshal() }
func (i *BigInt) UnmarshalAmino(bz []byte) error { return i.Unmarshal(bz) }

// intended to be used with require/assert:  require.True(IntEq(...))
func IntEq(t *testing.T, exp, got BigInt) (*testing.T, bool, string, string, string) {
	return t, exp.Equal(got), "expected:\t%v\ngot:\t\t%v", exp.String(), got.String()
}
