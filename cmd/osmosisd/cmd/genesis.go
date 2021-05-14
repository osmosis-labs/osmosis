package cmd

import (
	"encoding/json"
	"fmt"

	claimtypes "github.com/c-osmosis/osmosis/x/claim/types"
	appparams "github.com/c-osmosos/osmosis/app/params"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
)

func PrepareGenesisCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-genesis",
		Short: "Prepare a genesis file with initial setup",
		Long: `Prepare a genesis file with initial setup.
Examples include:
	- Setting module initial params
	- Setting denom metadata

Example:
	osmosisd validate-genesis
	- Check input genesis:
		file is at ~/.gaiad/config/genesis.json
`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)
			aminoCodec := clientCtx.LegacyAmino.Amino

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// bank module genesis
			bankGenState := banktypes.GetGenesisStateFromAppState(depCdc, appState)

			osmoMetadata := banktypes.Metadata{
				Description: fmt.Sprintf("The native token of Osmosis"),
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    appparams.BaseCoinUnit,
						Exponent: 0,
						Aliases: []string{
							"attopoolshare",
						},
					},
					{
						Denom:    appparams.HumanCoinUnit,
						Exponent: appparams.Exponent,
						Aliases:  nil,
					},
				},
				Base:    appparams.BaseCoinUnit,
				Display: appparams.HumanCoinUnit,
			}

			// bankGenState.DenomMetadata = banktypes.

			bankGenState.Balances = banktypes.SanitizeGenesisBalances(liquidBalances)
			bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}
			appState[banktypes.ModuleName] = bankGenStateBz

			// claim module genesis
			claimGenState := claimtypes.DefaultGenesis()
			claimGenState.ModuleAccountBalance = sdk.NewCoin(sdk.DefaultBondDenom, totalNormalizedOsmoBalance)
			claimGenState.InitialClaimables = claimBalances
			claimGenStateBz, err := cdc.MarshalJSON(claimGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal claim genesis state: %w", err)
			}
			appState[claimtypes.ModuleName] = claimGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}
			genDoc.AppState = appStateJSON

			err = genutil.ExportGenesisFile(genDoc, genFile)
			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flagOsmoSupply, "", "OSMO total genesis supply")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
