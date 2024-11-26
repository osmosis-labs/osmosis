package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
)

func NewCmdSubmitUpdatePoolIncentivesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-pool-incentives [gaugeIds] [weights]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit an update to the records for pool incentives",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
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

			content := types.NewUpdatePoolIncentivesProposal(proposalTitle, summary, records)

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)

	return cmd
}

func NewCmdSubmitReplacePoolIncentivesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace-pool-incentives [gaugeIds] [weights]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit a full replacement to the records for pool incentives",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
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

			content := types.NewReplacePoolIncentivesProposal(proposalTitle, summary, records)

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)

	return cmd
}
