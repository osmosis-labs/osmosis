package cli

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v7/x/claim/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	claimTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	claimTxCmd.AddCommand()

	return claimTxCmd
}
