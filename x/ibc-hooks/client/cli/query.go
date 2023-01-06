package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"strings"

	"github.com/osmosis-labs/osmosis/osmoutils/osmocli"
	"github.com/osmosis-labs/osmosis/x/ibc-hooks/types"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	cmd.AddCommand(
		GetCmdWasmSender(),
	)
	return cmd
}

// GetCmdPoolParams return pool params.
func GetCmdWasmSender() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm-sender <channelID> <originalSender>",
		Short: "Generate the local address for a wasm hooks sender",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Generate the local address for a wasm hooks sender.
Example:
$ %s query ibc-hooks wasm-hooks-sender channel-42 juno12smx2wdlyttvyzvzg54y2vnqwq2qjatezqwqxu
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelID := args[0]
			originalSender := args[1]

			senderStr := fmt.Sprintf("%s/%s", channelID, originalSender)
			senderHash32 := address.Hash("ibc-memo-action", []byte(senderStr))
			sender := sdk.AccAddress(senderHash32[:])
			// ToDo: Make this flexible as an arg
			prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
			senderBech32, err := sdk.Bech32ifyAddressBytes(prefix, sender)
			if err != nil {
				return err
			}
			fmt.Println(senderBech32)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
