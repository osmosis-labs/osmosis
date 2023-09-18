package osmomath_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
)

var (
	// minDecTolerance minimum tolerance for Dec, given its precision of 18.
	minDecTolerance = osmomath.MustNewDecFromStr("0.000000000000000001")
)

func TestExp2ChebyshevRationalApprox(t *testing.T) {
	// These values are used to test the approximated results close
	// to 0 and 1 boundaries.
	// With other types of approximations, there is a high likelyhood
	// of larger errors clsoer to the boundaries. This is known as Runge's phenomenon.
	// https://en.wikipedia.org/wiki/Runge%27s_phenomenon
	//
	// Chebyshev approximation should be able to handle this better.
	// Tests at the boundaries help to validate there is no Runge's phenomenon.
	smallValue := osmomath.MustNewBigDecFromStr("0.00001")
	smallerValue := osmomath.MustNewBigDecFromStr("0.00000000000000000001")

	tests := map[string]struct {
		exponent       osmomath.BigDec
		expectedResult osmomath.BigDec
		errTolerance   osmomath.ErrTolerance
		expectPanic    bool
	}{
		"exp2(0.5)": {
			exponent: osmomath.MustNewBigDecFromStr("0.5"),
			// https://www.wolframalpha.com/input?i=2%5E0.5+37+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1.414213562373095048801688724209698079"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       minDecTolerance,
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundDown,
			},
		},
		"exp2(0)": {
			exponent:       osmomath.ZeroBigDec(),
			expectedResult: osmomath.OneBigDec(),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.ZeroDec(),
				MultiplicativeTolerance: osmomath.ZeroDec(),
				RoundingDir:             osmomath.RoundDown,
			},
		},
		"exp2(1)": {
			exponent:       osmomath.OneBigDec(),
			expectedResult: osmomath.MustNewBigDecFromStr("2"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.ZeroDec(),
				MultiplicativeTolerance: osmomath.ZeroDec(),
				RoundingDir:             osmomath.RoundDown,
			},
		},
		"exp2(0.00001)": {
			exponent: smallValue,
			// https://www.wolframalpha.com/input?i=2%5E0.00001+37+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1.000006931495828305653209089800561681"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       minDecTolerance,
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(0.99999)": {
			exponent: osmomath.OneBigDec().Sub(smallValue),
			// https://www.wolframalpha.com/input?i=2%5E0.99999+37+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1.999986137104433991477606830496602898"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("0.00000000000000007"),
				MultiplicativeTolerance: minDecTolerance.Mul(osmomath.NewDec(100)),
				RoundingDir:             osmomath.RoundDown,
			},
		},
		"exp2(0.99999...)": {
			exponent: osmomath.OneBigDec().Sub(smallerValue),
			// https://www.wolframalpha.com/input?i=2%5E%281+-+0.00000000000000000001%29+37+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1.999999999999999999986137056388801094"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       minDecTolerance,
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundDown,
			},
		},
		"exp2(0.0000...1)": {
			exponent: osmomath.ZeroBigDec().Add(smallerValue),
			// https://www.wolframalpha.com/input?i=2%5E0.00000000000000000001+37+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1.000000000000000000006931471805599453"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       minDecTolerance,
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(0.3334567)": {
			exponent: osmomath.MustNewBigDecFromStr("0.3334567"),
			// https://www.wolframalpha.com/input?i=2%5E0.3334567+37+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1.260028791934303989065848870753742298"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("0.00000000000000007"),
				MultiplicativeTolerance: minDecTolerance.Mul(osmomath.NewDec(10)),
				RoundingDir:             osmomath.RoundDown,
			},
		},
		"exp2(0.84864288)": {
			exponent: osmomath.MustNewBigDecFromStr("0.84864288"),
			// https://www.wolframalpha.com/input?i=2%5E0.84864288+37+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1.800806138872630518880998772777747572"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("0.00000000000000002"),
				MultiplicativeTolerance: minDecTolerance.Mul(osmomath.NewDec(10)),
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(0.999999999999999999999999999999999956)": {
			exponent: osmomath.MustNewBigDecFromStr("0.999999999999999999999999999999999956"),
			// https://www.wolframalpha.com/input?i=2%5E0.999999999999999999999999999999999956+37+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1.999999999999999999999999999999999939"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       minDecTolerance,
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundDown,
			},
		},
		// out of bounds.
		"exponent < 0 - panic": {
			exponent:    osmomath.ZeroBigDec().Sub(smallValue),
			expectPanic: true,
		},
		"exponent > 1 - panic": {
			exponent:    osmomath.OneBigDec().Add(smallValue),
			expectPanic: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			osmomath.ConditionalPanic(t, tc.expectPanic, func() {
				// System under test.
				result := osmomath.Exp2ChebyshevRationalApprox(tc.exponent)

				// Reuse the same test cases for exp2 that is a wrapper around Exp2ChebyshevRationalApprox.
				// This is done to reduce boilerplate from duplicating test cases.
				resultExp2 := osmomath.Exp2(tc.exponent)
				require.Equal(t, result, resultExp2)

				osmoassert.Equal(t, tc.errTolerance, tc.expectedResult, result)
			})
		})
	}
}

