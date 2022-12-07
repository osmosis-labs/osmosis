package cli

import (
	"fmt"
	"strconv"

	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Concentrated liquidity transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewCreatePositionCmd(),
		NewWithdrawPositionCmd(),
	)

	return txCmd
}

func NewCreatePositionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-position [pool-id] [lower-tick] [upper-tick] [token-0] [token-1] [token-0-min-amount] [token-1-min-amount]",
		Short:   "create or add to existing concentrated liquidity position",
		Example: "create-position 1 200 38000 1000000000uosmo 1000000uion 1 1",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildCreatePositionMsg(clientCtx, args[0], args[1], args[2], args[3], args[4], args[5], args[6], txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewBuildCreatePositionMsg(clientCtx client.Context, poolIdStr, lowerTickStr, upperTickStr, tokenDesired0Str, tokenDesired1Str, tokenMinAmount0Str, tokenMinAmount1Str string, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := strconv.ParseUint(poolIdStr, 10, 64)
	if err != nil {
		return txf, nil, err
	}

	lowerTick, err := strconv.ParseInt(lowerTickStr, 10, 64)
	if err != nil {
		return txf, nil, err
	}

	upperTick, err := strconv.ParseInt(upperTickStr, 10, 64)
	if err != nil {
		return txf, nil, err
	}

	tokenDesired0, err := sdk.ParseCoinNormalized(tokenDesired0Str)
	if err != nil {
		return txf, nil, err
	}

	tokenDesired1, err := sdk.ParseCoinNormalized(tokenDesired1Str)
	if err != nil {
		return txf, nil, err
	}

	tokenMinAmount0, ok := sdk.NewIntFromString(tokenMinAmount0Str)
	if !ok {
		return txf, nil, fmt.Errorf("invalid token min amount 0, %s", tokenMinAmount0Str)
	}

	tokenMinAmount, ok := sdk.NewIntFromString(tokenMinAmount1Str)
	if !ok {
		return txf, nil, fmt.Errorf("invalid token min amount 1, %s", tokenMinAmount1Str)
	}

	msg := &types.MsgCreatePosition{
		PoolId:          poolId,
		Sender:          clientCtx.GetFromAddress().String(),
		LowerTick:       lowerTick,
		UpperTick:       upperTick,
		TokenDesired0:   tokenDesired0,
		TokenDesired1:   tokenDesired1,
		TokenMinAmount0: tokenMinAmount0,
		TokenMinAmount1: tokenMinAmount,
	}

	return txf, msg, nil
}

func NewWithdrawPositionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw-position [pool-id] [lower-tick] [upper-tick] [liquidity-out]",
		Short:   "withdraw from an existing concentrated liquidity position",
		Example: "withdraw-position 1 200 38000 1517818840",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildWithdrawPositionMsg(clientCtx, args[0], args[1], args[2], args[3], txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewBuildWithdrawPositionMsg(clientCtx client.Context, poolIdStr, lowerTickStr, upperTickStr, liquidityAmtStr string, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	poolId, err := strconv.ParseUint(poolIdStr, 10, 64)
	if err != nil {
		return txf, nil, err
	}

	lowerTick, err := strconv.ParseInt(lowerTickStr, 10, 64)
	if err != nil {
		return txf, nil, err
	}

	upperTick, err := strconv.ParseInt(upperTickStr, 10, 64)
	if err != nil {
		return txf, nil, err
	}

	liquidityAmt, ok := sdk.NewIntFromString(liquidityAmtStr)
	if !ok {
		return txf, nil, fmt.Errorf("invalid liquidity amount, %s", liquidityAmtStr)
	}

	msg := &types.MsgWithdrawPosition{
		PoolId:          poolId,
		Sender:          clientCtx.GetFromAddress().String(),
		LowerTick:       lowerTick,
		UpperTick:       upperTick,
		LiquidityAmount: liquidityAmt,
	}

	return txf, msg, nil
}
