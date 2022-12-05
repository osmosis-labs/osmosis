package cli

import (

	// "strings"

	"github.com/spf13/cobra"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/tokenfactory/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	cmd.AddCommand(
		osmocli.GetParams[*types.QueryParamsRequest](
			types.ModuleName, types.NewQueryClient),
		GetCmdDenomAuthorityMetadata(),
		GetCmdDenomsFromCreator(),
	)

	return cmd
}

func GetCmdDenomAuthorityMetadata() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryDenomAuthorityMetadataRequest](
		"denom-authority-metadata [denom] [flags]",
		"Get the authority metadata for a specific denom", "",
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdDenomsFromCreator() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryDenomsFromCreatorRequest](
		"denoms-from-creator [creator address] [flags]",
		"Returns a list of all tokens created by a specific creator address", "",
		types.ModuleName, types.NewQueryClient,
	)
}
