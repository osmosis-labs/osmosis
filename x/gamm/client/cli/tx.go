package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/c-osmosis/osmosis/x/gamm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Generalized automated market maker transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewCreatePoolCmd())

	return txCmd
}

func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool",
		Short: "create a new pool and provide the liquidity to it",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildCreatePoolMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreatePool())
	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagPoolBindTokens)
	_ = cmd.MarkFlagRequired(FlagPoolBindTokenWeights)
	_ = cmd.MarkFlagRequired(FlagSwapFee)

	return cmd
}

func NewBuildCreatePoolMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	bindTokenStrs, err := fs.GetStringArray(FlagPoolBindTokens)
	if err != nil {
		return txf, nil, err
	}
	if len(bindTokenStrs) < 2 {
		return txf, nil, fmt.Errorf("bind tokens should be more than 2")
	}

	bindTokenWeightStrs, err := fs.GetStringArray(FlagPoolBindTokenWeights)
	if err != nil {
		return txf, nil, err
	}
	if len(bindTokenStrs) != len(bindTokenWeightStrs) {
		return txf, nil, fmt.Errorf("bind tokens and token weight should have same length")
	}

	bindTokensSdk := sdk.Coins{}
	for i := 0; i < len(bindTokenStrs); i++ {
		parsed, err := sdk.ParseCoin(bindTokenStrs[i])
		if err != nil {
			return txf, nil, err
		}
		bindTokensSdk = append(bindTokensSdk, parsed)
	}

	var bindWeights []sdk.Dec
	for i := 0; i < len(bindTokenWeightStrs); i++ {
		parsed, err := sdk.NewDecFromStr(bindTokenWeightStrs[i])
		if err != nil {
			return txf, nil, err
		}
		bindWeights = append(bindWeights, parsed)
	}

	swapFeeStr, err := fs.GetString(FlagSwapFee)
	if err != nil {
		return txf, nil, err
	}
	swapFee, err := sdk.NewDecFromStr(swapFeeStr)
	if err != nil {
		return txf, nil, err
	}

	customDenom, err := fs.GetString(FlagPoolTokenCustomDenom)
	if err != nil {
		return txf, nil, err
	}

	description, err := fs.GetString(FlagPoolTokenDescription)
	if err != nil {
		return txf, nil, err
	}

	var bindTokens []types.BindTokenInfo
	for i := 0; i < len(bindTokensSdk); i++ {
		bindTokenSdk := bindTokensSdk[i]

		bindToken := types.BindTokenInfo{
			Denom:  bindTokenSdk.Denom,
			Weight: bindWeights[i],
			Amount: bindTokenSdk.Amount,
		}

		bindTokens = append(bindTokens, bindToken)
	}

	msg := &types.MsgCreatePool{
		Sender:  clientCtx.GetFromAddress(),
		SwapFee: swapFee,
		LpToken: types.LPTokenInfo{
			Denom:       customDenom,
			Description: description,
		},
		BindTokens: bindTokens,
	}

	return txf, msg, nil
}
