package wasm

import (
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app/wasm"
	wasmbindings "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDenom(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

	// Fund actor with 100 base denom creation fees
	actorAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, actor, actorAmount)

	specs := map[string]struct {
		createDenom *wasmbindings.CreateDenom
		expErr      bool
	}{
		"valid sub-denom": {
			createDenom: &wasmbindings.CreateDenom{
				Subdenom: "MOON",
			},
		},
		"empty sub-denom": {
			createDenom: &wasmbindings.CreateDenom{
				Subdenom: "",
			},
			expErr: false,
		},
		"invalid sub-denom": {
			createDenom: &wasmbindings.CreateDenom{
				Subdenom: "sub-denom_2",
			},
			expErr: true,
		},
		"null create denom": {
			createDenom: nil,
			expErr:      true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotErr := wasm.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, actor, spec.createDenom)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}

}

func TestMint(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

	// Fund actor with 100 base denom creation fees
	actorAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, actor, actorAmount)

	// Create denoms for valid mint tests
	validDenom := wasmbindings.CreateDenom{
		Subdenom: "MOON",
	}
	err := wasm.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, actor, &validDenom)
	require.NoError(t, err)

	emptyDenom := wasmbindings.CreateDenom{
		Subdenom: "",
	}
	err = wasm.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, actor, &emptyDenom)
	require.NoError(t, err)

	lucky := RandomAccountAddress()

	// lucky was broke
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	amount, ok := sdk.NewIntFromString("8080")
	require.True(t, ok)

	specs := map[string]struct {
		mint   *wasmbindings.MintTokens
		expErr bool
	}{
		"valid mint": {
			mint: &wasmbindings.MintTokens{
				Subdenom:  "MOON",
				Amount:    amount,
				Recipient: lucky.String(),
			},
		},
		"empty sub-denom": {
			mint: &wasmbindings.MintTokens{
				Subdenom:  "",
				Amount:    amount,
				Recipient: lucky.String(),
			},
			expErr: false,
		},
		"nonexistent sub-denom": {
			mint: &wasmbindings.MintTokens{
				Subdenom:  "SUN",
				Amount:    amount,
				Recipient: lucky.String(),
			},
			expErr: true,
		},
		"invalid sub-denom": {
			mint: &wasmbindings.MintTokens{
				Subdenom:  "sub-denom_2",
				Amount:    amount,
				Recipient: lucky.String(),
			},
			expErr: true,
		},
		"zero amount": {
			mint: &wasmbindings.MintTokens{
				Subdenom:  "MOON",
				Amount:    sdk.ZeroInt(),
				Recipient: lucky.String(),
			},
			expErr: true,
		},
		"negative amount": {
			mint: &wasmbindings.MintTokens{
				Subdenom:  "MOON",
				Amount:    amount.Neg(),
				Recipient: lucky.String(),
			},
			expErr: true,
		},
		"empty recipient": {
			mint: &wasmbindings.MintTokens{
				Subdenom:  "MOON",
				Amount:    amount,
				Recipient: "",
			},
			expErr: true,
		},
		"invalid recipient": {
			mint: &wasmbindings.MintTokens{
				Subdenom:  "MOON",
				Amount:    amount,
				Recipient: "invalid",
			},
			expErr: true,
		},
		"null mint": {
			mint:   nil,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotErr := wasm.PerformMint(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, actor, spec.mint)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}

}

