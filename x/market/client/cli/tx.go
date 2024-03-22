package cli

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	marketTxCmd := &cobra.Command{
		Use:                        "market",
		Short:                      "Market transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	marketTxCmd.AddCommand(
		GetSwapCmd(),
	)

	return marketTxCmd
}

// GetSwapCmd will create and send a MsgSwap
func GetSwapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [offer-coin] [ask-denom] [to-address]",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Atomically swap currencies at their target exchange rate",
		Long: strings.TrimSpace(`
Swap the offer-coin to the ask-denom currency at the oracle's effective exchange rate. 

$ osmosisd market swap "1000ukrw" "uusd"

The to-address can be specified. A default to-address is trader.

$ osmosisd market swap "1000ukrw" "uusd" "terra1..."
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Generate transaction factory for gas simulation
			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			offerCoinStr := args[0]
			offerCoin, err := sdk.ParseCoinNormalized(offerCoinStr)
			if err != nil {
				return err
			}

			askDenom := args[1]
			fromAddress := clientCtx.GetFromAddress()

			var msg sdk.Msg
			if len(args) == 3 {
				toAddress, err := sdk.AccAddressFromBech32(args[2])
				if err != nil {
					return err
				}

				msg = types.NewMsgSwapSend(fromAddress, toAddress, offerCoin, askDenom)
				if err = msg.ValidateBasic(); err != nil {
					return err
				}

				//if !clientCtx.GenerateOnly && txf.Fees().IsZero() {
				//	// estimate tax and gas
				//	stdFee, err := feeutils.ComputeFeesWithCmd(clientCtx, cmd.Flags(), msg)
				//	if err != nil {
				//		return err
				//	}
				//
				//	// override gas and fees
				//	txf = txf.
				//		WithFees(stdFee.Amount.String()).
				//		WithGas(stdFee.Gas).
				//		WithSimulateAndExecute(false).
				//		WithGasPrices("")
				//}
			} else {
				msg = types.NewMsgSwap(fromAddress, offerCoin, askDenom)
				if err = msg.ValidateBasic(); err != nil {
					return err
				}
			}

			// build and sign the transaction, then broadcast to Tendermint
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

//osmosis approach

//// GetTxCmd returns the transaction commands for this module
//func GetTxCmd() *cobra.Command {
//cmd := osmocli.TxIndexCmd(types.ModuleName)
//osmocli.AddTxCmd(cmd, NewSwapCmd)
//
//return cmd
//}

//// NewSwapCmd will create and send a MsgSwap
//func NewSwapCmd() (*osmocli.TxCliDesc, *types.MsgSwap) {
//	return &osmocli.TxCliDesc{
//		Use:     "swap [offer-coin] [ask-denom] [to-address]",
//		NumArgs: 2,
//		//Args:  cobra.RangeArgs(2, 3),
//		Short: "Atomically swap currencies at their target exchange rate",
//		Long: strings.TrimSpace(`
//    Swap the offer-coin to the ask-denom currency at the oracle's effective exchange rate.
//
//    $ osmosisd market swap "1000ukrw" "uusd"
//
//    The to-address can be specified. A default to-address is trader.
//
//    $ osmosisd market swap "1000ukrw" "uusd" "osmo1..."
//    `),
//		ParseAndBuildMsg: NewSwapMsg,
//	}, &types.MsgSwap{}
//}

//func NewSwapMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
//	offerCoinStr := args[0]
//	offerCoin, err := sdk.ParseCoinNormalized(offerCoinStr)
//	if err != nil {
//		return nil, err
//	}
//
//	askDenom := args[1]
//	fromAddress := clientCtx.GetFromAddress()
//
//	var msg sdk.Msg
//	if len(args) == 3 {
//		toAddress, err := sdk.AccAddressFromBech32(args[2])
//		if err != nil {
//			return nil, err
//		}
//
//		msg = types.NewMsgSwapSend(fromAddress, toAddress, offerCoin, askDenom)
//		if err = msg.ValidateBasic(); err != nil {
//			return nil, err
//		}
//	} else {
//		msg = types.NewMsgSwap(fromAddress, offerCoin, askDenom)
//		if err = msg.ValidateBasic(); err != nil {
//			return nil, err
//		}
//	}
//	return msg, nil
//}
