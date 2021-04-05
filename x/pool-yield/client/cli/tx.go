package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/c-osmosis/osmosis/x/pool-yield/types"
)

func NewCmdSubmitAddPoolIncentivesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-pool-incentives [farmIds] [weights]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit a new record for pool incentives",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var farmIds []uint64
			for _, farmIdStr := range strings.Split(args[0], ",") {
				farmIdStr = strings.TrimSpace(farmIdStr)

				parsed, err := strconv.ParseUint(farmIdStr, 10, 64)
				if err != nil {
					return err
				}
				farmIds = append(farmIds, parsed)
			}

			var weights []sdk.Int
			for _, weightStr := range strings.Split(args[1], ",") {
				weightStr = strings.TrimSpace(weightStr)

				parsed, err := strconv.ParseUint(weightStr, 10, 64)
				if err != nil {
					return err
				}
				weights = append(weights, sdk.NewIntFromUint64(parsed))
			}

			if len(farmIds) != len(weights) {
				return fmt.Errorf("the length of farm ids and wieghts not matched")
			}

			if len(farmIds) == 0 {
				return fmt.Errorf("records is empty")
			}

			var records []types.DistrRecord
			for i, farmId := range farmIds {
				records = append(records, types.DistrRecord{
					FarmId: farmId,
					Weight: weights[i],
				})
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			content := types.NewAddPoolIncentivesProposal(title, description, records)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.MarkFlagRequired(cli.FlagTitle)
	cmd.MarkFlagRequired(cli.FlagDescription)

	return cmd
}

func NewCmdSubmitRemovePoolIncentivesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-pool-incentives [indexes]",
		Args:  cobra.ExactArgs(1),
		Short: "Remove a specific index of record from pool incentives",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var indexes []uint64
			for _, indexStr := range strings.Split(args[0], ",") {
				indexStr = strings.TrimSpace(indexStr)

				parsed, err := strconv.ParseUint(indexStr, 10, 64)
				if err != nil {
					return err
				}
				indexes = append(indexes, parsed)
			}

			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			content := types.NewRemovePoolIncentivesProposal(title, description, indexes)

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.MarkFlagRequired(cli.FlagTitle)
	cmd.MarkFlagRequired(cli.FlagDescription)

	return cmd
}
