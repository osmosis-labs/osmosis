package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgCreatePool(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1, err := sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	require.NoError(t, err)
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgCreatePool) MsgCreatePool) MsgCreatePool {
		properMsg := MsgCreatePool{
			Sender: addr1,
			PoolParams: PoolParams{
				Lock:    false,
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDecWithPrec(1, 2),
			},
			Records: []Record{
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test", sdk.NewInt(100)),
				},
				{
					Weight: sdk.NewInt(100),
					Token:  sdk.NewCoin("test2", sdk.NewInt(100)),
				},
			},
		}

		return after(properMsg)
	}

	tests := []struct {
		name       string
		msg        MsgCreatePool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has no record",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has no record2",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records = []Record{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has one record",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records = []Record{
					msg.Records[0],
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes 0 weight",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Weight = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes the negative weight",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Weight = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes the negative weight",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Weight = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes the zero coin",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Token = sdk.NewCoin("test1", sdk.NewInt(0))
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the record that includes the negative coin",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.Records[0].Token = sdk.Coin{
					Denom:  "test1",
					Amount: sdk.NewInt(-10),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "locked pool",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.Lock = true
				return msg
			}),
			expectPass: false,
		},
		{
			name: "nagative swap fee",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.SwapFee = sdk.NewDecWithPrec(-1, 2)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "nagative exit fee",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.ExitFee = sdk.NewDecWithPrec(-1, 2)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero swap fee",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.SwapFee = sdk.NewDec(0)
				return msg
			}),
			expectPass: true,
		},
		{
			name: "zero exit fee",
			msg: createMsg(func(msg MsgCreatePool) MsgCreatePool {
				msg.PoolParams.ExitFee = sdk.NewDec(0)
				return msg
			}),
			expectPass: true,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgSwapExactAmountIn(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1, err := sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	require.NoError(t, err)
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn) MsgSwapExactAmountIn {
		properMsg := MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []SwapAmountInRoute{{
				PoolId:        0,
				TokenOutDenom: "test",
			}, {
				PoolId:        1,
				TokenOutDenom: "test2",
			}},
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			TokenOutMinAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	tests := []struct {
		name       string
		msg        MsgSwapExactAmountIn
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes = []SwapAmountInRoute{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.Routes[1].TokenOutDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom2",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenIn.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountIn) MsgSwapExactAmountIn {
				msg.TokenOutMinAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgSwapExactAmountOut(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1, err := sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	require.NoError(t, err)
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut) MsgSwapExactAmountOut {
		properMsg := MsgSwapExactAmountOut{
			Sender: addr1,
			Routes: []SwapAmountOutRoute{{
				PoolId:       0,
				TokenInDenom: "test",
			}, {
				PoolId:       1,
				TokenInDenom: "test2",
			}},
			TokenOut:         sdk.NewCoin("test", sdk.NewInt(100)),
			TokenInMaxAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	tests := []struct {
		name       string
		msg        MsgSwapExactAmountOut
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "empty routes2",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes = []SwapAmountOutRoute{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.Routes[1].TokenInDenom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Denom = "1"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount token",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount token",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenOut.Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount criteria",
			msg: createMsg(func(msg MsgSwapExactAmountOut) MsgSwapExactAmountOut {
				msg.TokenInMaxAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgJoinPool(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1, err := sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	require.NoError(t, err)
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgJoinPool) MsgJoinPool) MsgJoinPool {
		properMsg := MsgJoinPool{
			Sender:         addr1,
			PoolId:         1,
			ShareOutAmount: sdk.NewInt(10),
			TokenInMaxs:    sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}

		return after(properMsg)
	}

	tests := []struct {
		name       string
		msg        MsgJoinPool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "nagative requirement",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.ShareOutAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs[1].Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "nagative amount",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs[1].Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "'empty token max in' can pass",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs = nil
				return msg
			}),
			expectPass: true,
		},
		{
			name: "'empty token max in' can pass 2",
			msg: createMsg(func(msg MsgJoinPool) MsgJoinPool {
				msg.TokenInMaxs = sdk.Coins{}
				return msg
			}),
			expectPass: true,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func TestMsgExitPool(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1, err := sdk.Bech32ifyAddressBytes(sdk.Bech32PrefixAccAddr, pk1.Address().Bytes())
	require.NoError(t, err)
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgExitPool) MsgExitPool) MsgExitPool {
		properMsg := MsgExitPool{
			Sender:        addr1,
			PoolId:        1,
			ShareInAmount: sdk.NewInt(10),
			TokenOutMins:  sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(10)), sdk.NewCoin("test2", sdk.NewInt(20))),
		}
		return after(properMsg)
	}

	tests := []struct {
		name       string
		msg        MsgExitPool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "nagative requirement",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.ShareInAmount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins[1].Amount = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "nagative amount",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins[1].Amount = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "'empty token min out' can pass",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins = nil
				return msg
			}),
			expectPass: true,
		},
		{
			name: "'empty token min out' can pass 2",
			msg: createMsg(func(msg MsgExitPool) MsgExitPool {
				msg.TokenOutMins = sdk.Coins{}
				return msg
			}),
			expectPass: true,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}
