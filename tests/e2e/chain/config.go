package chain

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/viper"
	tmconfig "github.com/tendermint/tendermint/config"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

const (
	// common
	OsmoDenom     = "uosmo"
	StakeDenom    = "stake"
	IbcDenom      = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
	MinGasPrice   = "0.000"
	IbcSendAmount = 3300000000
	// chainA
	ChainAID      = "osmo-test-a"
	OsmoBalanceA  = 200000000000
	StakeBalanceA = 110000000000
	StakeAmountA  = 100000000000
	// chainB
	ChainBID      = "osmo-test-b"
	OsmoBalanceB  = 500000000000
	StakeBalanceB = 440000000000
	StakeAmountB  = 400000000000
)

var (
	StakeAmountIntA  = sdk.NewInt(StakeAmountA)
	StakeAmountCoinA = sdk.NewCoin(StakeDenom, StakeAmountIntA)
	StakeAmountIntB  = sdk.NewInt(StakeAmountB)
	StakeAmountCoinB = sdk.NewCoin(StakeDenom, StakeAmountIntB)

	InitBalanceStrA = fmt.Sprintf("%d%s,%d%s", OsmoBalanceA, OsmoDenom, StakeBalanceA, StakeDenom)
	InitBalanceStrB = fmt.Sprintf("%d%s,%d%s", OsmoBalanceB, OsmoDenom, StakeBalanceB, StakeDenom)
)

func addAccount(path, moniker, amountStr string, accAddr sdk.AccAddress) error {
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

	authGenState := authtypes.GetGenesisStateFromAppState(util.Cdc, appState)

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

	authGenStateBz, err := util.Cdc.MarshalJSON(&authGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	appState[authtypes.ModuleName] = authGenStateBz

	bankGenState := banktypes.GetGenesisStateFromAppState(util.Cdc, appState)
	bankGenState.Balances = append(bankGenState.Balances, balances)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	bankGenStateBz, err := util.Cdc.MarshalJSON(bankGenState)
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

func initGenesis(c *internalChain) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(c.validators[0].configDir())
	config.Moniker = c.validators[0].getMoniker()

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	if err != nil {
		return err
	}

	var bankGenState banktypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState); err != nil {
		return err
	}

	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     OsmoDenom,
		Base:        OsmoDenom,
		Symbol:      OsmoDenom,
		Name:        OsmoDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    OsmoDenom,
				Exponent: 0,
			},
		},
	})

	bz, err := util.Cdc.MarshalJSON(&bankGenState)
	if err != nil {
		return err
	}
	appGenState[banktypes.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState); err != nil {
		return err
	}

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(c.validators))
	for i, val := range c.validators {
		stakeAmountCoin := StakeAmountCoinA
		if c.chainMeta.Id != ChainAID {
			stakeAmountCoin = StakeAmountCoinB
		}
		createValmsg, err := val.buildCreateValidatorMsg(stakeAmountCoin)
		if err != nil {
			return err
		}

		signedTx, err := val.signMsg(createValmsg)
		if err != nil {
			return err
		}

		txRaw, err := util.Cdc.MarshalJSON(signedTx)
		if err != nil {
			return err
		}

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err = util.Cdc.MarshalJSON(&genUtilGenState)
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
	for _, val := range c.validators {
		if err := util.WriteFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz); err != nil {
			return err
		}
	}
	return nil
}

func initNodes(c *internalChain) error {
	if err := c.createAndInitValidators(2); err != nil {
		return err
	}

	// initialize a genesis file for the first validator
	val0ConfigDir := c.validators[0].configDir()
	for _, val := range c.validators {
		if c.chainMeta.Id == ChainAID {
			if err := addAccount(val0ConfigDir, "", InitBalanceStrA, val.getKeyInfo().GetAddress()); err != nil {
				return err
			}
		} else if c.chainMeta.Id == ChainBID {
			if err := addAccount(val0ConfigDir, "", InitBalanceStrB, val.getKeyInfo().GetAddress()); err != nil {
				return err
			}
		}
	}

	// copy the genesis file to the remaining validators
	for _, val := range c.validators[1:] {
		_, err := util.CopyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func initValidatorConfigs(c *internalChain) error {
	for i, val := range c.validators {
		tmCfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		if err := vpr.ReadInConfig(); err != nil {
			return err
		}

		valConfig := &tmconfig.Config{}
		if err := vpr.Unmarshal(valConfig); err != nil {
			return err
		}

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.instanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		var peers []string

		for j := 0; j < len(c.validators); j++ {
			if i == j {
				continue
			}

			peer := c.validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.getNodeKey().ID(), peer.getMoniker(), j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", MinGasPrice, OsmoDenom)

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
	return nil
}
