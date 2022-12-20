package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/spf13/cobra"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CmdSetProtoRevAdminAccountProposal implements the command to submit a SetProtoRevAdminAccountProposal
func CmdSetProtoRevAdminAccountProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-protorev-admin-account-proposal [sdk.AccAddress]",
		Args:    cobra.ExactArgs(1),
		Short:   "submit a set protorev admin account proposal",
		Long:    "submit a set protorev admin proposal followed by an sdk.AccAddress of the new admin account.",
		Example: fmt.Sprintf(`$ %s tx gov submit-proposal set-protorev-admin-account osmo123... --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			content := types.NewSetProtoRevAdminAccountProposal(title, description, args[0])

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

	return cmd
}

// CmdSetProtoRevEnabledProposal implements the command to submit a SetProtoRevEnabledProposal
func CmdSetProtoRevEnabledProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-protorev-enabled-proposal [boolean]",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit a SetProtoRevEnabled proposal",
		Long:    "Submit a SetProtoRevEnabled proposal followed by a boolean of whether or not the protocol should be enabled.",
		Example: fmt.Sprintf(`$ %s tx gov submit-proposal set-protorev-enabled true --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			res, err := strconv.ParseBool(args[0])
			if err != nil {
				return err
			}

			content := types.NewSetProtoRevEnabledProposal(title, description, res)

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

	return cmd
}
