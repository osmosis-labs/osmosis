package cli

import (

	// "strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)

	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdDenomAuthorityMetadata)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdDenomsFromCreator)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdAllBeforeSendHooks)

	cmd.AddCommand(
		osmocli.GetParams[*types.QueryParamsRequest](
			types.ModuleName, types.NewQueryClient),
	)

	return cmd
}

func GetCmdDenomAuthorityMetadata() (*osmocli.QueryDescriptor, *types.QueryDenomAuthorityMetadataRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denom-authority-metadata",
		Short: "Get the authority metadata for a specific denom",
		Long: `{{.Short}}{{.ExampleHeader}}
		{{.CommandPrefix}} uatom`,
	}, &types.QueryDenomAuthorityMetadataRequest{}
}

func GetCmdDenomsFromCreator() (*osmocli.QueryDescriptor, *types.QueryDenomsFromCreatorRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "denoms-from-creator",
		Short: "Returns a list of all tokens created by a specific creator address",
		Long: `{{.Short}}{{.ExampleHeader}}
		{{.CommandPrefix}} <address>`,
	}, &types.QueryDenomsFromCreatorRequest{}
}
func GetCmdAllBeforeSendHooks() (*osmocli.QueryDescriptor, *types.QueryAllBeforeSendHooksAddressesRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "all-before-send-hooks",
		Short: "Returns a list of all before send hooks registered",
	}, &types.QueryAllBeforeSendHooksAddressesRequest{}
}

// GetCmdDenomAuthorityMetadata returns the authority metadata for a queried denom
func GetCmdDenomBeforeSendHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denom-before-send-hook [denom] [flags]",
		Short: "Get the BeforeSend hook for a specific denom",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.BeforeSendHookAddress(cmd.Context(), &types.QueryBeforeSendHookAddressRequest{
				Denom: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
