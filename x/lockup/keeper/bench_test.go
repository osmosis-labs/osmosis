package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/osmosis-labs/osmosis/v8/app"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func benchmarkResetLogic(numLockups int, b *testing.B) {
	// b.ReportAllocs()
	b.StopTimer()

	blockStartTime := time.Now().UTC()
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: blockStartTime})

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	numAccts := 100
	numDenoms := 1

	denom := fmt.Sprintf("token%d", 0)

	// setup accounts with balances
	addrs := []sdk.AccAddress{}
	for i := 0; i < numAccts; i++ {
		addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		coins := sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000)}
		for j := 0; j < numDenoms; j++ {
			coins = coins.Add(sdk.NewInt64Coin(fmt.Sprintf("token%d", j), r.Int63n(100000000)))
		}
		_ = simapp.FundAccount(app.BankKeeper, ctx, addr, coins)
		app.AccountKeeper.SetAccount(ctx, authtypes.NewBaseAccount(addr, nil, 0, 0))
		addrs = append(addrs, addr)
	}

	// jump time to the future
	futureSecs := r.Intn(1 * 60 * 60 * 24 * 7)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(futureSecs) * time.Second))

	locks := make([]lockuptypes.PeriodLock, numLockups)
	// setup lockups
	for i := 0; i < numLockups; i++ {
		addr := addrs[r.Int()%numAccts]
		simCoins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(r.Int63n(100))))
		duration := time.Duration(r.Intn(1*60*60*24*7)) * time.Second
		lock := lockuptypes.NewPeriodLock(uint64(i+1), addr, duration, time.Time{}, simCoins)
		locks[i] = lock
	}

	b.StartTimer()
	b.ReportAllocs()
	// distribute coins from gauges to lockup owners
	_ = app.LockupKeeper.ResetAllLocks(ctx, locks)
}

func BenchmarkResetLogicMedium(b *testing.B) {
	benchmarkResetLogic(50000, b)
}
