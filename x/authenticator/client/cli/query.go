package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v23/x/authenticator/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdAuthenticators)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdParams)

	return cmd
}

func GetCmdAuthenticators() (*osmocli.QueryDescriptor, *types.GetAuthenticatorsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "authenticators",
		Short: "Query authenticators",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} osmo1asdadasd`,
	}, &types.GetAuthenticatorsRequest{}
}

func GetCmdParams() (*osmocli.QueryDescriptor, *types.QueryParamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "params",
		Short: "Query params",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} params`,
	}, &types.QueryParamsRequest{}
}
