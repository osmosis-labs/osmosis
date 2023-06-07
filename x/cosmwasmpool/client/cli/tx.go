package cli

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/cosmwasm/msg"
	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v16/x/cosmwasmpool/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewCreateCWPoolCmd)
	return txCmd
}

func NewCreateCWPoolCmd() (*osmocli.TxCliDesc, *model.MsgCreateCosmWasmPool) {
	return &osmocli.TxCliDesc{
		Use:              "create-pool [code-id] [instantiate-msg] [sender]",
		Short:            "create a concentrated liquidity pool with the given denom pair, tick spacing, and spread factor",
		Long:             "denom-1 (the quote denom), tick spacing, and spread factors must all be authorized by the concentrated liquidity module",
		Example:          "osmosisd tx concentratedliquidity create-pool uion uosmo 100 0.01 --from val --chain-id osmosis-1 -b block --keyring-backend test --fees 1000uosmo",
		NumArgs:          2,
		ParseAndBuildMsg: BuildCreatePoolMsg,
	}, &model.MsgCreateCosmWasmPool{}
}

func BuildCreatePoolMsg(clientCtx client.Context, args []string, flags *pflag.FlagSet) (sdk.Msg, error) {
	codeId, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return nil, err
	}

	denoms := strings.Split(args[1], ",")

	// Construct instantiate msg
	instantiateMsg := &msg.InstantiateMsg{
		PoolAssetDenoms: denoms,
	}
	msgBz, err := json.Marshal(instantiateMsg)
	if err != nil {
		return nil, err
	}

	return &model.MsgCreateCosmWasmPool{
		CodeId:         codeId,
		InstantiateMsg: msgBz,
		Sender:         clientCtx.GetFromAddress().String(),
	}, nil
}
