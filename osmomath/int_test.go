package osmomath

import (
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type intTestSuite struct {
	suite.Suite
}

func TestIntTestSuite(t *testing.T) {
	suite.Run(t, new(intTestSuite))
}

func (s *intTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *intTestSuite) TestFromInt64() {
	for n := 0; n < 20; n++ {
		r := rand.Int63()
		s.Require().Equal(r, NewBigInt(r).Int64())
	}
}

func (s *intTestSuite) TestFromUint64() {
	for n := 0; n < 20; n++ {
		r := rand.Uint64()
		s.Require().True(NewBigIntFromUint64(r).IsUint64())
		s.Require().Equal(r, NewBigIntFromUint64(r).Uint64())
	}
}

func (s *intTestSuite) TestIntPanic() {
	// Max Int = 2^1024-1 = 8.988466e+308
	// Min Int = -(2^1024-1) = -8.988466e+308
	s.Require().NotPanics(func() { NewBigIntWithDecimal(4, 307) })
	i1 := NewBigIntWithDecimal(4, 307)
	s.Require().NotPanics(func() { NewBigIntWithDecimal(5, 307) })
	i2 := NewBigIntWithDecimal(5, 307)
	s.Require().NotPanics(func() { NewBigIntWithDecimal(92, 306) })
	i3 := NewBigIntWithDecimal(92, 306)

	s.Require().Panics(func() { NewBigIntWithDecimal(2, 308) })
	s.Require().Panics(func() { NewBigIntWithDecimal(9, 340) })

	// Overflow check
	s.Require().NotPanics(func() { i1.Add(i1) })
	s.Require().NotPanics(func() { i2.Add(i2) })
	s.Require().Panics(func() { i3.Add(i3) })

	s.Require().NotPanics(func() { i1.Sub(i1.Neg()) })
	s.Require().NotPanics(func() { i2.Sub(i2.Neg()) })
	s.Require().Panics(func() { i3.Sub(i3.Neg()) })

	s.Require().Panics(func() { i1.Mul(i1) })
	s.Require().Panics(func() { i2.Mul(i2) })
	s.Require().Panics(func() { i3.Mul(i3) })

	s.Require().Panics(func() { i1.Neg().Mul(i1.Neg()) })
	s.Require().Panics(func() { i2.Neg().Mul(i2.Neg()) })
	s.Require().Panics(func() { i3.Neg().Mul(i3.Neg()) })

	// Underflow check
	i3n := i3.Neg()
	s.Require().NotPanics(func() { i3n.Sub(i1) })
	s.Require().NotPanics(func() { i3n.Sub(i2) })
	s.Require().Panics(func() { i3n.Sub(i3) })

	s.Require().NotPanics(func() { i3n.Add(i1.Neg()) })
	s.Require().NotPanics(func() { i3n.Add(i2.Neg()) })
	s.Require().Panics(func() { i3n.Add(i3.Neg()) })

	s.Require().Panics(func() { i1.Mul(i1.Neg()) })
	s.Require().Panics(func() { i2.Mul(i2.Neg()) })
	s.Require().Panics(func() { i3.Mul(i3.Neg()) })

	// Bound check
	intmax := NewBigIntFromBigInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(1024), nil), big.NewInt(1)))
	intmin := intmax.Neg()
	s.Require().NotPanics(func() { intmax.Add(ZeroBigInt()) })
	s.Require().NotPanics(func() { intmin.Sub(ZeroBigInt()) })
	s.Require().Panics(func() { intmax.Add(OneBigInt()) })
	s.Require().Panics(func() { intmin.Sub(OneBigInt()) })

	s.Require().NotPanics(func() { NewBigIntFromBigInt(nil) })
	s.Require().True(NewBigIntFromBigInt(nil).IsNil())

	// Division-by-zero check
	s.Require().Panics(func() { i1.Quo(NewBigInt(0)) })

	s.Require().NotPanics(func() { BigInt{}.BigInt() })
}

