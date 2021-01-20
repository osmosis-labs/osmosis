package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/c-osmosis/osmosis/x/farm/keeper"
	"github.com/c-osmosis/osmosis/x/farm/types"
)

const (
	testDenom = "test"
)

var (
	rewardPerBlock = sdk.NewInt(100)
)

func TestSimpleReward(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	keeper := keeper.NewKeeper(cdc, storeKey)

	farm, err := keeper.NewFarm(ctx, 0)
	require.NoError(t, err)

	acc1 := sdk.AccAddress{0x1}
	acc2 := sdk.AccAddress{0x2}
	acc3 := sdk.AccAddress{0x3}

	rewards, err := keeper.DepositShareToFarm(ctx, farm.FarmId, 1, acc1, sdk.NewInt(2))
	require.NoError(t, err)
	require.Equal(t, 0, len(rewards))

	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 1, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)
	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 2, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)
	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 3, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)

	// Until this, acc1 should have the 300test as rewards

	rewards, err = keeper.DepositShareToFarm(ctx, farm.FarmId, 4, acc2, sdk.NewInt(1))
	require.NoError(t, err)
	require.Equal(t, 0, len(rewards))

	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 4, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)
	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 5, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)
	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 6, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)

	// Until this, acc1 should have the 300test + (300 * 2 / 3 = 200)test as rewards
	// And, acc2 should have the 100test as rewards

	rewards, err = keeper.DepositShareToFarm(ctx, farm.FarmId, 7, acc3, sdk.NewInt(3))

	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 7, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)
	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 8, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)
	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 9, sdk.Coins{sdk.NewCoin(testDenom, rewardPerBlock)})
	require.NoError(t, err)

	// Until this, acc1 should have the 300test + (300 * 2 / 3 = 200)test + (300 * 2 / 6 = 100)test as rewards
	// And, acc2 should have the 100test + (300 / 6 = 50)test as rewards
	// And, acc3 should have the (300 / 2 = 150)test as rewards

	// Flush the current period, because the farmer can only the withdraw the rewards until the last period.
	err = keeper.AllocateAssetToFarm(ctx, farm.FarmId, 10, sdk.Coins{})
	require.NoError(t, err)

	acc1Rewards, err := keeper.WithdrawRewardsFromFarm(ctx, farm.FarmId, acc1)
	require.NoError(t, err)

	amount := acc1Rewards.AmountOf(testDenom)
	require.True(t, sdk.NewInt(600).Sub(amount).LTE(sdk.OneInt()))

	acc2Rewards, err := keeper.WithdrawRewardsFromFarm(ctx, farm.FarmId, acc2)
	require.NoError(t, err)

	amount = acc2Rewards.AmountOf(testDenom)
	require.True(t, sdk.NewInt(150).Sub(amount).LTE(sdk.OneInt()))

	acc3Rewards, err := keeper.WithdrawRewardsFromFarm(ctx, farm.FarmId, acc3)
	require.NoError(t, err)

	amount = acc3Rewards.AmountOf(testDenom)
	require.True(t, sdk.NewInt(150).Sub(amount).LTE(sdk.OneInt()))
}
