package cmd

import (
	"github.com/spf13/cobra"
)

// get cmd to convert any bech32 address to an osmo prefix.
func DebugStargateQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stargate-query [somethign]",
		Short: "Convert any bech32 string to the osmo prefix",
		Long: `Convert any bech32 string to the osmo prefix
Especially useful for converting cosmos addresses to osmo addresses

Example:
	osmosisd bech32-convert cosmos1ey69r37gfxvxg62sh4r0ktpuc46pzjrmz29g45
	`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// clientCtx := client.GetClientContextFromCmd(cmd)

			// // get cdc
			// depCdc := clientCtx.Codec
			// cdc := depCdc

			// // get grpc query router
			// grpcQueryRouter := baseapp.NewGRPCQueryRouter()
			// stargateQuerier := wasmbinding.StargateQuerier(*grpcQueryRouter, cdc)

			// requestPath := args[0]

			// cmd.Println(bech32Addr)

			return nil
		},
	}

	cmd.Flags().StringP(flagBech32Prefix, "p", "osmo", "Bech32 Prefix to encode to")

	return cmd
}

func ParsePathToStruct(path string) {

}
