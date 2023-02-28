package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v14/x/ibc-rate-limit/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)

	osmocli.AddTxCmd(cmd, NewCmdSetContractParam)

	return cmd
}

func NewCmdSetContractParam() (*osmocli.TxCliDesc, *types.MsgSetContractParam) {
	return &osmocli.TxCliDesc{
		Use:              "set-contract-param [address]",
		Short:            "set contract param",
		NumArgs:          1,
		ParseAndBuildMsg: ParseSetContractPfunc,
	}, &types.MsgSetContractParam{}
}

// func NewJoinPoolCmd() (*osmocli.TxCliDesc, *types.MsgJoinPool) {
// 	return &osmocli.TxCliDesc{
// 		Use:   "join-pool",
// 		Short: "join a new pool and provide the liquidity to it",
// 		CustomFlagOverrides: map[string]string{
// 			"poolid":         FlagPoolId,
// 			"ShareOutAmount": FlagShareAmountOut,
// 		},
// 		CustomFieldParsers: map[string]osmocli.CustomFieldParserFn{
// 			"TokenInMaxs": osmocli.FlagOnlyParser(maxAmountsInParser),
// 		},
// 		Flags: osmocli.FlagDesc{RequiredFlags: []*flag.FlagSet{FlagSetJoinPool()}},
// 	}, &types.MsgJoinPool{}
// }

// func NewMintCmd() *cobra.Command {
// 	return osmocli.BuildTxCli[*types.MsgMint](&osmocli.TxCliDesc{
// 		Use:   "mint [amount] [flags]",
// 		Short: "Mint a denom to an address. Must have admin authority to do so.",
// 	})
// }

func ParseSetContractPfunc(clientCtx client.Context, args []string, flags *pflag.FlagSet) (sdk.Msg, error) {
	fmt.Println(args)
	return &types.MsgSetContractParam{
		Address: args[0],
	}, nil
}
