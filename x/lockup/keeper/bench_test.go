package keeper_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
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

func benchmarkResetLogic(b *testing.B, numLockups int) {
	b.Helper()
	// b.ReportAllocs()
	b.StopTimer()

	blockStartTime := time.Now().UTC()
	app := app.Setup(false)
	ctx := app.BaseApp.NewContextLegacy(false, tmproto.Header{Height: 1, ChainID: "osmosis-1", Time: blockStartTime})

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
		_ = testutil.FundAccount(ctx, app.BankKeeper, addr, coins)
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
		simCoins := sdk.NewCoins(sdk.NewCoin(denom, osmomath.NewInt(r.Int63n(100))))
		duration := time.Duration(r.Intn(1*60*60*24*7)) * time.Second
		lock := lockuptypes.NewPeriodLock(uint64(i+1), addr, addr.String(), duration, time.Time{}, simCoins)
		locks[i] = lock
	}

	b.StartTimer()
	b.ReportAllocs()
	// distribute coins from gauges to lockup owners
	_ = app.LockupKeeper.InitializeAllLocks(ctx, locks)
}

func BenchmarkResetLogicMedium(b *testing.B) {
	benchmarkResetLogic(b, 50000)
}
