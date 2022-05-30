package wasm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/app"
	wasmbindings "github.com/osmosis-labs/osmosis/v7/app/wasm/bindings"
)

// func TestCreateDenomMsg(t *testing.T) {
// 	creator := RandomAccountAddress()
// 	osmosis, ctx := SetupCustomApp(t, creator)

// 	lucky := RandomAccountAddress()
// 	reflect := instantiateReflectContract(t, ctx, osmosis, lucky)
// 	require.NotEmpty(t, reflect)

// 	// Fund reflect contract with 100 base denom creation fees
// 	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
// 	fundAccount(t, ctx, osmosis, reflect, reflectAmount)

// 	msg := wasmbindings.OsmosisMsg{CreateDenom: &wasmbindings.CreateDenom{
// 		Subdenom: "SUN",
// 	}}
// 	err := executeCustom(t, ctx, osmosis, reflect, lucky, msg, []sdk.Coin{})
// 	require.NoError(t, err)

// 	// query the denom and see if it matches
// 	query := wasmbindings.OsmosisQuery{
// 		FullDenom: &wasmbindings.FullDenom{
// 			CreatorAddr: reflect.String(),
// 			Subdenom:    "SUN",
// 		},
// 	}
// 	resp := wasmbindings.FullDenomResponse{}
// 	queryCustom(t, ctx, osmosis, reflect, query, &resp)

// 	require.Equal(t, resp.Denom, fmt.Sprintf("factory/%s/SUN", reflect.String()))
// }

// func TestMintMsg(t *testing.T) {
// 	creator := RandomAccountAddress()
// 	osmosis, ctx := SetupCustomApp(t, creator)

// 	lucky := RandomAccountAddress()
// 	reflect := instantiateReflectContract(t, ctx, osmosis, lucky)
// 	require.NotEmpty(t, reflect)

// 	// Fund reflect contract with 100 base denom creation fees
// 	reflectAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)))
// 	fundAccount(t, ctx, osmosis, reflect, reflectAmount)

// 	// lucky was broke
// 	balances := osmosis.BankKeeper.GetAllBalances(ctx, lucky)
// 	require.Empty(t, balances)

// 	// Create denom for minting
// 	msg := wasmbindings.OsmosisMsg{CreateDenom: &wasmbindings.CreateDenom{
// 		Subdenom: "SUN",
// 	}}
// 	err := executeCustom(t, ctx, osmosis, reflect, lucky, msg, []sdk.Coin{})
// 	require.NoError(t, err)
// 	sunDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

// 	amount, ok := sdk.NewIntFromString("808010808")
// 	require.True(t, ok)
// 	msg = wasmbindings.OsmosisMsg{MintTokens: &wasmbindings.MintTokens{
// 		Denom:         sunDenom,
// 		Amount:        amount,
// 		MintToAddress: lucky.String(),
// 	}}
// 	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, []sdk.Coin{})
// 	require.NoError(t, err)

// 	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
// 	require.Len(t, balances, 1)
// 	coin := balances[0]
// 	require.Equal(t, amount, coin.Amount)
// 	require.Contains(t, coin.Denom, "factory/")

// 	// query the denom and see if it matches
// 	query := wasmbindings.OsmosisQuery{
// 		FullDenom: &wasmbindings.FullDenom{
// 			CreatorAddr: reflect.String(),
// 			Subdenom:    "SUN",
// 		},
// 	}
// 	resp := wasmbindings.FullDenomResponse{}
// 	queryCustom(t, ctx, osmosis, reflect, query, &resp)

// 	require.Equal(t, resp.Denom, coin.Denom)

// 	// mint the same denom again
// 	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, []sdk.Coin{})
// 	require.NoError(t, err)

// 	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
// 	require.Len(t, balances, 1)
// 	coin = balances[0]
// 	require.Equal(t, amount.MulRaw(2), coin.Amount)
// 	require.Contains(t, coin.Denom, "factory/")

// 	// query the denom and see if it matches
// 	query = wasmbindings.OsmosisQuery{
// 		FullDenom: &wasmbindings.FullDenom{
// 			CreatorAddr: reflect.String(),
// 			Subdenom:    "SUN",
// 		},
// 	}
// 	resp = wasmbindings.FullDenomResponse{}
// 	queryCustom(t, ctx, osmosis, reflect, query, &resp)

// 	require.Equal(t, resp.Denom, coin.Denom)

// 	// now mint another amount / denom
// 	// create it first
// 	msg = wasmbindings.OsmosisMsg{CreateDenom: &wasmbindings.CreateDenom{
// 		Subdenom: "MOON",
// 	}}
// 	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, []sdk.Coin{})
// 	require.NoError(t, err)
// 	moonDenom := fmt.Sprintf("factory/%s/%s", reflect.String(), msg.CreateDenom.Subdenom)