// Tests below uses randomness
// Since we are using *big.Int as underlying value
// and (U/)Int is immutable value(see TestImmutability(U/)Int)
// it is safe to use randomness in the tests
func (s *intTestSuite) TestIdentInt() {
	for d := 0; d < 1000; d++ {
		n := rand.Int63()
		i := NewBigInt(n)

		ifromstr, ok := NewBigIntFromString(strconv.FormatInt(n, 10))
		s.Require().True(ok)

		cases := []int64{
			i.Int64(),
			i.BigInt().Int64(),
			ifromstr.Int64(),
			NewBigIntFromBigInt(big.NewInt(n)).Int64(),
			NewBigIntWithDecimal(n, 0).Int64(),
		}

		for tcnum, tc := range cases {
			s.Require().Equal(n, tc, "Int is modified during conversion. tc #%d", tcnum)
		}
	}
}

func minint(i1, i2 int64) int64 {
	if i1 < i2 {
		return i1
	}
	return i2
}

func maxint(i1, i2 int64) int64 {
	if i1 > i2 {
		return i1
	}
	return i2
}

func (s *intTestSuite) TestArithInt() {
	for d := 0; d < 1000; d++ {
		n1 := int64(rand.Int31())
		i1 := NewBigInt(n1)
		n2 := int64(rand.Int31())
		i2 := NewBigInt(n2)

		cases := []struct {
			ires BigInt
			nres int64
		}{
			{i1.Add(i2), n1 + n2},
			{i1.Sub(i2), n1 - n2},
			{i1.Mul(i2), n1 * n2},
			{i1.Quo(i2), n1 / n2},
			{i1.AddRaw(n2), n1 + n2},
			{i1.SubRaw(n2), n1 - n2},
			{i1.MulRaw(n2), n1 * n2},
			{i1.QuoRaw(n2), n1 / n2},
			{MinBigInt(i1, i2), minint(n1, n2)},
			{MaxBigInt(i1, i2), maxint(n1, n2)},
			{i1.Neg(), -n1},
			{i1.Abs(), n1},
			{i1.Neg().Abs(), n1},
		}

		for tcnum, tc := range cases {
			s.Require().Equal(tc.nres, tc.ires.Int64(), "Int arithmetic operation does not match with int64 operation. tc #%d", tcnum)
		}
	}
}

func (s *intTestSuite) TestCompInt() {
	for d := 0; d < 1000; d++ {
		n1 := int64(rand.Int31())
		i1 := NewBigInt(n1)
		n2 := int64(rand.Int31())
		i2 := NewBigInt(n2)

		cases := []struct {
			ires bool
			nres bool
		}{
			{i1.Equal(i2), n1 == n2},
			{i1.GT(i2), n1 > n2},
			{i1.LT(i2), n1 < n2},
			{i1.LTE(i2), n1 <= n2},
		}

		for tcnum, tc := range cases {
			s.Require().Equal(tc.nres, tc.ires, "Int comparison operation does not match with int64 operation. tc #%d", tcnum)
		}
	}
}

func randint() BigInt {
	return NewBigInt(rand.Int63())
}

func (s *intTestSuite) TestImmutabilityAllInt() {
	ops := []func(*BigInt){
		func(i *BigInt) { _ = i.Add(randint()) },
		func(i *BigInt) { _ = i.Sub(randint()) },
		func(i *BigInt) { _ = i.Mul(randint()) },
		func(i *BigInt) { _ = i.Quo(randint()) },
		func(i *BigInt) { _ = i.AddRaw(rand.Int63()) },
		func(i *BigInt) { _ = i.SubRaw(rand.Int63()) },
		func(i *BigInt) { _ = i.MulRaw(rand.Int63()) },
		func(i *BigInt) { _ = i.QuoRaw(rand.Int63()) },
		func(i *BigInt) { _ = i.Neg() },
		func(i *BigInt) { _ = i.Abs() },
		func(i *BigInt) { _ = i.IsZero() },
		func(i *BigInt) { _ = i.Sign() },
		func(i *BigInt) { _ = i.Equal(randint()) },
		func(i *BigInt) { _ = i.GT(randint()) },
		func(i *BigInt) { _ = i.LT(randint()) },
		func(i *BigInt) { _ = i.String() },
	}

	for i := 0; i < 1000; i++ {
		n := rand.Int63()
		ni := NewBigInt(n)

		for opnum, op := range ops {
			op(&ni)

			s.Require().Equal(n, ni.Int64(), "Int is modified by operation. tc #%d", opnum)
			s.Require().Equal(NewBigInt(n), ni, "Int is modified by operation. tc #%d", opnum)
		}
	}
}

