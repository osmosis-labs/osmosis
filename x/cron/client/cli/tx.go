package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v29/x/cron/types"
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

	cmd.AddCommand(CmdRegisterCron(),
		CmdUpdateCronJob(),
		CmdDeleteCronJob(),
		CmdToggleCronJob())

	return cmd
}

func CmdRegisterCron() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-cron [name] [description] [contract_address] [json_msg]",
		Short: "Register New Cron",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgRegisterCron(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				args[2],
				args[3],
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpdateCronJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-cron-job [id] [contract_address] [json_msg]",
		Short: "Update cron job",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cronID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("cron-id '%s' not a valid uint", args[0])
			}
			msg := types.NewMsgUpdateCronJob(
				clientCtx.GetFromAddress().String(),
				cronID,
				args[1],
				args[2],
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdDeleteCronJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-cron-job [id] [contract_address]",
		Short: "Delete cron job of a contract",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cronID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("cron-id '%s' not a valid uint", args[0])
			}
			msg := types.NewMsgDeleteCronJob(
				clientCtx.GetFromAddress().String(),
				cronID,
				args[1],
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdToggleCronJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "toggle-cron-job [id]",
		Short: "Toggle cron job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cronID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("cron-id '%s' not a valid uint", args[0])
			}
			msg := types.NewMsgToggleCronJob(
				clientCtx.GetFromAddress().String(),
				cronID,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
