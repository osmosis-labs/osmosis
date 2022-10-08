package balancer_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	appParams "github.com/osmosis-labs/osmosis/v12/app/params"
	balancer "github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func TestMsgCreateBalancerPool(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")
	_, invalidAccErr := sdk.AccAddressFromBech32(invalidAddr.String())

	createMsg := func(after func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
		testPoolAsset := []balancer.PoolAsset{
			{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("test", sdk.NewInt(100)),
			},
			{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin("test2", sdk.NewInt(100)),
			},
		}

		poolParams := &balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}

		msg := &balancer.MsgCreateBalancerPool{
			Sender:             addr1,
			PoolParams:         poolParams,
			PoolAssets:         testPoolAsset,
			FuturePoolGovernor: "",
		}

		return after(*msg)
	}

	default_msg := createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
		// Do nothing
		return msg
	})

	require.Equal(t, default_msg.Route(), types.RouterKey)
	require.Equal(t, default_msg.Type(), "create_balancer_pool")
	signers := default_msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1)

	tests := []struct {
		name       string
		msg        balancer.MsgCreateBalancerPool
		expectErr error
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				// Do nothing
				return msg
			}),
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectErr: sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", invalidAccErr),
		},
		{
			name: "has no PoolAsset",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets = nil
				return msg
			}),
			expectErr: types.ErrTooFewPoolAssets,
		},
		{
			name: "has no PoolAsset2",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets = []balancer.PoolAsset{}
				return msg
			}),
			expectErr: types.ErrTooFewPoolAssets,
		},
		{
			name: "has one Pool Asset",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets = []balancer.PoolAsset{
					msg.PoolAssets[0],
				}
				return msg
			}),
			expectErr: types.ErrTooFewPoolAssets,
		},
		{
			name: "has the PoolAsset that includes 0 weight",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Weight = sdk.NewInt(0)
				return msg
			}),
			expectErr: sdkerrors.Wrap(types.ErrNotPositiveWeight, sdk.ZeroInt().String()),
		},
		{
			name: "has a PoolAsset that includes a negative weight",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Weight = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(types.ErrNotPositiveWeight, sdk.NewInt(-10).String()),
		},
		{
			name: "has a PoolAsset that includes a negative weight",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Weight = sdk.NewInt(-10)
				return msg
			}),
			expectErr: sdkerrors.Wrap(types.ErrNotPositiveWeight, sdk.NewInt(-10).String()),
		},
		{
			name: "has a PoolAsset that includes a zero coin",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Token = sdk.NewCoin("test1", sdk.NewInt(0))
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, sdk.Coin{"test1", sdk.NewInt(0)}.String()),
		},
		{
			name: "has a PoolAsset that includes a negative coin",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Token = sdk.Coin{
					Denom:  "test1",
					Amount: sdk.NewInt(-10),
				}
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, sdk.Coin{"test1", sdk.NewInt(-10)}.String()),
		},
		{
			name: "negative swap fee with zero exit fee",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolParams = &balancer.PoolParams{
					SwapFee: sdk.NewDecWithPrec(-1, 2),
					ExitFee: sdk.NewDecWithPrec(0, 0),
				}
				return msg
			}),
			expectErr: types.ErrNegativeSwapFee,
		},
		{
			name: "invalid governor",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.FuturePoolGovernor = "invalid_cosmos_address"
				return msg
			}),
			expectErr: sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid future governor: %s", "invalid_cosmos_address")),
		},
		{
			name: "valid governor: lptoken and lock",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.FuturePoolGovernor = "lptoken,1000h"
				return msg
			}),
		},
		{
			name: "valid governor: just lock duration for pool token",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.FuturePoolGovernor = "1000h"
				return msg
			}),
		},
		{
			name: "valid governor: address",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.FuturePoolGovernor = "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
				return msg
			}),
		},
		{
			name: "zero swap fee, zero exit fee",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolParams = &balancer.PoolParams{
					ExitFee: sdk.NewDecWithPrec(0, 0),
					SwapFee: sdk.NewDecWithPrec(0, 0),
				}
				return msg
			}),
		},
		{
			name: "too large of a weight",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Weight = sdk.NewInt(1 << 21)
				return msg
			}),
			expectErr: sdkerrors.Wrap(types.ErrWeightTooLarge, sdk.NewInt(1 << 21).String()),
		},
		// {
		// 	name: "Create an LBP",
		// 	msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
		// 		msg.PoolParams.SmoothWeightChangeParams = &SmoothWeightChangeParams{
		// 			StartTime: time.Now(),
		// 			Duration:  time.Hour,
		// 			TargetPoolWeights: []PoolAsset{
		// 				{
		// 					Weight: sdk.NewInt(200),
		// 					Token:  sdk.NewCoin("test", sdk.NewInt(1)),
		// 				},
		// 				{
		// 					Weight: sdk.NewInt(50),
		// 					Token:  sdk.NewCoin("test2", sdk.NewInt(1)),
		// 				},
		// 			},
		// 		}
		// 		return msg
		// 	}),
		// 	expectPass: true,
		// },
	}

	for _, test := range tests {
		err := test.msg.ValidateBasic()
		if test.expectErr == nil {
			require.NoError(t, err, "test: %v", test.name)
		} else {
			require.Error(t, err, "test: %v", test.name)
			require.ErrorAs(t, test.expectErr, &err)
		}
	}
}
