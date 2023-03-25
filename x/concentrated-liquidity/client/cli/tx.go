package cli

import (
	flag "github.com/spf13/pflag"

	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	clmodel "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewCreatePositionCmd)
	osmocli.AddTxCmd(txCmd, NewWithdrawPositionCmd)
	osmocli.AddTxCmd(txCmd, NewCreateConcentratedPoolCmd)
	osmocli.AddTxCmd(txCmd, NewCollectFeesCmd)
	osmocli.AddTxCmd(txCmd, NewCollectIncentivesCmd)
	osmocli.AddTxCmd(txCmd, NewCreateIncentiveCmd)
	return txCmd
}

var poolIdFlagOverride = map[string]string{
	"poolid": FlagPoolId,
}

func NewCreateConcentratedPoolCmd() (*osmocli.TxCliDesc, *clmodel.MsgCreateConcentratedPool) {
	return &osmocli.TxCliDesc{
		Use:     "create-concentrated-pool [denom-0] [denom-1] [tick-spacing] [swap-fee]",
		Short:   "create a concentrated liquidity pool with the given tick spacing",
		Example: "create-concentrated-pool uion uosmo 1 \"[-1]\" 0.01 --from val --chain-id osmosis-1",
	}, &clmodel.MsgCreateConcentratedPool{}
}

func NewCreatePositionCmd() (*osmocli.TxCliDesc, *types.MsgCreatePosition) {
	return &osmocli.TxCliDesc{
		Use:                 "create-position [lower-tick] [upper-tick] [token-0] [token-1] [token-0-min-amount] [token-1-min-amount] [freeze-duration]",
		Short:               "create or add to existing concentrated liquidity position",
		Example:             "create-position [-69082] 69082 1000000000uosmo 10000000uion 0 0 24h --pool-id 1 --from val --chain-id osmosis-1",
		CustomFlagOverrides: poolIdFlagOverride,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetJustPoolId()}},
	}, &types.MsgCreatePosition{}
}

func NewWithdrawPositionCmd() (*osmocli.TxCliDesc, *types.MsgWithdrawPosition) {
	return &osmocli.TxCliDesc{
		Use:     "withdraw-position [position-id] [liquidity]",
		Short:   "withdraw from an existing concentrated liquidity position",
		Example: "withdraw-position 1 100317215 --from val --chain-id osmosis-1",
	}, &types.MsgWithdrawPosition{}
}

func NewCollectFeesCmd() (*osmocli.TxCliDesc, *types.MsgCollectFees) {
	return &osmocli.TxCliDesc{
		Use:     "collect-fees [position-id]",
		Short:   "collect fees from a liquidity position",
		Example: "collect-fees 1 --from val --chain-id osmosis-1",
	}, &types.MsgCollectFees{}
}

func NewCollectIncentivesCmd() (*osmocli.TxCliDesc, *types.MsgCollectIncentives) {
	return &osmocli.TxCliDesc{
		Use:     "collect-incentives [position-id]",
		Short:   "collect incentives from a liquidity position",
		Example: "collect-incentives 1 --from val --chain-id osmosis-1",
	}, &types.MsgCollectIncentives{}
}

func NewCreateIncentiveCmd() (*osmocli.TxCliDesc, *types.MsgCreateIncentive) {
	return &osmocli.TxCliDesc{
		Use:                 "create-incentive [incentive-denom] [incentive-amount] [emission-rate] [start-time] [min-uptime]",
		Short:               "create an incentive record to emit incentives (per second) to a given pool",
		Example:             "create-incentive uosmo 69082 0.02 100 2023-03-03 03:20:35.419543805 24h --pool-id 1 --from val --chain-id osmosis-1",
		CustomFlagOverrides: poolIdFlagOverride,
		Flags:               osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetJustPoolId()}},
	}, &types.MsgCreateIncentive{}
}
