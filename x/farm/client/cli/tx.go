package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/c-osmosis/osmosis/x/farm/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Farm transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewAllocateAssetsTxCmd())
	txCmd.AddCommand(NewWithdrawRewardsMsg())

	return txCmd
}

func NewAllocateAssetsTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allocate [from_key_or_address] [farm_id] [amount]",
		Short: "Allow external accounts to provide additional incentives by adding assets to specific farm reward pools",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			farmId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			assets, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgAllocateAssets(
				clientCtx.GetFromAddress(),
				farmId,
				assets,
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

func NewWithdrawRewardsMsg() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards [farm_id]",
		Short: "Withdraw rewards from specific farm",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			farmId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawRewards(
				clientCtx.GetFromAddress(),
				farmId,
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
