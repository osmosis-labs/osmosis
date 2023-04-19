package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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
		Use:     "create-concentrated-pool [denom-0] [denom-1] [tick-spacing] [exponent-at-price-one] [swap-fee]",
		Short:   "create a concentrated liquidity pool with the given tick spacing",
		Example: "create-concentrated-pool uion uosmo 1 \"[-1]\" 0.01 --from val --chain-id osmosis-1",
	}, &clmodel.MsgCreateConcentratedPool{}
}

func NewCreatePositionCmd() (*osmocli.TxCliDesc, *types.MsgCreatePosition) {
	return &osmocli.TxCliDesc{
		Use:                 "create-position [lower-tick] [upper-tick] [token-0] [token-1] [token-0-min-amount] [token-1-min-amount]",
		Short:               "create or add to existing concentrated liquidity position",
		Example:             "create-position [-69082] 69082 1000000000uosmo 10000000uion 0 0 --pool-id 1 --from val --chain-id osmosis-1",
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
		Use:     "collect-fees [position-ids]",
		Short:   "collect fees from liquidity position(s)",
		Example: "collect-fees 1,5,7 --from val --chain-id osmosis-1",
	}, &types.MsgCollectFees{}
}

func NewCollectIncentivesCmd() (*osmocli.TxCliDesc, *types.MsgCollectIncentives) {
	return &osmocli.TxCliDesc{
		Use:     "collect-incentives [position-ids]",
		Short:   "collect incentives from liquidity position(s)",
		Example: "collect-incentives 1,5,7 --from val --chain-id osmosis-1",
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

// NewCmdCreateConcentratedLiquidityPoolProposal implements a command handler for create concentrated liquidity pool proposal
func NewCmdCreateConcentratedLiquidityPoolProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-concentratedliquidity-pool-proposal [denom0] [denom1] [tick-spacing] [exponent-at-price-one] [swap-fee] [flags]",
		Args:  cobra.ExactArgs(5),
		Short: "Submit a create concentrated liquidity pool proposal",
		Long: strings.TrimSpace(`Submit a create concentrated liquidity pool proposal.

		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			tickSpacing, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}
			exponentAtPriceOne, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("Failed to parse exponent at price one to sdk.Int")
			}
			swapFee, err := sdk.NewDecFromStr(args[4])
			if err != nil {
				return err
			}

			content, err := parseCreateConcentratedLiquidityPoolArgsToContent(cmd, args[0], args[1], tickSpacing, exponentAtPriceOne, swapFee)
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

	return cmd
}

func parseCreateConcentratedLiquidityPoolArgsToContent(cmd *cobra.Command, denom0, denom1 string, tickSpacing uint64, exponentAtPriceOne sdk.Int, swapFee sdk.Dec) (govtypes.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagDescription)
	if err != nil {
		return nil, err
	}

	content := &types.CreateConcentratedLiquidityPoolProposal{
		Title:              title,
		Description:        description,
		Denom0:             denom0,
		Denom1:             denom1,
		TickSpacing:        tickSpacing,
		ExponentAtPriceOne: exponentAtPriceOne,
		SwapFee:            swapFee,
	}

	return content, nil
}
