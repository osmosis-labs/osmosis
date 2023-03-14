package osmoutils_test

import (
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
)

var expectedOutOfGasError = types.ErrorOutOfGas{Descriptor: "my func"}

func consumeGas(ctx sdk.Context, gas uint64, numTimes int) error {
	for i := 0; i < numTimes; i++ {
		ctx.GasMeter().ConsumeGas(gas, "my func")
	}
	return nil
}

func (s *TestSuite) TestCacheCtxConsumeGas() {
	// every test case adds 1k gas 10 times
	testcases := map[string]struct {
		gasLimit       uint64
		gasUsedPreCtx  uint64
		gasUsedPostCtx uint64
		expectPanic    bool
	}{
		"no gas limit hit": {
			gasLimit:       15_000,
			gasUsedPreCtx:  111,
			gasUsedPostCtx: 111 + 10_000,
			expectPanic:    false,
		},
		"gas limit hit": {
			gasLimit:       10_000,
			gasUsedPreCtx:  111,
			gasUsedPostCtx: 111 + 10_000,
			expectPanic:    true,
		},
	}
	for name, tc := range testcases {
		s.Run(name, func() {
			ctx := s.ctx.WithGasMeter(sdk.NewGasMeter(tc.gasLimit))
			ctx.GasMeter().ConsumeGas(tc.gasUsedPreCtx, "pre ctx")
			var err error
			f := func() {
				osmoutils.ApplyFuncIfNoError(ctx, func(c sdk.Context) error {
					return consumeGas(c, 1000, 10)
				})
			}
			if tc.expectPanic {
				s.PanicsWithValue(expectedOutOfGasError, f)
			} else {
				f()
				s.Require().NoError(err)
			}
			s.Equal(tc.gasUsedPostCtx, ctx.GasMeter().GasConsumed())
		})
	}
}
