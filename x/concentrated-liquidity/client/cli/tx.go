package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	clmodel "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewCreatePositionCmd)
	osmocli.AddTxCmd(txCmd, NewAddToPositionCmd)
	osmocli.AddTxCmd(txCmd, NewWithdrawPositionCmd)
	osmocli.AddTxCmd(txCmd, NewCreateConcentratedPoolCmd)
	osmocli.AddTxCmd(txCmd, NewCollectSpreadRewardsCmd)
	osmocli.AddTxCmd(txCmd, NewCollectIncentivesCmd)
	osmocli.AddTxCmd(txCmd, NewFungifyChargedPositionsCmd)
	osmocli.AddTxCmd(txCmd, NewTransferPositionsCmd)
	return txCmd
}

var poolIdFlagOverride = map[string]string{
	"poolid": FlagPoolId,
}

func NewCreateConcentratedPoolCmd() (*osmocli.TxCliDesc, *clmodel.MsgCreateConcentratedPool) {
	return &osmocli.TxCliDesc{
		Use:     "create-pool",
		Short:   "create a concentrated liquidity pool with the given denom pair, tick spacing, and spread factor",
		Long:    "denom-1 (the quote denom), tick spacing, and spread factors must all be authorized by the concentrated liquidity module",
		Example: "osmosisd tx concentratedliquidity create-pool uion uosmo 100 0.01 --from val --chain-id osmosis-1 -b block --keyring-backend test --fees 1000uosmo",
	}, &clmodel.MsgCreateConcentratedPool{}
}

func NewCreatePositionCmd() (*osmocli.TxCliDesc, *types.MsgCreatePosition) {
	return &osmocli.TxCliDesc{
		Use:     "create-position",
		Short:   "create or add to existing concentrated liquidity position",
		Example: "osmosisd tx concentratedliquidity create-position 1 \"[-69082]\" 69082 10000uosmo,10000uion 0 0 --from val --chain-id osmosis-1 -b block --keyring-backend test --fees 1000uosmo",
	}, &types.MsgCreatePosition{}
}

func NewAddToPositionCmd() (*osmocli.TxCliDesc, *types.MsgAddToPosition) {
	return &osmocli.TxCliDesc{
		Use:     "add-to-position",
		Short:   "add to an existing concentrated liquidity position",
		Example: "osmosisd tx concentratedliquidity add-to-position 10 1000000000uosmo 10000000uion --from val --chain-id localosmosis -b block --keyring-backend test --fees 1000000uosmo",
	}, &types.MsgAddToPosition{}
}

func NewWithdrawPositionCmd() (*osmocli.TxCliDesc, *types.MsgWithdrawPosition) {
	return &osmocli.TxCliDesc{
		Use:     "withdraw-position",
		Short:   "withdraw from an existing concentrated liquidity position",
		Example: "osmosisd tx concentratedliquidity withdraw-position 1 1000 --from val --chain-id localosmosis --keyring-backend=test --fees=1000uosmo",
	}, &types.MsgWithdrawPosition{}
}

func NewCollectSpreadRewardsCmd() (*osmocli.TxCliDesc, *types.MsgCollectSpreadRewards) {
	return &osmocli.TxCliDesc{
		Use:     "collect-spread-rewards",
		Short:   "collect spread rewards from liquidity position(s)",
		Example: "osmosisd tx concentratedliquidity collect-spread-rewards 998 --from val --chain-id localosmosis -b block --keyring-backend test --fees 1000000uosmo",
	}, &types.MsgCollectSpreadRewards{}
}

func NewCollectIncentivesCmd() (*osmocli.TxCliDesc, *types.MsgCollectIncentives) {
	return &osmocli.TxCliDesc{
		Use:     "collect-incentives",
		Short:   "collect incentives from liquidity position(s)",
		Example: "osmosisd tx concentratedliquidity collect-incentives 1 --from val --chain-id localosmosis -b block --keyring-backend test --fees 10000uosmo",
	}, &types.MsgCollectIncentives{}
}

func NewFungifyChargedPositionsCmd() (*osmocli.TxCliDesc, *types.MsgFungifyChargedPositions) {
	return &osmocli.TxCliDesc{
		Use:     "fungify-positions",
		Short:   "Combine fully charged positions within the same range into a new single fully charged position",
		Example: "osmosisd tx concentratedliquidity fungify-positions 1,2 --from val --keyring-backend test -b=block --chain-id=localosmosis --gas=1000000 --fees 20000uosmo",
	}, &types.MsgFungifyChargedPositions{}
}

func NewTransferPositionsCmd() (*osmocli.TxCliDesc, *types.MsgTransferPositions) {
	return &osmocli.TxCliDesc{
		Use:     "transfer-positions",
		Short:   "transfer a list of concentrated liquidity positions to a new owner",
		Example: "osmosisd tx concentratedliquidity transfer-positions 56,89,1011 osmo10fhdy8zhepstpwsr9l4a8yxuyggqmpqx4ktheq --from val --chain-id osmosis-1 -b block --keyring-backend test --fees 1000uosmo",
	}, &types.MsgTransferPositions{}
}

