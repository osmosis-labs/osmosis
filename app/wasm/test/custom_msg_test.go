package wasm

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/osmosis-labs/osmosis/v7/x/tokenfactory/types"

	"github.com/stretchr/testify/require"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app"
	wasmbindings "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
)

func TestCreateDenomMsg(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	lucky := RandomAccountAddress()
	reflect := instantiateReflectContract(t, ctx, osmosis, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, reflect, reflectAmount)

	msg := wasmbindings.OsmosisMsg{CreateDenom: &wasmbindings.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	// query the denom and see if it matches
	query := wasmbindings.OsmosisQuery{
		FullDenom: &wasmbindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp := wasmbindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, fmt.Sprintf("factory/%s/SUN", reflect.String()))
}

func TestMintMsg(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	lucky := RandomAccountAddress()
	reflect := instantiateReflectContract(t, ctx, osmosis, lucky)
	require.NotEmpty(t, reflect)

	// Fund reflect contract with 100 base denom creation fees
	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
	fundAccount(t, ctx, osmosis, reflect, reflectAmount)

	// lucky was broke
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Empty(t, balances)

	// Create denom for minting
	msg := wasmbindings.OsmosisMsg{CreateDenom: &wasmbindings.CreateDenom{
		Subdenom: "SUN",
	}}
	err := executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount, ok := sdk.NewIntFromString("808010808")
	require.True(t, ok)
	msg = wasmbindings.OsmosisMsg{MintTokens: &wasmbindings.MintTokens{
		Denom:         sunDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 1)
	coin := balances[0]
	require.Equal(t, amount, coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query := wasmbindings.OsmosisQuery{
		FullDenom: &wasmbindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp := wasmbindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// mint the same denom again
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 1)
	coin = balances[0]
	require.Equal(t, amount.MulRaw(2), coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = wasmbindings.OsmosisQuery{
		FullDenom: &wasmbindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp = wasmbindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// now mint another amount / denom
	// create it first
	msg = wasmbindings.OsmosisMsg{CreateDenom: &wasmbindings.CreateDenom{
		Subdenom: "MOON",
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)
	moonDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

	amount = amount.SubRaw(1)
	msg = wasmbindings.OsmosisMsg{MintTokens: &wasmbindings.MintTokens{
		Denom:         moonDenom,
		Amount:        amount,
		MintToAddress: lucky.String(),
	}}
	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	require.NoError(t, err)

	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	require.Len(t, balances, 2)
	coin = balances[0]
	require.Equal(t, amount, coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = wasmbindings.OsmosisQuery{
		FullDenom: &wasmbindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "MOON",
		},
	}
	resp = wasmbindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)

	// and check the first denom is unchanged
	coin = balances[1]
	require.Equal(t, amount.AddRaw(1).MulRaw(2), coin.Amount)
	require.Contains(t, coin.Denom, "factory/")

	// query the denom and see if it matches
	query = wasmbindings.OsmosisQuery{
		FullDenom: &wasmbindings.FullDenom{
			CreatorAddr: reflect.String(),
			Subdenom:    "SUN",
		},
	}
	resp = wasmbindings.FullDenomResponse{}
	queryCustom(t, ctx, osmosis, reflect, query, &resp)

	require.Equal(t, resp.Denom, coin.Denom)
}

type BaseState struct {
	StarPool  uint64
	AtomPool  uint64
	RegenPool uint64
}

func TestSwapMsg(t *testing.T) {
	// table tests with this setup
	cases := []struct {
		name       string
		msg        func(BaseState) *wasmbindings.SwapMsg
		expectErr  bool
		initFunds  sdk.Coin
		finalFunds []sdk.Coin
	}{
		{
			name: "exact in: simple swap works",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.StarPool,
						DenomIn:  "uosmo",
						DenomOut: "ustar",
					},
					// Note: you must use empty array, not nil, for valid Rust JSON
					Route: []wasmbindings.Step{},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactIn: &wasmbindings.ExactIn{
							Input:     sdk.NewInt(12000000),
							MinOutput: sdk.NewInt(5000000),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("uosmo", 13000000),
			finalFunds: []sdk.Coin{
				sdk.NewInt64Coin("uosmo", 1000000),
				sdk.NewInt64Coin("ustar", 120000000),
			},
		},
		{
			name: "exact in: price too low",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.StarPool,
						DenomIn:  "uosmo",
						DenomOut: "ustar",
					},
					// Note: you must use empty array, not nil, for valid Rust JSON
					Route: []wasmbindings.Step{},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactIn: &wasmbindings.ExactIn{
							Input:     sdk.NewInt(12000000),
							MinOutput: sdk.NewInt(555000000),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("uosmo", 13000000),
			expectErr: true,
		},
		{
			name: "exact in: not enough funds to swap",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.StarPool,
						DenomIn:  "uosmo",
						DenomOut: "ustar",
					},
					// Note: you must use empty array, not nil, for valid Rust JSON
					Route: []wasmbindings.Step{},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactIn: &wasmbindings.ExactIn{
							Input:     sdk.NewInt(12000000),
							MinOutput: sdk.NewInt(5000000),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("uosmo", 7000000),
			expectErr: true,
		},
		{
			name: "exact in: invalidPool",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.StarPool,
						DenomIn:  "uosmo",
						DenomOut: "uatom",
					},
					// Note: you must use empty array, not nil, for valid Rust JSON
					Route: []wasmbindings.Step{},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactIn: &wasmbindings.ExactIn{
							Input:     sdk.NewInt(12000000),
							MinOutput: sdk.NewInt(100000),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("uosmo", 13000000),
			expectErr: true,
		},

		// FIXME: this panics in GAMM module !?! hits a known TODO
		// https://github.com/osmosis-labs/osmosis/blob/a380ab2fcd39fb94c2b10411e07daf664911257a/osmomath/math.go#L47-L51
		//"exact out: panics if too much swapped": {
		//	msg: func(state BaseState) *wasmbindings.SwapMsg {
		//		return &wasmbindings.SwapMsg{
		//			First: wasmbindings.Swap{
		//				PoolId:   state.StarPool,
		//				DenomIn:  "uosmo",
		//				DenomOut: "ustar",
		//			},
		//			// Note: you must use empty array, not nil, for valid Rust JSON
		//			Route: []wasmbindings.Step{},
		//			Amount: wasmbindings.SwapAmountWithLimit{
		//				ExactOut: &wasmbindings.ExactOut{
		//					MaxInput: sdk.NewInt(22000000),
		//					Output:   sdk.NewInt(120000000),
		//				},
		//			},
		//		}
		//	},
		//	initFunds: sdk.NewInt64Coin("uosmo", 15000000),
		//	finalFunds: []sdk.Coin{
		//		sdk.NewInt64Coin("uosmo", 3000000),
		//		sdk.NewInt64Coin("ustar", 120000000),
		//	},
		//},
		{
			name: "exact out: simple swap works",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.AtomPool,
						DenomIn:  "uosmo",
						DenomOut: "uatom",
					},
					// Note: you must use empty array, not nil, for valid Rust JSON
					Route: []wasmbindings.Step{},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactOut: &wasmbindings.ExactOut{
							// 12 OSMO * 6 ATOM == 18 OSMO * 4 ATOM (+6 OSMO, -2 ATOM)
							MaxInput: sdk.NewInt(7000000),
							Output:   sdk.NewInt(2000000),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("uosmo", 8000000),
			finalFunds: []sdk.Coin{
				sdk.NewInt64Coin("uatom", 2000000),
				sdk.NewInt64Coin("uosmo", 2000000),
			},
		},
		{
			name: "exact in: 2 step multi-hop",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.StarPool,
						DenomIn:  "ustar",
						DenomOut: "uosmo",
					},
					Route: []wasmbindings.Step{{
						PoolId:   state.AtomPool,
						DenomOut: "uatom",
					}},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactIn: &wasmbindings.ExactIn{
							Input:     sdk.NewInt(240000000),
							MinOutput: sdk.NewInt(1999000),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("ustar", 240000000),
			finalFunds: []sdk.Coin{
				// 240 STAR -> 6 OSMO
				// 6 OSMO -> 2 ATOM (with minor rounding)
				sdk.NewInt64Coin("uatom", 1999999),
			},
		},
		{
			name: "exact out: 2 step multi-hop",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.AtomPool,
						DenomIn:  "uosmo",
						DenomOut: "uatom",
					},
					Route: []wasmbindings.Step{{
						PoolId:   state.RegenPool,
						DenomOut: "uregen",
					}},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactOut: &wasmbindings.ExactOut{
							MaxInput: sdk.NewInt(2000000),
							Output:   sdk.NewInt(12000000 - 12),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("uosmo", 2000000),
			finalFunds: []sdk.Coin{
				// 2 OSMO -> 1.2 ATOM
				// 1.2 ATOM -> 12 REGEN (with minor rounding)
				sdk.NewInt64Coin("uosmo", 2),
				sdk.NewInt64Coin("uregen", 12000000-12),
			},
		},
		// FIXME: this panics in GAMM module !?! hits a known TODO
		// https://github.com/osmosis-labs/osmosis/blob/a380ab2fcd39fb94c2b10411e07daf664911257a/osmomath/math.go#L47-L51
		// {
		// 	name: "exact out: panics on math power stuff",
		// 	msg: func(state BaseState) *wasmbindings.SwapMsg {
		// 		return &wasmbindings.SwapMsg{
		// 			First: wasmbindings.Swap{
		// 				PoolId:   state.StarPool,
		// 				DenomIn:  "ustar",
		// 				DenomOut: "uosmo",
		// 			},
		// 			Route: []wasmbindings.Step{{
		// 				PoolId:   state.AtomPool,
		// 				DenomOut: "uatom",
		// 			}},
		// 			Amount: wasmbindings.SwapAmountWithLimit{
		// 				ExactOut: &wasmbindings.ExactOut{
		// 					MaxInput: sdk.NewInt(240005000),
		// 					Output:   sdk.NewInt(2000000),
		// 				},
		// 			},
		// 		}
		// 	},
		// 	initFunds: sdk.NewInt64Coin("ustar", 240005000),
		// 	finalFunds: []sdk.Coin{
		// 		// 240 STAR -> 6 OSMO
		// 		// 6 OSMO -> 2 ATOM (with minor rounding)
		// 		sdk.NewInt64Coin("uatom", 2000000),
		// 		sdk.NewInt64Coin("ustar", 5000),
		// 	},
		// },
		{
			name: "exact in: 3 step multi-hop",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.StarPool,
						DenomIn:  "ustar",
						DenomOut: "uosmo",
					},
					Route: []wasmbindings.Step{{
						PoolId:   state.AtomPool,
						DenomOut: "uatom",
					}, {
						PoolId:   state.RegenPool,
						DenomOut: "uregen",
					}},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactIn: &wasmbindings.ExactIn{
							Input:     sdk.NewInt(240000000),
							MinOutput: sdk.NewInt(23900000),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("ustar", 240000000),
			finalFunds: []sdk.Coin{
				// 240 STAR -> 6 OSMO
				// 6 OSMO -> 2 ATOM
				// 2 ATOM -> 24 REGEN (with minor rounding)
				sdk.NewInt64Coin("uregen", 23999990),
			},
		},
		{
			name: "exact out: 3 step multi-hop",
			msg: func(state BaseState) *wasmbindings.SwapMsg {
				return &wasmbindings.SwapMsg{
					First: wasmbindings.Swap{
						PoolId:   state.StarPool,
						DenomIn:  "ustar",
						DenomOut: "uosmo",
					},
					Route: []wasmbindings.Step{{
						PoolId:   state.AtomPool,
						DenomOut: "uatom",
					}, {
						PoolId:   state.RegenPool,
						DenomOut: "uregen",
					}},
					Amount: wasmbindings.SwapAmountWithLimit{
						ExactOut: &wasmbindings.ExactOut{
							MaxInput: sdk.NewInt(50000000),
							Output:   sdk.NewInt(12000000),
						},
					},
				}
			},
			initFunds: sdk.NewInt64Coin("ustar", 50000000),
			finalFunds: []sdk.Coin{
				// ~48 STAR -> 2 OSMO
				// 2 OSMO -> .857 ATOM
				// .857 ATOM -> 12 REGEN (with minor rounding)
				sdk.NewInt64Coin("uregen", 12000000),
				sdk.NewInt64Coin("ustar", 1999971),
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			creator := RandomAccountAddress()
			osmosis, ctx := SetupCustomApp(t, creator)
			state := prepareBaseState(t, ctx, osmosis)

			trader := RandomAccountAddress()
			fundAccount(t, ctx, osmosis, trader, []sdk.Coin{tc.initFunds})
			reflect := instantiateReflectContract(t, ctx, osmosis, trader)
			require.NotEmpty(t, reflect)

			msg := wasmbindings.OsmosisMsg{Swap: tc.msg(state)}
			err := executeCustom(t, ctx, osmosis, reflect, trader, msg, tc.initFunds)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				balances := osmosis.BankKeeper.GetAllBalances(ctx, reflect)
				// uncomment these to debug any confusing results (show balances, not (*big.Int)(0x140005e51e0))
				// fmt.Printf("Expected: %s\n", tc.finalFunds)
				// fmt.Printf("Got: %s\n", balances)
				require.EqualValues(t, tc.finalFunds, balances)
			}
		})
	}
}

// test setup for each run through the table test above
func prepareBaseState(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp) BaseState {
	actor := RandomAccountAddress()

	swapperFunds := sdk.NewCoins(
		sdk.NewInt64Coin("uatom", 333000000),
		sdk.NewInt64Coin("uosmo", 555000000+3*poolFee),
		sdk.NewInt64Coin("uregen", 777000000),
		sdk.NewInt64Coin("ustar", 999000000),
	)
	fundAccount(t, ctx, osmosis, actor, swapperFunds)

	// 20 star to 1 osmo
	funds1 := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12000000),
		sdk.NewInt64Coin("ustar", 240000000),
	}
	starPool := preparePool(t, ctx, osmosis, actor, funds1)

	// 2 osmo to 1 atom
	funds2 := []sdk.Coin{
		sdk.NewInt64Coin("uatom", 6000000),
		sdk.NewInt64Coin("uosmo", 12000000),
	}
	atomPool := preparePool(t, ctx, osmosis, actor, funds2)

	// 16 regen to 1 atom
	funds3 := []sdk.Coin{
		sdk.NewInt64Coin("uatom", 6000000),
		sdk.NewInt64Coin("uregen", 96000000),
	}
	regenPool := preparePool(t, ctx, osmosis, actor, funds3)

	return BaseState{
		StarPool:  starPool,
		AtomPool:  atomPool,
		RegenPool: regenPool,
	}
}

type ReflectExec struct {
	ReflectMsg    *ReflectMsgs    `json:"reflect_msg,omitempty"`
	ReflectSubMsg *ReflectSubMsgs `json:"reflect_sub_msg,omitempty"`
}

type ReflectMsgs struct {
	Msgs []wasmvmtypes.CosmosMsg `json:"msgs"`
}

type ReflectSubMsgs struct {
	Msgs []wasmvmtypes.SubMsg `json:"msgs"`
}

func executeCustom(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, contract sdk.AccAddress, sender sdk.AccAddress, msg wasmbindings.OsmosisMsg, funds sdk.Coin) error {
	customBz, err := json.Marshal(msg)
	require.NoError(t, err)
	reflectMsg := ReflectExec{
		ReflectMsg: &ReflectMsgs{
			Msgs: []wasmvmtypes.CosmosMsg{{
				Custom: customBz,
			}},
		},
	}
	reflectBz, err := json.Marshal(reflectMsg)
	require.NoError(t, err)

	// no funds sent if amount is 0
	var coins sdk.Coins
	if !funds.Amount.IsNil() {
		coins = sdk.Coins{funds}
	}

	contractKeeper := keeper.NewDefaultPermissionKeeper(osmosis.WasmKeeper)
	_, err = contractKeeper.Execute(ctx, contract, sender, reflectBz, coins)
	return err
}

//func TestLockTokensMsg(t *testing.T) {
//	cases := []struct {
//		name       string
//		msg        func(BaseState) *wasmbindings.LockTokensMsg
//		expectErr  bool
//		initFunds  sdk.Coin
//		finalFunds []sdk.Coin
//		locked     []sdk.Coin
//	}{
//		{
//			name: "simple lock works",
//			msg: func(state BaseState) *wasmbindings.LockTokensMsg {
//				tmp, _ := time.ParseDuration("3600")
//				duration := wasmbindings.Duration(tmp)
//				return &wasmbindings.LockTokensMsg{
//					"gamm/pool/1",
//					sdk.NewInt(1),
//					duration,
//				}
//			},
//			initFunds: sdk.NewInt64Coin("gamm/pool/1", 10),
//			finalFunds: []sdk.Coin{
//				sdk.NewInt64Coin("gamm/pool/1", 9),
//			},
//			locked: []sdk.Coin{
//				sdk.NewInt64Coin("gamm/pool/1", 1),
//			},
//		},
//	}
//	for _, tc := range cases {
//		tc := tc
//		t.Run(tc.name, func(t *testing.T) {
//			creator := RandomAccountAddress()
//			osmosis, ctx := SetupCustomApp(t, creator)
//			state := prepareBaseState(t, ctx, osmosis)
//
//			lp := RandomAccountAddress()
//			fundAccount(t, ctx, osmosis, lp, []sdk.Coin{tc.initFunds})
//			reflect := instantiateReflectContract(t, ctx, osmosis, lp)
//			require.NotEmpty(t, reflect)
//
//			msg := wasmbindings.OsmosisMsg{LockTokens: tc.msg(state)}
//			err := executeCustom(t, ctx, osmosis, reflect, lp, msg, tc.initFunds)
//			if tc.expectErr {
//				require.Error(t, err)
//			} else {
//				require.NoError(t, err)
//				balances := osmosis.BankKeeper.GetAllBalances(ctx, reflect)
//				// uncomment these to debug any confusing results (show balances, not (*big.Int)(0x140005e51e0))
//				// fmt.Printf("Expected: %s\n", tc.finalFunds)
//				fmt.Printf("Got: %s\n", balances)
//				require.EqualValues(t, tc.finalFunds, balances)
//			}
//		})
//	}
//}

func TestLockTokensMsg(t *testing.T) {
	creator := RandomAccountAddress()

	osmosis, ctx := SetupCustomApp(t, creator)
	lp := RandomAccountAddress()
	initFunds := sdk.NewCoins(
		sdk.NewInt64Coin("uosmo", 555000000+3*poolFee),
		sdk.NewInt64Coin("ustar", 999000000),
	)
	fundAccount(t, ctx, osmosis, lp, initFunds)

	// 20 star to 1 osmo
	funds1 := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12000000),
		sdk.NewInt64Coin("ustar", 240000000),
	}
	preparePool(t, ctx, osmosis, lp, funds1)

	reflect := instantiateReflectContract(t, ctx, osmosis, lp)
	require.NotEmpty(t, reflect)

	// lp has some pool tokens
	balances := osmosis.BankKeeper.GetAllBalances(ctx, lp)
	require.Len(t, balances, 3)
	coin := balances[0]
	require.Equal(t, coin.Denom, "gamm/pool/1")

	amount, ok := sdk.NewIntFromString("200")
	require.True(t, ok)
	duration, _ := time.ParseDuration("3600")
	msg := wasmbindings.OsmosisMsg{LockTokens: &wasmbindings.LockTokensMsg{
		Denom:    "gamm/pool/1",
		Amount:   amount,
		Duration: wasmbindings.Duration(duration),
	}}
	err := executeCustom(t, ctx, osmosis, reflect, lp, msg, sdk.Coin{})
	require.NoError(t, err)

	//
	//balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	//require.Len(t, balances, 1)
	//coin := balances[0]
	//require.Equal(t, amount, coin.Amount)
	//require.Contains(t, coin.Denom, "factory/")
	//
	//// query the denom and see if it matches
	//query := wasmbindings.OsmosisQuery{
	//	FullDenom: &wasmbindings.FullDenom{
	//		Contract: reflect.String(),
	//		SubDenom: "SUN",
	//	},
	//}
	//resp := wasmbindings.FullDenomResponse{}
	//queryCustom(t, ctx, osmosis, reflect, query, &resp)
	//
	//require.Equal(t, resp.Denom, coin.Denom)
	//
	//// mint the same denom again
	//err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	//require.NoError(t, err)
	//
	//balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	//require.Len(t, balances, 1)
	//coin = balances[0]
	//require.Equal(t, amount.MulRaw(2), coin.Amount)
	//require.Contains(t, coin.Denom, "factory/")
	//
	//// query the denom and see if it matches
	//query = wasmbindings.OsmosisQuery{
	//	FullDenom: &wasmbindings.FullDenom{
	//		Contract: reflect.String(),
	//		SubDenom: "SUN",
	//	},
	//}
	//resp = wasmbindings.FullDenomResponse{}
	//queryCustom(t, ctx, osmosis, reflect, query, &resp)
	//
	//require.Equal(t, resp.Denom, coin.Denom)
	//
	//// now mint another amount / denom
	//amount = amount.SubRaw(1)
	//msg = wasmbindings.OsmosisMsg{MintTokens: &wasmbindings.MintTokens{
	//	SubDenom:  "MOON",
	//	Amount:    amount,
	//	Recipient: lucky.String(),
	//}}
	//err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, sdk.Coin{})
	//require.NoError(t, err)
	//
	//balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
	//require.Len(t, balances, 2)
	//coin = balances[0]
	//require.Equal(t, amount, coin.Amount)
	//require.Contains(t, coin.Denom, "factory/")
	//
	//// query the denom and see if it matches
	//query = wasmbindings.OsmosisQuery{
	//	FullDenom: &wasmbindings.FullDenom{
	//		Contract: reflect.String(),
	//		SubDenom: "MOON",
	//	},
	//}
	//resp = wasmbindings.FullDenomResponse{}
	//queryCustom(t, ctx, osmosis, reflect, query, &resp)
	//
	//require.Equal(t, resp.Denom, coin.Denom)
	//
	//// and check the first denom is unchanged
	//coin = balances[1]
	//require.Equal(t, amount.AddRaw(1).MulRaw(2), coin.Amount)
	//require.Contains(t, coin.Denom, "factory/")
	//
	//// query the denom and see if it matches
	//query = wasmbindings.OsmosisQuery{
	//	FullDenom: &wasmbindings.FullDenom{
	//		Contract: reflect.String(),
	//		SubDenom: "SUN",
	//	},
	//}
	//resp = wasmbindings.FullDenomResponse{}
	//queryCustom(t, ctx, osmosis, reflect, query, &resp)
	//
	//require.Equal(t, resp.Denom, coin.Denom)
}
