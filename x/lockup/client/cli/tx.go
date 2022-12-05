package cli

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	cmd.AddCommand(
		NewLockTokensCmd(),
		NewBeginUnlockingAllCmd(),
		NewBeginUnlockByIDCmd(),
		NewForceUnlockByIdCmd(),
	)

	return cmd
}

func NewLockTokensCmd() *cobra.Command {
	cmd := osmocli.BuildTxCli[*types.MsgLockTokens](&osmocli.TxCliDesc{
		Use:   "lock-tokens [tokens]",
		Short: "lock tokens into lockup pool from user account",
		CustomFlagOverrides: map[string]string{
			"duration": FlagDuration,
		},
	})

	cmd.Flags().AddFlagSet(FlagSetLockTokens())
	err := cmd.MarkFlagRequired(FlagDuration)
	if err != nil {
		panic(err)
	}
	return cmd
}

// TODO: We should change the Use string to be unlock-all
func NewBeginUnlockingAllCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgBeginUnlockingAll](&osmocli.TxCliDesc{
		Use:   "begin-unlock-tokens",
		Short: "begin unlock not unlocking tokens from lockup pool for sender",
	})
}

// NewBeginUnlockByIDCmd unlocks individual period lock by ID.
func NewBeginUnlockByIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "begin-unlock-by-id [id]",
		Short: "begin unlock individual period lock by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			coins := sdk.Coins(nil)
			amountStr, err := cmd.Flags().GetString(FlagAmount)
			if err != nil {
				return err
			}

			if amountStr != "" {
				coins, err = sdk.ParseCoinsNormalized(amountStr)
				if err != nil {
					return err
				}
			}

			msg := types.NewMsgBeginUnlocking(
				clientCtx.GetFromAddress(),
				uint64(id),
				coins,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetUnlockTokens())

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewForceUnlockByIdCmd force unlocks individual period lock by ID if proper permissions exist.
func NewForceUnlockByIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "force-unlock-by-id [id]",
		Short: "force unlocks individual period lock by ID",
		Long:  "force unlocks individual period lock by ID. if no amount provided, entire lock is unlocked",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			coins := sdk.Coins(nil)
			amountStr, err := cmd.Flags().GetString(FlagAmount)
			if err != nil {
				return err
			}

			if amountStr != "" {
				coins, err = sdk.ParseCoinsNormalized(amountStr)
				if err != nil {
					return err
				}
			}

			msg := types.NewMsgForceUnlock(
				clientCtx.GetFromAddress(),
				uint64(id),
				coins,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetUnlockTokens())

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
