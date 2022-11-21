package twapcli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v12/x/twap/client/queryproto"
	"github.com/osmosis-labs/osmosis/v12/x/twap/client/v2queryproto"
	"github.com/osmosis-labs/osmosis/v12/x/twap/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	// Group superfluid queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(GetQueryTwapCommand())
	cmd.AddCommand(GetQueryTwapLegacyCommand())

	return cmd
}

// GetQueryTwapCommand returns multiplier of an asset by denom.
func GetQueryTwapCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "twap [poolid] [base denom] [start time] [end time]",
		Short: "Query twap",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query twap for pool. Start time must be unix time. End time can be unix time or duration.

Example:
$ %s q twap 1 uosmo 1667088000 24h
$ %s q twap 1 uosmo 1667088000 1667174400
`,
				version.AppName, version.AppName,
			),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// boilerplate parse fields
			poolId, baseDenom, startTime, endTime, err := twapQueryParseArgs(args)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := v2queryproto.NewQueryClient(clientCtx)
			gammClient := gammtypes.NewQueryClient(clientCtx)
			liquidity, err := gammClient.TotalPoolLiquidity(cmd.Context(), &gammtypes.QueryTotalPoolLiquidityRequest{PoolId: poolId})
			if err != nil {
				return err
			}
			if len(liquidity.Liquidity) != 2 {
				return fmt.Errorf("pool %d has %d assets of liquidity, CLI support only exists for 2 assets right now.", poolId, len(liquidity.Liquidity))
			}
			quoteDenom := ""
			if liquidity.Liquidity[0].Denom == baseDenom {
				quoteDenom = liquidity.Liquidity[1].Denom
			} else if liquidity.Liquidity[1].Denom == baseDenom {
				quoteDenom = liquidity.Liquidity[0].Denom
			} else {
				return fmt.Errorf("pool %d doesn't have provided baseDenom %s, has %s and %s",
					poolId, baseDenom, liquidity.Liquidity[0], liquidity.Liquidity[1])
			}

			res, err := queryClient.ArithmeticTwap(cmd.Context(), &v2queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  baseDenom,
				QuoteAsset: quoteDenom,
				StartTime:  startTime,
				EndTime:    &endTime,
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

// GetQueryTwapCommand returns multiplier of an asset by denom.
func GetQueryTwapLegacyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "twap-legacy [poolid] [base denom] [start time] [end time]",
		Short: "Query twap",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query twap for pool. Start time must be unix time. End time can be unix time or duration.

Example:
$ %s q twap 1 uosmo 1667088000 24h
$ %s q twap 1 uosmo 1667088000 1667174400
`,
				version.AppName, version.AppName,
			),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			poolId, baseDenom, startTime, endTime, err := twapQueryParseArgs(args)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := queryproto.NewQueryClient(clientCtx)
			gammClient := gammtypes.NewQueryClient(clientCtx)
			liquidity, err := gammClient.TotalPoolLiquidity(cmd.Context(), &gammtypes.QueryTotalPoolLiquidityRequest{PoolId: poolId})
			if err != nil {
				return err
			}
			if len(liquidity.Liquidity) != 2 {
				return fmt.Errorf("pool %d has %d assets of liquidity, CLI support only exists for 2 assets right now.", poolId, len(liquidity.Liquidity))
			}
			quoteDenom := ""
			if liquidity.Liquidity[0].Denom == baseDenom {
				quoteDenom = liquidity.Liquidity[1].Denom
			} else if liquidity.Liquidity[1].Denom == baseDenom {
				quoteDenom = liquidity.Liquidity[0].Denom
			} else {
				return fmt.Errorf("pool %d doesn't have provided baseDenom %s, has %s and %s",
					poolId, baseDenom, liquidity.Liquidity[0], liquidity.Liquidity[1])
			}

			res, err := queryClient.ArithmeticTwap(cmd.Context(), &queryproto.ArithmeticTwapRequest{
				PoolId:     poolId,
				BaseAsset:  baseDenom,
				QuoteAsset: quoteDenom,
				StartTime:  startTime,
				EndTime:    &endTime,
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

func twapQueryParseArgs(args []string) (poolId uint64, baseDenom string, startTime time.Time, endTime time.Time, err error) {
	// boilerplate parse fields
	// <UINT PARSE>
	poolId, err = parseUint(args[0], "poolId")
	if err != nil {
		return
	}

	// <DENOM PARSE>
	baseDenom = strings.TrimSpace(args[1])

	// <UNIX TIME PARSE>
	startTime, err = parseUnixTime(args[2], "start time")
	if err != nil {
		return
	}

	// END TIME PARSE: ONEOF {<UNIX TIME PARSE>, <DURATION>}
	// try parsing in unix time, if failed try parsing in duration
	endTime, err = parseUnixTime(args[3], "end time")
	if err != nil {
		// TODO if we don't use protoreflect:
		// make better error combiner, rather than just returning last error
		duration, err2 := time.ParseDuration(args[3])
		if err2 != nil {
			err = err2
			return
		}
		endTime = startTime.Add(duration)
	}
	return poolId, baseDenom, startTime, endTime, nil
}

func parseUint(arg string, fieldName string) (uint64, error) {
	v, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s as uint for field %s: %w", arg, fieldName, err)
	}
	return v, nil
}

func parseUnixTime(arg string, fieldName string) (time.Time, error) {
	timeUnix, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse %s as unix time for field %s: %w", arg, fieldName, err)
	}
	startTime := time.Unix(timeUnix, 0)
	return startTime, nil
}
