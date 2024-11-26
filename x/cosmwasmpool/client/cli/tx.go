package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/model"
	"github.com/osmosis-labs/osmosis/v27/x/cosmwasmpool/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(txCmd, NewCreateCWPoolCmd)
	return txCmd
}

func NewCreateCWPoolCmd() (*osmocli.TxCliDesc, *model.MsgCreateCosmWasmPool) {
	return &osmocli.TxCliDesc{
		Use:              "create-pool",
		Short:            "create a cosmwasm pool",
		Example:          "osmosisd tx cosmwasmpool create-pool 1 '{\"pool_asset_denoms\":[\"uion\",\"uosmo\"]}' --from lo-test1 --keyring-backend test --chain-id localosmosis --fees 875uosmo -b=block",
		NumArgs:          2,
		ParseAndBuildMsg: BuildCreatePoolMsg,
	}, &model.MsgCreateCosmWasmPool{}
}

func BuildCreatePoolMsg(clientCtx client.Context, args []string, flags *pflag.FlagSet) (sdk.Msg, error) {
	codeId, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return nil, err
	}

	instantiateMsg := args[1]

	// Check JSON format for instantiateMsg
	var jsonCheck map[string]interface{}
	err = json.Unmarshal([]byte(instantiateMsg), &jsonCheck)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON format for instantiateMsg: %v", err)
	}

	// Turn instantiateMsg to bytes
	msgBz := []byte(instantiateMsg)

	return &model.MsgCreateCosmWasmPool{
		CodeId:         codeId,
		InstantiateMsg: msgBz,
		Sender:         clientCtx.GetFromAddress().String(),
	}, nil
}

func NewCmdUploadCodeIdAndWhitelistProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upload-code-id-and-whitelist [wasm-file-path] [flags]",
		Args:    cobra.ExactArgs(1),
		Short:   "Submit an upload code id and whitelist proposal",
		Example: "osmosisd tx gov submit-proposal upload-code-id-and-whitelist x/cosmwasmpool/bytecode/transmuter.wasm --from lo-test1 --keyring-backend test --title \"Test\" --summary \"Test\" -b=block --chain-id localosmosis --fees=100000uosmo --gas=20000000",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseUploadCodeIdAndWhitelistProposal(cmd, args[0])
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)

	return cmd
}

func parseUploadCodeIdAndWhitelistProposal(cmd *cobra.Command, fileName string) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	wasm, err := parseWasmByteCode(fileName)
	if err != nil {
		return nil, err
	}

	content := &types.UploadCosmWasmPoolCodeAndWhiteListProposal{
		Title:        title,
		Description:  description,
		WASMByteCode: wasm,
	}

	return content, nil
}

func parseWasmByteCode(fileName string) ([]byte, error) {
	if len(fileName) == 0 {
		return nil, fmt.Errorf("invalid input file. Provide file argument")
	}

	wasm, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	// gzip the wasm file
	if ioutils.IsWasm(wasm) {
		wasm, err = ioutils.GzipIt(wasm)

		if err != nil {
			return nil, err
		}
	} else if !ioutils.IsGzip(wasm) {
		return nil, fmt.Errorf("invalid input file. Use wasm binary or gzip")
	}

	return wasm, nil
}

func NewCmdMigratePoolContractsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-cw-pool-contracts [pool-ids] [new-code-id] [wasm-file-path] [flags]",
		Args:  cobra.ExactArgs(3),
		Short: "Submit a migrate cw pool contracts proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, proposalTitle, summary, deposit, isExpedited, authority, err := osmocli.GetProposalInfo(cmd)
			if err != nil {
				return err
			}

			content, err := parseMigratePoolContractsProposal(cmd, args)
			if err != nil {
				return err
			}

			contentMsg, err := v1.NewLegacyContent(content, authority.String())
			if err != nil {
				return err
			}

			msg := v1.NewMsgExecLegacyContent(contentMsg.Content, authority.String())

			proposalMsg, err := v1.NewMsgSubmitProposal([]sdk.Msg{msg}, deposit, clientCtx.GetFromAddress().String(), "", proposalTitle, summary, isExpedited)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposalMsg)
		},
	}
	osmocli.AddCommonProposalFlags(cmd)

	return cmd
}

func parseMigratePoolContractsProposal(cmd *cobra.Command, args []string) (govtypesv1beta1.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagSummary)
	if err != nil {
		return nil, err
	}

	poolIdsStr := strings.Split(args[0], ",")
	poolIds := make([]uint64, len(poolIdsStr))
	for i, poolIdStr := range poolIdsStr {
		poolId, err := strconv.ParseUint(poolIdStr, 10, 64)
		if err != nil {
			return nil, err
		}
		poolIds[i] = poolId
	}

	newCodeId, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return nil, err
	}

	wasm := []byte{}
	// Only attempt to parse the bytecode if code ID is not
	// given (i.e. 0)
	if newCodeId == 0 {
		byteCodeFileName := args[2]
		wasm, err = parseWasmByteCode(byteCodeFileName)
		if err != nil {
			return nil, err
		}
	}

	// TODO: implement this later if needed.
	emptyMigrateMsg, err := json.Marshal(struct{}{})
	if err != nil {
		return nil, err
	}

	content := &types.MigratePoolContractsProposal{
		Title:        title,
		Description:  description,
		PoolIds:      poolIds,
		NewCodeId:    newCodeId,
		WASMByteCode: wasm,
		MigrateMsg:   emptyMigrateMsg,
	}

	return content, nil
}
