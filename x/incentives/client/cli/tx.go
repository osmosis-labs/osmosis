package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCreateGaugeCmd(),
		NewAddToGaugeCmd(),
	)

	return cmd
}

// NewCreateGaugeCmd broadcast MsgCreateGauge
func NewCreateGaugeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-gauge [denom] [reward] [flags]",
		Short: "create a gauge to distribute rewards to users",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom := args[0]

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			startTime := time.Time{}
			timeStr, err := cmd.Flags().GetString(FlagStartTime)
			if err != nil {
				return err
			}
			if timeStr == "" { // empty start time
				// do nothing
			} else if timeUnix, err := strconv.ParseInt(timeStr, 10, 64); err != nil { // unix time
				startTime = time.Unix(timeUnix, 0)
			} else if timeRFC, err := time.Parse(time.RFC3339, timeStr); err != nil { // RFC time
				startTime = timeRFC
			} else { // invalid input
				return errors.New("Invalid start time format")
			}

			epochs, err := cmd.Flags().GetUint64(FlagEpochs)
			if err != nil {
				return err
			}

			epochsDuration, err := cmd.Flags().GetDuration(FlagEpochsDuration)
			if err != nil {
				return err
			}

			perpetual, err := cmd.Flags().GetBool(FlagPerpetual)
			if err != nil {
				return err
			}

			if epochs != 0 && epochsDuration != time.Duration(0) {
				return errors.New("Cannot provide both --epochs and --epochs-duration flags")
			}

			var epochIdentifier string
			if epochs == 0 {
				if epochsDuration != time.Duration(0) {
					// BEGIN epoch info query logic
					paramsQueryClient := proposal.NewQueryClient(clientCtx)
					params := proposal.QueryParamsRequest{
						Subspace: types.ModuleName,
						Key:      string(types.KeyDistrEpochIdentifier),
					}
					epochIdentifierRes, err := paramsQueryClient.Params(context.Background(), &params)
					if err != nil {
						return err
					}
					epochIdentifier = epochIdentifierRes.Param.Value

					epochsQueryClient := epochtypes.NewQueryClient(clientCtx)
					epochInfoRes, err := epochsQueryClient.EpochInfos(cmd.Context(), &epochtypes.QueryEpochsInfoRequest{})
					if err != nil {
						return err
					}
					// END

					for _, epochInfo := range epochInfoRes.Epochs {
						if epochInfo.Identifier == epochIdentifier {
							epochs = uint64(epochsDuration / epochInfo.Duration)
							break
						}
					}
				} else if perpetual {
					epochs = 1
				} else {
					return errors.New("None of --epochs, --epochs-duration, --perpetual are set")
				}
			}

			duration, err := cmd.Flags().GetDuration(FlagDuration)
			if err != nil {
				return err
			}

			distributeTo := lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         denom,
				Duration:      duration,
				Timestamp:     time.Unix(0, 0), // XXX check
			}

			isPerpetual := false
			if epochs == 1 {
				isPerpetual = true
			}

			perpetualStr := ""
			if isPerpetual {
				perpetualStr = "perpetual "
			}
			durationStr := ""
			if duration > time.Duration(0) {
				durationStr = fmt.Sprintf(" over %v", duration)
			}
			fmt.Printf("Creating %sgauge for locked %s%s\n", perpetualStr, denom, durationStr)
			if !isPerpetual {
				fmt.Printf("Distributed over %d %s(%v)", epochs, epochIdentifier, duration)
			}

			msg := types.NewMsgCreateGauge(
				isPerpetual,
				clientCtx.GetFromAddress(),
				distributeTo,
				coins,
				startTime,
				epochs,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateGauge())
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewAddToGaugeCmd broadcast MsgAddToGauge
func NewAddToGaugeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-to-gauge [gauge_id] [rewards] [flags]",
		Short: "add coins to gauge to distribute more rewards to users",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			gaugeId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			rewards, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgAddToGauge(
				clientCtx.GetFromAddress(),
				gaugeId,
				rewards,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateGauge())
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
