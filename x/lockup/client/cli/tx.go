package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(cmd, NewLockTokensCmd)
	osmocli.AddTxCmd(cmd, NewBeginUnlockingAllCmd)
	osmocli.AddTxCmd(cmd, NewBeginUnlockByIDCmd)
	osmocli.AddTxCmd(cmd, NewForceUnlockByIdCmd)
	osmocli.AddTxCmd(cmd, NewSetRewardReceiverAddress)

	return cmd
}

func NewLockTokensCmd() (*osmocli.TxCliDesc, *types.MsgLockTokens) {
	return &osmocli.TxCliDesc{
		Use:   "lock-tokens",
		Short: "lock tokens into lockup pool from user account",
		CustomFlagOverrides: map[string]string{
			"duration": FlagDuration,
		},
		Flags: osmocli.FlagDesc{RequiredFlags: []*pflag.FlagSet{FlagSetLockTokens()}},
	}, &types.MsgLockTokens{}
}

// TODO: We should change the Use string to be unlock-all
func NewBeginUnlockingAllCmd() (*osmocli.TxCliDesc, *types.MsgBeginUnlockingAll) {
	return &osmocli.TxCliDesc{
		Use:   "begin-unlock-tokens",
		Short: "begin unlock not unlocking tokens from lockup pool for sender",
	}, &types.MsgBeginUnlockingAll{}
}

// NewBeginUnlockByIDCmd unlocks individual period lock by ID.
func NewBeginUnlockByIDCmd() (*osmocli.TxCliDesc, *types.MsgBeginUnlocking) {
	return &osmocli.TxCliDesc{
		Use:   "begin-unlock-by-id",
		Short: "begin unlock individual period lock by ID",
		CustomFlagOverrides: map[string]string{
			"coins": FlagAmount,
		},
		Flags: osmocli.FlagDesc{OptionalFlags: []*pflag.FlagSet{FlagSetUnlockTokens()}},
	}, &types.MsgBeginUnlocking{}
}

// NewForceUnlockByIdCmd force unlocks individual period lock by ID if proper permissions exist.
func NewForceUnlockByIdCmd() (*osmocli.TxCliDesc, *types.MsgForceUnlock) {
	return &osmocli.TxCliDesc{
		Use:   "force-unlock-by-id",
		Short: "force unlocks individual period lock by ID",
		Long:  "force unlocks individual period lock by ID. if no amount provided, entire lock is unlocked",
		CustomFlagOverrides: map[string]string{
			"coins": FlagAmount,
		},
		Flags: osmocli.FlagDesc{OptionalFlags: []*pflag.FlagSet{FlagSetUnlockTokens()}},
	}, &types.MsgForceUnlock{}
}

// NewSetRewardReceiverAddress sets the reward receiver address.
func NewSetRewardReceiverAddress() (*osmocli.TxCliDesc, *types.MsgSetRewardReceiverAddress) {
	return &osmocli.TxCliDesc{
		Use:   "set-reward-receiver-address",
		Short: "sets reward receiver address for the designated lock id",
		Long:  "sets reward receiver address for the designated lock id",
	}, &types.MsgSetRewardReceiverAddress{}
}
