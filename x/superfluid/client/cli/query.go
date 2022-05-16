package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/osmosis-labs/osmosis/v8/x/superfluid/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group superfluid queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdAllSuperfluidAssets(),
		GetCmdAssetMultiplier(),
		GetCmdAllIntermediaryAccounts(),
		GetCmdConnectedIntermediaryAccount(),
		GetCmdSuperfluidDelegationAmount(),
		GetCmdSuperfluidDelegationsByDelegator(),
		GetCmdSuperfluidUndelegationsByDelegator(),
		GetCmdTotalSuperfluidDelegations(),
	)

	return cmd
}

// GetCmdQueryParams implements a command to fetch superfluid parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current superfluid parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query parameters for the superfluid module:

$ <appd> query superfluid params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAllSuperfluidAssets returns all superfluid enabled assets
func GetCmdAllSuperfluidAssets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-superfluid-assets",
		Short: "Query all superfluid assets",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all superfluid assets.

Example:
$ %s query superfluid all-superfluid-assets
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

			res, err := queryClient.AllAssets(cmd.Context(), &types.AllAssetsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAssetMultiplier returns multiplier of an asset by denom
func GetCmdAssetMultiplier() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset-multiplier [denom]",
		Short: "Query asset multiplier by denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query asset multiplier by denom.

Example:
$ %s query superfluid asset-multiplier gamm/pool/1
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

			res, err := queryClient.AssetMultiplier(cmd.Context(), &types.AssetMultiplierRequest{
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

// GetCmdAllIntermediaryAccounts returns all superfluid intermediary accounts
func GetCmdAllIntermediaryAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-intermediary-accounts",
		Short: "Query all superfluid intermediary accounts",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all superfluid intermediary accounts.

Example:
$ %s query superfluid all-intermediary-accounts
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

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.AllIntermediaryAccounts(cmd.Context(), &types.AllIntermediaryAccountsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "superfluid")

	return cmd
}

// GetCmdConnectedIntermediaryAccount returns connected intermediary account
func GetCmdConnectedIntermediaryAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connected-intermediary-account [lock_id]",
		Short: "Query connected intermediary account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query connected intermediary account.

Example:
$ %s query superfluid connected-intermediary-account 1
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

			lockId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.ConnectedIntermediaryAccount(cmd.Context(), &types.ConnectedIntermediaryAccountRequest{
				LockId: uint64(lockId),
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

// GetCmdSuperfluidDelegationAmount returns the coins superfluid delegated for a
// delegator, validator, denom
func GetCmdSuperfluidDelegationAmount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "superfluid-delegation-amount [delegator_address] [validator_address] [denom]",
		Short: "Query coins superfluid delegated for a delegator, validator, denom",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SuperfluidDelegationAmount(cmd.Context(), &types.SuperfluidDelegationAmountRequest{
				DelegatorAddress: args[0],
				ValidatorAddress: args[1],
				Denom:            args[2],
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

// GetCmdSuperfluidDelegationsByDelegator returns the coins superfluid delegated for the specified delegator
func GetCmdSuperfluidDelegationsByDelegator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "superfluid-delegation-by-delegator [delegator_address]",
		Short: "Query coins superfluid delegated for the specified delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SuperfluidDelegationsByDelegator(cmd.Context(), &types.SuperfluidDelegationsByDelegatorRequest{
				DelegatorAddress: args[0],
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

// GetCmdSuperfluidUndelegationsByDelegator returns the coins superfluid undelegated for the specified delegator
func GetCmdSuperfluidUndelegationsByDelegator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "superfluid-undelegation-by-delegator [delegator_address]",
		Short: "Query coins superfluid undelegated for the specified delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SuperfluidUndelegationsByDelegator(cmd.Context(), &types.SuperfluidUndelegationsByDelegatorRequest{
				DelegatorAddress: args[0],
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

// GetCmdTotalSuperfluidDelegations returns total amount of base denom delegated via superfluid staking
func GetCmdTotalSuperfluidDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-superfluid-delegations",
		Short: "Query total amount of osmo delegated via superfluid staking",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TotalSuperfluidDelegations(cmd.Context(), &types.TotalSuperfluidDelegationsRequest{})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
