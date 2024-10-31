package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	cmd.AddCommand(
		NewCreateDenomCmd(),
		NewMintCmd(),
		NewBurnCmd(),
		// NewForceTransferCmd(),
		NewChangeAdminCmd(),
		NewSetBeforeSendHookCmd(),
		NewMsgSetDenomMetadata(),
	)

	return cmd
}

func NewMsgSetDenomMetadata() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgSetDenomMetadata](&osmocli.TxCliDesc{
		Use:   "set-denom-metadata",
		Short: "overwriting of the denom metadata in the bank module.",
	})
}

func NewCreateDenomCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgCreateDenom](&osmocli.TxCliDesc{
		Use:   "create-denom",
		Short: "create a new denom from an account. (osmo to create tokens is charged through gas consumption)",
	})
}

func NewMintCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgMint](&osmocli.TxCliDesc{
		Use:   "mint",
		Short: "Mint a denom to an address. Must have admin authority to do so.",
	})
}

func NewBurnCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgBurn](&osmocli.TxCliDesc{
		Use:   "burn",
		Short: "Burn tokens from an address. Must have admin authority to do so.",
	})
}

func NewChangeAdminCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgChangeAdmin](&osmocli.TxCliDesc{
		Use:   "change-admin",
		Short: "Changes the admin address for a factory-created denom. Must have admin authority to do so.",
	})
}

// NewChangeAdminCmd broadcast MsgChangeAdmin
func NewSetBeforeSendHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-beforesend-hook [denom] [cosmwasm-address] [flags]",
		Short: "Set a cosmwasm contract to be the beforesend hook for a factory-created denom. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			msg := types.NewMsgSetBeforeSendHook(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
