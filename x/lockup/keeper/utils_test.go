package keeper

import (
	"bytes"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/stretchr/testify/require"
)

func TestCombineKeys(t *testing.T) {
	key1 := []byte{0x11}
	key2 := []byte{0x12}
	key3 := []byte{0x13}
	key := combineKeys(key1, key2, key3)
	require.Len(t, key, 3+2) // 2 is separator
	require.Equal(t, key[0], key1[0])
	require.Equal(t, key[1], types.KeyIndexSeparator[0])
	require.Equal(t, key[2], key2[0])
	require.Equal(t, key[3], types.KeyIndexSeparator[0])
	require.Equal(t, key[4], key3[0])
}

func TestFindIndex(t *testing.T) {
	IDs := []uint64{1, 2, 3, 4, 5}
	require.Equal(t, findIndex(IDs, 1), 0)
	require.Equal(t, findIndex(IDs, 3), 2)
	require.Equal(t, findIndex(IDs, 5), 4)
	require.Equal(t, findIndex(IDs, 6), -1)
}

func TestRemoveValue(t *testing.T) {
	IDs := []uint64{1, 2, 3, 4, 5}
	IDs, index1 := removeValue(IDs, 5)
	require.Len(t, IDs, 4)
	require.Equal(t, index1, 4)
	IDs, index2 := removeValue(IDs, 3)
	require.Len(t, IDs, 3)
	require.Equal(t, index2, 2)
	IDs, index3 := removeValue(IDs, 1)
	require.Len(t, IDs, 2)
	require.Equal(t, index3, 0)
	IDs, index4 := removeValue(IDs, 6)
	require.Len(t, IDs, 2)
	require.Equal(t, index4, -1)
}

func TestGetTimeKey(t *testing.T) {
	now := time.Now()
	timeKey := getTimeKey(now)
	require.True(t, bytes.HasPrefix(timeKey, types.KeyPrefixTimestamp))
	require.True(t, bytes.HasSuffix(timeKey, sdk.FormatTimeBytes(now)))
}

func TestGetDurationKey(t *testing.T) {
	durationKey := getDurationKey(time.Second)
	require.True(t, bytes.HasPrefix(durationKey, types.KeyPrefixDuration))
	require.True(t, bytes.HasSuffix(durationKey, sdk.Uint64ToBigEndian(uint64(time.Second))))
	durationKeyNeg := getDurationKey(-time.Second)
	require.True(t, bytes.HasPrefix(durationKeyNeg, types.KeyPrefixDuration))
	require.True(t, bytes.HasSuffix(durationKeyNeg, sdk.Uint64ToBigEndian(0)))
}

func TestLockRefKeys(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	// empty address and 1 coin
	lock1 := types.NewPeriodLock(1, sdk.AccAddress{}, time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	_, err := lockRefKeys(lock1)
	require.Error(t, err)
	// empty address and 2 coins
	lock2 := types.NewPeriodLock(1, sdk.AccAddress{}, time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("atom", 1)})
	_, err = lockRefKeys(lock2)
	require.Error(t, err)

	// not empty address and 1 coin
	lock3 := types.NewPeriodLock(1, addr1, time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	keys3, err := lockRefKeys(lock3)
	require.Len(t, keys3, 8)
	// not empty address and empty coin
	lock4 := types.NewPeriodLock(1, addr1, time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	keys4, err := lockRefKeys(lock4)
	require.Len(t, keys4, 8)
	// not empty address and 2 coins
	lock5 := types.NewPeriodLock(1, addr1, time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("atom", 1)})
	keys5, err := lockRefKeys(lock5)
	require.Len(t, keys5, 12)
}
