package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	cmd.AddCommand(
		NewInboundTransferCmd(),
		NewOutboundTransferCmd(),
		NewUpdateParamsCmd(),
		NewChangeAssetStatusCmd(),
	)

	return cmd
}

func NewInboundTransferCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgInboundTransfer](&osmocli.TxCliDesc{
		Use:   "inbound-transfer",
		Short: "Make an inbound transfer from the external chain to osmosis.",
	})
}

func NewOutboundTransferCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgOutboundTransfer](&osmocli.TxCliDesc{
		Use:   "outbound-transfer",
		Short: "Make an outbound transfer from osmosis to the external chain.",
	})
}

func NewUpdateParamsCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgUpdateParams](&osmocli.TxCliDesc{
		Use:   "update-params",
		Short: "Update the x/bridge module params.",
	})
}

func NewChangeAssetStatusCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgChangeAssetStatus](&osmocli.TxCliDesc{
		Use:   "change-asset-status",
		Short: "Change the asset status to the one specified in the call.",
		Long: `Change the asset status to the one specified in the call.
Available statuses: 
ASSET_STATUS_OK
ASSET_STATUS_BLOCKED_INBOUND
ASSET_STATUS_BLOCKED_OUTBOUND
ASSET_STATUS_BLOCKED_BOTH`,
	})
}
