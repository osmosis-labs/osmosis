package cli

import (
	"fmt"
	"strings"

	"github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	claimTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	claimTxCmd.AddCommand(
		NewCmdClaim(),
	)

	return claimTxCmd
}

// NewCmdClaim implements claim transaction command.
func NewCmdClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim",
		Short: "Claim airdrop for sender account.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim airdrop for sender account.
Example:
$ %s tx claim claim --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaim(clientCtx.GetFromAddress())
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
