package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	gammtypes "github.com/osmosis-labs/osmosis/v28/x/gamm/types"
)

// GetPoolsCmd returns a CLI command to get information about liquidity pools
func GetPoolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools [pool-id]",
		Short: "Query liquidity pool information",
		Long: `Query information about liquidity pools. If pool-id is provided, returns information about specific pool.
If no pool-id is provided, returns information about all pools.

Example:
$ osmosisd query pools         # Query all pools
$ osmosisd query pools 1       # Query specific pool with ID 1
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := gammtypes.NewQueryClient(clientCtx)
			ctx := cmd.Context()

			if len(args) == 0 {
				// Query all pools
				res, err := queryClient.Pools(ctx, &gammtypes.QueryPoolsRequest{})
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			}

			// Query specific pool
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("pool-id %s not a valid uint", args[0])
			}

			res, err := queryClient.Pool(ctx, &gammtypes.QueryPoolRequest{
				PoolId: poolID,
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
