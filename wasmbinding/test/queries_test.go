package wasmbinding

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/wasmbinding"
	"github.com/osmosis-labs/osmosis/v10/wasmbinding/bindings"
)

func TestFullDenom(t *testing.T) {
	actor := RandomAccountAddress()

	specs := map[string]struct {
		addr         string
		subdenom     string
		expFullDenom string
		expErr       bool
	}{
		"valid address": {
			addr:         actor.String(),
			subdenom:     "subDenom1",
			expFullDenom: fmt.Sprintf("factory/%s/subDenom1", actor.String()),
		},
		"empty address": {
			addr:     "",
			subdenom: "subDenom1",
			expErr:   true,
		},
		"invalid address": {
			addr:     "invalid",
			subdenom: "subDenom1",
			expErr:   true,
		},
		"empty sub-denom": {
			addr:         actor.String(),
			subdenom:     "",
			expFullDenom: fmt.Sprintf("factory/%s/", actor.String()),
		},
		"invalid sub-denom (contains underscore)": {
			addr:     actor.String(),
			subdenom: "sub_denom",
			expErr:   true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotFullDenom, gotErr := wasmbinding.GetFullDenom(spec.addr, spec.subdenom)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expFullDenom, gotFullDenom, "exp %s but got %s", spec.expFullDenom, gotFullDenom)
		})
	}
}

func TestDenomAdmin(t *testing.T) {
	addr := RandomAccountAddress()
	app, ctx := SetupCustomApp(t, addr)

	// set token creation fee to zero to make testing easier
	tfParams := app.TokenFactoryKeeper.GetParams(ctx)
	tfParams.DenomCreationFee = sdk.NewCoins()
	app.TokenFactoryKeeper.SetParams(ctx, tfParams)

	// create a subdenom via the token factory
	admin := sdk.AccAddress([]byte("addr1_______________"))
	tfDenom, err := app.TokenFactoryKeeper.CreateDenom(ctx, admin.String(), "subdenom")
	require.NoError(t, err)
	require.NotEmpty(t, tfDenom)

	queryPlugin := wasmbinding.NewQueryPlugin(app.GAMMKeeper, app.TwapKeeper, app.TokenFactoryKeeper)

	testCases := []struct {
		name        string
		denom       string
		expectErr   bool
		expectAdmin string
	}{
		{
			name:        "valid token factory denom",
			denom:       tfDenom,
			expectAdmin: admin.String(),
		},
		{
			name:        "invalid token factory denom",
			denom:       "uosmo",
			expectErr:   false,
			expectAdmin: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			resp, err := queryPlugin.GetDenomAdmin(ctx, tc.denom)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.expectAdmin, resp.Admin)
			}
		})
	}
}

func TestPoolState(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12000000),
		sdk.NewInt64Coin("ustar", 240000000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	// FIXME: Derive / obtain these values
	starSharesDenom := fmt.Sprintf("gamm/pool/%d", starPool)
	starSharedAmount, _ := sdk.NewIntFromString("100_000_000_000_000_000_000")

	queryPlugin := wasmbinding.NewQueryPlugin(osmosis.GAMMKeeper, osmosis.TwapKeeper, osmosis.TokenFactoryKeeper)

	specs := map[string]struct {
		poolId       uint64
		expPoolState *bindings.PoolAssets
		expErr       bool
	}{
		"existent pool id": {
			poolId: starPool,
			expPoolState: &bindings.PoolAssets{
				Assets: poolFunds,
				Shares: sdk.NewCoin(starSharesDenom, starSharedAmount),
			},
		},
		"non-existent pool id": {
			poolId: starPool + 1,
			expErr: true,
		},
		"zero pool id": {
			poolId: 0,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotPoolState, gotErr := queryPlugin.GetPoolState(ctx, spec.poolId)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, spec.expPoolState, gotPoolState, "exp %s but got %s", spec.expPoolState, gotPoolState)
		})
	}
}