func TestExp2(t *testing.T) {
	tests := map[string]struct {
		exponent       osmomath.BigDec
		expectedResult osmomath.BigDec
		errTolerance   osmomath.ErrTolerance
		expectPanic    bool
	}{
		"exp2(28.5)": {
			exponent: osmomath.MustNewBigDecFromStr("28.5"),
			// https://www.wolframalpha.com/input?i=2%5E%2828.5%29+45+digits
			expectedResult: osmomath.MustNewBigDecFromStr("379625062.497006211556423566253288543343173698"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       minDecTolerance,
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(63.84864288)": {
			exponent: osmomath.MustNewBigDecFromStr("63.84864288"),
			// https://www.wolframalpha.com/input?i=2%5E%2863.84864288%29+56+digits
			expectedResult: osmomath.MustNewBigDecFromStr("16609504985074238416.013387053450559984846024066925604094"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("0.00042"),
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(64.5)": {
			exponent: osmomath.MustNewBigDecFromStr("64.5"),
			// https://www.wolframalpha.com/input?i=2%5E%2864.5%29+56+digits
			expectedResult: osmomath.MustNewBigDecFromStr("26087635650665564424.699143612505016737766552579185717157"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("0.000000000000000008"),
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(80.5)": {
			exponent: osmomath.MustNewBigDecFromStr("80.5"),
			// https://www.wolframalpha.com/input?i=2%5E%2880.5%29+61+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1709679290002018430137083.075789128776926268789829515159631571"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("0.0000000000006"),
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(100.5)": {
			exponent: osmomath.MustNewBigDecFromStr("100.5"),
			// https://www.wolframalpha.com/input?i=2%5E%28100.5%29+67+digits
			expectedResult: osmomath.MustNewBigDecFromStr("1792728671193156477399422023278.661496394239222564273688025833797661"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("0.0000006"),
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(128.5)": {
			exponent: osmomath.MustNewBigDecFromStr("128.5"),
			// https://www.wolframalpha.com/input?i=2%5E%28128.5%29+75+digits
			expectedResult: osmomath.MustNewBigDecFromStr("481231938336009023090067544955250113854.229961482126296754016435255422777776"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("146.5"),
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"exp2(127.999999999999999999999999999999999999)": {
			exponent: osmomath.MustNewBigDecFromStr("127.999999999999999999999999999999999999"),
			// https://www.wolframalpha.com/input?i=2%5E%28127.999999999999999999999999999999999999%29+75+digits
			expectedResult: osmomath.MustNewBigDecFromStr("340282366920938463463374607431768211220.134236774486705862055857235845515682"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("15044647266406936"),
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundDown,
			},
		},
		"exp2(127.84864288)": {
			exponent: osmomath.MustNewBigDecFromStr("127.84864288"),
			// https://www.wolframalpha.com/input?i=2%5E%28127.84864288%29+75+digits
			expectedResult: osmomath.MustNewBigDecFromStr("306391287650667462068703337664945630660.398687487527674545778353588077174571"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       osmomath.MustNewDecFromStr("7707157415597963"),
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundUnconstrained,
			},
		},
		"panic, too large - positive": {
			exponent:    osmomath.MaxSupportedExponent.Add(osmomath.OneBigDec()),
			expectPanic: true,
		},
		"panic - negative exponent": {
			exponent:    osmomath.OneBigDec().Neg(),
			expectPanic: true,
		},
		"at exponent boundary - positive": {
			exponent: osmomath.MaxSupportedExponent,
			// https://www.wolframalpha.com/input?i=2%5E%282%5E9%29
			expectedResult: osmomath.MustNewBigDecFromStr("13407807929942597099574024998205846127479365820592393377723561443721764030073546976801874298166903427690031858186486050853753882811946569946433649006084096"),

			errTolerance: osmomath.ErrTolerance{
				AdditiveTolerance:       minDecTolerance,
				MultiplicativeTolerance: minDecTolerance,
				RoundingDir:             osmomath.RoundDown,
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			osmomath.ConditionalPanic(t, tc.expectPanic, func() {

				// System under test.
				result := osmomath.Exp2(tc.exponent)

				osmoassert.Equal(t, tc.errTolerance, tc.expectedResult, result)
			})
		})
	}
}