// 	amount = amount.SubRaw(1)
// 	msg = wasmbindings.OsmosisMsg{MintTokens: &wasmbindings.MintTokens{
// 		Denom:         moonDenom,
// 		Amount:        amount,
// 		MintToAddress: lucky.String(),
// 	}}
// 	err = executeCustom(t, ctx, osmosis, reflect, lucky, msg, []sdk.Coin{})
// 	require.NoError(t, err)

// 	balances = osmosis.BankKeeper.GetAllBalances(ctx, lucky)
// 	require.Len(t, balances, 2)
// 	coin = balances[0]
// 	require.Equal(t, amount, coin.Amount)
// 	require.Contains(t, coin.Denom, "factory/")

// 	// query the denom and see if it matches
// 	query = wasmbindings.OsmosisQuery{
// 		FullDenom: &wasmbindings.FullDenom{
// 			CreatorAddr: reflect.String(),
// 			Subdenom:    "MOON",
// 		},
// 	}
// 	resp = wasmbindings.FullDenomResponse{}
// 	queryCustom(t, ctx, osmosis, reflect, query, &resp)

// 	require.Equal(t, resp.Denom, coin.Denom)

// 	// and check the first denom is unchanged
// 	coin = balances[1]
// 	require.Equal(t, amount.AddRaw(1).MulRaw(2), coin.Amount)
// 	require.Contains(t, coin.Denom, "factory/")

// 	// query the denom and see if it matches
// 	query = wasmbindings.OsmosisQuery{
// 		FullDenom: &wasmbindings.FullDenom{
// 			CreatorAddr: reflect.String(),
// 			Subdenom:    "SUN",
// 		},
// 	}
// 	resp = wasmbindings.FullDenomResponse{}
// 	queryCustom(t, ctx, osmosis, reflect, query, &resp)

// 	require.Equal(t, resp.Denom, coin.Denom)
// }

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
			state := prepareSwapState(t, ctx, osmosis)

			trader := RandomAccountAddress()
			fundAccount(t, ctx, osmosis, trader, []sdk.Coin{tc.initFunds})
			reflect := instantiateReflectContract(t, ctx, osmosis, trader)
			require.NotEmpty(t, reflect)

			msg := wasmbindings.OsmosisMsg{Swap: tc.msg(state)}
			err := executeCustom(t, ctx, osmosis, reflect, trader, msg, []sdk.Coin{tc.initFunds})
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

