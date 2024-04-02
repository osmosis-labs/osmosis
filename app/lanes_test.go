package app_test

import (
	"testing"

	"github.com/osmosis-labs/osmosis/v24/app"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// TestWithdrawStakingRewardsMatchHandler
func TestWithdrawStakingRewardsMatchHandler(t *testing.T) {
	txConfig := app.GetEncodingConfig().TxConfig
	handler := app.WithdrawStakingRewardsMatchHandler()
	t.Run("test non WithdrawStakingRewards tx", func(t *testing.T) {
		msg := banktypes.MsgSend{}
		fac := txConfig.NewTxBuilder()

		require.NoError(t, fac.SetMsgs([]sdk.Msg{&msg}...))
		require.False(t, handler(sdk.Context{}, fac.GetTx()))
	})

	t.Run("test WithdrawStakingRewards tx w/ other msgs", func(t *testing.T) {
		msg := banktypes.MsgSend{}
		fac := txConfig.NewTxBuilder()

		require.NoError(t, fac.SetMsgs([]sdk.Msg{&msg, &distrtypes.MsgWithdrawDelegatorReward{}}...))
		require.False(t, handler(sdk.Context{}, fac.GetTx()))
	})

	t.Run("test WithdrawStakingRewards tx w/ single msg", func(t *testing.T) {
		fac := txConfig.NewTxBuilder()

		require.NoError(t, fac.SetMsgs([]sdk.Msg{&distrtypes.MsgWithdrawDelegatorReward{}}...))
		require.True(t, handler(sdk.Context{}, fac.GetTx()))
	})
}
