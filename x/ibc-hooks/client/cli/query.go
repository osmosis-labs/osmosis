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
			config := sdk.GetConfig()
			channelID := args[0]
			originalSender := args[1]

			bech32PrefixArg, err := cmd.Flags().GetString(FlagBech32Prefix)
			if err != nil {
				return err
			}

			hashPrefixArg, err := cmd.Flags().GetString(FlagHashPrefix)
			if err != nil {
				return err
			}

			var hashPrefix string
			if hashPrefixArg == "" {
				hashPrefix = types.SenderPrefix
			} else {
				hashPrefix = hashPrefixArg
			}

			var bech32Prefix string
			if bech32PrefixArg == "" {
				bech32Prefix = config.GetBech32AccountAddrPrefix()
			} else {
				bech32Prefix, err = getBech32CustomPrefix(config, bech32PrefixArg)
				if err != nil {
					return err
				}
			}

			senderBech32, err := keeper.DeriveIntermediateSender(channelID, originalSender, bech32Prefix, hashPrefix)
			if err != nil {
				return err
			}
			fmt.Println(senderBech32)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagBech32Prefix, "", "bech32 prefix to use in derivation")
	cmd.Flags().String(FlagHashPrefix, "", "hash prefix to use in derivation")

	return cmd
}