func (s *intTestSuite) TestEncodingTableInt() {
	var i BigInt

	cases := []struct {
		i      BigInt
		jsonBz []byte
		rawBz  []byte
	}{
		{
			NewBigInt(0),
			[]byte("\"0\""),
			[]byte{0x30},
		},
		{
			NewBigInt(100),
			[]byte("\"100\""),
			[]byte{0x31, 0x30, 0x30},
		},
		{
			NewBigInt(-100),
			[]byte("\"-100\""),
			[]byte{0x2d, 0x31, 0x30, 0x30},
		},
		{
			NewBigInt(51842),
			[]byte("\"51842\""),
			[]byte{0x35, 0x31, 0x38, 0x34, 0x32},
		},
		{
			NewBigInt(-51842),
			[]byte("\"-51842\""),
			[]byte{0x2d, 0x35, 0x31, 0x38, 0x34, 0x32},
		},
		{
			NewBigInt(19513368),
			[]byte("\"19513368\""),
			[]byte{0x31, 0x39, 0x35, 0x31, 0x33, 0x33, 0x36, 0x38},
		},
		{
			NewBigInt(-19513368),
			[]byte("\"-19513368\""),
			[]byte{0x2d, 0x31, 0x39, 0x35, 0x31, 0x33, 0x33, 0x36, 0x38},
		},
		{
			NewBigInt(999999999999),
			[]byte("\"999999999999\""),
			[]byte{0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39},
		},
		{
			NewBigInt(-999999999999),
			[]byte("\"-999999999999\""),
			[]byte{0x2d, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39},
		},
	}

	for tcnum, tc := range cases {
		bz, err := tc.i.MarshalJSON()
		s.Require().Nil(err, "Error marshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.jsonBz, bz, "Marshaled value is different from exported. tc #%d", tcnum)

		err = (&i).UnmarshalJSON(bz)
		s.Require().Nil(err, "Error unmarshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.i, i, "Unmarshaled value is different from exported. tc #%d", tcnum)

		bz, err = tc.i.Marshal()
		s.Require().Nil(err, "Error marshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.rawBz, bz, "Marshaled value is different from exported. tc #%d", tcnum)

		err = (&i).Unmarshal(bz)
		s.Require().Nil(err, "Error unmarshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.i, i, "Unmarshaled value is different from exported. tc #%d", tcnum)
	}
}

func (s *intTestSuite) TestEncodingTableUint() {
	var i Uint

	cases := []struct {
		i      Uint
		jsonBz []byte
		rawBz  []byte
	}{
		{
			NewUint(0),
			[]byte("\"0\""),
			[]byte{0x30},
		},
		{
			NewUint(100),
			[]byte("\"100\""),
			[]byte{0x31, 0x30, 0x30},
		},
		{
			NewUint(51842),
			[]byte("\"51842\""),
			[]byte{0x35, 0x31, 0x38, 0x34, 0x32},
		},
		{
			NewUint(19513368),
			[]byte("\"19513368\""),
			[]byte{0x31, 0x39, 0x35, 0x31, 0x33, 0x33, 0x36, 0x38},
		},
		{
			NewUint(999999999999),
			[]byte("\"999999999999\""),
			[]byte{0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39, 0x39},
		},
	}

	for tcnum, tc := range cases {
		bz, err := tc.i.MarshalJSON()
		s.Require().Nil(err, "Error marshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.jsonBz, bz, "Marshaled value is different from exported. tc #%d", tcnum)

		err = (&i).UnmarshalJSON(bz)
		s.Require().Nil(err, "Error unmarshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.i, i, "Unmarshaled value is different from exported. tc #%d", tcnum)

		bz, err = tc.i.Marshal()
		s.Require().Nil(err, "Error marshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.rawBz, bz, "Marshaled value is different from exported. tc #%d", tcnum)

		err = (&i).Unmarshal(bz)
		s.Require().Nil(err, "Error unmarshaling Int. tc #%d, err %s", tcnum, err)
		s.Require().Equal(tc.i, i, "Unmarshaled value is different from exported. tc #%d", tcnum)
	}
}

func (s *intTestSuite) TestIntMod() {
	tests := []struct {
		name      string
		x         int64
		y         int64
		ret       int64
		wantPanic bool
	}{
		{"3 % 10", 3, 10, 3, false},
		{"10 % 3", 10, 3, 1, false},
		{"4 % 2", 4, 2, 0, false},
		{"2 % 0", 2, 0, 0, true},
	}

	for _, tt := range tests {
		if tt.wantPanic {
			s.Require().Panics(func() { NewBigInt(tt.x).Mod(NewBigInt(tt.y)) })
			s.Require().Panics(func() { NewBigInt(tt.x).ModRaw(tt.y) })
			return
		}
		s.Require().True(NewBigInt(tt.x).Mod(NewBigInt(tt.y)).Equal(NewBigInt(tt.ret)))
		s.Require().True(NewBigInt(tt.x).ModRaw(tt.y).Equal(NewBigInt(tt.ret)))
	}
}

func (s *intTestSuite) TestIntEq() {
	_, resp, _, _, _ := BigIntEq(s.T(), ZeroBigInt(), ZeroBigInt())
	s.Require().True(resp)
	_, resp, _, _, _ = BigIntEq(s.T(), OneBigInt(), ZeroBigInt())
	s.Require().False(resp)
}

func TestRoundTripMarshalToInt(t *testing.T) {
	values := []int64{
		0,
		1,
		1 << 10,
		1<<10 - 3,
		1<<63 - 1,
		1<<32 - 7,
		1<<22 - 8,
	}

	for _, value := range values {
		value := value
		t.Run(fmt.Sprintf("%d", value), func(t *testing.T) {
			t.Parallel()

			var scratch [20]byte
			iv := NewBigInt(value)
			n, err := iv.MarshalTo(scratch[:])
			if err != nil {
				t.Fatal(err)
			}
			rt := new(BigInt)
			if err := rt.Unmarshal(scratch[:n]); err != nil {
				t.Fatal(err)
			}
			if !rt.Equal(iv) {
				t.Fatalf("roundtrip=%q != original=%q", rt, iv)
			}
		})
	}
}

func (s *intTestSuite) TestEncodingRandom() {
	for i := 0; i < 1000; i++ {
		n := rand.Int63()
		ni := NewBigInt(n)
		var ri BigInt

		str, err := ni.Marshal()
		s.Require().Nil(err)
		err = (&ri).Unmarshal(str)
		s.Require().Nil(err)

		s.Require().Equal(ni, ri, "binary mismatch; tc #%d, expected %s, actual %s", i, ni.String(), ri.String())
		s.Require().True(ni.i != ri.i, "pointer addresses are equal; tc #%d", i)

		bz, err := ni.MarshalJSON()
		s.Require().Nil(err)
		err = (&ri).UnmarshalJSON(bz)
		s.Require().Nil(err)

		s.Require().Equal(ni, ri, "json mismatch; tc #%d, expected %s, actual %s", i, ni.String(), ri.String())
		s.Require().True(ni.i != ri.i, "pointer addresses are equal; tc #%d", i)
	}

	for i := 0; i < 1000; i++ {
		n := rand.Uint64()
		ni := NewUint(n)

		var ri Uint

		str, err := ni.Marshal()
		s.Require().Nil(err)
		err = (&ri).Unmarshal(str)
		s.Require().Nil(err)

		s.Require().Equal(ni, ri, "binary mismatch; tc #%d, expected %s, actual %s", i, ni.String(), ri.String())

		bz, err := ni.MarshalJSON()
		s.Require().Nil(err)
		err = (&ri).UnmarshalJSON(bz)
		s.Require().Nil(err)

		s.Require().Equal(ni, ri, "json mismatch; tc #%d, expected %s, actual %s", i, ni.String(), ri.String())
	}
}
