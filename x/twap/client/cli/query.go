package twapcli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	cosmossdk_io_math "cosmossdk.io/math"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	poolmanager "github.com/osmosis-labs/osmosis/v20/x/poolmanager/client/queryproto"
	"github.com/osmosis-labs/osmosis/v20/x/twap/client/queryproto"
	"github.com/osmosis-labs/osmosis/v20/x/twap/types"
)

// twapQueryParseArgs represents the outcome
// of parsing the arguments for twap query command.
type twapQueryArgs struct {
	PoolId    uint64
	BaseDenom string
	StartTime time.Time
	EndTime   time.Time
}

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	cmd.AddCommand(GetQueryArithmeticCommand())
	cmd.AddCommand(GetQueryGeometricCommand())
	cmd.AddCommand(GetQueryExtraArithmeticCommand())
	return cmd
}

// GetQueryArithmeticCommand returns an arithmetic twap query command.
func GetQueryArithmeticCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "arithmetic [poolid] [base denom] [start time] [end time]",
		Short:   "Query arithmetic twap",
		Aliases: []string{"twap"},
		Long: osmocli.FormatLongDescDirect(`Query arithmetic twap for pool. Start time must be unix time. End time can be unix time or duration.

Example:
{{.CommandPrefix}} arithmetic 1 uosmo 1667088000 24h
{{.CommandPrefix}} arithmetic 1 uosmo 1667088000 1667174400
`, types.ModuleName),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// boilerplate parse fields
			twapArgs, err := twapQueryParseArgs(args)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			quoteDenom, err := getQuoteDenomFromLiquidity(cmd.Context(), clientCtx, twapArgs.PoolId, twapArgs.BaseDenom)
			if err != nil {
				return err
			}

			queryClient := queryproto.NewQueryClient(clientCtx)
			res, err := queryClient.ArithmeticTwap(cmd.Context(), &queryproto.ArithmeticTwapRequest{
				PoolId:     twapArgs.PoolId,
				BaseAsset:  twapArgs.BaseDenom,
				QuoteAsset: quoteDenom,
				StartTime:  twapArgs.StartTime,
				EndTime:    &twapArgs.EndTime,
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

// GetQueryExtraArithmeticCommand returns an arithmetic twap query command with no [start, end] range limit.
func GetQueryExtraArithmeticCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extra-arithmetic [poolid] [base denom] [start time] [end time]",
		Short:   "Query extra arithmetic twap",
		Aliases: []string{"extra-twap"},
		Long: osmocli.FormatLongDescDirect(`Query arithmetic twap for pool. Start time must be unix time. End time can be unix time or duration.

Example:
{{.CommandPrefix}} extra-arithmetic 1 uosmo 1667088000 24h
{{.CommandPrefix}} extra-arithmetic 1 uosmo 1667088000 1667174400
`, types.ModuleName),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// boilerplate parse fields
			twapArgs, err := twapQueryParseArgs(args)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			quoteDenom, err := getQuoteDenomFromLiquidity(cmd.Context(), clientCtx, twapArgs.PoolId, twapArgs.BaseDenom)
			if err != nil {
				return err
			}

			node, err := clientCtx.GetNode()
			if err != nil {
				return err
			}
			nodeStatus, err := node.Status(context.Background())
			if err != nil {
				return err
			}

			latestHeight := nodeStatus.SyncInfo.LatestBlockHeight
			diffAccum := cosmossdk_io_math.LegacyZeroDec()
			diffTime := types.CanonicalTimeMs(twapArgs.EndTime) - types.CanonicalTimeMs(twapArgs.StartTime)

			// Get twap of 2 days chunk then calculate final twap
			for twapArgs.StartTime.Before(twapArgs.EndTime) {
				tempStartTime := twapArgs.StartTime
				tempEndTime := twapArgs.EndTime

				if twapArgs.EndTime.Add(-time.Hour * 48).After(twapArgs.StartTime) {
					tempStartTime = twapArgs.EndTime.Add(-time.Hour * 48)
				}

				height, err := findBlockByTime(clientCtx, tempEndTime.UTC(), latestHeight)
				if err != nil {
					return err
				}

				clientCtx, err = getClientQueryContextFromHeight(cmd, height)
				if err != nil {
					return err
				}

				queryClient := queryproto.NewQueryClient(clientCtx)

				res, err := queryClient.ArithmeticTwap(cmd.Context(), &queryproto.ArithmeticTwapRequest{
					PoolId:     twapArgs.PoolId,
					BaseAsset:  twapArgs.BaseDenom,
					QuoteAsset: quoteDenom,
					StartTime:  tempStartTime,
					EndTime:    &tempEndTime,
				})
				if err != nil {
					return err
				}

				timeDelta := types.CanonicalTimeMs(tempEndTime) - types.CanonicalTimeMs(tempStartTime)
				diffAccum = diffAccum.Add(res.ArithmeticTwap.MulInt64(timeDelta))
				twapArgs.EndTime = twapArgs.EndTime.Add(-time.Hour * 48)
				latestHeight = height
			}

			twap := diffAccum.QuoInt64(diffTime)

			return clientCtx.PrintProto(&queryproto.ArithmeticTwapResponse{ArithmeticTwap: twap})
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Get Client Query Context at a given height
func getClientQueryContextFromHeight(cmd *cobra.Command, height int64) (client.Context, error) {
	ctx := client.GetClientContextFromCmd(cmd)
	ctx = ctx.WithHeight(height)
	return client.ReadPersistentCommandFlags(ctx, cmd.Flags())
}

// Try to find the block containing the given timestamp
func findBlockByTime(clientCtx client.Context, time time.Time, currentHeight int64) (int64, error) {
	// Currently we hardcode this value
	// TODO: Fetch block time from somewhere
	blockTime := 5.84
	client, _ := clientCtx.GetNode()

	currentBlockResult, err := client.Block(context.Background(), &currentHeight)
	if err != nil {
		return -1, err
	}
	diffTime := currentBlockResult.Block.Time.Sub(time)
	estimateBlock := currentHeight - int64(diffTime.Seconds()/5.84)
	estimateBlockResult, err := client.Block(context.Background(), &estimateBlock)
	if err != nil {
		return -1, err
	}

	estimateBlockDelta := int64(time.Sub(estimateBlockResult.Block.Time).Seconds() / blockTime)
	for estimateBlockDelta != 0 {
		if estimateBlockDelta > 0 && estimateBlockDelta < 5 {
			estimateBlock += 1
		} else if estimateBlockDelta < 0 && estimateBlockDelta > -5 {
			estimateBlock -= 1
		} else {
			estimateBlock += estimateBlockDelta
		}

		estimateBlockResult, err = client.Block(context.Background(), &estimateBlock)
		if err != nil {
			return -1, err
		}
		estimateBlockDelta = int64(time.Sub(estimateBlockResult.Block.Time).Seconds() / blockTime)
	}

	// If target time after estimate block time, target block is next block
	// If target time equal estimate block time => target time equal to end block time, return estimate block
	// If target time before estimate block time => target time inside estimate block, return estimate block
	if time.After(estimateBlockResult.Block.Time) {
		return estimateBlock + 1, nil
	} else {
		return estimateBlock, nil
	}
}

// GetQueryGeometricCommand returns a geometric twap query command.
func GetQueryGeometricCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "geometric [poolid] [base denom] [start time] [end time]",
		Short: "Query geometric twap",
		Long: osmocli.FormatLongDescDirect(`Query geometric twap for pool. Start time must be unix time. End time can be unix time or duration.

Example:
{{.CommandPrefix}} geometric 1 uosmo 1667088000 24h
{{.CommandPrefix}} geometric 1 uosmo 1667088000 1667174400
`, types.ModuleName),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// boilerplate parse fields
			twapArgs, err := twapQueryParseArgs(args)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			quoteDenom, err := getQuoteDenomFromLiquidity(cmd.Context(), clientCtx, twapArgs.PoolId, twapArgs.BaseDenom)
			if err != nil {
				return err
			}
			queryClient := queryproto.NewQueryClient(clientCtx)
			if err != nil {
				return err
			}

			res, err := queryClient.GeometricTwap(cmd.Context(), &queryproto.GeometricTwapRequest{
				PoolId:     twapArgs.PoolId,
				BaseAsset:  twapArgs.BaseDenom,
				QuoteAsset: quoteDenom,
				StartTime:  twapArgs.StartTime,
				EndTime:    &twapArgs.EndTime,
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

// getQuoteDenomFromLiquidity gets the quote liquidity denom from the pool. In addition, validates that base denom
// exists in the pool. Fails if not.
func getQuoteDenomFromLiquidity(ctx context.Context, clientCtx client.Context, poolId uint64, baseDenom string) (string, error) {
	poolmanagerClient := poolmanager.NewQueryClient(clientCtx)
	liquidity, err := poolmanagerClient.TotalPoolLiquidity(ctx, &poolmanager.TotalPoolLiquidityRequest{PoolId: poolId})
	if err != nil {
		return "", err
	}
	if len(liquidity.Liquidity) != 2 {
		return "", fmt.Errorf("pool %d has %d assets of liquidity, CLI support only exists for 2 assets right now.", poolId, len(liquidity.Liquidity))
	}

	quoteDenom := ""
	if liquidity.Liquidity[0].Denom == baseDenom {
		quoteDenom = liquidity.Liquidity[1].Denom
	} else if liquidity.Liquidity[1].Denom == baseDenom {
		quoteDenom = liquidity.Liquidity[0].Denom
	} else {
		return "", fmt.Errorf("pool %d doesn't have provided baseDenom %s, has %s and %s",
			poolId, baseDenom, liquidity.Liquidity[0], liquidity.Liquidity[1])
	}
	return quoteDenom, nil
}

func twapQueryParseArgs(args []string) (twapQueryArgs, error) {
	// boilerplate parse fields
	// <UINT PARSE>
	poolId, err := osmocli.ParseUint(args[0], "poolId")
	if err != nil {
		return twapQueryArgs{}, err
	}

	// <DENOM PARSE>
	baseDenom := strings.TrimSpace(args[1])

	// <UNIX TIME PARSE>
	startTime, err := osmocli.ParseUnixTime(args[2], "start time")
	if err != nil {
		return twapQueryArgs{}, err
	}

	// END TIME PARSE: ONEOF {<UNIX TIME PARSE>, <DURATION>}
	// try parsing in unix time, if failed try parsing in duration
	endTime, err := osmocli.ParseUnixTime(args[3], "end time")
	if err != nil {
		// TODO if we don't use protoreflect:
		// make better error combiner, rather than just returning last error
		duration, err2 := time.ParseDuration(args[3])
		if err2 != nil {
			err = err2
			return twapQueryArgs{}, err
		}
		endTime = startTime.Add(duration)
	}
	return twapQueryArgs{
		PoolId:    poolId,
		BaseDenom: baseDenom,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}
