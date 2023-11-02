package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v20/x/interchainqueries/types"
)

const (
	flagOwners       = "owners"
	flagConnectionID = "connection_id"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group interchainqueries queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryRegisteredQueries())
	cmd.AddCommand(CmdQueryRegisteredQuery())
	cmd.AddCommand(CmdQueryRegisteredQueryResult())
	cmd.AddCommand(CmdQueryLastRemoteHeight())

	return cmd
}

func CmdQueryRegisteredQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registered-query [id]",
		Short: "queries all the interchain queries in the module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			queryID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse query id: %w", err)
			}

			res, err := queryClient.RegisteredQuery(context.Background(), &types.QueryRegisteredQueryRequest{QueryId: queryID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryRegisteredQueries() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registered-queries",
		Short: "queries all the interchain queries in the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			owners, _ := cmd.Flags().GetStringArray(flagOwners)
			connectionID, _ := cmd.Flags().GetString(flagConnectionID)

			res, err := queryClient.RegisteredQueries(context.Background(), &types.QueryRegisteredQueriesRequest{
				Pagination:   pageReq,
				Owners:       owners,
				ConnectionId: connectionID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().StringArray(flagOwners, []string{}, "(optional) filter by query owners")
	cmd.Flags().String(flagConnectionID, "", "(optional) filter by connection id")
	flags.AddPaginationFlagsToCmd(cmd, "registered queries")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryRegisteredQueryResult() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-result [query-id]",
		Short: "queries result for registered query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			queryID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse query id: %w", err)
			}

			res, err := queryClient.QueryResult(context.Background(), &types.QueryRegisteredQueryResultRequest{QueryId: queryID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryLastRemoteHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-last-remote-height [connection-id]",
		Short: "queries last remote height by connection id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			connectionID := args[0]
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.LastRemoteHeight(context.Background(), &types.QueryLastRemoteHeight{ConnectionId: connectionID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
