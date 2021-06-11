package keeper_test

import (
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/lockup/keeper"
)

// TODO: integrate to simulation

func (suite *KeeperTestSuite) TestBehaviours() {
	suite.SetupTest()

	denoms := make([]string, 10)
	for i := range denoms {
		denoms[i] = fmt.Sprintf("token%d", i)
	}

	accounts := make([]*keeper.BehaviourAccount, 20)
	for i := range accounts {
		address := sdk.AccAddress([]byte(fmt.Sprintf("testaddr%2d----------", i)))
		accounts[i] = keeper.NewBehaviourAccount(address)
		var coins sdk.Coins
		for _, denom := range denoms {
			if rand.Int()%2==0 {
				continue
			}
			coins = append(coins, sdk.NewInt64Coin(denom, rand.Int63n(100000000)))
		}
		suite.app.BankKeeper.SetBalances(suite.ctx, address, coins)
	}

	blockLimit := 100
	txLimit := 20
	blockTime := time.Second*time.Duration(5)
	for i := 0; i < blockLimit; i++ {
		for i := 0; i < txLimit; i++ {
			account := accounts[rand.Intn(len(accounts))]
			flip := rand.Intn(100)
			if flip < 80 {
				account.GenerateBehaviourLockToken(suite.ctx, suite.app.LockupKeeper, suite.app.BankKeeper, blockTime*time.Duration(blockLimit/2))(suite.ctx, suite.Suite)
			}
			if flip < 90 {
				account.GenerateBehaviourBeginUnlocking(suite.app.LockupKeeper)(suite.ctx, suite.Suite)
			}
			if flip < 100 {
				account.GenerateBehaviourUnlock(suite.app.LockupKeeper)(suite.ctx, suite.Suite)
			}
		}
		suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(blockTime))
	}
}
