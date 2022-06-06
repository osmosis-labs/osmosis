package wasm

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app/wasm"
	wasmbindings "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
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
			gotFullDenom, gotErr := wasm.GetFullDenom(spec.addr, spec.subdenom)
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

	queryPlugin := wasm.NewQueryPlugin(app.GAMMKeeper, app.TokenFactoryKeeper)

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

	queryPlugin := wasm.NewQueryPlugin(osmosis.GAMMKeeper, osmosis.TokenFactoryKeeper)

	specs := map[string]struct {
		poolId       uint64
		expPoolState *wasmbindings.PoolAssets
		expErr       bool
	}{
		"existent pool id": {
			poolId: starPool,
			expPoolState: &wasmbindings.PoolAssets{
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

	queryPlugin := wasm.NewQueryPlugin(osmosis.GAMMKeeper, osmosis.TokenFactoryKeeper)

	specs := map[string]struct {
		spotPrice *wasmbindings.SpotPrice
		expPrice  *sdk.Dec
		expErr    bool
	}{
		"valid spot price": {
			spotPrice: &wasmbindings.SpotPrice{
				Swap: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expPrice: &starPrice,
		},
		"valid spot price with fee": {
			spotPrice: &wasmbindings.SpotPrice{
				Swap: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				WithSwapFee: true,
			},
			expPrice: &starPriceWithFee,
		},
		"non-existent pool id": {
			spotPrice: &wasmbindings.SpotPrice{
				Swap: wasmbindings.Swap{
					PoolId:   starPool + 2,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"zero pool id": {
			spotPrice: &wasmbindings.SpotPrice{
				Swap: wasmbindings.Swap{
					PoolId:   0,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"invalid denom in": {
			spotPrice: &wasmbindings.SpotPrice{
				Swap: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "invalid",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"empty denom in": {
			spotPrice: &wasmbindings.SpotPrice{
				Swap: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "",
					DenomOut: "ustar",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"invalid denom out": {
			spotPrice: &wasmbindings.SpotPrice{
				Swap: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "invalid",
				},
				WithSwapFee: false,
			},
			expErr: true,
		},
		"empty denom out": {
			spotPrice: &wasmbindings.SpotPrice{
				Swap: wasmbindings.Swap{
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

	starSwapAmount := wasmbindings.SwapAmount{Out: &starAmount}

	queryPlugin := wasm.NewQueryPlugin(osmosis.GAMMKeeper, osmosis.TokenFactoryKeeper)

	specs := map[string]struct {
		estimateSwap *wasmbindings.EstimateSwap
		expCost      *wasmbindings.SwapAmount
		expErr       bool
	}{
		"valid estimate swap (exact in)": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					In: &amountIn,
				},
			},
			expCost: &starSwapAmount,
		},
		"non-existent pool id": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool + 3,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"zero pool id": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   0,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid denom in": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "invalid",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"empty denom in": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"invalid denom out": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "invalid",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					In: &amountIn,
				},
			},
			expErr: true,
		},
		"empty denom out": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "ustar",
					DenomOut: "",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
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
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route:  nil,
				Amount: wasmbindings.SwapAmount{},
			},
			expErr: true,
		},
		"zero amount in": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					In: &zeroAmount,
				},
			},
			expErr: true,
		},
		"zero amount out": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					Out: &zeroAmount,
				},
			},
			expErr: true,
		},
		"negative amount in": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
					In: &negativeAmount,
				},
			},
			expErr: true,
		},
		"negative amount out": {
			estimateSwap: &wasmbindings.EstimateSwap{
				Sender: actor.String(),
				First: wasmbindings.Swap{
					PoolId:   starPool,
					DenomIn:  "uosmo",
					DenomOut: "ustar",
				},
				Route: nil,
				Amount: wasmbindings.SwapAmount{
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

func TestJoinPoolShares(t *testing.T) {
	actor := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, actor)

	fundAccount(t, ctx, osmosis, actor, defaultFunds)

	poolFunds := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12000000),
		sdk.NewInt64Coin("ustar", 240000000),
	}
	// 20 star to 1 osmo
	starPool := preparePool(t, ctx, osmosis, actor, poolFunds)

	queryPlugin := wasm.NewQueryPlugin(osmosis.GAMMKeeper, osmosis.TokenFactoryKeeper)

	specs := map[string]struct {
		joinPoolShares *wasmbindings.JoinPoolShares
		err            error
	}{
		"valid join pool shares one coin": {
			joinPoolShares: &wasmbindings.JoinPoolShares{
				PoolId: starPool,
				Coins:  sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(120_000))),
			},
		},
		"valid join pool shares two coin": {
			joinPoolShares: &wasmbindings.JoinPoolShares{
				PoolId: starPool,
				Coins:  sdk.NewCoins(sdk.NewCoin("ustar", sdk.NewInt(2_400_000)), sdk.NewCoin("uosmo", sdk.NewInt(120_000))),
			},
		},
		"non-existent pool id": {
			joinPoolShares: &wasmbindings.JoinPoolShares{
				PoolId: starPool + 10,
				Coins:  sdk.NewCoins(sdk.NewCoin("ustar", sdk.NewInt(2_400_000)), sdk.NewCoin("uosmo", sdk.NewInt(120_000))),
			},
			err: errors.New("Invalid pool"),
		},
		"zero pool id": {
			joinPoolShares: &wasmbindings.JoinPoolShares{
				PoolId: 0,
				Coins:  sdk.NewCoins(sdk.NewCoin("ustar", sdk.NewInt(2_400_000)), sdk.NewCoin("uosmo", sdk.NewInt(120_000))),
			},
			err: errors.New("Invalid pool"),
		},
		"empty coins in": {
			joinPoolShares: &wasmbindings.JoinPoolShares{
				PoolId: starPool,
				Coins:  sdk.NewCoins(),
			},
			err: errors.New("balancer pool only supports LP'ing with one asset or all assets in pool"),
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			// when
			_, gotErr := queryPlugin.GetJoinPoolShares(ctx, spec.joinPoolShares)
			// then
			if spec.err != nil {
				require.EqualError(t, gotErr, spec.err.Error())
				return
			}
			require.NoError(t, gotErr)
		})
	}
}
