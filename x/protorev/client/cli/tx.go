package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/spf13/cobra"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewCmdTx returns the cli transaction commands for this module
func NewCmdTx() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, CmdSetDeveloperAccount)
	osmocli.AddTxCmd(txCmd, CmdSetMaxPoolPointsPerTx)
	osmocli.AddTxCmd(txCmd, CmdSetMaxPoolPointsPerBlock)
	osmocli.AddTxCmd(txCmd, CmdSetPoolWeights)
	osmocli.AddTxCmd(txCmd, CmdSetBaseDenoms)
	txCmd.AddCommand(
		CmdSetDeveloperHotRoutes().BuildCommandCustomFn(),
		CmdSetProtoRevAdminAccountProposal(),
		CmdSetProtoRevEnabledProposal(),
	)
	return txCmd
}

// CmdSetDeveloperHotRoutes() implements the command to set the protorev hot routes
func CmdSetDeveloperHotRoutes() *osmocli.TxCliDesc {
	desc := osmocli.TxCliDesc{
		Use:   "set-protorev-hot-routes [path/to/routes.json]",
		Short: "set the protorev hot routes",
		Long: `Must provide a json file with all of the routes that will be set. 
		Sample json file:
		[
			{
				"token_in": "uosmo",
				"token_out": "ibc/123...",
				"arb_routes" : [
					{
						"trades": [
							{
								"pool": 1,
								"token_in": "uosmo",
								"token_out": "uatom",
							},
							{
								"pool": 2,
								"token_in": "uatom",
								"token_out": "ibc/123...",
							},
							{
								"pool": 3,
								"token_in": "ibc/123...",
								"token_out": "uosmo",
							},
						],
						"step_size": 1000000,
					}
				]
			}
		]
		`,
		Example:          fmt.Sprintf(`$ %s tx protorev set-protorev-hot-routes routes.json --from mykey`, version.AppName),
		NumArgs:          1,
		ParseAndBuildMsg: BuildSetHotRoutesMsg,
	}

	return &desc
}

// CmdSetDeveloperAccount() implements the command to set the protorev developer account
func CmdSetDeveloperAccount() (*osmocli.TxCliDesc, *types.MsgSetDeveloperAccount) {
	return &osmocli.TxCliDesc{
		Use:  "set-protorev-developer-account [sdk.AccAddress]",
		Args: cobra.ExactArgs(1),
	}
}

// CmdSetMaxPoolPointsPerTx implements the command to set the max pool points per tx
func CmdSetMaxPoolPointsPerTx() (*osmocli.TxCliDesc, *types.MsgSetMaxPoolPointsPerTx) {
	return nil, nil
}

// CmdSetMaxPoolPointsPerBlock implements the command to set the max pool points per block
func CmdSetMaxPoolPointsPerBlock() (*osmocli.TxCliDesc, *types.MsgSetMaxPoolPointsPerBlock) {
	return nil, nil
}

// CmdSetPoolWeights implements the command to set the pool weights used to estimate execution costs
func CmdSetPoolWeights() (*osmocli.TxCliDesc, *types.MsgSetPoolWeights) {
	return nil, nil
}

// CmdSetBaseDenoms implements the command to set the base denoms used in the highest liquidity method
func CmdSetBaseDenoms() (*osmocli.TxCliDesc, *types.MsgSetBaseDenoms) {
	return nil, nil
}

// CmdSetProtoRevAdminAccountProposal implements the command to submit a SetProtoRevAdminAccountProposal
func CmdSetProtoRevAdminAccountProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-protorev-admin-account-proposal [sdk.AccAddress]",
		Args:    cobra.ExactArgs(1),
		Short:   "submit a set protorev admin account proposal to set the admin account for x/protorev",
		Example: fmt.Sprintf(`$ %s tx protorev set-protorev-admin-account osmo123... --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			createContent := func(title string, description string, args ...string) (govtypes.Content, error) {
				return types.NewSetProtoRevAdminAccountProposal(title, description, args[0]), nil
			}

			return ProposalExecute(cmd, args, createContent)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(cli.FlagTitle)
	_ = cmd.MarkFlagRequired(cli.FlagDescription)

	return cmd
}

// CmdSetProtoRevEnabledProposal implements the command to submit a SetProtoRevEnabledProposal
func CmdSetProtoRevEnabledProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-protorev-enabled-proposal [boolean]",
		Args:    cobra.ExactArgs(1),
		Short:   "submit a set protorev enabled proposal to enable or disable the protocol",
		Example: fmt.Sprintf(`$ %s tx protorev set-protorev-enabled true --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			createContent := func(title string, description string, args ...string) (govtypes.Content, error) {
				res, err := strconv.ParseBool(args[0])
				if err != nil {
					return nil, err
				}

				content := types.NewSetProtoRevEnabledProposal(title, description, res)
				return content, nil
			}

			return ProposalExecute(cmd, args, createContent)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(cli.FlagTitle)
	_ = cmd.MarkFlagRequired(cli.FlagDescription)

	return cmd
}

// ProposalExecute is a helper function to execute a proposal command. It takes in a function to create the proposal content.
func ProposalExecute(cmd *cobra.Command, args []string, createContent func(title string, description string, args ...string) (govtypes.Content, error)) error {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return err
	}

	title, err := cmd.Flags().GetString(cli.FlagTitle)
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString(cli.FlagDescription)
	if err != nil {
		return err
	}

	depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
	if err != nil {
		return err
	}

	deposit, err := sdk.ParseCoinsNormalized(depositStr)
	if err != nil {
		return err
	}

	from := clientCtx.GetFromAddress()

	content, err := createContent(title, description, args...)
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
}
