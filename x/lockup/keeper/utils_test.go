package keeper

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	lock1 := types.NewPeriodLock(1, sdk.AccAddress{}, sdk.AccAddress{}.String(), time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	_, err := lockRefKeys(lock1)
	require.Error(t, err)
	// empty address and 2 coins
	lock2 := types.NewPeriodLock(1, sdk.AccAddress{}, sdk.AccAddress{}.String(), time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("atom", 1)})
	_, err = lockRefKeys(lock2)
	require.Error(t, err)

	// not empty address and 1 coin
	lock3 := types.NewPeriodLock(1, addr1, addr1.String(), time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	keys3, err := lockRefKeys(lock3)
	require.NoError(t, err)
	require.Len(t, keys3, 8)
	// not empty address and empty coin
	lock4 := types.NewPeriodLock(1, addr1, addr1.String(), time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	keys4, err := lockRefKeys(lock4)
	require.NoError(t, err)
	require.Len(t, keys4, 8)
	// not empty address and 2 coins
	lock5 := types.NewPeriodLock(1, addr1, addr1.String(), time.Second, time.Now(), sdk.Coins{sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("atom", 1)})
	keys5, err := lockRefKeys(lock5)
	require.NoError(t, err)
	require.Len(t, keys5, 12)
}
