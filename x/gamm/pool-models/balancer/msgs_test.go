package balancer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v16/app/params"
	balancer "github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

func TestMsgCreateBalancerPool_ValidateBasic(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	invalidAddr := sdk.AccAddress("invalid")

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
			ExitFee: sdk.ZeroDec(),
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
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has no PoolAsset",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has no PoolAsset2",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets = []balancer.PoolAsset{}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has one Pool Asset",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets = []balancer.PoolAsset{
					msg.PoolAssets[0],
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has the PoolAsset that includes 0 weight",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Weight = sdk.NewInt(0)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has a PoolAsset that includes a negative weight",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Weight = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has a PoolAsset that includes a negative weight",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Weight = sdk.NewInt(-10)
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has a PoolAsset that includes a zero coin",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Token = sdk.NewCoin("test1", sdk.NewInt(0))
				return msg
			}),
			expectPass: false,
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
			expectPass: false,
		},
		{
			name: "negative spread factor with zero exit fee",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolParams = &balancer.PoolParams{
					SwapFee: sdk.NewDecWithPrec(-1, 2),
					ExitFee: sdk.NewDecWithPrec(0, 0),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid governor",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.FuturePoolGovernor = "invalid_cosmos_address"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid governor: lptoken and lock",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.FuturePoolGovernor = "lptoken,1000h"
				return msg
			}),
			expectPass: true,
		},
		{
			name: "valid governor: just lock duration for pool token",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.FuturePoolGovernor = "1000h"
				return msg
			}),
			expectPass: true,
		},
		{
			name: "valid governor: address",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.FuturePoolGovernor = "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
				return msg
			}),
			expectPass: true,
		},
		{
			name: "zero spread factor, zero exit fee",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolParams = &balancer.PoolParams{
					ExitFee: sdk.NewDecWithPrec(0, 0),
					SwapFee: sdk.NewDecWithPrec(0, 0),
				}
				return msg
			}),
			expectPass: true,
		},
		{
			name: "too large of a weight",
			msg: createMsg(func(msg balancer.MsgCreateBalancerPool) balancer.MsgCreateBalancerPool {
				msg.PoolAssets[0].Weight = sdk.NewInt(1 << 21)
				return msg
			}),
			expectPass: false,
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
		if test.expectPass {
			require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
		}
	}
}

func (s *KeeperTestSuite) TestMsgCreateBalancerPool() {
	tests := map[string]struct {
		msg         balancer.MsgCreateBalancerPool
		poolId      uint64
		expectError bool
	}{
		"basic success test": {
			msg: balancer.MsgCreateBalancerPool{
				Sender:             s.TestAccs[0].String(),
				PoolParams:         &balancer.PoolParams{SwapFee: sdk.NewDecWithPrec(1, 2), ExitFee: sdk.ZeroDec()},
				PoolAssets:         apptesting.DefaultPoolAssets,
				FuturePoolGovernor: "",
			},
			poolId: 1,
		},
		"error due to negative spread factor": {
			msg: balancer.MsgCreateBalancerPool{
				Sender:             s.TestAccs[0].String(),
				PoolParams:         &balancer.PoolParams{SwapFee: sdk.NewDecWithPrec(1, 2).Neg(), ExitFee: sdk.ZeroDec()},
				PoolAssets:         apptesting.DefaultPoolAssets,
				FuturePoolGovernor: "",
			},
			poolId:      2,
			expectError: true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			pool, err := tc.msg.CreatePool(s.Ctx, 1)

			if tc.expectError {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			s.Require().Equal(tc.poolId, pool.GetId())
			expectedPoolLiquidity := sdk.NewCoins()
			for _, asset := range tc.msg.PoolAssets {
				expectedPoolLiquidity = expectedPoolLiquidity.Add(asset.Token)
			}

			cfmmPool, ok := pool.(types.CFMMPoolI)
			s.Require().True(ok)

			s.Require().Equal(expectedPoolLiquidity, cfmmPool.GetTotalPoolLiquidity(s.Ctx))
			s.Require().Equal(types.InitPoolSharesSupply, cfmmPool.GetTotalShares())
		})
	}
}