func TestSwap(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)
	epsilon := 1e-3

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12_000_000),
		sdk.NewInt64Coin("ustar", 240_000_000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	// Estimate swap rate
	uosmo := poolFunds[0].Amount.ToDec().MustFloat64()
	ustar := poolFunds[1].Amount.ToDec().MustFloat64()
	swapRate := ustar / uosmo

	amountIn := wasmbindings.ExactIn{
		Input:     sdk.NewInt(10000),
		MinOutput: sdk.OneInt(),
	}
	zeroAmountIn := amountIn
	zeroAmountIn.Input = sdk.ZeroInt()
	negativeAmountIn := amountIn
	negativeAmountIn.Input = negativeAmountIn.Input.Neg()

	amountOut := wasmbindings.ExactOut{
		MaxInput: sdk.NewInt(math.MaxInt64),
		Output:   sdk.NewInt(10000),
	}
	zeroAmountOut := amountOut
	zeroAmountOut.Output = sdk.ZeroInt()
	negativeAmountOut := amountOut
	negativeAmountOut.Output = negativeAmountOut.Output.Neg()

	amount := amountIn.Input.ToDec().MustFloat64()
	starAmount := sdk.NewInt(int64(amount * swapRate))

	starSwapAmount := wasmbindings.SwapAmount{Out: &starAmount}

	specs := map[string]struct {
		swap    *wasmbindings.SwapMsg
		expCost *wasmbindings.SwapAmount
		expErr  bool
	}{
		"valid swap (exact in)": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expCost: &starSwapAmount,
		},
		"non-existent pool id": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool + 4,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"zero pool id": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   0,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid denom in": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "invalid",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"empty denom in": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid denom out": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "invalid",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"empty denom out": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"null swap": {
			swap:   nil,
			expErr: true,
		},
		"empty swap amount": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "",
				},
				Route:  nil,
				Amount: wasmbindings.SwapAmountWithLimit{},
			},
			expErr: true,
		},
		"zero amount in": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &zeroAmountIn,
				},
			},
			expErr: true,
		},
		"zero amount out": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactOut: &zeroAmountOut,
				},
			},
			expErr: true,
		},
		"negative amount in": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &negativeAmountIn,
				},
			},
			expErr: true,
		},
		"negative amount out": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactOut: &negativeAmountOut,
				},
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotAmount, gotErr := wasm.PerformSwap(osmosis.GAMMKeeper, ctx, actor, spec.swap)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.InEpsilonf(t, (*spec.expCost.Out).ToDec().MustFloat64(), (*gotAmount.Out).ToDec().MustFloat64(), epsilon, "exp %s but got %s", spec.expCost.Out.String(), gotAmount.Out.String())
		})
	}
}

func TestSwapMultiHop(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)
	epsilon := 1e-3

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12_000_000),
		sdk.NewInt64Coin("ustar", 240_000_000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	// 2 osmo to 1 atom
	poolFunds2 := []sdk.Coin{
		sdk.NewInt64Coin("uatom", 6_000_000),
		sdk.NewInt64Coin("uosmo", 12_000_000),
	}
	atomPool := preparePool(t, ctx, osmosis, actor, poolFunds2)

	amountIn := wasmbindings.ExactIn{
		Input:     sdk.NewInt(1_000_000),
		MinOutput: sdk.NewInt(20_000),
	}

	// Multi-hop
	// Estimate 1st swap rate
	uosmo := poolFunds[0].Amount.ToDec().MustFloat64()
	ustar := poolFunds[1].Amount.ToDec().MustFloat64()
	expectedOut1 := uosmo - uosmo*ustar/(ustar+amountIn.Input.ToDec().MustFloat64())

	// Estimate 2nd swap rate
	uatom2 := poolFunds2[0].Amount.ToDec().MustFloat64()
	uosmo2 := poolFunds2[1].Amount.ToDec().MustFloat64()
	expectedOut2 := uatom2 - uosmo2*uatom2/(uosmo2+expectedOut1)

	atomAmount := sdk.NewInt(int64(expectedOut2))
	atomSwapAmount := wasmbindings.SwapAmount{Out: &atomAmount}

	specs := map[string]struct {
		swap    *wasmbindings.SwapMsg
		expCost *wasmbindings.SwapAmount
		expErr  bool
	}{
		"valid swap (exact in, 2 step multi-hop)": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []wasmbindings.Step{{
					PoolId:   atomPool,
					DenomOut: "uatom",
				}},
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expCost: &atomSwapAmount,
		},
		"non-existent step pool id": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []wasmbindings.Step{{
					PoolId:   atomPool + 2,
					DenomOut: "uatom",
				}},
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"zero step pool id": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []wasmbindings.Step{{
					PoolId:   0,
					DenomOut: "uatom",
				}},
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"wrong step denom out": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []wasmbindings.Step{{
					PoolId:   atomPool,
					DenomOut: "ATOM",
				}},
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"self-swap not allowed": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []wasmbindings.Step{{
					PoolId:   atomPool,
					DenomOut: "uosmo", // this is same as the input (output of first swap)
				}},
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid step denom out": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []wasmbindings.Step{{
					PoolId:   atomPool,
					DenomOut: "invalid",
				}},
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"empty step denom out": {
			swap: &wasmbindings.SwapMsg{
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []wasmbindings.Step{{
					PoolId:   atomPool,
					DenomOut: "",
				}},
				Amount: wasmbindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// use scratch context to avoid interference between tests
			subCtx, _ := ctx.CacheContext()
			// when
			gotAmount, gotErr := wasm.PerformSwap(osmosis.GAMMKeeper, subCtx, actor, spec.swap)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.InEpsilonf(t, (*spec.expCost.Out).ToDec().MustFloat64(), (*gotAmount.Out).ToDec().MustFloat64(), epsilon, "exp %s but got %s", spec.expCost.Out.String(), gotAmount.Out.String())
		})
	}
}
