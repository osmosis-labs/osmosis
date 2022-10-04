package osmomath

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

type decimalTestSuite struct {
	suite.Suite
}

func TestDecimalTestSuite(t *testing.T) {
	suite.Run(t, new(decimalTestSuite))
}

func TestDecApproxEq(t *testing.T) {
	// d1 = 0.55, d2 = 0.6, tol = 0.1
	d1 := NewDecWithPrec(55, 2)
	d2 := NewDecWithPrec(6, 1)
	tol := NewDecWithPrec(1, 1)

	require.True(DecApproxEq(t, d1, d2, tol))

	// d1 = 0.55, d2 = 0.6, tol = 1E-5
	d1 = NewDecWithPrec(55, 2)
	d2 = NewDecWithPrec(6, 1)
	tol = NewDecWithPrec(1, 5)

	require.False(DecApproxEq(t, d1, d2, tol))

	// d1 = 0.6, d2 = 0.61, tol = 0.01
	d1 = NewDecWithPrec(6, 1)
	d2 = NewDecWithPrec(61, 2)
	tol = NewDecWithPrec(1, 2)

	require.True(DecApproxEq(t, d1, d2, tol))
}

// create a decimal from a decimal string (ex. "1234.5678")
func (s *decimalTestSuite) mustNewDecFromStr(str string) (d BigDec) {
	d, err := NewDecFromStr(str)
	s.Require().NoError(err)

	return d
}

