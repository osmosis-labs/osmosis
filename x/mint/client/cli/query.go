package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v31/x/mint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for the minting module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	cmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryEpochProvisions(),
		GetCmdQueryInflation(),
		GetCmdQueryBurnedSupply(),
		GetCmdQueryTotalSupply(),
		GetCmdQueryRestrictedSupply(),
		GetCmdQueryCirculatingSupply(),
	)

	return cmd
}

// GetCmdQueryParams implements a command to return the current minting
// parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current minting parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryEpochProvisions implements a command to return the current minting
// epoch provisions value.
func GetCmdQueryEpochProvisions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-provisions",
		Short: "Query the current minting epoch provisions value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryEpochProvisionsRequest{}
			res, err := queryClient.EpochProvisions(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.EpochProvisions))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryInflation implements a command to return the current minting
// inflation value.
func GetCmdQueryInflation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation",
		Short: "Query the current minting inflation value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryInflationRequest{}
			res, err := queryClient.Inflation(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.Inflation))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBurnedSupply implements a command to return the amount of
// mint-denom held in the burn address.
func GetCmdQueryBurnedSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burned-supply",
		Short: "Query the total amount of burned tokens",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.BurnedSupply(context.Background(), &types.QueryBurnedRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.Burned))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTotalSupply implements a command to return the total supply
// (minted - burned).
func GetCmdQueryTotalSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-supply",
		Short: "Query the total supply (minted minus burned)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TotalSupply(context.Background(), &types.QueryTotalSupplyRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.TotalSupply))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRestrictedSupply implements a command to return the supply held in
// restricted addresses.
func GetCmdQueryRestrictedSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restricted-supply",
		Short: "Query the supply held in restricted addresses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RestrictedSupply(context.Background(), &types.QueryRestrictedSupplyRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.RestrictedSupply))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryCirculatingSupply implements a command to return the circulating
// supply (minted - burned - restricted).
func GetCmdQueryCirculatingSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "circulating-supply",
		Short: "Query the circulating supply (minted minus burned minus restricted)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.CirculatingSupply(context.Background(), &types.QueryCirculatingSupplyRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.CirculatingSupply))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
