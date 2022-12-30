package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdSpotPrice)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPool)
	osmocli.AddQueryCmd(cmd, types.NewQueryClient, GetCmdPools)
	cmd.AddCommand(
		GetCmdPoolParams(),
		GetCmdTotalShares(),
		GetCmdQueryTotalLiquidity(),
		GetCmdTotalPoolLiquidity(),
		GetCmdQueryPoolsWithFilter(),
		GetCmdPoolType(),
	)

	return cmd
}

func GetCmdPool() (*osmocli.QueryDescriptor, *types.QueryPoolRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pool [poolID]",
		Short: "Query pool",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pool 1`}, &types.QueryPoolRequest{}
}

// TODO: Push this to the SDK.
func writeOutputBoilerplate(ctx client.Context, out []byte) error {
	writer := ctx.Output
	if writer == nil {
		writer = os.Stdout
	}

	_, err := writer.Write(out)
	if err != nil {
		return err
	}

	if ctx.OutputFormat != "text" {
		// append new-line for formats besides YAML
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCmdPools() (*osmocli.QueryDescriptor, *types.QueryPoolsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pools",
		Short: "Query pools",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} pools`}, &types.QueryPoolsRequest{}
}

// GetCmdPoolParams return pool params.
func GetCmdPoolParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool-params <poolID>",
		Short: "Query pool-params",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pool-params.
Example:
$ %s query gamm pool-params 1
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

			poolID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.PoolParams(cmd.Context(), &types.QueryPoolParamsRequest{
				PoolId: uint64(poolID),
			})
			if err != nil {
				return err
			}

			if clientCtx.OutputFormat == "text" {
				poolParams := &balancer.PoolParams{}
				if err := poolParams.Unmarshal(res.GetParams().Value); err != nil {
					return err
				}

				out, err := yaml.Marshal(poolParams)
				if err != nil {
					return err
				}
				return writeOutputBoilerplate(clientCtx, out)
			} else {
				out, err := clientCtx.Codec.MarshalJSON(res)
				if err != nil {
					return err
				}

				return writeOutputBoilerplate(clientCtx, out)
			}
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdTotalPoolLiquidity() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryTotalPoolLiquidityRequest](
		"total-pool-liquidity [poolID]",
		"Query total-pool-liquidity",
		`Query total-pool-liquidity.
Example:
{{.CommandPrefix}} total-pool-liquidity 1
`,
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdTotalShares() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryTotalSharesRequest](
		"total-share [poolID]",
		"Query total-share",
		`Query total-share.
Example:
{{.CommandPrefix}} total-share 1
`,
		types.ModuleName, types.NewQueryClient,
	)
}

func GetCmdQueryTotalLiquidity() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryTotalLiquidityRequest](
		"total-liquidity",
		"Query total-liquidity",
		`Query total-liquidity.
Example:
{{.CommandPrefix}} total-liquidity
`,
		types.ModuleName, types.NewQueryClient,
	)
}

//nolint:staticcheck
func GetCmdSpotPrice() (*osmocli.QueryDescriptor, *types.QuerySpotPriceRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "spot-price <pool-ID> [quote-asset-denom] [base-asset-denom]",
		Short: "Query spot-price (LEGACY, arguments are reversed!!)",
		Long: `Query spot price (Legacy).{{.ExampleHeader}}
{{.CommandPrefix}} spot-price 1 uosmo ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
`}, &types.QuerySpotPriceRequest{}
}

// GetCmdQueryPoolsWithFilter returns pool with filter
func GetCmdQueryPoolsWithFilter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools-with-filter <min_liquidity> <pool_type>",
		Short: "Query pools with filter",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pools with filter. The possible filter options are:

1. By Pool type: Either "Balancer" or "Stableswap"
2. By min pool liquidity by providing min coins

Note that if both filters are to be applied, "min_liquidity" always needs to be provided as the first argument.

Example:
$ %s query gamm pools-with-filter <min_liquidity> <pool_type> 
`,
				version.AppName,
			),
		),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			var pool_type string
			min_liquidity := args[0]
			if len(args) > 1 {
				pool_type = args[1]
			}
			res, err := queryClient.PoolsWithFilter(cmd.Context(), &types.QueryPoolsWithFilterRequest{
				MinLiquidity: min_liquidity,
				PoolType:     pool_type,
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

// GetCmdPoolType returns pool type given pool id.
func GetCmdPoolType() *cobra.Command {
	return osmocli.SimpleQueryCmd[*types.QueryPoolTypeRequest](
		"pool-type <pool_id>",
		"Query pool type",
		`Query pool type
Example:
{{.CommandPrefix}} pool-type <pool_id>
`,
		types.ModuleName, types.NewQueryClient,
	)
}
