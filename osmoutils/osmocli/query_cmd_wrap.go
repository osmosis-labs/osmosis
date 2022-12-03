package osmocli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func QueryIndexCmd(moduleName string) *cobra.Command {
	return &cobra.Command{
		Use:                        moduleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", moduleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       indexRunCmd,
	}
}

func indexRunCmd(cmd *cobra.Command, args []string) error {
	usageTemplate := `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}
  
{{if .HasAvailableSubCommands}}Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
	cmd.SetUsageTemplate(usageTemplate)
	return cmd.Help()
}