func (s *decimalTestSuite) TestNewDecFromStr() {
	largeBigInt, success := new(big.Int).SetString("3144605511029693144278234343371835", 10)
	s.Require().True(success)

	tests := []struct {
		decimalStr string
		expErr     bool
		exp        BigDec
	}{
		{"", true, BigDec{}},
		{"0.-75", true, BigDec{}},
		{"0", false, NewBigDec(0)},
		{"1", false, NewBigDec(1)},
		{"1.1", false, NewDecWithPrec(11, 1)},
		{"0.75", false, NewDecWithPrec(75, 2)},
		{"0.8", false, NewDecWithPrec(8, 1)},
		{"0.11111", false, NewDecWithPrec(11111, 5)},
		{"314460551102969.31442782343433718353144278234343371835", true, NewBigDec(3141203149163817869)},
		{
			"314460551102969314427823434337.18357180924882313501835718092488231350",
			true, NewDecFromBigIntWithPrec(largeBigInt, 4),
		},
		{
			"314460551102969314427823434337.1835",
			false, NewDecFromBigIntWithPrec(largeBigInt, 4),
		},
		{".", true, BigDec{}},
		{".0", true, NewBigDec(0)},
		{"1.", true, NewBigDec(1)},
		{"foobar", true, BigDec{}},
		{"0.foobar", true, BigDec{}},
		{"0.foobar.", true, BigDec{}},
		{"179769313486231590772930519078902473361797697894230657273430081157732675805500963132708477322407536021120113879871393357658789768814416622492847430639474124377767893424865485276302219601246094119453082952085005768838150682342462881473913110540827237163350510684586298239947245938479716304835356329624224137216", true, BigDec{}},
	}

	for tcIndex, tc := range tests {
		res, err := NewDecFromStr(tc.decimalStr)
		if tc.expErr {
			s.Require().NotNil(err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			s.Require().Nil(err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			s.Require().True(res.Equal(tc.exp), "equality was incorrect, res %v, exp %v, tc %v", res, tc.exp, tcIndex)
		}

		// negative tc
		res, err = NewDecFromStr("-" + tc.decimalStr)
		if tc.expErr {
			s.Require().NotNil(err, "error expected, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
		} else {
			s.Require().Nil(err, "unexpected error, decimalStr %v, tc %v", tc.decimalStr, tcIndex)
			exp := tc.exp.Mul(NewBigDec(-1))
			s.Require().True(res.Equal(exp), "equality was incorrect, res %v, exp %v, tc %v", res, exp, tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestDecString() {
	tests := []struct {
		d    BigDec
		want string
	}{
		{NewBigDec(0), "0.000000000000000000000000000000000000"},
		{NewBigDec(1), "1.000000000000000000000000000000000000"},
		{NewBigDec(10), "10.000000000000000000000000000000000000"},
		{NewBigDec(12340), "12340.000000000000000000000000000000000000"},
		{NewDecWithPrec(12340, 4), "1.234000000000000000000000000000000000"},
		{NewDecWithPrec(12340, 5), "0.123400000000000000000000000000000000"},
		{NewDecWithPrec(12340, 8), "0.000123400000000000000000000000000000"},
		{NewDecWithPrec(1009009009009009009, 17), "10.090090090090090090000000000000000000"},
		{MustNewDecFromStr("10.090090090090090090090090090090090090"), "10.090090090090090090090090090090090090"},
	}
	for tcIndex, tc := range tests {
		s.Require().Equal(tc.want, tc.d.String(), "bad String(), index: %v", tcIndex)
	}
}

func (s *decimalTestSuite) TestDecFloat64() {
	tests := []struct {
		d    BigDec
		want float64
	}{
		{NewBigDec(0), 0.000000000000000000},
		{NewBigDec(1), 1.000000000000000000},
		{NewBigDec(10), 10.000000000000000000},
		{NewBigDec(12340), 12340.000000000000000000},
		{NewDecWithPrec(12340, 4), 1.234000000000000000},
		{NewDecWithPrec(12340, 5), 0.123400000000000000},
		{NewDecWithPrec(12340, 8), 0.000123400000000000},
		{NewDecWithPrec(1009009009009009009, 17), 10.090090090090090090},
	}
	for tcIndex, tc := range tests {
		value, err := tc.d.Float64()
		s.Require().Nil(err, "error getting Float64(), index: %v", tcIndex)
		s.Require().Equal(tc.want, value, "bad Float64(), index: %v", tcIndex)
		s.Require().Equal(tc.want, tc.d.MustFloat64(), "bad MustFloat64(), index: %v", tcIndex)
	}
}

func (s *decimalTestSuite) TestSdkDec() {
	tests := []struct {
		d    BigDec
		want sdk.Dec
		expPanic bool
	}{
		{NewBigDec(0), sdk.MustNewDecFromStr("0.000000000000000000"), false},
		{NewBigDec(1), sdk.MustNewDecFromStr("1.000000000000000000"), false},
		{NewBigDec(10), sdk.MustNewDecFromStr("10.000000000000000000"), false},
		{NewBigDec(12340), sdk.MustNewDecFromStr("12340.000000000000000000"), false},
		{NewDecWithPrec(12340, 4), sdk.MustNewDecFromStr("1.234000000000000000"), false},
		{NewDecWithPrec(12340, 5), sdk.MustNewDecFromStr("0.123400000000000000"), false},
		{NewDecWithPrec(12340, 8), sdk.MustNewDecFromStr("0.000123400000000000"), false},
		{NewDecWithPrec(1009009009009009009, 17), sdk.MustNewDecFromStr("10.090090090090090090"), false},
	}
	for tcIndex, tc := range tests {
		if tc.expPanic {
			s.Require().Panics(func() { tc.d.SDKDec() })
		} else {
			value := tc.d.SDKDec()
			s.Require().Equal(tc.want, value, "bad SdkDec(), index: %v", tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestBigDecFromSdkDec() {
	tests := []struct {
		d    sdk.Dec
		want BigDec
		expPanic bool
	}{
		{sdk.MustNewDecFromStr("0.000000000000000000"), NewBigDec(0), false},
		{sdk.MustNewDecFromStr("1.000000000000000000"), NewBigDec(1), false},
		{sdk.MustNewDecFromStr("10.000000000000000000"), NewBigDec(10), false},
		{sdk.MustNewDecFromStr("12340.000000000000000000"), NewBigDec(12340), false},
		{sdk.MustNewDecFromStr("1.234000000000000000"), NewDecWithPrec(12340, 4), false},
		{sdk.MustNewDecFromStr("0.123400000000000000"), NewDecWithPrec(12340, 5), false},
		{sdk.MustNewDecFromStr("0.000123400000000000"), NewDecWithPrec(12340, 8), false},
		{sdk.MustNewDecFromStr("10.090090090090090090"), NewDecWithPrec(1009009009009009009, 17), false},
	}
	for tcIndex, tc := range tests {
		if tc.expPanic {
			s.Require().Panics(func() { BigDecFromSDKDec(tc.d) })
		} else {
			value := BigDecFromSDKDec(tc.d)
			s.Require().Equal(tc.want, value, "bad BigDecFromSdkDec(), index: %v", tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestBigDecFromSdkDecSlice() {
	tests := []struct {
		d    []sdk.Dec
		want []BigDec
		expPanic bool
	}{
		{[]sdk.Dec{sdk.MustNewDecFromStr("0.000000000000000000")}, []BigDec{NewBigDec(0)}, false},
		{[]sdk.Dec{sdk.MustNewDecFromStr("0.000000000000000000"), sdk.MustNewDecFromStr("1.000000000000000000")}, []BigDec{NewBigDec(0), NewBigDec(1)}, false},
		{[]sdk.Dec{sdk.MustNewDecFromStr("1.000000000000000000"), sdk.MustNewDecFromStr("0.000000000000000000"), sdk.MustNewDecFromStr("0.000123400000000000")}, []BigDec{NewBigDec(1), NewBigDec(0), NewDecWithPrec(12340, 8)}, false},
		{[]sdk.Dec{sdk.MustNewDecFromStr("10.000000000000000000")}, []BigDec{NewBigDec(10)}, false},
		{[]sdk.Dec{sdk.MustNewDecFromStr("12340.000000000000000000")}, []BigDec{NewBigDec(12340)}, false},
		{[]sdk.Dec{sdk.MustNewDecFromStr("1.234000000000000000"), sdk.MustNewDecFromStr("12340.000000000000000000")}, []BigDec{NewDecWithPrec(12340, 4), NewBigDec(12340)}, false},
		{[]sdk.Dec{sdk.MustNewDecFromStr("0.123400000000000000"), sdk.MustNewDecFromStr("12340.000000000000000000")}, []BigDec{NewDecWithPrec(12340, 5), NewBigDec(12340)}, false},
		{[]sdk.Dec{sdk.MustNewDecFromStr("0.000123400000000000"), sdk.MustNewDecFromStr("10.090090090090090090")}, []BigDec{NewDecWithPrec(12340, 8), NewDecWithPrec(1009009009009009009, 17)}, false},
		{[]sdk.Dec{sdk.MustNewDecFromStr("10.090090090090090090"), sdk.MustNewDecFromStr("10.090090090090090090")}, []BigDec{NewDecWithPrec(1009009009009009009, 17), NewDecWithPrec(1009009009009009009, 17)}, false},
	}
	for tcIndex, tc := range tests {
		if tc.expPanic {
			s.Require().Panics(func() { BigDecFromSDKDecSlice(tc.d) })
		} else {
			value := BigDecFromSDKDecSlice(tc.d)
			s.Require().Equal(tc.want, value, "bad BigDecFromSdkDec(), index: %v", tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestEqualities() {
	tests := []struct {
		d1, d2     BigDec
		gt, lt, eq bool
	}{
		{NewBigDec(0), NewBigDec(0), false, false, true},
		{NewDecWithPrec(0, 2), NewDecWithPrec(0, 4), false, false, true},
		{NewDecWithPrec(100, 0), NewDecWithPrec(100, 0), false, false, true},
		{NewDecWithPrec(-100, 0), NewDecWithPrec(-100, 0), false, false, true},
		{NewDecWithPrec(-1, 1), NewDecWithPrec(-1, 1), false, false, true},
		{NewDecWithPrec(3333, 3), NewDecWithPrec(3333, 3), false, false, true},

		{NewDecWithPrec(0, 0), NewDecWithPrec(3333, 3), false, true, false},
		{NewDecWithPrec(0, 0), NewDecWithPrec(100, 0), false, true, false},
		{NewDecWithPrec(-1, 0), NewDecWithPrec(3333, 3), false, true, false},
		{NewDecWithPrec(-1, 0), NewDecWithPrec(100, 0), false, true, false},
		{NewDecWithPrec(1111, 3), NewDecWithPrec(100, 0), false, true, false},
		{NewDecWithPrec(1111, 3), NewDecWithPrec(3333, 3), false, true, false},
		{NewDecWithPrec(-3333, 3), NewDecWithPrec(-1111, 3), false, true, false},

		{NewDecWithPrec(3333, 3), NewDecWithPrec(0, 0), true, false, false},
		{NewDecWithPrec(100, 0), NewDecWithPrec(0, 0), true, false, false},
		{NewDecWithPrec(3333, 3), NewDecWithPrec(-1, 0), true, false, false},
		{NewDecWithPrec(100, 0), NewDecWithPrec(-1, 0), true, false, false},
		{NewDecWithPrec(100, 0), NewDecWithPrec(1111, 3), true, false, false},
		{NewDecWithPrec(3333, 3), NewDecWithPrec(1111, 3), true, false, false},
		{NewDecWithPrec(-1111, 3), NewDecWithPrec(-3333, 3), true, false, false},
	}

	for tcIndex, tc := range tests {
		s.Require().Equal(tc.gt, tc.d1.GT(tc.d2), "GT result is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.lt, tc.d1.LT(tc.d2), "LT result is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.eq, tc.d1.Equal(tc.d2), "equality result is incorrect, tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestDecsEqual() {
	tests := []struct {
		d1s, d2s []BigDec
		eq       bool
	}{
		{[]BigDec{NewBigDec(0)}, []BigDec{NewBigDec(0)}, true},
		{[]BigDec{NewBigDec(0)}, []BigDec{NewBigDec(1)}, false},
		{[]BigDec{NewBigDec(0)}, []BigDec{}, false},
		{[]BigDec{NewBigDec(0), NewBigDec(1)}, []BigDec{NewBigDec(0), NewBigDec(1)}, true},
		{[]BigDec{NewBigDec(1), NewBigDec(0)}, []BigDec{NewBigDec(1), NewBigDec(0)}, true},
		{[]BigDec{NewBigDec(1), NewBigDec(0)}, []BigDec{NewBigDec(0), NewBigDec(1)}, false},
		{[]BigDec{NewBigDec(1), NewBigDec(0)}, []BigDec{NewBigDec(1)}, false},
		{[]BigDec{NewBigDec(1), NewBigDec(2)}, []BigDec{NewBigDec(2), NewBigDec(4)}, false},
		{[]BigDec{NewBigDec(3), NewBigDec(18)}, []BigDec{NewBigDec(1), NewBigDec(6)}, false},
	}

	for tcIndex, tc := range tests {
		s.Require().Equal(tc.eq, DecsEqual(tc.d1s, tc.d2s), "equality of decional arrays is incorrect, tc %d", tcIndex)
		s.Require().Equal(tc.eq, DecsEqual(tc.d2s, tc.d1s), "equality of decional arrays is incorrect (converse), tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestArithmetic() {
	tests := []struct {
		d1, d2                                BigDec
		expMul, expMulTruncate                BigDec
		expQuo, expQuoRoundUp, expQuoTruncate BigDec
		expAdd, expSub                        BigDec
	}{
		//  d1         d2         MUL    MulTruncate    QUO    QUORoundUp QUOTrunctate  ADD         SUB
		{NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0)},
		{NewBigDec(1), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(1), NewBigDec(1)},
		{NewBigDec(0), NewBigDec(1), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(1), NewBigDec(-1)},
		{NewBigDec(0), NewBigDec(-1), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(-1), NewBigDec(1)},
		{NewBigDec(-1), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(0), NewBigDec(-1), NewBigDec(-1)},

		{NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(2), NewBigDec(0)},
		{NewBigDec(-1), NewBigDec(-1), NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(-2), NewBigDec(0)},
		{NewBigDec(1), NewBigDec(-1), NewBigDec(-1), NewBigDec(-1), NewBigDec(-1), NewBigDec(-1), NewBigDec(-1), NewBigDec(0), NewBigDec(2)},
		{NewBigDec(-1), NewBigDec(1), NewBigDec(-1), NewBigDec(-1), NewBigDec(-1), NewBigDec(-1), NewBigDec(-1), NewBigDec(0), NewBigDec(-2)},

		{
			NewBigDec(3), NewBigDec(7), NewBigDec(21), NewBigDec(21),
			MustNewDecFromStr("0.428571428571428571428571428571428571"), MustNewDecFromStr("0.428571428571428571428571428571428572"), MustNewDecFromStr("0.428571428571428571428571428571428571"),
			NewBigDec(10), NewBigDec(-4),
		},
		{
			NewBigDec(2), NewBigDec(4), NewBigDec(8), NewBigDec(8), NewDecWithPrec(5, 1), NewDecWithPrec(5, 1), NewDecWithPrec(5, 1),
			NewBigDec(6), NewBigDec(-2),
		},

		{NewBigDec(100), NewBigDec(100), NewBigDec(10000), NewBigDec(10000), NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(200), NewBigDec(0)},

		{
			NewDecWithPrec(15, 1), NewDecWithPrec(15, 1), NewDecWithPrec(225, 2), NewDecWithPrec(225, 2),
			NewBigDec(1), NewBigDec(1), NewBigDec(1), NewBigDec(3), NewBigDec(0),
		},
		{
			NewDecWithPrec(3333, 4), NewDecWithPrec(333, 4), NewDecWithPrec(1109889, 8), NewDecWithPrec(1109889, 8),
			MustNewDecFromStr("10.009009009009009009009009009009009009"), MustNewDecFromStr("10.009009009009009009009009009009009010"), MustNewDecFromStr("10.009009009009009009009009009009009009"),
			NewDecWithPrec(3666, 4), NewDecWithPrec(3, 1),
		},
	}

	for tcIndex, tc := range tests {
		tc := tc
		resAdd := tc.d1.Add(tc.d2)
		resSub := tc.d1.Sub(tc.d2)
		resMul := tc.d1.Mul(tc.d2)
		resMulTruncate := tc.d1.MulTruncate(tc.d2)
		s.Require().True(tc.expAdd.Equal(resAdd), "exp %v, res %v, tc %d", tc.expAdd, resAdd, tcIndex)
		s.Require().True(tc.expSub.Equal(resSub), "exp %v, res %v, tc %d", tc.expSub, resSub, tcIndex)
		s.Require().True(tc.expMul.Equal(resMul), "exp %v, res %v, tc %d", tc.expMul, resMul, tcIndex)
		s.Require().True(tc.expMulTruncate.Equal(resMulTruncate), "exp %v, res %v, tc %d", tc.expMulTruncate, resMulTruncate, tcIndex)

		if tc.d2.IsZero() { // panic for divide by zero
			s.Require().Panics(func() { tc.d1.Quo(tc.d2) })
		} else {
			resQuo := tc.d1.Quo(tc.d2)
			s.Require().True(tc.expQuo.Equal(resQuo), "exp %v, res %v, tc %d", tc.expQuo.String(), resQuo.String(), tcIndex)

			resQuoRoundUp := tc.d1.QuoRoundUp(tc.d2)
			s.Require().True(tc.expQuoRoundUp.Equal(resQuoRoundUp), "exp %v, res %v, tc %d",
				tc.expQuoRoundUp.String(), resQuoRoundUp.String(), tcIndex)

			resQuoTruncate := tc.d1.QuoTruncate(tc.d2)
			s.Require().True(tc.expQuoTruncate.Equal(resQuoTruncate), "exp %v, res %v, tc %d",
				tc.expQuoTruncate.String(), resQuoTruncate.String(), tcIndex)
		}
	}
}

func (s *decimalTestSuite) TestBankerRoundChop() {
	tests := []struct {
		d1  BigDec
		exp int64
	}{
		{s.mustNewDecFromStr("0.25"), 0},
		{s.mustNewDecFromStr("0"), 0},
		{s.mustNewDecFromStr("1"), 1},
		{s.mustNewDecFromStr("0.75"), 1},
		{s.mustNewDecFromStr("0.5"), 0},
		{s.mustNewDecFromStr("7.5"), 8},
		{s.mustNewDecFromStr("1.5"), 2},
		{s.mustNewDecFromStr("2.5"), 2},
		{s.mustNewDecFromStr("0.545"), 1}, // 0.545-> 1 even though 5 is first decimal and 1 not even
		{s.mustNewDecFromStr("1.545"), 2},
	}

	for tcIndex, tc := range tests {
		resNeg := tc.d1.Neg().RoundInt64()
		s.Require().Equal(-1*tc.exp, resNeg, "negative tc %d", tcIndex)

		resPos := tc.d1.RoundInt64()
		s.Require().Equal(tc.exp, resPos, "positive tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestTruncate() {
	tests := []struct {
		d1  BigDec
		exp int64
	}{
		{s.mustNewDecFromStr("0"), 0},
		{s.mustNewDecFromStr("0.25"), 0},
		{s.mustNewDecFromStr("0.75"), 0},
		{s.mustNewDecFromStr("1"), 1},
		{s.mustNewDecFromStr("1.5"), 1},
		{s.mustNewDecFromStr("7.5"), 7},
		{s.mustNewDecFromStr("7.6"), 7},
		{s.mustNewDecFromStr("7.4"), 7},
		{s.mustNewDecFromStr("100.1"), 100},
		{s.mustNewDecFromStr("1000.1"), 1000},
	}

	for tcIndex, tc := range tests {
		resNeg := tc.d1.Neg().TruncateInt64()
		s.Require().Equal(-1*tc.exp, resNeg, "negative tc %d", tcIndex)

		resPos := tc.d1.TruncateInt64()
		s.Require().Equal(tc.exp, resPos, "positive tc %d", tcIndex)
	}
}

func (s *decimalTestSuite) TestStringOverflow() {
	// two random 64 bit primes
	dec1, err := NewDecFromStr("51643150036226787134389711697696177267")
	s.Require().NoError(err)
	dec2, err := NewDecFromStr("-31798496660535729618459429845579852627")
	s.Require().NoError(err)
	dec3 := dec1.Add(dec2)
	s.Require().Equal(
		"19844653375691057515930281852116324640.000000000000000000000000000000000000",
		dec3.String(),
	)
}

func (s *decimalTestSuite) TestDecMulInt() {
	tests := []struct {
		sdkDec BigDec
		sdkInt BigInt
		want   BigDec
	}{
		{NewBigDec(10), NewInt(2), NewBigDec(20)},
		{NewBigDec(1000000), NewInt(100), NewBigDec(100000000)},
		{NewDecWithPrec(1, 1), NewInt(10), NewBigDec(1)},
		{NewDecWithPrec(1, 5), NewInt(20), NewDecWithPrec(2, 4)},
	}
	for i, tc := range tests {
		got := tc.sdkDec.MulInt(tc.sdkInt)
		s.Require().Equal(tc.want, got, "Incorrect result on test case %d", i)
	}
}

func (s *decimalTestSuite) TestDecCeil() {
	testCases := []struct {
		input    BigDec
		expected BigDec
	}{
		{MustNewDecFromStr("0.001"), NewBigDec(1)}, // 0.001 => 1.0
		{MustNewDecFromStr("-0.001"), ZeroDec()},   // -0.001 => 0.0
		{ZeroDec(), ZeroDec()}, // 0.0 => 0.0
		{MustNewDecFromStr("0.9"), NewBigDec(1)},    // 0.9 => 1.0
		{MustNewDecFromStr("4.001"), NewBigDec(5)},   // 4.001 => 5.0
		{MustNewDecFromStr("-4.001"), NewBigDec(-4)}, // -4.001 => -4.0
		{MustNewDecFromStr("4.7"), NewBigDec(5)},   // 4.7 => 5.0
		{MustNewDecFromStr("-4.7"), NewBigDec(-4)}, // -4.7 => -4.0
	}

	for i, tc := range testCases {
		res := tc.input.Ceil()
		s.Require().Equal(tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestPower() {
	testCases := []struct {
		input    BigDec
		power    uint64
		expected BigDec
	}{
		{OneDec(), 10, OneDec()},                                               // 1.0 ^ (10) => 1.0
		{NewDecWithPrec(5, 1), 2, NewDecWithPrec(25, 2)},                       // 0.5 ^ 2 => 0.25
		{NewDecWithPrec(2, 1), 2, NewDecWithPrec(4, 2)},                        // 0.2 ^ 2 => 0.04
		{NewDecFromInt(NewInt(3)), 3, NewDecFromInt(NewInt(27))},               // 3 ^ 3 => 27
		{NewDecFromInt(NewInt(-3)), 4, NewDecFromInt(NewInt(81))},              // -3 ^ 4 = 81
		{MustNewDecFromStr("1.414213562373095048801688724209698079"), 2, NewDecFromInt(NewInt(2))}, // 1.414213562373095048801688724209698079 ^ 2 = 2
	}

	for i, tc := range testCases {
		res := tc.input.Power(tc.power)
		s.Require().True(tc.expected.Sub(res).Abs().LTE(SmallestDec()), "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestApproxRoot() {
	testCases := []struct {
		input    BigDec
		root     uint64
		expected BigDec
	}{
		{OneDec(), 10, OneDec()},                                                       // 1.0 ^ (0.1) => 1.0
		{NewDecWithPrec(25, 2), 2, NewDecWithPrec(5, 1)},                               // 0.25 ^ (0.5) => 0.5
		{NewDecWithPrec(4, 2), 2, NewDecWithPrec(2, 1)},                                // 0.04 ^ (0.5) => 0.2
		{NewDecFromInt(NewInt(27)), 3, NewDecFromInt(NewInt(3))},                       // 27 ^ (1/3) => 3
		{NewDecFromInt(NewInt(-81)), 4, NewDecFromInt(NewInt(-3))},                     // -81 ^ (0.25) => -3
		{NewDecFromInt(NewInt(2)), 2, MustNewDecFromStr("1.414213562373095048801688724209698079")},         // 2 ^ (0.5) => 1.414213562373095048801688724209698079
		{NewDecWithPrec(1005, 3), 31536000, MustNewDecFromStr("1.000000000158153903837946258002096839")}, // 1.005 ^ (1/31536000) ≈ 1.000000000158153903837946258002096839
		{SmallestDec(), 2, NewDecWithPrec(1, 18)},                                       // 1e-36 ^ (0.5) => 1e-18
		{SmallestDec(), 3, MustNewDecFromStr("0.000000000001000000000000000002431786")},                  // 1e-36 ^ (1/3) => 1e-12
		{NewDecWithPrec(1, 8), 3, MustNewDecFromStr("0.002154434690031883721759293566519280")},           // 1e-8 ^ (1/3) ≈ 0.002154434690031883721759293566519
	}

	// In the case of 1e-8 ^ (1/3), the result repeats every 5 iterations starting from iteration 24
	// (i.e. 24, 29, 34, ... give the same result) and never converges enough. The maximum number of
	// iterations (100) causes the result at iteration 100 to be returned, regardless of convergence.

	for i, tc := range testCases {
		res, err := tc.input.ApproxRoot(tc.root)
		s.Require().NoError(err)
		s.Require().True(tc.expected.Sub(res).Abs().LTE(SmallestDec()), "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestApproxSqrt() {
	testCases := []struct {
		input    BigDec
		expected BigDec
	}{
		{OneDec(), OneDec()},                                                // 1.0 => 1.0
		{NewDecWithPrec(25, 2), NewDecWithPrec(5, 1)},                       // 0.25 => 0.5
		{NewDecWithPrec(4, 2), NewDecWithPrec(2, 1)},                        // 0.09 => 0.3
		{NewDecFromInt(NewInt(9)), NewDecFromInt(NewInt(3))},                // 9 => 3
		{NewDecFromInt(NewInt(-9)), NewDecFromInt(NewInt(-3))},              // -9 => -3
		{NewDecFromInt(NewInt(2)), MustNewDecFromStr("1.414213562373095048801688724209698079")}, // 2 => 1.414213562373095048801688724209698079
	}

	for i, tc := range testCases {
		res, err := tc.input.ApproxSqrt()
		s.Require().NoError(err)
		s.Require().Equal(tc.expected, res, "unexpected result for test case %d, input: %v", i, tc.input)
	}
}

func (s *decimalTestSuite) TestDecSortableBytes() {
	tests := []struct {
		d    BigDec
		want []byte
	}{
		{NewBigDec(0), []byte("000000000000000000000000000000000000.000000000000000000000000000000000000")},
		{NewBigDec(1), []byte("000000000000000000000000000000000001.000000000000000000000000000000000000")},
		{NewBigDec(10), []byte("000000000000000000000000000000000010.000000000000000000000000000000000000")},
		{NewBigDec(12340), []byte("000000000000000000000000000000012340.000000000000000000000000000000000000")},
		{NewDecWithPrec(12340, 4), []byte("000000000000000000000000000000000001.234000000000000000000000000000000000")},
		{NewDecWithPrec(12340, 5), []byte("000000000000000000000000000000000000.123400000000000000000000000000000000")},
		{NewDecWithPrec(12340, 8), []byte("000000000000000000000000000000000000.000123400000000000000000000000000000")},
		{NewDecWithPrec(1009009009009009009, 17), []byte("000000000000000000000000000000000010.090090090090090090000000000000000000")},
		{NewDecWithPrec(-1009009009009009009, 17), []byte("-000000000000000000000000000000000010.090090090090090090000000000000000000")},
		{MustNewDecFromStr("1000000000000000000000000000000000000"), []byte("max")},
		{MustNewDecFromStr("-1000000000000000000000000000000000000"), []byte("--")},
	}
	for tcIndex, tc := range tests {
		s.Require().Equal(tc.want, SortableDecBytes(tc.d), "bad String(), index: %v", tcIndex)
	}

	s.Require().Panics(func() { SortableDecBytes(MustNewDecFromStr("1000000000000000000000000000000000001")) })
	s.Require().Panics(func() { SortableDecBytes(MustNewDecFromStr("-1000000000000000000000000000000000001")) })
}

func (s *decimalTestSuite) TestDecEncoding() {
	testCases := []struct {
		input   BigDec
		rawBz   string
		jsonStr string
		yamlStr string
	}{
		{
			NewBigDec(0), "30",
			"\"0.000000000000000000000000000000000000\"",
			"\"0.000000000000000000000000000000000000\"\n",
		},
		{
			NewDecWithPrec(4, 2),
			"3430303030303030303030303030303030303030303030303030303030303030303030",
			"\"0.040000000000000000000000000000000000\"",
			"\"0.040000000000000000000000000000000000\"\n",
		},
		{
			NewDecWithPrec(-4, 2),
			"2D3430303030303030303030303030303030303030303030303030303030303030303030",
			"\"-0.040000000000000000000000000000000000\"",
			"\"-0.040000000000000000000000000000000000\"\n",
		},
		{
			MustNewDecFromStr("1.414213562373095048801688724209698079"),
			"31343134323133353632333733303935303438383031363838373234323039363938303739",
			"\"1.414213562373095048801688724209698079\"",
			"\"1.414213562373095048801688724209698079\"\n",
		},
		{
			MustNewDecFromStr("-1.414213562373095048801688724209698079"),
			"2D31343134323133353632333733303935303438383031363838373234323039363938303739",
			"\"-1.414213562373095048801688724209698079\"",
			"\"-1.414213562373095048801688724209698079\"\n",
		},
	}

	for _, tc := range testCases {
		bz, err := tc.input.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(tc.rawBz, fmt.Sprintf("%X", bz))

		var other BigDec
		s.Require().NoError((&other).Unmarshal(bz))
		s.Require().True(tc.input.Equal(other))

		bz, err = json.Marshal(tc.input)
		s.Require().NoError(err)
		s.Require().Equal(tc.jsonStr, string(bz))
		s.Require().NoError(json.Unmarshal(bz, &other))
		s.Require().True(tc.input.Equal(other))

		bz, err = yaml.Marshal(tc.input)
		s.Require().NoError(err)
		s.Require().Equal(tc.yamlStr, string(bz))
	}
}

// Showcase that different orders of operations causes different results.
func (s *decimalTestSuite) TestOperationOrders() {
	n1 := NewBigDec(10)
	n2 := NewBigDec(1000000010)
	s.Require().Equal(n1.Mul(n2).Quo(n2), NewBigDec(10))
	s.Require().NotEqual(n1.Mul(n2).Quo(n2), n1.Quo(n2).Mul(n2))
}

func BenchmarkMarshalTo(b *testing.B) {
	b.ReportAllocs()
	bis := []struct {
		in   BigDec
		want []byte
	}{
		{
			NewBigDec(1e8), []byte{
				0x31, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
				0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30,
			},
		},
		{NewBigDec(0), []byte{0x30}},
	}
	data := make([]byte, 100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, bi := range bis {
			if n, err := bi.in.MarshalTo(data); err != nil {
				b.Fatal(err)
			} else {
				if !bytes.Equal(data[:n], bi.want) {
					b.Fatalf("Mismatch\nGot:  % x\nWant: % x\n", data[:n], bi.want)
				}
			}
		}
	}
}
