package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v7/osmoutils"
	"github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
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

			// TODO: Make a parse uint64 slice function
			gaugeIds, err := osmoutils.ParseUint64SliceFromString(args[0], ",")
			if err != nil {
				return err
			}

			weights, err := osmoutils.ParseSdkIntFromString(args[1], ",")
			if err != nil {
				return err
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

			from := clientCtx.GetFromAddress()

			proposal, err := osmoutils.ParseProposalFlags(cmd.Flags())
			if err != nil {
				return fmt.Errorf("failed to parse proposal: %w", err)
			}
			
			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			content := types.NewUpdatePoolIncentivesProposal(proposal.Title, proposal.Deposit, records)

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

	cmd.Flags().String(cli.FlagProposal, "", "Proposal file path (if this path is given, other proposal flags are ignored)")

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

			gaugeIds, err := osmoutils.ParseUint64SliceFromString(args[0], ",")
			if err != nil {
				return err
			}

			weights, err := osmoutils.ParseSdkIntFromString(args[1], ",")
			if err != nil {
				return err
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

			from := clientCtx.GetFromAddress()

			proposal, err := osmoutils.ParseProposalFlags(cmd.Flags())
			if err != nil {
				return fmt.Errorf("failed to parse proposal: %w", err)
			}
			
			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			content := types.NewReplacePoolIncentivesProposal(proposal.Title, proposal.Description, records)

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

	cmd.Flags().String(cli.FlagProposal, "", "Proposal file path (if this path is given, other proposal flags are ignored)")

	return cmd
}
