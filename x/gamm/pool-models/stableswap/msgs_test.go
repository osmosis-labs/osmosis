package stableswap

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appParams "github.com/osmosis-labs/osmosis/v11/app/params"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/types"
)

func TestMsgCreateStableswapPool(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg MsgCreateStableswapPool) MsgCreateStableswapPool) MsgCreateStableswapPool {
		testPoolAsset := sdk.Coins{
			sdk.NewCoin("osmo", sdk.NewInt(100)),
			sdk.NewCoin("atom", sdk.NewInt(100)),
		}

		poolParams := &PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}

		msg := &MsgCreateStableswapPool{
			Sender:               addr1,
			PoolParams:           poolParams,
			InitialPoolLiquidity: testPoolAsset,
			FuturePoolGovernor:   "",
		}

		return after(*msg)
	}

	default_msg := createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
		// Do nothing
		return msg
	})

	require.Equal(t, default_msg.Route(), types.RouterKey)
	require.Equal(t, default_msg.Type(), "create_stableswap_pool")
	signers := default_msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        MsgCreateStableswapPool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has nil InitialPoolLiquidity ",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has one coin in InitialPoolLiquidity",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("osmo", sdk.NewInt(100)),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has three coins in InitialPoolLiquidity",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("osmo", sdk.NewInt(100)),
					sdk.NewCoin("atom", sdk.NewInt(100)),
					sdk.NewCoin("usdt", sdk.NewInt(100)),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative swap fee with zero exit fee",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.PoolParams = &PoolParams{
					SwapFee: sdk.NewDecWithPrec(-1, 2),
					ExitFee: sdk.NewDecWithPrec(0, 0),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid governor",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "invalid_cosmos_address"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid governor : len governor > 2",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "lptoken,1000h,invalid_cosmos_address"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid governor : len governor > 2",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "lptoken,1000h,invalid_cosmos_address"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid governor: err when parse duration ",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "lptoken, invalid_duration"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid governor: just lock duration for pool token",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "1000h"
				return msg
			}),
			expectPass: true,
		},
		{
			name: "valid governor: address",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
				return msg
			}),
			expectPass: true,
		},
		{
			name: "valid governor: address",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.FuturePoolGovernor = ""
				return msg
			}),
			expectPass: true,
		},
		{
			name: "zero swap fee, zero exit fee",
			msg: createMsg(func(msg MsgCreateStableswapPool) MsgCreateStableswapPool {
				msg.PoolParams = &PoolParams{
					ExitFee: sdk.NewDecWithPrec(0, 0),
					SwapFee: sdk.NewDecWithPrec(0, 0),
				}
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