// NewCmdCreateConcentratedLiquidityPoolsProposal implements a command handler for create concentrated liquidity pool proposal
func NewCmdCreateConcentratedLiquidityPoolsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-concentratedliquidity-pool-proposal [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a create concentrated liquidity pool proposal",
		Long: strings.TrimSpace(`Submit a create concentrated liquidity pool proposal.

Passing in FlagPoolRecords separated by commas would be parsed automatically to pairs of pool records.
Ex) --pool-records=uion,uosmo,100,0.003,stake,uosmo,1000,0.005 ->
[uion<>uosmo, tickSpacing 100, spreadFactor 0.3%]
[stake<>uosmo, tickSpacing 1000, spreadFactor 0.5%]

		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseCreateConcentratedLiquidityPoolArgsToContent(cmd)
			if err != nil {
				return err
			}

			msg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}
			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)
	cmd.Flags().String(FlagPoolRecords, "", "The pool records array")

	return cmd
}

func NewTickSpacingDecreaseProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tick-spacing-decrease-proposal [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a tick spacing decrease proposal",
		Long: strings.TrimSpace(`Submit a tick spacing decrease proposal.

Passing in FlagPoolIdToTickSpacingRecords separated by commas would be parsed automatically to pairs of PoolIdToTickSpacing records.
Ex) --pool-tick-spacing-records=1,10,5,1 -> [(poolId 1, newTickSpacing 10), (poolId 5, newTickSpacing 1)]
Note: The new tick spacing value must be less than the current tick spacing value.

		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parsePoolIdToTickSpacingRecordsArgsToContent(cmd)
			if err != nil {
				return err
			}

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
	cmd.Flags().String(FlagPoolIdToTickSpacingRecords, "", "The pool ID to new tick spacing records array")

	return cmd
}

func parseCreateConcentratedLiquidityPoolArgsToContent(cmd *cobra.Command) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	poolRecords, err := parsePoolRecords(cmd)
	if err != nil {
		return nil, err
	}

	content := &types.CreateConcentratedLiquidityPoolsProposal{
		Title:       title,
		Description: description,
		PoolRecords: poolRecords,
	}

	return content, nil
}

func parsePoolIdToTickSpacingRecordsArgsToContent(cmd *cobra.Command) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	poolIdToTickSpacingRecords, err := parsePoolIdToTickSpacingRecords(cmd)
	if err != nil {
		return nil, err
	}

	content := &types.TickSpacingDecreaseProposal{
		Title:                      title,
		Description:                description,
		PoolIdToTickSpacingRecords: poolIdToTickSpacingRecords,
	}
	return content, nil
}

func parsePoolIdToTickSpacingRecords(cmd *cobra.Command) ([]types.PoolIdToTickSpacingRecord, error) {
	assetsStr, err := cmd.Flags().GetString(FlagPoolIdToTickSpacingRecords)
	if err != nil {
		return nil, err
	}

	assets := strings.Split(assetsStr, ",")

	if len(assets)%2 != 0 {
		return nil, fmt.Errorf("poolIdToTickSpacingRecords must be a list of pairs of poolId and newTickSpacing")
	}

	poolIdToTickSpacingRecords := []types.PoolIdToTickSpacingRecord{}
	i := 0
	for i < len(assets) {
		poolId, err := strconv.Atoi(assets[i])
		if err != nil {
			return nil, err
		}
		newTickSpacing, err := strconv.Atoi(assets[i+1])
		if err != nil {
			return nil, err
		}

		poolIdToTickSpacingRecords = append(poolIdToTickSpacingRecords, types.PoolIdToTickSpacingRecord{
			PoolId:         uint64(poolId),
			NewTickSpacing: uint64(newTickSpacing),
		})

		// increase counter by the next 2
		i = i + 2
	}

	return poolIdToTickSpacingRecords, nil
}

func parsePoolRecords(cmd *cobra.Command) ([]types.PoolRecord, error) {
	poolRecordsStr, err := cmd.Flags().GetString(FlagPoolRecords)
	if err != nil {
		return nil, err
	}

	poolRecords := strings.Split(poolRecordsStr, ",")

	if len(poolRecords)%4 != 0 {
		return nil, fmt.Errorf("poolRecords must be a list of denom0, denom1, tickSpacing, and spreadFactor")
	}

	finalPoolRecords := []types.PoolRecord{}
	i := 0
	for i < len(poolRecords) {
		denom0 := poolRecords[i]
		denom1 := poolRecords[i+1]

		tickSpacing, err := strconv.Atoi(poolRecords[i+2])
		if err != nil {
			return nil, err
		}

		spreadFactorStr := poolRecords[i+3]
		spreadFactor, err := osmomath.NewDecFromStr(spreadFactorStr)
		if err != nil {
			return nil, err
		}

		finalPoolRecords = append(finalPoolRecords, types.PoolRecord{
			Denom0:       denom0,
			Denom1:       denom1,
			TickSpacing:  uint64(tickSpacing),
			SpreadFactor: spreadFactor,
		})

		// increase counter by the next 4
		i = i + 4
	}

	return finalPoolRecords, nil
}
