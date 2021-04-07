package cli

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/c-osmosis/osmosis/x/incentives/types"
	lockuptypes "github.com/c-osmosis/osmosis/x/lockup/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		NewCreatePotCmd(),
		NewAddToPotCmd(),
	)

	return cmd
}

// NewCreatePotCmd broadcast MsgCreatePot
func NewCreatePotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pot [coins] [start_time] [num_epochs] [flags]",
		Short: "create a pot to distribute rewards to users",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			coins, err := sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return err
			}

			timeUnix, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			startTime := time.Unix(timeUnix, 0)

			numEpochs, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			queryTypeStr, err := cmd.Flags().GetString(FlagLockQueryType)
			if err != nil {
				return err
			}
			queryType, ok := lockuptypes.LockQueryType_value[queryTypeStr]
			if !ok {
				return errors.New("invalid lock query type")
			}
			denom, err := cmd.Flags().GetString(FlagDenom)
			if err != nil {
				return err
			}
			durationStr, err := cmd.Flags().GetString(FlagDuration)
			if err != nil {
				return err
			}
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return err
			}
			timestamp, err := cmd.Flags().GetInt64(FlagTimestamp)
			if err != nil {
				return err
			}

			msg := &types.MsgCreatePot{
				Owner: clientCtx.GetFromAddress(),
				DistributeTo: lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.LockQueryType(queryType),
					Denom:         denom,
					Duration:      duration,
					Timestamp:     time.Unix(timestamp, 0),
				},
				Coins:     coins,
				StartTime: startTime,
				NumEpochs: numEpochs,
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreatePot())
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewAddToPotCmd broadcast MsgAddToPot
func NewAddToPotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-to-pot [pot_id] [rewards] [flags]",
		Short: "add coins to pot to distribute more rewards to users",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			potId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			rewards, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgAddToPot{
				Owner:   clientCtx.GetFromAddress(),
				PotId:   potId,
				Rewards: rewards,
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreatePot())
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
