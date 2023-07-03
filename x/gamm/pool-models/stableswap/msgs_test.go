package stableswap_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appParams "github.com/osmosis-labs/osmosis/v16/app/params"
	stableswap "github.com/osmosis-labs/osmosis/v16/x/gamm/pool-models/stableswap"
	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
)

func baseCreatePoolMsgGen(sender sdk.AccAddress) *stableswap.MsgCreateStableswapPool {
	testPoolAsset := sdk.Coins{
		sdk.NewCoin("atom", sdk.NewInt(100)),
		sdk.NewCoin("osmo", sdk.NewInt(100)),
	}

	poolParams := &stableswap.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.ZeroDec(),
	}

	msg := &stableswap.MsgCreateStableswapPool{
		Sender:               sender.String(),
		PoolParams:           poolParams,
		InitialPoolLiquidity: testPoolAsset,
		ScalingFactors:       []uint64{1, 1},
		FuturePoolGovernor:   "",
	}

	return msg
}

func TestMsgCreateStableswapPoolValidateBasic(t *testing.T) {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())
	invalidAddr := sdk.AccAddress("invalid")

	default_msg := baseCreatePoolMsgGen(addr1)
	updateMsg := func(f func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
		m := *baseCreatePoolMsgGen(addr1)
		return f(m)
	}

	require.Equal(t, default_msg.Route(), types.RouterKey)
	require.Equal(t, default_msg.Type(), "create_stableswap_pool")
	signers := default_msg.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), addr1.String())

	tests := []struct {
		name       string
		msg        stableswap.MsgCreateStableswapPool
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				// Do nothing
				return msg
			}),
			expectPass: true,
		},
		{
			name: "no scaling factors",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.ScalingFactors = []uint64{}
				return msg
			}),
			expectPass: true,
		},
		{
			name: "invalid sender",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.Sender = invalidAddr.String()
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has nil InitialPoolLiquidity ",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = nil
				return msg
			}),
			expectPass: false,
		},
		{
			name: "has one coin in InitialPoolLiquidity",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("osmo", sdk.NewInt(100)),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "have assets in excess of cap",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("akt", sdk.NewInt(100)),
					sdk.NewCoin("atom", sdk.NewInt(100)),
					sdk.NewCoin("band", sdk.NewInt(100)),
					sdk.NewCoin("evmos", sdk.NewInt(100)),
					sdk.NewCoin("juno", sdk.NewInt(100)),
					sdk.NewCoin("osmo", sdk.NewInt(100)),
					sdk.NewCoin("regen", sdk.NewInt(100)),
					sdk.NewCoin("usdt", sdk.NewInt(100)),
					sdk.NewCoin("usdc", sdk.NewInt(100)),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative spread factor with zero exit fee",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.PoolParams = &stableswap.PoolParams{
					SwapFee: sdk.NewDecWithPrec(-1, 2),
					ExitFee: sdk.NewDecWithPrec(0, 0),
				}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "scaling factors with invalid length",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.ScalingFactors = []uint64{1, 2, 3}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid governor",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "invalid_cosmos_address"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid governor : len governor > 2",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "lptoken,1000h,invalid_cosmos_address"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid governor : len governor > 2",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "lptoken,1000h,invalid_cosmos_address"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid governor: err when parse duration ",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "lptoken, invalid_duration"
				return msg
			}),
			expectPass: false,
		},
		{
			name: "valid governor: just lock duration for pool token",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "1000h"
				return msg
			}),
			expectPass: true,
		},
		{
			name: "valid governor: address",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.FuturePoolGovernor = "osmo1fqlr98d45v5ysqgp6h56kpujcj4cvsjnjq9nck"
				return msg
			}),
			expectPass: true,
		},
		{
			name: "valid governor: address",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.FuturePoolGovernor = ""
				return msg
			}),
			expectPass: true,
		},
		{
			name: "zero spread factor, zero exit fee",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.PoolParams = &stableswap.PoolParams{
					ExitFee: sdk.NewDecWithPrec(0, 0),
					SwapFee: sdk.NewDecWithPrec(0, 0),
				}
				return msg
			}),
			expectPass: true,
		},
		{
			name: "multi assets pool",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("atom", sdk.NewInt(100)),
					sdk.NewCoin("osmo", sdk.NewInt(100)),
					sdk.NewCoin("usdc", sdk.NewInt(100)),
					sdk.NewCoin("usdt", sdk.NewInt(100)),
				}
				msg.ScalingFactors = []uint64{1, 1, 1, 1}
				return msg
			}),
			expectPass: true,
		},
		{
			name: "post-scaled asset amount less than 1",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("osmo", sdk.NewInt(100)),
					sdk.NewCoin("atom", sdk.NewInt(100)),
					sdk.NewCoin("usdt", sdk.NewInt(100)),
					sdk.NewCoin("usdc", sdk.NewInt(100)),
				}
				msg.ScalingFactors = []uint64{1000, 1, 1, 1}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "max asset amounts",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("akt", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("atom", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("band", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("juno", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("osmo", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("regen", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("usdc", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("usdt", types.StableswapMaxScaledAmtPerAsset),
				}
				msg.ScalingFactors = []uint64{1, 1, 1, 1, 1, 1, 1, 1}
				return msg
			}),
			expectPass: true,
		},
		{
			name: "greater than max post-scaled amount with regular scaling factors",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("osmo", types.StableswapMaxScaledAmtPerAsset.Add(sdk.OneInt())),
					sdk.NewCoin("atom", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("usdt", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("usdc", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("juno", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("akt", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("regen", types.StableswapMaxScaledAmtPerAsset),
					sdk.NewCoin("band", types.StableswapMaxScaledAmtPerAsset),
				}
				msg.ScalingFactors = []uint64{1, 1, 1, 1, 1, 1, 1, 1}
				return msg
			}),
			expectPass: false,
		},
		{
			name: "100B token 8-asset pool using large scaling factors (6 decimal precision per asset)",
			msg: updateMsg(func(msg stableswap.MsgCreateStableswapPool) stableswap.MsgCreateStableswapPool {
				msg.InitialPoolLiquidity = sdk.Coins{
					sdk.NewCoin("akt", sdk.NewInt(100_000_000_000_000_000)),
					sdk.NewCoin("atom", sdk.NewInt(100_000_000_000_000_000)),
					sdk.NewCoin("band", sdk.NewInt(100_000_000_000_000_000)),
					sdk.NewCoin("juno", sdk.NewInt(100_000_000_000_000_000)),
					sdk.NewCoin("osmo", sdk.NewInt(100_000_000_000_000_000)),
					sdk.NewCoin("regen", sdk.NewInt(100_000_000_000_000_000)),
					sdk.NewCoin("usdc", sdk.NewInt(100_000_000_000_000_000)),
					sdk.NewCoin("usdt", sdk.NewInt(100_000_000_000_000_000)),
				}
				msg.ScalingFactors = []uint64{10000000, 10000000, 10000000, 10000000, 10000000, 10000000, 10000000, 10000000}
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

func (suite *TestSuite) TestMsgCreateStableswapPool() {
	suite.SetupTest()

	var (
		validParams           = &stableswap.PoolParams{SwapFee: sdk.NewDecWithPrec(1, 2), ExitFee: sdk.ZeroDec()}
		validInitialLiquidity = sdk.NewCoins(sdk.NewCoin("usdc", sdk.NewInt(1_000_000)), sdk.NewCoin("usdt", sdk.NewInt(2_000_000)))
		validScalingFactors   = []uint64{1, 1}
		invalidScalingFactors = []uint64{1, 1, 1}
	)

	tests := map[string]struct {
		msg         stableswap.MsgCreateStableswapPool
		poolId      uint64
		expectError bool
	}{
		"basic success test": {
			msg: stableswap.MsgCreateStableswapPool{
				Sender:                  suite.TestAccs[0].String(),
				PoolParams:              validParams,
				InitialPoolLiquidity:    validInitialLiquidity,
				ScalingFactors:          validScalingFactors,
				FuturePoolGovernor:      "",
				ScalingFactorController: "",
			},
			poolId: 1,
		},
		"error test - more scaling factors than initial liquidity": {
			msg: stableswap.MsgCreateStableswapPool{
				Sender:                  suite.TestAccs[0].String(),
				PoolParams:              validParams,
				InitialPoolLiquidity:    validInitialLiquidity,
				ScalingFactors:          invalidScalingFactors,
				FuturePoolGovernor:      "",
				ScalingFactorController: "",
			},
			poolId:      2,
			expectError: true,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			pool, err := tc.msg.CreatePool(suite.Ctx, 1)

			if tc.expectError {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)

			suite.Require().Equal(tc.poolId, pool.GetId())

			cfmmPool, ok := pool.(types.CFMMPoolI)
			suite.Require().True(ok)

			suite.Require().Equal(tc.msg.InitialPoolLiquidity, cfmmPool.GetTotalPoolLiquidity(suite.Ctx))
			suite.Require().Equal(types.InitPoolSharesSupply, cfmmPool.GetTotalShares())
		})
	}
}
