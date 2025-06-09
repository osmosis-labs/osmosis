package cli

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v27/x/stablestaking/types"
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
		NewStakeTokensCmd(),
		NewUnstakeTokensCmd(),
	)

	return cmd
}

// NewStakeTokensCmd returns a CLI command handler for staking tokens
func NewStakeTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake-tokens [amount] [denom]",
		Short: "Stake tokens in the stable staking module",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse amount
			amount, ok := math.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[0])
			}

			// Validate denom
			denom := args[1]
			if err := sdk.ValidateDenom(denom); err != nil {
				return fmt.Errorf("invalid denom: %s", err)
			}

			// Create a message
			msg := types.NewMsgStakeTokens(
				clientCtx.GetFromAddress(),
				sdk.NewCoin(denom, amount),
			)

			// Generate transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewUnstakeTokensCmd returns a CLI command handler for unstaking tokens
func NewUnstakeTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstake-tokens [amount] [denom]",
		Short: "Unstake tokens from the stable staking module",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse amount
			amount, ok := math.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[0])
			}

			// Validate denom
			denom := args[1]
			if err := sdk.ValidateDenom(denom); err != nil {
				return fmt.Errorf("invalid denom: %s", err)
			}

			// Create a message
			msg := types.NewMsgUnstakeTokens(
				clientCtx.GetFromAddress(),
				sdk.NewCoin(denom, amount),
			)

			// Generate transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
