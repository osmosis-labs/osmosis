package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/osmosis-labs/osmosis/v13/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v13/x/tokenfactory/types"
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

// NewMintCmd broadcast MsgMint
func NewMintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [amount] [flags]",
		Short: "Mint a denom to an address. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgMint(
				clientCtx.GetFromAddress().String(),
				amount,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewBurnCmd broadcast MsgBurn
func NewBurnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn [amount] [flags]",
		Short: "Burn tokens from an address. Must have admin authority to do so.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgBurn(
				clientCtx.GetFromAddress().String(),
				amount,
			)

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// // NewForceTransferCmd broadcast MsgForceTransfer
// func NewForceTransferCmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "force-transfer [amount] [transfer-from-address] [transfer-to-address] [flags]",
// 		Short: "Force transfer tokens from one address to another address. Must have admin authority to do so.",
// 		Args:  cobra.ExactArgs(3),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			clientCtx, err := client.GetClientTxContext(cmd)
// 			if err != nil {
// 				return err
// 			}

// 			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

// 			amount, err := sdk.ParseCoinNormalized(args[0])
// 			if err != nil {
// 				return err
// 			}

// 			msg := types.NewMsgForceTransfer(
// 				clientCtx.GetFromAddress().String(),
// 				amount,
// 				args[1],
// 				args[2],
// 			)

// 			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
// 		},
// 	}

// 	flags.AddTxFlagsToCmd(cmd)
// 	return cmd
// }

func NewChangeAdminCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgChangeAdmin](&osmocli.TxCliDesc{
		Use:   "change-admin [denom] [new-admin-address] [flags]",
		Short: "Changes the admin address for a factory-created denom. Must have admin authority to do so.",
	})
}
