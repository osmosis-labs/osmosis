package cli

import (
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/types"
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
	return txCmd
}

func NewCreateCWPoolCmd() (*osmocli.TxCliDesc, *model.MsgCreateCosmWasmPool) {
	return &osmocli.TxCliDesc{
		Use:     "create-pool [code-id] [denom-1] [tick-spacing] [spread-factor]",
		Short:   "create a concentrated liquidity pool with the given denom pair, tick spacing, and spread factor",
		Long:    "denom-1 (the quote denom), tick spacing, and spread factors must all be authorized by the concentrated liquidity module",
		Example: "osmosisd tx concentratedliquidity create-pool uion uosmo 100 0.01 --from val --chain-id osmosis-1 -b block --keyring-backend test --fees 1000uosmo",
	}, &model.MsgCreateCosmWasmPool{}
}
