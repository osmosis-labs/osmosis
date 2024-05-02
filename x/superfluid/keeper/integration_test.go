package keeper_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	balancertypes "github.com/osmosis-labs/osmosis/v25/x/gamm/pool-models/balancer"
	minttypes "github.com/osmosis-labs/osmosis/v25/x/mint/types"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/types"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestSuite struct {
	KeeperTestSuite
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupTest() {
	s.KeeperTestSuite.SetupTest()

	// make pool creation fees be paid in the bond denom. Also make them low.
	poolmanagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
	poolmanagerParams.PoolCreationFee = sdk.NewCoins(sdk.NewInt64Coin(s.App.StakingKeeper.BondDenom(s.Ctx), 1))
	s.App.PoolManagerKeeper.SetParams(s.Ctx, poolmanagerParams)
}

func createPoolMsgGen(sender sdk.AccAddress, assets sdk.Coins) *balancertypes.MsgCreateBalancerPool {
	if len(assets) != 2 {
		panic("baseCreatePoolMsg requires 2 assets")
	}
	poolAssets := []balancertypes.PoolAsset{
		{
			Weight: osmomath.NewInt(1),
			Token:  assets[0],
		},
		{
			Weight: osmomath.NewInt(1),
			Token:  assets[1],
		},
	}

	poolParams := &balancertypes.PoolParams{
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.ZeroDec(),
	}

	msg := &balancertypes.MsgCreateBalancerPool{
		Sender:             sender.String(),
		PoolAssets:         poolAssets,
		PoolParams:         poolParams,
		FuturePoolGovernor: "",
	}

	return msg
}

func (s *TestSuite) mintToAccount(amount osmomath.Int, denom string, acc sdk.AccAddress) {
	err := s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(denom, amount)))
	s.Require().NoError(err)
	// send the coins to user1
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc, sdk.NewCoins(sdk.NewCoin(denom, amount)))
	s.Require().NoError(err)
}

func (s *TestSuite) TestGammSuperfluid() {
	s.SetupTest()
	//addr1 := sdk.AccAddress(pk1.Address())

	// denoms
	btcDenom := "btc" // Asset to superfluid stake
	bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)

	// accounts
	// pool creator
	lpKey := ed25519.GenPrivKey().PubKey()
	lpAddr := sdk.AccAddress(lpKey.Address())
	userKey := ed25519.GenPrivKey().PubKey()
	userAddr := sdk.AccAddress(userKey.Address())

	osmoPoolAmount := sdk.NewInt(1_000_000_000_000)
	btcPoolAmount := sdk.NewInt(10_000_000_000)
	// default bond denom

	// mint necessary tokens
	s.mintToAccount(btcPoolAmount, btcDenom, lpAddr)
	s.mintToAccount(osmoPoolAmount.Mul(osmomath.NewInt(2)), bondDenom, lpAddr)
	s.mintToAccount(sdk.NewInt(100_000_000), bondDenom, userAddr)
	s.mintToAccount(sdk.NewInt(1_000_000), btcDenom, userAddr)

	nextPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) // the pool id we'll create

	// create an bondDenom/btcDenom pool
	createPoolMsg := createPoolMsgGen(
		lpAddr,
		sdk.NewCoins(sdk.NewCoin(btcDenom, btcPoolAmount), sdk.NewCoin(bondDenom, osmoPoolAmount)),
	)

	_, err := s.RunMsg(createPoolMsg)
	s.Require().NoError(err)
	gammToken := fmt.Sprintf("gamm/pool/%d", nextPoolId)

	// get twap
	//price, err := s.App.TwapKeeper.GetArithmeticTwapToNow(s.Ctx, 1, btcDenom, bondDenom, s.Ctx.BlockTime())
	//s.Require().NoError(err)
	//fmt.Println("price", price)

	// Creating a native type without a pool should fail
	//err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: btcDenom, AssetType: types.SuperfluidAssetTypeNative})
	//s.Require().Error(err)

	// Add btcDenom as an allowed superfluid asset
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: gammToken, AssetType: types.SuperfluidAssetTypeLPShare})
	s.Require().NoError(err)

	balances := s.App.BankKeeper.GetAllBalances(s.Ctx, lpAddr)
	fmt.Println("balances", balances)

	// superfluid stake btcDenom
	validator := s.App.StakingKeeper.GetAllValidators(s.Ctx)[0]
	delegateMsg := &types.MsgLockAndSuperfluidDelegate{
		Sender:  lpAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(gammToken, sdk.NewInt(1000000000000000000))),
		ValAddr: validator.GetOperator().String(),
	}
	_, err = s.RunMsg(delegateMsg)
	s.Require().NoError(err)

	// Run epoch
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(2*time.Hour + time.Second))
	s.EndBlock()
	s.BeginNewBlock(true)
	//s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{})
	//s.App.BeginBlocker(s.Ctx, abci.RequestBeginBlock{})

	// TODO: How do I check distribution happened properly?
}