func TestJoinPoolMsg(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	provider := RandomAccountAddress()

	providerFunds := sdk.NewCoins(
		sdk.NewInt64Coin("uatom", 333000000),
		sdk.NewInt64Coin("uosmo", 555000000+3*poolFee),
		sdk.NewInt64Coin("uregen", 777000000),
		sdk.NewInt64Coin("ustar", 999000000),
	)
	fundAccount(t, ctx, osmosis, creator, providerFunds)
	fundAccount(t, ctx, osmosis, provider, providerFunds)

	// 20 star to 1 osmo
	funds1 := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12000000),
		sdk.NewInt64Coin("ustar", 240000000),
	}
	starPool := preparePool(t, ctx, osmosis, creator, funds1)

	poolDenom := "gamm/pool/1"

	creatorBalance := osmosis.BankKeeper.GetBalance(ctx, creator, poolDenom)

	require.Equal(t, "100000000000000000000", creatorBalance.Amount.String())

	poolInfo, err := osmosis.GAMMKeeper.GetPoolAndPoke(ctx, starPool)

	require.NoError(t, err)

	coinsInPool := poolInfo.GetTotalPoolLiquidity(ctx)

	require.Equal(t, sdk.NewCoin("uosmo", sdk.NewInt(12000000)), coinsInPool[0])
	require.Equal(t, sdk.NewCoin("ustar", sdk.NewInt(240000000)), coinsInPool[1])

	totalSharesBefore := poolInfo.GetTotalShares()

	require.Equal(t, "100000000000000000000", totalSharesBefore.String())

	osmoStarLiquidity := sdk.NewCoins(sdk.NewInt64Coin("uosmo", 12_000), sdk.NewInt64Coin("ustar", 240_000))
	reflect := instantiateReflectContract(t, ctx, osmosis, provider)

	invalidMsg := wasmbindings.OsmosisMsg{JoinPoolNoSwap: &wasmbindings.JoinPoolNoSwap{
		PoolId:         starPool,
		ShareOutAmount: sdk.NewInt(100000000000000000),
		TokenInMaxs:    sdk.NewCoins(sdk.NewCoin("random", sdk.NewInt(10))),
	}}
	expectedErr := executeCustom(t, ctx, osmosis, reflect, provider, invalidMsg, osmoStarLiquidity)
	require.Error(t, expectedErr)
	require.Equal(t, "dispatch: submessages: join pool no swap: TokenInMaxs is less than the needed LP liquidity to this JoinPoolNoSwap, upperbound: 10random, needed 12000uosmo,240000ustar: calculated amount is larger than max amount", expectedErr.Error())

	//ShareOutAmount = TotalShares * tokenInAmount / poolAsset.amount
	//Either asset can be used for tokenInAmount and poolAsset.amount can be used to calculate this amount
	//ShareOutAmount = 100000000000000000000 * 12000 / 12000000 = 100000000000000000000 * 240000 / 240000000

	msg := wasmbindings.OsmosisMsg{JoinPoolNoSwap: &wasmbindings.JoinPoolNoSwap{
		PoolId:         starPool,
		ShareOutAmount: sdk.NewInt(100000000000000000),
		TokenInMaxs:    osmoStarLiquidity,
	}}
	err = executeCustom(t, ctx, osmosis, reflect, provider, msg, sdk.NewCoins(sdk.NewCoin("random", sdk.NewInt(10))))

	require.Error(t, err)
	require.Equal(t, "0random is smaller than 10random: insufficient funds", err.Error())

	err = executeCustom(t, ctx, osmosis, reflect, provider, msg, osmoStarLiquidity)

	require.NoError(t, err)

	reflectBalanceAfterJoining := osmosis.BankKeeper.GetBalance(ctx, reflect, poolDenom)

	require.Equal(t, "100000000000000000", reflectBalanceAfterJoining.Amount.String())

	poolInfoAfterDeposit, err := osmosis.GAMMKeeper.GetPoolAndPoke(ctx, starPool)

	require.NoError(t, err)

	coinsInPoolAfter := poolInfoAfterDeposit.GetTotalPoolLiquidity(ctx)

	require.Equal(t, sdk.NewCoin("uosmo", sdk.NewInt(12012000)), coinsInPoolAfter[0])
	require.Equal(t, sdk.NewCoin("ustar", sdk.NewInt(240240000)), coinsInPoolAfter[1])

	totalSharesAfter := poolInfoAfterDeposit.GetTotalShares()

	require.Equal(t, "100100000000000000000", totalSharesAfter.String())
}

