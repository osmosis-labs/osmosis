package wasmbinding

import (
	"errors"
	"fmt"
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/wasmbinding"
	"github.com/osmosis-labs/osmosis/v7/wasmbinding/bindings"
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
		createDenom *bindings.CreateDenom
		expErr      bool
	}{
		"valid sub-denom": {
			createDenom: &bindings.CreateDenom{
				Subdenom: "MOON",
			},
		},
		"empty sub-denom": {
			createDenom: &bindings.CreateDenom{
				Subdenom: "",
			},
			expErr: false,
		},
		"invalid sub-denom": {
			createDenom: &bindings.CreateDenom{
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
			gotErr := wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, actor, spec.createDenom)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestChangeAdmin(t *testing.T) {
	const validDenom = "validdenom"

	tokenCreator := RandomAccountAddress()

	specs := map[string]struct {
		actor       sdk.AccAddress
		changeAdmin *bindings.ChangeAdmin

		expErrMsg string
	}{
		"valid": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor: tokenCreator,
		},
		"typo in factory in denom name": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("facory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "denom prefix is incorrect. Is: facory.  Should be: factory: invalid denom",
		},
		"invalid address in denom": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", RandomBech32AccountAddress(), validDenom),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "failed changing admin from message: unauthorized account",
		},
		"other denom name in 3 part name": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), "invalid denom"),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: fmt.Sprintf("invalid denom: factory/%s/invalid denom", tokenCreator.String()),
		},
		"empty denom": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           "",
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     tokenCreator,
			expErrMsg: "invalid denom: ",
		},
		"empty address": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: "",
			},
			actor:     tokenCreator,
			expErrMsg: "address from bech32: empty address string is not allowed",
		},
		"creator is a different address": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: RandomBech32AccountAddress(),
			},
			actor:     RandomAccountAddress(),
			expErrMsg: "failed changing admin from message: unauthorized account",
		},
		"change to the same address": {
			changeAdmin: &bindings.ChangeAdmin{
				Denom:           fmt.Sprintf("factory/%s/%s", tokenCreator.String(), validDenom),
				NewAdminAddress: tokenCreator.String(),
			},
			actor: tokenCreator,
		},
		"nil binding": {
			actor:     tokenCreator,
			expErrMsg: "invalid request: changeAdmin is nil - original request: ",
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// Setup
			osmosis, ctx := SetupCustomApp(t, tokenCreator)

			// Fund actor with 100 base denom creation fees
			actorAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
			fundAccount(t, ctx, osmosis, tokenCreator, actorAmount)

			err := wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, tokenCreator, &bindings.CreateDenom{
				Subdenom: validDenom,
			})
			require.NoError(t, err)

			err = wasmbinding.ChangeAdmin(osmosis.TokenFactoryKeeper, ctx, spec.actor, spec.changeAdmin)
			if len(spec.expErrMsg) > 0 {
				require.Error(t, err)
				actualErrMsg := err.Error()
				require.Equal(t, spec.expErrMsg, actualErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMint(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	// Fund actor with 100 base denom creation fees
	tokenCreationFeeAmt := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, creator, tokenCreationFeeAmt)

	// Create denoms for valid mint tests
	validDenom := bindings.CreateDenom{
		Subdenom: "MOON",
	}
	err := wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, &validDenom)
	require.NoError(t, err)

	emptyDenom := bindings.CreateDenom{
		Subdenom: "",
	}
	err = wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, &emptyDenom)
	require.NoError(t, err)

	validDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), validDenom.Subdenom)
	emptyDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), emptyDenom.Subdenom)

	lucky := RandomAccountAddress()

	// lucky was broke
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	amount, ok := sdk.NewIntFromString("8080")
	require.True(t, ok)

	specs := map[string]struct {
		mint   *bindings.MintTokens
		expErr bool
	}{
		"valid mint": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
		},
		"empty sub-denom": {
			mint: &bindings.MintTokens{
				Denom:         emptyDenomStr,
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: false,
		},
		"nonexistent sub-denom": {
			mint: &bindings.MintTokens{
				Denom:         fmt.Sprintf("factory/%s/%s", creator.String(), "SUN"),
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"invalid sub-denom": {
			mint: &bindings.MintTokens{
				Denom:         "sub-denom_2",
				Amount:        amount,
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"zero amount": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        sdk.ZeroInt(),
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"negative amount": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount.Neg(),
				MintToAddress: lucky.String(),
			},
			expErr: true,
		},
		"empty recipient": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: "",
			},
			expErr: true,
		},
		"invalid recipient": {
			mint: &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        amount,
				MintToAddress: "invalid",
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
			gotErr := wasmbinding.PerformMint(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, spec.mint)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestBurn(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	// Fund actor with 100 base denom creation fees
	tokenCreationFeeAmt := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, creator, tokenCreationFeeAmt)

	// Create denoms for valid burn tests
	validDenom := bindings.CreateDenom{
		Subdenom: "MOON",
	}
	err := wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, &validDenom)
	require.NoError(t, err)

	emptyDenom := bindings.CreateDenom{
		Subdenom: "",
	}
	err = wasmbinding.PerformCreateDenom(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, &emptyDenom)
	require.NoError(t, err)

	lucky := RandomAccountAddress()

	// lucky was broke
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	validDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), validDenom.Subdenom)
	emptyDenomStr := fmt.Sprintf("factory/%s/%s", creator.String(), emptyDenom.Subdenom)
	mintAmount, ok := sdk.NewIntFromString("8080")
	require.True(t, ok)

	specs := map[string]struct {
		burn   *bindings.BurnTokens
		expErr bool
	}{
		"valid burn": {
			burn: &bindings.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: false,
		},
		"non admin address": {
			burn: &bindings.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: lucky.String(),
			},
			expErr: true,
		},
		"empty sub-denom": {
			burn: &bindings.BurnTokens{
				Denom:           emptyDenomStr,
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: false,
		},
		"invalid sub-denom": {
			burn: &bindings.BurnTokens{
				Denom:           "sub-denom_2",
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"non-minted denom": {
			burn: &bindings.BurnTokens{
				Denom:           fmt.Sprintf("factory/%s/%s", creator.String(), "SUN"),
				Amount:          mintAmount,
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"zero amount": {
			burn: &bindings.BurnTokens{
				Denom:           validDenomStr,
				Amount:          sdk.ZeroInt(),
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
		"negative amount": {
			burn:   nil,
			expErr: true,
		},
		"null burn": {
			burn: &bindings.BurnTokens{
				Denom:           validDenomStr,
				Amount:          mintAmount.Neg(),
				BurnFromAddress: creator.String(),
			},
			expErr: true,
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// Mint valid denom str and empty denom string for burn test
			mintBinding := &bindings.MintTokens{
				Denom:         validDenomStr,
				Amount:        mintAmount,
				MintToAddress: creator.String(),
			}
			err := wasmbinding.PerformMint(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, mintBinding)
			require.NoError(t, err)

			emptyDenomMintBinding := &bindings.MintTokens{
				Denom:         emptyDenomStr,
				Amount:        mintAmount,
				MintToAddress: creator.String(),
			}
			err = wasmbinding.PerformMint(osmosis.TokenFactoryKeeper, osmosis.BankKeeper, ctx, creator, emptyDenomMintBinding)
			require.NoError(t, err)

			// when
			gotErr := wasmbinding.PerformBurn(osmosis.TokenFactoryKeeper, ctx, creator, spec.burn)
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

	amountIn := bindings.ExactIn{
		Input:     sdk.NewInt(10000),
		MinOutput: sdk.OneInt(),
	}
	zeroAmountIn := amountIn
	zeroAmountIn.Input = sdk.ZeroInt()
	negativeAmountIn := amountIn
	negativeAmountIn.Input = negativeAmountIn.Input.Neg()

	amountOut := bindings.ExactOut{
		MaxInput: sdk.NewInt(math.MaxInt64),
		Output:   sdk.NewInt(10000),
	}
	zeroAmountOut := amountOut
	zeroAmountOut.Output = sdk.ZeroInt()
	negativeAmountOut := amountOut
	negativeAmountOut.Output = negativeAmountOut.Output.Neg()

	amount := amountIn.Input.ToDec().MustFloat64()
	starAmount := sdk.NewInt(int64(amount * swapRate))

	starSwapAmount := bindings.SwapAmount{Out: &starAmount}

	specs := map[string]struct {
		swap    *bindings.SwapMsg
		expCost *bindings.SwapAmount
		expErr  bool
	}{
		"valid swap (exact in)": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expCost: &starSwapAmount,
		},
		"non-existent pool id": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool + 4,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"zero pool id": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   0,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid denom in": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "invalid",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"empty denom in": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid denom out": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "invalid",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"empty denom out": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
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
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "",
				},
				Route:  nil,
				Amount: bindings.SwapAmountWithLimit{},
			},
			expErr: true,
		},
		"zero amount in": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &zeroAmountIn,
				},
			},
			expErr: true,
		},
		"zero amount out": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactOut: &zeroAmountOut,
				},
			},
			expErr: true,
		},
		"negative amount in": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &negativeAmountIn,
				},
			},
			expErr: true,
		},
		"negative amount out": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmountWithLimit{
					ExactOut: &negativeAmountOut,
				},
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotAmount, gotErr := wasmbinding.PerformSwap(osmosis.GAMMKeeper, ctx, actor, spec.swap)
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

	amountIn := bindings.ExactIn{
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
	atomSwapAmount := bindings.SwapAmount{Out: &atomAmount}

	specs := map[string]struct {
		swap    *bindings.SwapMsg
		expCost *bindings.SwapAmount
		expErr  bool
	}{
		"valid swap (exact in, 2 step multi-hop)": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []bindings.Step{{
					PoolId:   atomPool,
					DenomOut: "uatom",
				}},
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expCost: &atomSwapAmount,
		},
		"non-existent step pool id": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []bindings.Step{{
					PoolId:   atomPool + 2,
					DenomOut: "uatom",
				}},
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"zero step pool id": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []bindings.Step{{
					PoolId:   0,
					DenomOut: "uatom",
				}},
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"wrong step denom out": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []bindings.Step{{
					PoolId:   atomPool,
					DenomOut: "ATOM",
				}},
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"self-swap not allowed": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []bindings.Step{{
					PoolId:   atomPool,
					DenomOut: "uosmo", // this is same as the input (output of first swap)
				}},
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid step denom out": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []bindings.Step{{
					PoolId:   atomPool,
					DenomOut: "invalid",
				}},
				Amount: bindings.SwapAmountWithLimit{
					ExactIn: &amountIn,
				},
			},
			expErr: true,
		},
		"empty step denom out": {
			swap: &bindings.SwapMsg{
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "uosmo",
				},
				Route: []bindings.Step{{
					PoolId:   atomPool,
					DenomOut: "",
				}},
				Amount: bindings.SwapAmountWithLimit{
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
			gotAmount, gotErr := wasmbinding.PerformSwap(osmosis.GAMMKeeper, subCtx, actor, spec.swap)
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

func TestJoinPoolNoSwap(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12_000_000),
		sdk.NewInt64Coin("ustar", 240_000_000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	osmoStarLiquidity := sdk.NewCoins(sdk.NewInt64Coin("uosmo", 12_000),
		sdk.NewInt64Coin("ustar", 240_000))

	specs := map[string]struct {
		JoinPoolNoSwap *bindings.JoinPoolNoSwap
		expErr         error
	}{
		"zero share out amount": {
			JoinPoolNoSwap: &bindings.JoinPoolNoSwap{
				PoolId:         starPool,
				ShareOutAmount: sdk.NewInt(0),
				TokenInMaxs:    osmoStarLiquidity,
			},
			expErr: errors.New("share ratio is zero or negative: invalid calculated result"),
		},
		"incorrect pool id": {
			JoinPoolNoSwap: &bindings.JoinPoolNoSwap{
				PoolId:         starPool + uint64(10),
				ShareOutAmount: sdk.NewInt(1000000),
				TokenInMaxs:    osmoStarLiquidity,
			},
			expErr: errors.New("pool with ID 11 does not exist"),
		},
		"sending one coin": {
			JoinPoolNoSwap: &bindings.JoinPoolNoSwap{
				PoolId:         starPool,
				ShareOutAmount: sdk.NewInt(1000000),
				TokenInMaxs:    sdk.NewCoins(sdk.NewCoin("osmo", sdk.NewInt(10))),
			},
			expErr: errors.New("TokenInMaxs is less than the needed LP liquidity to this JoinPoolNoSwap, upperbound: 10osmo, needed 1uosmo,1ustar: calculated amount is larger than max amount"),
		},
		"valid join pool": {
			JoinPoolNoSwap: &bindings.JoinPoolNoSwap{
				PoolId:         starPool,
				ShareOutAmount: sdk.NewInt(1000000),
				TokenInMaxs:    osmoStarLiquidity,
			},
			expErr: nil,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// use scratch context to avoid interference between tests
			subCtx, _ := ctx.CacheContext()
			// when
			gotErr := wasmbinding.PerformJoinPoolNoSwap(osmosis.GAMMKeeper, subCtx, actor, spec.JoinPoolNoSwap)
			// then
			if spec.expErr != nil {
				require.EqualError(t, gotErr, spec.expErr.Error())
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestExitPool(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12_000_000),
		sdk.NewInt64Coin("ustar", 240_000_000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	err := osmosis.GAMMKeeper.JoinPoolNoSwap(ctx, actor, starPool, sdk.NewInt(100000000000000), sdk.NewCoins(sdk.NewInt64Coin("uosmo", 48000_000), sdk.NewInt64Coin("ustar", 9600000_000)))

	require.NoError(t, err)

	osmoStarLiquidity := sdk.NewCoins(sdk.NewInt64Coin("uosmo", 12_000),
		sdk.NewInt64Coin("ustar", 240_000))

	specs := map[string]struct {
		ExitPool *bindings.ExitPool
		expErr   error
	}{
		"valid exit pool": {
			ExitPool: &bindings.ExitPool{
				PoolId:        starPool,
				ShareInAmount: sdk.NewInt(100000000000000000),
				TokenOutMins:  sdk.NewCoins(sdk.NewInt64Coin("uosmo", 11999), sdk.NewInt64Coin("ustar", 239999)),
			}},
		"zero share in amount": {
			ExitPool: &bindings.ExitPool{
				PoolId:        starPool,
				ShareInAmount: sdk.NewInt(0),
				TokenOutMins:  osmoStarLiquidity,
			},
			expErr: errors.New("share ratio is zero or negative: invalid calculated result"),
		},
		"incorrect pool id": {
			ExitPool: &bindings.ExitPool{
				PoolId:        starPool + uint64(10),
				ShareInAmount: sdk.NewInt(1000000),
				TokenOutMins:  osmoStarLiquidity,
			},
			expErr: errors.New("pool with ID 11 does not exist"),
		},
		"sending one coin": {
			ExitPool: &bindings.ExitPool{
				PoolId:        starPool,
				ShareInAmount: sdk.NewInt(1000000),
				TokenOutMins:  sdk.NewCoins(sdk.NewCoin("osmo", sdk.NewInt(10))),
			},
			expErr: errors.New("Exit pool returned  , minimum tokens out specified as 10osmo: calculated amount is lesser than min amount"),
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// use scratch context to avoid interference between tests
			subCtx, _ := ctx.CacheContext()
			// when
			gotErr := wasmbinding.PerformExit(osmosis.GAMMKeeper, subCtx, actor, spec.ExitPool)
			// then
			if spec.expErr != nil {
				require.EqualError(t, gotErr, spec.expErr.Error())
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestJoinSwapExactAmountIn(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12_000_000),
		sdk.NewInt64Coin("ustar", 240_000_000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	err := osmosis.GAMMKeeper.JoinPoolNoSwap(ctx, actor, starPool, sdk.NewInt(100000000000000), sdk.NewCoins(sdk.NewInt64Coin("uosmo", 48000_000), sdk.NewInt64Coin("ustar", 9600000_000)))

	require.NoError(t, err)

	// osmoStarLiquidity := sdk.NewCoins(sdk.NewInt64Coin("uosmo", 12_000),
	// 	sdk.NewInt64Coin("ustar", 240_000))

	specs := map[string]struct {
		JoinSwapExactAmountIn *bindings.JoinSwapExactAmountIn
		expErr                error
	}{
		"excessive share out min amount": {
			JoinSwapExactAmountIn: &bindings.JoinSwapExactAmountIn{
				PoolId:            starPool,
				ShareOutMinAmount: sdk.NewInt(math.MaxInt64),
				TokenIn:           sdk.NewCoin("uosmo", sdk.NewInt(1_000_000))},
			expErr: errors.New("too much slippage; needed a minimum of 9223372036854775807 shares to pass, got 4083300053402688006: calculated amount is lesser than min amount"),
		},
		"incorrect pool id": {
			JoinSwapExactAmountIn: &bindings.JoinSwapExactAmountIn{
				PoolId:            starPool + 10,
				ShareOutMinAmount: sdk.NewInt(0),
				TokenIn:           sdk.NewCoin("uosmo", sdk.NewInt(1_000_000))},
			expErr: errors.New("pool with ID 11 does not exist"),
		},
		"valid join swap exact amount in": {
			JoinSwapExactAmountIn: &bindings.JoinSwapExactAmountIn{
				PoolId:            starPool,
				ShareOutMinAmount: sdk.NewInt(0),
				TokenIn:           sdk.NewCoin("uosmo", sdk.NewInt(1_000_000))},
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// use scratch context to avoid interference between tests
			subCtx, _ := ctx.CacheContext()
			// when
			_, gotErr := wasmbinding.PerformJoinSwapExactAmountIn(osmosis.GAMMKeeper, subCtx, actor, spec.JoinSwapExactAmountIn)
			// then
			if spec.expErr != nil {
				require.EqualError(t, gotErr, spec.expErr.Error())
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestExitSwapShareAmountIn(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12_000_000),
		sdk.NewInt64Coin("ustar", 240_000_000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	err := osmosis.GAMMKeeper.JoinPoolNoSwap(ctx, actor, starPool, sdk.NewInt(100000000000000), sdk.NewCoins(sdk.NewInt64Coin("uosmo", 48000_000), sdk.NewInt64Coin("ustar", 9600000_000)))

	require.NoError(t, err)

	// osmoStarLiquidity := sdk.NewCoins(sdk.NewInt64Coin("uosmo", 12_000),
	// 	sdk.NewInt64Coin("ustar", 240_000))

	specs := map[string]struct {
		ExitSwapShareAmountIn *bindings.ExitSwapShareAmountIn
		expErr                error
	}{
		"zero share in amount": {
			ExitSwapShareAmountIn: &bindings.ExitSwapShareAmountIn{
				PoolId:            starPool,
				TokenOutDenom:     "ustar",
				ShareInAmount:     sdk.NewInt(0),
				TokenOutMinAmount: sdk.NewInt(479000),
			},
			expErr: errors.New("exit swap share amount in message: share ratio is zero or negative: invalid calculated result"),
		},
		"incorrect pool id": {
			ExitSwapShareAmountIn: &bindings.ExitSwapShareAmountIn{
				PoolId:            starPool + 10,
				TokenOutDenom:     "ustar",
				ShareInAmount:     sdk.NewInt(100000000000000000),
				TokenOutMinAmount: sdk.NewInt(479000),
			},
			expErr: errors.New("exit swap share amount in message: pool with ID 11 does not exist"),
		},
		"valid exit swap share amount in": {
			ExitSwapShareAmountIn: &bindings.ExitSwapShareAmountIn{
				PoolId:            starPool,
				TokenOutDenom:     "ustar",
				ShareInAmount:     sdk.NewInt(100000000000000000),
				TokenOutMinAmount: sdk.NewInt(479000),
			},
		},
	}

	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// use scratch context to avoid interference between tests
			subCtx, _ := ctx.CacheContext()
			// when
			_, gotErr := wasmbinding.PerformExitSwapShareAmountIn(osmosis.GAMMKeeper, subCtx, actor, spec.ExitSwapShareAmountIn)
			// then
			if spec.expErr != nil {
				require.EqualError(t, gotErr, spec.expErr.Error())
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