func TestSpotPrice(t *testing.T) {
	actor := RandomAccountAddress()
	swapFee := 0. // FIXME: Set / support an actual fee
	epsilon := 1e-6
	osmosis, ctx := SetupCustomApp(t, actor)

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12000000),
		sdk.NewInt64Coin("ustar", 240000000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	uosmo := poolFunds[0].Amount.ToDec().MustFloat64()
	ustar := poolFunds[1].Amount.ToDec().MustFloat64()

	starPrice := sdk.MustNewDecFromStr(fmt.Sprintf("%f", uosmo/ustar))
	starFee := sdk.MustNewDecFromStr(fmt.Sprintf("%f", swapFee))
	starPriceWithFee := starPrice.Add(starFee)

	queryPlugin := wasmbinding.NewQueryPlugin(osmosis.GAMMKeeper, osmosis.TwapKeeper, osmosis.TokenFactoryKeeper)

	specs := map[string]struct {
		spotPrice *bindings.SpotPrice
		expPrice  *sdk.Dec
		expErr    bool
	}{
		"valid spot price": {
			spotPrice: &bindings.SpotPrice{
				Swap: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expPrice: &starPrice,
		},
		"valid spot price with fee": {
			spotPrice: &bindings.SpotPrice{
				Swap: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				WithSwapFee: true,
			},
			expPrice: &starPriceWithFee,
		},
		"non-existent pool id": {
			spotPrice: &bindings.SpotPrice{
				Swap: bindings.Swap{
					PoolId:   starPool + 2,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"zero pool id": {
			spotPrice: &bindings.SpotPrice{
				Swap: bindings.Swap{
					PoolId:   0,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"invalid denom in": {
			spotPrice: &bindings.SpotPrice{
				Swap: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "invalid",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"empty denom in": {
			spotPrice: &bindings.SpotPrice{
				Swap: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"invalid denom out": {
			spotPrice: &bindings.SpotPrice{
				Swap: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "invalid",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"empty denom out": {
			spotPrice: &bindings.SpotPrice{
				Swap: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"null spot price": {
			spotPrice: nil,
			expErr:    true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotPrice, gotErr := queryPlugin.GetSpotPrice(ctx, spec.spotPrice)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.InEpsilonf(t, spec.expPrice.MustFloat64(), gotPrice.MustFloat64(), epsilon, "exp %s but got %s", spec.expPrice.String(), gotPrice.String())
		})
	}
}

func TestEstimateSwap(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)
	epsilon := 1e-3

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12000000),
		sdk.NewInt64Coin("ustar", 240000000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	// Estimate swap rate
	uosmo := poolFunds[0].Amount.ToDec().MustFloat64()
	ustar := poolFunds[1].Amount.ToDec().MustFloat64()
	swapRate := ustar / uosmo

	amountIn := sdk.NewInt(10000)
	zeroAmount := sdk.ZeroInt()
	negativeAmount := amountIn.Neg()

	amount := amountIn.ToDec().MustFloat64()
	starAmount := sdk.NewInt(int64(amount * swapRate))

	starSwapAmount := bindings.SwapAmount{Out: &starAmount}

	queryPlugin := wasmbinding.NewQueryPlugin(osmosis.GAMMKeeper, osmosis.TwapKeeper, osmosis.TokenFactoryKeeper)

	specs := map[string]struct {
		estimateSwap *bindings.EstimateSwap
		expCost      *bindings.SwapAmount
		expErr       bool
	}{
		"valid estimate swap (exact in)": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &amountIn,
				},
			},
			expCost: &starSwapAmount,
		},
		"non-existent pool id": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool + 3,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"zero pool id": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   0,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid denom in": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "invalid",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"empty denom in": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid denom out": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "invalid",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"empty denom out": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"null estimate swap": {
			estimateSwap: nil,
			expErr:       true,
		},
		"empty swap amount": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route:  nil,
				Amount: bindings.SwapAmount{},
			},
			expErr: true,
		},
		"zero amount in": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &zeroAmount,
				},
			},
			expErr: true,
		},
		"zero amount out": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					Out: &zeroAmount,
				},
			},
			expErr: true,
		},
		"negative amount in": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					In: &negativeAmount,
				},
			},
			expErr: true,
		},
		"negative amount out": {
			estimateSwap: &bindings.EstimateSwap{
				Sender: actor.String(),
				First: bindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: bindings.SwapAmount{
					Out: &negativeAmount,
				},
			},
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			gotCost, gotErr := queryPlugin.EstimateSwap(ctx, spec.estimateSwap)
			// then
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.InEpsilonf(t, (*spec.expCost.Out).ToDec().MustFloat64(), (*gotCost.Out).ToDec().MustFloat64(), epsilon, "exp %s but got %s", spec.expCost.Out.String(), gotCost.Out.String())
		})
	}
}
