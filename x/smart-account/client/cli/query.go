package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/smart-account/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdAuthenticators)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdAuthenticator)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdParams)

	return cmd
}

func GetCmdAuthenticators() (*osmocli.QueryDescriptor, *types.GetAuthenticatorsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "authenticators",
		Short: "Query authenticators by account",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj`,
	}, &types.GetAuthenticatorsRequest{}
}

func GetCmdAuthenticator() (*osmocli.QueryDescriptor, *types.GetAuthenticatorRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "authenticator",
		Short: "Query authenticator by account and id",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj 17`,
	}, &types.GetAuthenticatorRequest{}
}

func GetCmdParams() (*osmocli.QueryDescriptor, *types.QueryParamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "params",
		Short: "Query smartaccount params",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} params`,
	}, &types.QueryParamsRequest{}
}
