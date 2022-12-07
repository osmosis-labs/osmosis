package osmocli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TxIndexCmd(moduleName string) *cobra.Command {
	cmd := IndexCmd(moduleName)
	cmd.Short = fmt.Sprintf("%s transactions subcommands", moduleName)
	return cmd
}

type TxCliDesc struct {
	Use     string
	Short   string
	Long    string
	Example string

	NumArgs int
	// Contract: len(args) = NumArgs
	ParseAndBuildMsg  func(clientCtx client.Context, args []string, flags *pflag.FlagSet) (sdk.Msg, error)
	TxSignerFieldName string

	Flags FlagDesc
	// Map of FieldName -> FlagName
	CustomFlagOverrides map[string]string
	// Map of FieldName -> CustomParseFn
	CustomFieldParsers map[string]CustomFieldParserFn
}

func AddTxCmd[M sdk.Msg](cmd *cobra.Command, f func() (*TxCliDesc, M)) {
	desc, _ := f()
	subCmd := BuildTxCli[M](desc)
	cmd.AddCommand(subCmd)
}

func BuildTxCli[M sdk.Msg](desc *TxCliDesc) *cobra.Command {
	desc.TxSignerFieldName = strings.ToLower(desc.TxSignerFieldName)
	if desc.NumArgs == 0 {
		// NumArgs = NumFields - 1, since 1 field is from the msg
		desc.NumArgs = ParseNumFields[M]() - 1 - len(desc.CustomFlagOverrides) - len(desc.CustomFieldParsers)
	}
	if desc.ParseAndBuildMsg == nil {
		desc.ParseAndBuildMsg = func(clientCtx client.Context, args []string, flags *pflag.FlagSet) (sdk.Msg, error) {
			flagAdvice := FlagAdvice{
				IsTx:                true,
				TxSenderFieldName:   desc.TxSignerFieldName,
				FromValue:           clientCtx.GetFromAddress().String(),
				CustomFlagOverrides: desc.CustomFlagOverrides,
				CustomFieldParsers:  desc.CustomFieldParsers,
			}.Sanitize()
			return ParseFieldsFromFlagsAndArgs[M](flagAdvice, flags, args)
		}
	}
	return desc.BuildCommandCustomFn()
}

// Creates a new cobra command given the description.
// Its up to then caller to add CLI flags, aside from `flags.AddTxFlagsToCmd(cmd)`
func (desc TxCliDesc) BuildCommandCustomFn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   desc.Use,
		Short: desc.Short,
		Args:  cobra.ExactArgs(desc.NumArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			msg, err := desc.ParseAndBuildMsg(clientCtx, args, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	if desc.Example != "" {
		cmd.Example = desc.Example
	}
	if desc.Long != "" {
		cmd.Long = desc.Long
	}

	flags.AddTxFlagsToCmd(cmd)
	AddFlags(cmd, desc.Flags)
	return cmd
}
