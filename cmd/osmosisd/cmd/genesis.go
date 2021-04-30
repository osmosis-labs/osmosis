package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	claimtypes "github.com/c-osmosis/osmosis/x/claim/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	appparams "github.com/c-osmosis/osmosis/app/params"
)

func GenerateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-genesis [input-snapshot-file]",
		Short: "Export a genesis from fairdrop snapshot",
		Long: `Export a genesis from fairdrop snapshot
Example:
	osmosisd export-genesis ../snapshot.json
	- Check input genesis:
		file is at ~/.gaiad/config/genesis.json
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)
			aminoCodec := clientCtx.LegacyAmino.Amino

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			snapshotInput := args[0]
			osdenom := "uosmo"

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			// Read snapshot file
			snapshotJson, err := os.Open(snapshotInput)
			if err != nil {
				return err
			}
			defer snapshotJson.Close()

			byteValue, _ := ioutil.ReadAll(snapshotJson)

			// Produce the map of address to total atom balance, both staked and unstaked
			snapshot := make(map[string]SnapshotFields)
			err = aminoCodec.UnmarshalJSON(byteValue, &snapshot)
			if err != nil {
				return err
			}

			balances := []banktypes.Balance{}
			feeBalances := []banktypes.Balance{}

			totalOsmoBalance := sdk.NewInt(0)
			for _, acc := range snapshot {
				// calculate total osmo balance
				totalOsmoBalance = totalOsmoBalance.Add(acc.OsmoBalance)

				// set atom bech32 prefixes
				setCosmosBech32Prefixes()
				
				// read address from snapshot
				address, err := sdk.AccAddressFromBech32(acc.AtomAddress)
				if err != nil {
					return err
				}
 
				// set osmo bech32 prefixes
				appparams.SetBech32Prefixes()

				// airdrop balances
				coins := sdk.NewCoins(sdk.NewCoin(osdenom, acc.OsmoNormalizedBalance))
				balances = append(balances, banktypes.Balance{Address: address.String(), Coins: coins})

				// transaction fee balances
				feeCoins := sdk.NewCoins(sdk.NewCoin(osdenom, sdk.NewInt(1e6))) // 1 OSMO = 10^6 uosmo
				feeBalances = append(feeBalances, banktypes.Balance{Address: address.String(), Coins: feeCoins})
			}

			// auth module genesis
			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to convert accounts into any's: %w", err)
			}
			authGenState.Accounts = genAccs
			authGenStateBz, err := cdc.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}
			appState[authtypes.ModuleName] = authGenStateBz

			// bank module genesis
			bankGenState := banktypes.GetGenesisStateFromAppState(depCdc, appState)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(feeBalances)
			bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}
			appState[banktypes.ModuleName] = bankGenStateBz

			// claim module genesis
			claimGenState := claimtypes.DefaultGenesis()
			claimGenState.AirdropAmount = totalOsmoBalance
			claimGenState.Claimables = balances
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

	cmd.Flags().String(flagOsmoSupply, "", "OSMO total genesis supply")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
