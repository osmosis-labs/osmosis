package cli

import (
	"errors"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v19/x/txfees/types"
)

const FlagFeeTokens = "fee-tokens"

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	return txCmd
}

func NewCmdSubmitUpdateFeeTokenProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-fee-token [flags]",
		Args:    cobra.ExactArgs(0),
		Example: "update-fee-token --fee-tokens uosmo,1,uion,2,ufoo,0 --from val --chain-id osmosis-1",
		Short:   "Submit a update fee token record proposal",
		Long: strings.TrimSpace(`Submit a update fee token record proposal.

Passing in denom,poolID pairs separated by commas would be parsed automatically to pairs of fee token records.
Ex) uosmo,1,uion,2,ufoo,0 -> [Adds uosmo<>pool1, uion<>pool2, Removes ufoo as a fee token]

		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			content, err := parseFeeTokenRecordsArgsToContent(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(govcli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

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

	cmd.Flags().String(govcli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(govcli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(govcli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Bool(govcli.FlagIsExpedited, false, "If true, makes the proposal an expedited one")
	cmd.Flags().String(govcli.FlagProposal, "", "Proposal file path (if this path is given, other proposal flags are ignored)")
	cmd.Flags().String(FlagFeeTokens, "", "The fee token records array")

	return cmd
}

func parseFeeTokenRecords(cmd *cobra.Command) ([]types.FeeToken, error) {
	feeTokensStr, err := cmd.Flags().GetString(FlagFeeTokens)
	if err != nil {
		return nil, err
	}

	feeTokens := strings.Split(feeTokensStr, ",")

	if len(feeTokens)%2 != 0 {
		return nil, errors.New("fee denom records should be a comma separated list of denom and poolId pairs")
	}

	feeTokenRecords := []types.FeeToken{}
	i := 0
	for i < len(feeTokens) {
		denom := feeTokens[i]
		poolId, err := strconv.Atoi(feeTokens[i+1])
		if err != nil {
			return nil, err
		}

		feeTokenRecords = append(feeTokenRecords, types.FeeToken{
			Denom:  denom,
			PoolID: uint64(poolId),
		})

		// increase counter by the next 2
		i = i + 2
	}

	return feeTokenRecords, nil
}

func parseFeeTokenRecordsArgsToContent(cmd *cobra.Command) (govtypes.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagDescription)
	if err != nil {
		return nil, err
	}

	feeTokenRecords, err := parseFeeTokenRecords(cmd)
	if err != nil {
		return nil, err
	}

	content := &types.UpdateFeeTokenProposal{
		Title:       title,
		Description: description,
		Feetokens:   feeTokenRecords,
	}
	return content, nil
}
