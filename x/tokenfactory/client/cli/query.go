package cli

import (

	// "strings"

	"github.com/spf13/cobra"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/tokenfactory/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdDenomAuthorityMetadata)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdDenomAuthorityMetadata)

	cmd.AddCommand(
		osmocli.GetParams[*types.QueryParamsRequest](
			types.ModuleName, types.NewQueryClient),
	)

	return cmd
}

func GetCmdDenomAuthorityMetadata() (*osmocli.QueryDescriptor, *types.QueryDenomAuthorityMetadataRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denom-authority-metadata [denom] [flags]",
		Short: "Get the authority metadata for a specific denom",
		Long: `{{.Short}}{{.ExampleHeader}}
		{{.CommandPrefix}} uatom`,
	}, &types.QueryDenomAuthorityMetadataRequest{}
}

func GetCmdDenomsFromCreator() (*osmocli.QueryDescriptor, *types.QueryDenomsFromCreatorRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denoms-from-creator [creator address] [flags]",
		Short: "Returns a list of all tokens created by a specific creator address",
		Long: `{{.Short}}{{.ExampleHeader}}
		{{.CommandPrefix}} <address>`,
	}, &types.QueryDenomsFromCreatorRequest{}
}
