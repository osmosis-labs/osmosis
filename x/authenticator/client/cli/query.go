package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v20/x/authenticator/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdAuthenticators)

	return cmd
}

func GetCmdAuthenticators() (*osmocli.QueryDescriptor, *types.GetAuthenticatorsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "authenticators",
		Short: "Query authenticators",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pool 1`,
	}, &types.GetAuthenticatorsRequest{}
}
