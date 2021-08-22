package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/x/bech32ibc/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdHrpIbcRecords(),
		GetCmdHrpSourceChannel(),
		GetCmdNativeHrp(),
	)

	return cmd
}

// GetCmdGaugeIds takes the pool id and returns the matching gauge ids and durations
func GetCmdHrpIbcRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hrp-ibc-records",
		Short: "Query the entire mapping of bech32 prefixes to ibc source channels",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the entire mapping of bech32 prefixes to ibc source channels
Example:
$ %s query bech32ibc hrp-ibc-records
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.HrpIbcRecords(cmd.Context(), &types.QueryHrpIbcRecordsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdHrpSourceChannel returns the source channel associated with a specific bech32 prefix
func GetCmdHrpSourceChannel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hrp-ibc-record",
		Short: "Query the pool id associated with a specific whitelisted fee token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the pool id associated with a specific fee token
Example:
$ %s query bech32ibc hrp-ibc-record [source-channel]
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.HrpSourceChannel(cmd.Context(), &types.QueryHrpSourceChannelRequest{
				Hrp: args[0],
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

// GetCmdNativeHrp returns the native bech32 prefix for the chain
func GetCmdNativeHrp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "native-hrp",
		Short: "Query the native bech32 prefix for the chain.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the native bech32 prefix for the chain.
Example:
$ %s query bech32ibc native-hrp
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.NativeHrp(cmd.Context(), &types.QueryNativeHrpRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