func TestExitPoolMsg(t *testing.T) {
	creator := RandomAccountAddress()
	osmosis, ctx := SetupCustomApp(t, creator)

	provider := RandomAccountAddress()

	providerFunds := sdk.NewCoins(
		sdk.NewInt64Coin("uosmo", 555000000+3*poolFee),
		sdk.NewInt64Coin("ustar", 999000000),
	)
	fundAccount(t, ctx, osmosis, creator, providerFunds)
	fundAccount(t, ctx, osmosis, provider, providerFunds)

	// 20 star to 1 osmo
	funds1 := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 12000000),
		sdk.NewInt64Coin("ustar", 240000000),
	}
	starPool := preparePool(t, ctx, osmosis, creator, funds1)

	poolInfo, err := osmosis.GAMMKeeper.GetPoolAndPoke(ctx, starPool)

	poolDenom := "gamm/pool/1"

	creatorBalance := osmosis.BankKeeper.GetBalance(ctx, creator, poolDenom)

	require.Equal(t, "100000000000000000000", creatorBalance.Amount.String())

	require.NoError(t, err)

	coinsInPool := poolInfo.GetTotalPoolLiquidity(ctx)

	require.Equal(t, sdk.NewCoin("uosmo", sdk.NewInt(12000000)), coinsInPool[0])
	require.Equal(t, sdk.NewCoin("ustar", sdk.NewInt(240000000)), coinsInPool[1])

	totalSharesBefore := poolInfo.GetTotalShares()

	require.Equal(t, "100000000000000000000", totalSharesBefore.String())

	osmoStarLiquidity := sdk.NewCoins(sdk.NewInt64Coin("uosmo", 12_000), sdk.NewInt64Coin("ustar", 240_000))
	reflect := instantiateReflectContract(t, ctx, osmosis, provider)

	invalidMsg := wasmbindings.OsmosisMsg{JoinPoolNoSwap: &wasmbindings.JoinPoolNoSwap{
		PoolId:         starPool,
		ShareOutAmount: sdk.NewInt(100000000000000000),
		TokenInMaxs:    sdk.NewCoins(sdk.NewCoin("random", sdk.NewInt(10))),
	}}
	expectedErr := executeCustom(t, ctx, osmosis, reflect, provider, invalidMsg, sdk.NewCoins())
	require.Error(t, expectedErr)
	require.Equal(t, "dispatch: submessages: join pool no swap: TokenInMaxs is less than the needed LP liquidity to this JoinPoolNoSwap, upperbound: 10random, needed 12000uosmo,240000ustar: calculated amount is larger than max amount", expectedErr.Error())

	//ShareOutAmount = TotalShares * tokenInAmount / poolAsset.amount
	//Either asset can be used for tokenInAmount and poolAsset.amount can be used to calculate this amount
	//ShareOutAmount = 100000000000000000000 * 12000 / 12000000 = 100000000000000000000 * 240000 / 240000000

	msg := wasmbindings.OsmosisMsg{JoinPoolNoSwap: &wasmbindings.JoinPoolNoSwap{
		PoolId:         starPool,
		ShareOutAmount: sdk.NewInt(100000000000000000),
		TokenInMaxs:    osmoStarLiquidity,
	}}
	err = executeCustom(t, ctx, osmosis, reflect, provider, msg, sdk.NewCoins(sdk.NewCoin("random", sdk.NewInt(10))))

	require.Error(t, err)
	require.Equal(t, "0random is smaller than 10random: insufficient funds", err.Error())

	err = executeCustom(t, ctx, osmosis, reflect, provider, msg, osmoStarLiquidity)

	require.NoError(t, err)

	providerUstarBalanceAfterJoining := osmosis.BankKeeper.GetBalance(ctx, provider, "ustar")

	// 999_000_000 - 240_000 ustar
	require.Equal(t, "998760000", providerUstarBalanceAfterJoining.Amount.String())

	poolInfoAfterDeposit, err := osmosis.GAMMKeeper.GetPoolAndPoke(ctx, starPool)

	require.NoError(t, err)

	coinsInPoolAfter := poolInfoAfterDeposit.GetTotalPoolLiquidity(ctx)

	require.Equal(t, sdk.NewCoin("uosmo", sdk.NewInt(12012000)), coinsInPoolAfter[0])
	require.Equal(t, sdk.NewCoin("ustar", sdk.NewInt(240240000)), coinsInPoolAfter[1])

	totalSharesAfter := poolInfoAfterDeposit.GetTotalShares()

	require.Equal(t, "100100000000000000000", totalSharesAfter.String())

	contractBalanceAfterJoining := osmosis.BankKeeper.GetBalance(ctx, reflect, poolDenom)

	require.Equal(t, "100000000000000000", contractBalanceAfterJoining.Amount.String())

	//Simualate sending share tokens to user
	err = osmosis.BankKeeper.SendCoins(ctx, reflect, provider, sdk.NewCoins(sdk.NewCoin(poolDenom, sdk.NewInt(100000000000000000))))

	providerBalanceAfterJoining := osmosis.BankKeeper.GetBalance(ctx, provider, poolDenom)

	require.Equal(t, "100000000000000000", providerBalanceAfterJoining.Amount.String())

	providerUatomBalanceBefore := osmosis.BankKeeper.GetBalance(ctx, provider, "uatom")

	require.Equal(t, "0", providerUatomBalanceBefore.Amount.String())

	exitPoolMsg := wasmbindings.OsmosisMsg{ExitPool: &wasmbindings.ExitPool{
		PoolId:        starPool,
		ShareInAmount: sdk.NewInt(100000000000000000),
		TokenOutMins:  sdk.NewCoins(),
	}}

	err = executeCustom(t, ctx, osmosis, reflect, provider, exitPoolMsg, sdk.NewCoins(sdk.NewCoin(poolDenom, sdk.NewInt(100000000000000000))))

	require.NoError(t, err)

	providerLpTokenBalanceAfter := osmosis.BankKeeper.GetBalance(ctx, provider, poolDenom)

	require.Equal(t, "0", providerLpTokenBalanceAfter.Amount.String())

	providerUosmoBalanceAfter := osmosis.BankKeeper.GetBalance(ctx, reflect, "uosmo")

	require.Equal(t, "11999", providerUosmoBalanceAfter.Amount.String())

	providerUstarBalanceAfter := osmosis.BankKeeper.GetBalance(ctx, reflect, "ustar")

	require.Equal(t, "239999", providerUstarBalanceAfter.Amount.String())

}

// test setup for each run through the table test above
func prepareSwapState(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp) BaseState {
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

func executeCustom(t *testing.T, ctx sdk.Context, osmosis *app.OsmosisApp, contract sdk.AccAddress, sender sdk.AccAddress, msg wasmbindings.OsmosisMsg, funds []sdk.Coin) error {
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
	coins = funds

	contractKeeper := keeper.NewDefaultPermissionKeeper(osmosis.WasmKeeper)
	_, err = contractKeeper.Execute(ctx, contract, sender, reflectBz, coins)
	return err
}
