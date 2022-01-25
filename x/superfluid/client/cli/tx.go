package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/superfluid/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewSuperfluidDelegateCmd(),
		NewSuperfluidUndelegateCmd(),
		NewSuperfluidRedelegateCmd(),
	)

	return cmd
}

// NewSuperfluidDelegateCmd broadcast MsgSuperfluidDelegate
func NewSuperfluidDelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate [lock_id] [val_addr] [flags]",
		Short: "superfluid delegate a lock to a validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			lockId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgSuperfluidDelegate(
				clientCtx.GetFromAddress(),
				uint64(lockId),
				valAddr,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSuperfluidUndelegateCmd broadcast MsgSuperfluidUndelegate
func NewSuperfluidUndelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate [lock_id] [flags]",
		Short: "superfluid undelegate a lock from a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			lockId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSuperfluidUndelegate(
				clientCtx.GetFromAddress(),
				uint64(lockId),
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSuperfluidRedelegateCmd broadcast MsgSuperfluidRedelegate
func NewSuperfluidRedelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate [lock_id] [val_addr] [flags]",
		Short: "superfluid redelegate a lock to a new validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			lockId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgSuperfluidRedelegate(
				clientCtx.GetFromAddress(),
				uint64(lockId),
				valAddr,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
