package genesis

import (
	"encoding/json"
	"path/filepath"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/server"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/common"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func AddAccount(path, moniker, amountStr string, accAddr sdk.AccAddress) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(path)
	config.Moniker = moniker

	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	balances := banktypes.Balance{Address: accAddr.String(), Coins: coins.Sort()}
	genAccount := authtypes.NewBaseAccount(accAddr, nil, 0, 0)

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	authGenState := authtypes.GetGenesisStateFromAppState(common.Cdc, appState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return fmt.Errorf("failed to get accounts from any: %w", err)
	}

	if accs.Contains(accAddr) {
		return fmt.Errorf("failed to add account to genesis state; account already exists: %s", accAddr)
	}

	// Add the new account to the set of genesis accounts and sanitize the
	// accounts afterwards.
	accs = append(accs, genAccount)
	accs = authtypes.SanitizeGenesisAccounts(accs)

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		return fmt.Errorf("failed to convert accounts into any's: %w", err)
	}

	authGenState.Accounts = genAccs

	authGenStateBz, err := common.Cdc.MarshalJSON(&authGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	appState[authtypes.ModuleName] = authGenStateBz

	bankGenState := banktypes.GetGenesisStateFromAppState(common.Cdc, appState)
	bankGenState.Balances = append(bankGenState.Balances, balances)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	bankGenStateBz, err := common.Cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}

	appState[banktypes.ModuleName] = bankGenStateBz

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}

	genDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(genDoc, genFile)
}

func Init(c *chain.Chain) error {
	_, cdc := common.InitEncodingConfigAndCdc()

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(c.Validators[0].ConfigDir())
	config.Moniker = c.Validators[0].GetMoniker()

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	if err != nil { 
		return err
	}

	var bankGenState banktypes.GenesisState
	if err := cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState); err != nil {
		return err
	}

	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     common.OsmoDenom,
		Base:        common.OsmoDenom,
		Symbol:      common.OsmoDenom,
		Name:        common.OsmoDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    common.OsmoDenom,
				Exponent: 0,
			},
		},
	})

	bz, err := cdc.MarshalJSON(&bankGenState)
	if err != nil {
		return err
	}
	appGenState[banktypes.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	if err := cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState); err != nil {
		return err
	}

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(c.Validators))
	for i, val := range c.Validators {
		stakeAmountCoin := common.StakeAmountCoinA
		if c.Id != common.ChainAID {
			stakeAmountCoin = common.StakeAmountCoinB
		}
		createValmsg, err := val.BuildCreateValidatorMsg(stakeAmountCoin)
		if err != nil {
			return err
		}

		signedTx, err := val.SignMsg(createValmsg)
		if err != nil {
			return err
		}

		txRaw, err := cdc.MarshalJSON(signedTx)
		if err != nil {
			return err
		}

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err = cdc.MarshalJSON(&genUtilGenState)
	if err != nil {
		return err
	}
	appGenState[genutiltypes.ModuleName] = bz

	bz, err = json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc.AppState = bz

	bz, err = tmjson.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}

	// write the updated genesis file to each validator
	for _, val := range c.Validators {
		common.WriteFile(filepath.Join(val.ConfigDir(), "config", "genesis.json"), bz)
	}
	return nil
}
