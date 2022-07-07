package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewLockTokensCmd(),
		NewBeginUnlockingCmd(),
		NewBeginUnlockByIDCmd(),
	)

	return cmd
}

// NewLockTokensCmd creates a new lock with the specified duration and tokens from the user's account.
func NewLockTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-tokens [tokens]",
		Short: "lock tokens into lockup pool from user account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			durationStr, err := cmd.Flags().GetString(FlagDuration)
			if err != nil {
				return err
			}

			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgLockTokens(
				clientCtx.GetFromAddress(),
				duration,
				coins,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetLockTokens())
	flags.AddTxFlagsToCmd(cmd)
	err := cmd.MarkFlagRequired(FlagDuration)
	if err != nil {
		panic(err)
	}
	return cmd
}

// NewBeginUnlockingCmd starts unlocking all unlockable locks from user's account.
func NewBeginUnlockingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "begin-unlock-tokens",
		Short: "begin unlock not unlocking tokens from lockup pool for an account",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			msg := types.NewMsgBeginUnlockingAll(
				clientCtx.GetFromAddress(),
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
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
