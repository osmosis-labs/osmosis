package superfluid_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v25/app/params"
	balancertypes "github.com/osmosis-labs/osmosis/v25/x/gamm/pool-models/balancer"
	minttypes "github.com/osmosis-labs/osmosis/v25/x/mint/types"
	"github.com/osmosis-labs/osmosis/v25/x/superfluid/types"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/osmosis-labs/osmosis/v25/app/apptesting"
)

type TestSuite struct {
	apptesting.KeeperTestHelper
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) SetupTest() {
	suite.Setup()
}

func createPoolMsgGen(sender sdk.AccAddress, assets sdk.Coins) *balancertypes.MsgCreateBalancerPool {
	if len(assets) != 2 {
		panic("baseCreatePoolMsg requires 2 assets")
	}
	poolAssets := []balancertypes.PoolAsset{
		balancertypes.PoolAsset{
			Weight: osmomath.NewInt(1),
			Token:  assets[0],
		},
		balancertypes.PoolAsset{
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

func (s *TestSuite) TestNativeSuperfluid() {
	s.SetupTest()
	//addr1 := sdk.AccAddress(pk1.Address())

	// denoms
	bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)
	stakeDenom := "btc"                     // Asset to superfluid stake
	osmoDenom := appparams.DefaultBondDenom // used for paying pool creation fees

	// accounts
	// pool creator
	lpKey := ed25519.GenPrivKey().PubKey()
	lpAddr := sdk.AccAddress(lpKey.Address())
	userKey := ed25519.GenPrivKey().PubKey()
	userAddr := sdk.AccAddress(userKey.Address())

	bondPoolAmount := sdk.NewInt(1_000_000_000_000)
	stakePoolAmount := sdk.NewInt(10_000_000_000)
	// default bond denom

	// mint necessary tokens
	s.mintToAccount(stakePoolAmount, stakeDenom, lpAddr)
	s.mintToAccount(stakePoolAmount, osmoDenom, lpAddr)
	s.mintToAccount(bondPoolAmount.Mul(osmomath.NewInt(2)), bondDenom, lpAddr)
	s.mintToAccount(sdk.NewInt(100_000_000), bondDenom, userAddr)
	s.mintToAccount(sdk.NewInt(1_000_000), bondDenom, userAddr)

	nextPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) // the pool id we'll create

	// create an bondDenom/stakeDenom pool
	createPoolMsg := createPoolMsgGen(
		lpAddr,
		sdk.NewCoins(sdk.NewCoin(stakeDenom, stakePoolAmount), sdk.NewCoin(bondDenom, bondPoolAmount)),
	)

	_, err := s.RunMsg(createPoolMsg)
	s.Require().NoError(err)

	//s.EndBlock()
	//s.BeginNewBlock(false)

	// get twap
	//price, err := s.App.TwapKeeper.GetArithmeticTwapToNow(s.Ctx, 1, stakeDenom, bondDenom, s.Ctx.BlockTime())
	//s.Require().NoError(err)
	//fmt.Println("price", price)

	// Creating without a pool should fail
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: stakeDenom})
	s.Require().Error(err)

	// Add stakeDenom as an allowed superfluid asset
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{Denom: stakeDenom, PricePoolId: nextPoolId})
	s.Require().NoError(err)

}
