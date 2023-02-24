package cli

import (
	"github.com/spf13/cobra"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v15/x/tokenfactory/types"
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
	)

	return cmd
}

func NewCreateDenomCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgCreateDenom](&osmocli.TxCliDesc{
		Use:   "create-denom [subdenom] [flags]",
		Short: "create a new denom from an account. (Costs osmo though!)",
	})
}

func NewMintCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgMint](&osmocli.TxCliDesc{
		Use:   "mint [amount] [flags]",
		Short: "Mint a denom to an address. Must have admin authority to do so.",
	})
}

func NewBurnCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgBurn](&osmocli.TxCliDesc{
		Use:   "burn [amount] [flags]",
		Short: "Burn tokens from an address. Must have admin authority to do so.",
	})
}

func NewChangeAdminCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgChangeAdmin](&osmocli.TxCliDesc{
		Use:   "change-admin [denom] [new-admin-address] [flags]",
		Short: "Changes the admin address for a factory-created denom. Must have admin authority to do so.",
	})
}
