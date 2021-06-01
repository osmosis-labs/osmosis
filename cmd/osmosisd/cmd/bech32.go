package cmd

// DONTCOVER

import (
	appparams "github.com/osmosis-labs/osmosis/app/params"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// get cmd to convert any bech32 address to an osmo prefix
func ConvertBech32Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bech32-convert [bech32 string]",
		Short: "Convert any bech32 string to the osmo prefix",
		Long: `Convert any bech32 string to the osmo prefix
Especially useful for converting cosmos addresses to osmo addresses

Example:
	osmosisd bech32-convert cosmos1ey69r37gfxvxg62sh4r0ktpuc46pzjrmz29g45
	`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			appparams.SetAddressPrefixes()

			_, bz, err := bech32.DecodeAndConvert(args[0])
			if err != nil {
				return err
			}

			addr := sdk.AccAddress(bz)

			cmd.Printf("Osmo Bech32: %s\n", addr.String())

			return nil
		},
	}

	return cmd
}
