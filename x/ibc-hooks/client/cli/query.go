package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/x/ibc-hooks/keeper"

	"github.com/osmosis-labs/osmosis/x/ibc-hooks/types"
)

func indexRunCmd(cmd *cobra.Command, args []string) error {
	usageTemplate := `Usage:{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}
  
{{if .HasAvailableSubCommands}}Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	cmd.SetUsageTemplate(usageTemplate)
	return cmd.Help()
}

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       indexRunCmd,
	}

	cmd.AddCommand(
		GetCmdWasmSender(),
	)
	return cmd
}

// GetCmdWasmSender returns a generated local address for a wasm hooks sender.
func GetCmdWasmSender() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm-sender <channelID> <originalSender>",
		Short: "Generate the local address for a wasm hooks sender",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Generate the local address for a wasm hooks sender.
Example:
$ %s query ibchooks wasm-hooks-sender channel-42 juno12smx2wdlyttvyzvzg54y2vnqwq2qjatezqwqxu
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelID := args[0]
			originalSender := args[1]

			prefixArg, err := cmd.Flags().GetString(FlagBech32Prefix)
			if err != nil {
				return err
			}

			config := sdk.GetConfig()

			var prefix string
			if prefixArg == "" {
				prefix = config.GetBech32AccountAddrPrefix()
			} else {
				prefix, err = getBech32CustomPrefix(config, prefixArg)
				if err != nil {
					return err
				}
			}

			senderBech32, err := keeper.DeriveIntermediateSender(channelID, originalSender, prefix)
			if err != nil {
				return err
			}
			fmt.Println(senderBech32)
			return nil
		},
	}

	cmd.Flags().String(FlagBech32Prefix, "", "bech32 prefix to use in derivation")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
