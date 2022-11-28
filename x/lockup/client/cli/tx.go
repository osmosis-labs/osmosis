package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v3/x/lockup/types"
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
		NewLockTokensCmd(),
		NewBeginUnlockingCmd(),
		NewUnlockTokensCmd(),
		NewBeginUnlockByIDCmd(),
		NewUnlockByIDCmd(),
	)

	return cmd
}

// NewLockTokensCmd lock tokens into lockup pool from user's account
func NewLockTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-tokens",
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

// NewBeginUnlockingCmd unlock all unlockable tokens from user's account
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

// NewUnlockTokensCmd unlock all unlockable tokens from user's account
func NewUnlockTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock-tokens",
		Short: "unlock tokens from lockup pool for an account",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			msg := types.NewMsgUnlockTokens(
				clientCtx.GetFromAddress(),
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewBeginUnlockByIDCmd unlock individual period lock by ID
func NewBeginUnlockByIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "begin-unlock-by-id",
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

			msg := types.NewMsgBeginUnlocking(
				clientCtx.GetFromAddress(),
				uint64(id),
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewUnlockByIDCmd unlock individual period lock by ID
func NewUnlockByIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock-by-id",
		Short: "unlock individual period lock by ID",
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

			msg := types.NewMsgUnlockPeriodLock(
				clientCtx.GetFromAddress(),
				uint64(id),
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
