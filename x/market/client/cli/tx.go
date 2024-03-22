package cli

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v23/x/market/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(cmd, NewSwapCmd)

	return cmd
}

// NewSwapCmd will create and send a MsgSwap
func NewSwapCmd() (*osmocli.TxCliDesc, *types.MsgSwap) {
	return &osmocli.TxCliDesc{
		Use:     "swap",
		NumArgs: 3,
		//Args:  cobra.RangeArgs(2, 3),
		Short: "Atomically swap currencies at their target exchange rate",
		Long: strings.TrimSpace(`
   Swap the offer-coin to the ask-denom currency at the oracle's effective exchange rate.

   $ osmosisd market swap osmo1fr2x4cdvka7yfs8q9gqh0gzmh4hkmktpqwqj63 1000stake uosmo

   The to-address can be specified. A default to-address is trader.

   $ osmosisd market swap "osmo1..." "1000stake" "uosmo"
   `),
		ParseAndBuildMsg: NewSwapMsg,
	}, &types.MsgSwap{}
}

func NewSwapMsg(clientCtx client.Context, args []string, fs *flag.FlagSet) (sdk.Msg, error) {
	offerCoinStr := args[1]
	offerCoin, err := sdk.ParseCoinNormalized(offerCoinStr)
	if err != nil {
		return nil, err
	}

	askDenom := args[2]
	fromAddress := clientCtx.GetFromAddress()

	var msg sdk.Msg
	if len(args) == 3 {
		toAddress, err := sdk.AccAddressFromBech32(args[0])
		if err != nil {
			return nil, err
		}

		msg = types.NewMsgSwapSend(fromAddress, toAddress, offerCoin, askDenom)
		if err = msg.ValidateBasic(); err != nil {
			return nil, err
		}
	} else {
		msg = types.NewMsgSwap(fromAddress, offerCoin, askDenom)
		if err = msg.ValidateBasic(); err != nil {
			return nil, err
		}
	}
	return msg, nil
}
