package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v27/x/stablestaking/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryUserStake(),
		GetCmdQueryUserTotalStake(),
		GetCmdQueryStablePool(),
		GetCmdQueryStablePools(),
		GetCmdQueryUserUnbonding(),
		GetCmdQueryUserTotalUnbonding(),
	)

	return cmd
}

// GetCmdQueryParams implements the params query command
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current stablestaking parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryUserStake implements the user stake query command
func GetCmdQueryUserStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-stake [address] [token]",
		Short: "Query user's stake for a specific token",
		Long: fmt.Sprintf(`Query user's stake for a specific token.

Example:
$ %s query stablestaking user-stake symphony1... uusd
`, version.AppName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.UserStake(cmd.Context(), &types.QueryUserStakeRequest{
				Address: args[0],
				Denom:   args[1],
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

// GetCmdQueryUserTotalStake implements the user total stake query command
func GetCmdQueryUserTotalStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-total-stake [address]",
		Short: "Query user's total stake across all tokens",
		Long: fmt.Sprintf(`Query user's total stake across all tokens.

Example:
$ %s query stablestaking user-total-stake symphony1...
`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.UserTotalStake(cmd.Context(), &types.QueryUserTotalStakeRequest{
				Address: args[0],
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

// GetCmdQueryStablePool implements the stable pool query command
func GetCmdQueryStablePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stable-pool [denom]",
		Short: "Query stable pool information for a specific token",
		Long: fmt.Sprintf(`Query stable pool information for a specific token.

Example:
$ %s query stablestaking stable-pool uusd
`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.StablePool(cmd.Context(), &types.QueryPoolRequest{
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

// GetCmdQueryStablePools implements the stable pools query command
func GetCmdQueryStablePools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stable-pools",
		Short: "Query all stable pools information",
		Long: fmt.Sprintf(`Query all stable pools information.

Example:
$ %s query stablestaking stable-pools
`, version.AppName),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.StablePools(cmd.Context(), &types.QueryPoolsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryUserUnbonding implements the user unbonding query command
func GetCmdQueryUserUnbonding() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-unbonding [address] [denom]",
		Short: "Query user's unbonding information for a specific token",
		Long: fmt.Sprintf(`Query user's unbonding information for a specific token.

Example:
$ %s query stablestaking user-unbonding symphony1... uusd
`, version.AppName),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.UserUnbonding(cmd.Context(), &types.QueryUserUnbondingRequest{
				Address: args[0],
				Denom:   args[1],
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

// GetCmdQueryUserTotalUnbonding implements the user total unbonding query command
func GetCmdQueryUserTotalUnbonding() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-total-unbonding [address]",
		Short: "Query user's total unbonding information across all tokens",
		Long: fmt.Sprintf(`Query user's total unbonding information across all tokens.

Example:
$ %s query stablestaking user-total-unbonding symphony1...
`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.UserTotalUnbonding(cmd.Context(), &types.QueryUserTotalUnbondingRequest{
				Address: args[0],
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
