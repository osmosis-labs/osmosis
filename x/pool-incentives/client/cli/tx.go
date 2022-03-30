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

	"github.com/osmosis-labs/osmosis/v3/x/pool-incentives/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "pool incentives transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewCmdSubmitUpdatePoolIncentivesProposal(),
		NewCmdSubmitReplacePoolIncentivesProposal(),
	)

	return txCmd
}

func NewCmdSubmitUpdatePoolIncentivesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-pool-incentives [gaugeIds] [weights]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit an update to the records for pool incentives",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var gaugeIds []uint64
			for _, gaugeIdStr := range strings.Split(args[0], ",") {
				gaugeIdStr = strings.TrimSpace(gaugeIdStr)

				parsed, err := strconv.ParseUint(gaugeIdStr, 10, 64)
				if err != nil {
					return err
				}
				gaugeIds = append(gaugeIds, parsed)
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

			if len(gaugeIds) != len(weights) {
				return fmt.Errorf("the length of gauge ids and weights not matched")
			}

			if len(gaugeIds) == 0 {
				return fmt.Errorf("records is empty")
			}

			var records []types.DistrRecord
			for i, gaugeId := range gaugeIds {
				records = append(records, types.DistrRecord{
					GaugeId: gaugeId,
					Weight:  weights[i],
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

			content := types.NewUpdatePoolIncentivesProposal(title, description, records)

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
	err := cmd.MarkFlagRequired(cli.FlagTitle)
	if err != nil {
		panic(err)
	}
	err = cmd.MarkFlagRequired(cli.FlagDescription)
	if err != nil {
		panic(err)
	}

	return cmd
}

func NewCmdSubmitReplacePoolIncentivesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace-pool-incentives [gaugeIds] [weights]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit a full replacement to the records for pool incentives",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var gaugeIds []uint64
			for _, gaugeIdStr := range strings.Split(args[0], ",") {
				gaugeIdStr = strings.TrimSpace(gaugeIdStr)

				parsed, err := strconv.ParseUint(gaugeIdStr, 10, 64)
				if err != nil {
					return err
				}
				gaugeIds = append(gaugeIds, parsed)
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

			if len(gaugeIds) != len(weights) {
				return fmt.Errorf("the length of gauge ids and weights not matched")
			}

			if len(gaugeIds) == 0 {
				return fmt.Errorf("records is empty")
			}

			var records []types.DistrRecord
			for i, gaugeId := range gaugeIds {
				records = append(records, types.DistrRecord{
					GaugeId: gaugeId,
					Weight:  weights[i],
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

			content := types.NewReplacePoolIncentivesProposal(title, description, records)

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
	err := cmd.MarkFlagRequired(cli.FlagTitle)
	if err != nil {
		panic(err)
	}
	err = cmd.MarkFlagRequired(cli.FlagDescription)
	if err != nil {
		panic(err)
	}

	return cmd
}
