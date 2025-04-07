package cmd

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	sims "github.com/cosmos/cosmos-sdk/testutil/sims"
	rosettaCmd "github.com/cosmos/rosetta/cmd"
	"github.com/prometheus/client_golang/prometheus"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	cosmosdb "github.com/cosmos/cosmos-db"

	confixcmd "cosmossdk.io/tools/confix/cmd"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/params"

	"cosmossdk.io/log"
	tmcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/bytes"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/joho/godotenv"

	symphony "github.com/osmosis-labs/osmosis/v27/app"
)

type AssetList struct {
	Assets []Asset `json:"assets"`
}
type Asset struct {
	DenomUnits []DenomUnit `json:"denom_units"`
	Symbol     string      `json:"symbol"`
	Base       string      `json:"base"`
	Traces     []Trace     `json:"traces"`
}

type DenomUnit struct {
	Denom    string   `json:"denom"`
	Exponent uint64   `json:"exponent"`
	Aliases  []string `json:"aliases"`
}

type Trace struct {
	Type         string `json:"type"`
	Counterparty struct {
		BaseDenom string `json:"base_denom"`
	} `json:"counterparty"`
}

type DenomUnitMap struct {
	Base     string
	Exponent uint64 `json:"exponent"`
}

type SectionKeyValue struct {
	Section string
	Key     string
	Value   any
}

var (
	recommendedAppTomlValues = []SectionKeyValue{
		{
			Section: "",
			Key:     "minimum-gas-prices",
			Value:   "0uosmo",
		},
		{
			Section: "symphony-mempool",
			Key:     "arbitrage-min-gas-fee",
			Value:   "0.1",
		},
		{
			Section: "symphony-mempool",
			Key:     "max-gas-wanted-per-tx",
			Value:   "60000000",
		},
		{
			Section: "wasm",
			Key:     "memory_cache_size",
			Value:   1000,
		},
	}

	recommendedConfigTomlValues = []SectionKeyValue{
		{
			Section: "p2p",
			Key:     "flush_throttle_timeout",
			Value:   "80ms",
		},
		{
			Section: "consensus",
			Key:     "timeout_commit",
			Value:   "500ms",
		},
		{
			Section: "consensus",
			Key:     "timeout_propose",
			Value:   "1.8s",
		},
		{
			Section: "consensus",
			Key:     "peer_gossip_sleep_duration",
			Value:   "50ms",
		},
	}
)

var (
	//go:embed "symphony-1-assetlist.json" "melody-test-5-assetlist.json"
	assetFS   embed.FS
	mainnetId = "symphony-1"
	testnetId = "symphony-test-5"
)

func loadAssetList(initClientCtx client.Context, cmd *cobra.Command, basedenomToIBC, IBCtoBasedenom bool) (map[string]DenomUnitMap, map[string]string) {
	var assetList AssetList

	chainId := GetChainId(initClientCtx, cmd)
	homeDir := initClientCtx.HomeDir

	fileName := ""
	if chainId == mainnetId || chainId == "" {
		fileName = filepath.Join(homeDir, "config", "symphony-1-assetlist-manual.json")
	} else if chainId == testnetId {
		fileName = filepath.Join(homeDir, "config", "osmo-test-5-assetlist-manual.json")
	} else {
		return nil, nil
	}

	// The order of precedence for asset list is:
	//  - If the manually generated asset list (generated via the `update-asset-list` cli cmd) exists (noted by -manual.json ending), use it.
	//  - If the manually generated asset list does not exist, fall back to the embedded asset list.
	localFile, err := os.Open(fileName)
	if err != nil {
		// If we can't open the local file, fall back to the embedded file.
		if chainId == mainnetId || chainId == "" {
			fileName = "symphony-1-assetlist.json"
		} else if chainId == testnetId {
			fileName = "osmo-test-5-assetlist.json"
		} else {
			return nil, nil
		}
		embeddedFile, err := assetFS.Open(fileName)
		if err != nil {
			return nil, nil
		}
		defer embeddedFile.Close()

		byteValue, _ := io.ReadAll(embeddedFile)
		err = json.Unmarshal(byteValue, &assetList)
		if err != nil {
			return nil, nil
		}
	} else {
		// If the local file opens successfully, use it instead of the embedded file.
		defer localFile.Close()
		byteValue, _ := io.ReadAll(localFile)
		err = json.Unmarshal(byteValue, &assetList)
		if err != nil {
			return nil, nil
		}
	}

	baseMap := make(map[string]DenomUnitMap)
	baseMapRev := make(map[string]string)

	if basedenomToIBC {
		for _, asset := range assetList.Assets {
			// Each asset has a list of denom units. A majority of them have 2 entries, one being the base 0 exponent denom and the other being a larger exponent denom.
			// An example for tether:
			// * Exponent 0: uusdt
			// * Exponent 6: usdt
			// This implies that if a usdt value is given, in order to convert it to it's base denom (uusdt), we need to multiply the provided value by 10^6.
			for i, denomUnit := range asset.DenomUnits {
				DenomUnitMap := DenomUnitMap{
					Base:     asset.Base,
					Exponent: asset.DenomUnits[i].Exponent,
				}
				// The 0 exponent denom is the base denom.
				if asset.DenomUnits[i].Exponent == 0 {
					// To make everyone's life harder, some assets have multiple base denom aliases. For example, the asset list has the following base aliases for the asset "luna":
					// * uluna
					// * microluna
					for _, alias := range denomUnit.Aliases {
						baseMap[strings.ToLower(alias)] = DenomUnitMap
					}
				} else {
					// Otherwise we just store the denom alias for that exponent.
					baseMap[strings.ToLower(denomUnit.Denom)] = DenomUnitMap
				}
			}
		}
	}
	if IBCtoBasedenom {
		// We just store a link from the first base denom alias to the IBC denom. This is just used for display purposes on the terminal's output.
		for _, asset := range assetList.Assets {
			if len(asset.DenomUnits) > 0 && asset.DenomUnits[0].Exponent == 0 && len(asset.DenomUnits[0].Aliases) > 0 {
				baseDenom := asset.DenomUnits[0].Aliases[0]
				baseMapRev[asset.Base] = strings.ToLower(baseDenom)
			}
		}
	}
	return baseMap, baseMapRev
}

type customWriter struct {
	originalOut io.Writer
	baseMap     map[string]string
}

func (cw *customWriter) Write(p []byte) (n int, err error) {
	// Convert byte slice to string.
	s := string(p)

	// Buffer to hold the new string.
	var buf strings.Builder

	// Index where the current denom starts. -1 if we're not currently in a denom.
	denomStart := -1

	// Counter for slashes encountered
	slashCounter := 0

	re, err := regexp.Compile("[^a-zA-Z0-9/-]")
	if err != nil {
		return 0, err
	}

	for i := 0; i < len(s); i++ {
		if denomStart == -1 {
			// If we're not currently in a denom, check if this character starts a new denom.
			// Check for "ibc/" or "factory/" prefix.
			if strings.HasPrefix(s[i:], "ibc/") || strings.HasPrefix(s[i:], "factory/") {
				slashCounter = 0
				denomStart = i
				continue
			}
			// Write the character to the buffer.
			buf.WriteByte(s[i])
		} else {
			// For factory denoms, we keep track of slashes to find the second slash.
			if s[i] == '/' {
				slashCounter++
			}

			// For "ibc/" we find the end by length, for "factory/" we find the end by second slash and regex.
			if (strings.HasPrefix(s[denomStart:], "ibc/") && i-denomStart == 68) || (strings.HasPrefix(s[denomStart:], "factory/") && slashCounter == 2 && re.MatchString(string(s[i]))) {
				// We've reached the end of the line containing the denom.
				denom := s[denomStart:i]
				if replacement, ok := cw.baseMap[denom]; ok {
					// If the denom is in the map, write the replacement to the buffer.
					buf.WriteString(replacement)
				} else {
					// If the denom is not in the map, write the original denom to the buffer.
					buf.WriteString(denom)
				}
				// Write the new line character to the buffer.
				buf.WriteByte(s[i])

				// We're no longer in a denom.
				denomStart = -1
				slashCounter = 0
			}
		}
	}

	// If we're still in a denom at the end of the string, write the rest of the denom to the buffer.
	if denomStart != -1 {
		denom := s[denomStart:]
		if replacement, ok := cw.baseMap[denom]; ok {
			buf.WriteString(replacement)
		} else {
			buf.WriteString(denom)
		}
	}

	// Write the new string to the original output.
	return cw.originalOut.Write([]byte(buf.String()))
}

// NewRootCmd creates a new root command for simd. It is called once in the
// main function.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	encodingConfig := symphony.MakeEncodingConfig()
	homeEnvironment := getHomeEnvironment()
	homeDir, err := environmentNameToPath(homeEnvironment)
	if err != nil {
		// Failed to convert home environment to home path, using default home
		homeDir = symphony.DefaultNodeHome
	}

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(homeDir).
		WithViper("SYMPHONY")

	tempDir := tempDir()
	tempApp := symphony.NewSymphonyApp(
		log.NewNopLogger(),
		cosmosdb.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		tempDir,
		5,
		sims.EmptyAppOptions{},
		symphony.EmptyWasmOpts,
		baseapp.SetChainID("symphony-1"),
	)

	// Allows you to add extra params to your client.toml
	// gas, gas-price, gas-adjustment, and human-readable-denoms
	SetCustomEnvVariablesFromClientToml(initClientCtx)
	humanReadableDenomsInput, humanReadableDenomsOutput := GetHumanReadableDenomEnvVariables()

	rootCmd := &cobra.Command{
		Use:   "symphonyd",
		Short: "Start symphony app",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// If not calling the set-env command, this is a no-op.
			err := changeEnvPriorToSetup(cmd, &initClientCtx, args, homeDir)
			if err != nil {
				return err
			}

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			// Only loads asset list into a map if human readable denoms are enabled.
			assetMap, assetMapRev := map[string]DenomUnitMap{}, map[string]string{}
			if humanReadableDenomsInput || humanReadableDenomsOutput {
				assetMap, assetMapRev = loadAssetList(initClientCtx, cmd, humanReadableDenomsInput, humanReadableDenomsOutput)
			}

			// If enabled, CLI output will be parsed and human readable denominations will be used in place of ibc denoms.
			if humanReadableDenomsOutput {
				initClientCtx.Output = &customWriter{originalOut: os.Stdout, baseMap: assetMapRev}
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}
			customAppTemplate, customAppConfig := initAppConfig()

			// If enabled, CLI input will be parsed and human readable denominations will be automatically converted to ibc denoms.
			if humanReadableDenomsInput {
				// Parse and replace denoms in args
				for i, arg := range args {
					argArray := strings.Split(arg, ",")

					re := regexp.MustCompile(`^([\d.]+)(\D+)$`)

					for i, singleArg := range argArray {
						lowerCaseArg := strings.ToLower(singleArg)
						match := re.FindStringSubmatch(lowerCaseArg)
						if len(match) == 3 {
							value, denom := match[1], match[2]
							// If the index has a length of 3 then it has a number and a denom (this is a coin object)
							// Note, index 0 is the entire string, index 1 is the number, and index 2 is the denom
							transformedCoin, err := transformCoinValueToBaseInt(value, denom, assetMap)
							if err != nil {
								continue
							}
							argArray[i] = transformedCoin
						} else {
							if _, ok := assetMap[lowerCaseArg]; ok {
								// In this case, we just need to replace the denom with the base denom
								argArray[i] = assetMap[lowerCaseArg].Base
							}
						}
					}
					args[i] = strings.Join(argArray, ",")
				}

				// Parse and replace denoms in flags
				cmd.Flags().VisitAll(func(flag *pflag.Flag) {
					lowerCaseFlagValue := strings.ToLower(flag.Value.String())
					lowerCaseFlagValueArray := strings.Split(lowerCaseFlagValue, ",")

					re := regexp.MustCompile(`^([\d.]+)(\D+)$`)

					for i, lowerCaseFlagValue := range lowerCaseFlagValueArray {
						match := re.FindStringSubmatch(lowerCaseFlagValue)
						if len(match) == 3 {
							value, denom := match[1], match[2]
							// If the index has a length of 3 then it has a number and a denom (this is a coin object)
							// Note, index 0 is the entire string, index 1 is the number, and index 2 is the denom
							transformedCoin, err := transformCoinValueToBaseInt(value, denom, assetMap)
							if err != nil {
								continue
							}
							lowerCaseFlagValueArray[i] = transformedCoin
						} else {
							if _, ok := assetMap[lowerCaseFlagValue]; ok {
								// Otherwise, we just need to replace the denom with the base denom
								lowerCaseFlagValueArray[i] = assetMap[lowerCaseFlagValue].Base
							}
						}
					}
					newLowerCaseFlagValue := strings.Join(lowerCaseFlagValueArray, ",")
					if lowerCaseFlagValue != newLowerCaseFlagValue {
						if err := cmd.Flags().Set(flag.Name, newLowerCaseFlagValue); err != nil {
							fmt.Println("Failed to set flag:", err)
						}
					}
				})
			}
			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, tmcfg.DefaultConfig())
		},
		SilenceUsage: true,
	}

	genAutoCompleteCmd(rootCmd)

	initRootCmd(rootCmd, encodingConfig, tempApp)

	if err := autoCliOpts(initClientCtx, tempApp).EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd, encodingConfig
}

// tempDir create a temporary directory to initialize the command line client
func tempDir() string {
	dir, err := os.MkdirTemp("", "symphonyd")
	if err != nil {
		dir = symphony.DefaultNodeHome
	}
	defer os.RemoveAll(dir)

	return dir
}

// overwriteConfigTomlValues overwrites config.toml values. Returns error if config.toml does not exist
//
// Currently, overwrites:
// - timeout_commit
//
// Also overwrites the respective viper config value.
//
// Silently handles and skips any error/panic due to write permission issues.
// No-op otherwise.
func overwriteConfigTomlValues(serverCtx *server.Context) error {
	// Get paths to config.toml and config parent directory
	rootDir := serverCtx.Viper.GetString(tmcli.HomeFlag)

	configParentDirPath := filepath.Join(rootDir, "config")
	configFilePath := filepath.Join(configParentDirPath, "config.toml")

	fileInfo, err := os.Stat(configFilePath)
	if err != nil {
		// something besides a does not exist error
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read in %s: %w", configFilePath, err)
		}
	} else {
		// config.toml exists

		// Check if each key is already set to the recommended value
		// If it is, we don't need to overwrite it and can also skip the app.toml overwrite
		var sectionKeyValuesToWrite []SectionKeyValue

		// Set aside which keys need to be updated in the config.toml
		for _, rec := range recommendedConfigTomlValues {
			currentValue := serverCtx.Viper.Get(rec.Section + "." + rec.Key)
			if currentValue != rec.Value {
				// Current value in config.toml is not the recommended value
				// Set the value in viper to the recommended value
				// and add it to the list of key values we will overwrite in the config.toml
				serverCtx.Viper.Set(rec.Section+"."+rec.Key, rec.Value)
				sectionKeyValuesToWrite = append(sectionKeyValuesToWrite, rec)
			}
		}

		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("failed to write to %s: %s\n", configFilePath, err)
			}
		}()

		// Check if the file is writable
		if fileInfo.Mode()&os.FileMode(0o200) != 0 {
			// It will be re-read in server.InterceptConfigsPreRunHandler
			// this may panic for permissions issues. So we catch the panic.
			// Note that this exits with a non-zero exit code if fails to write the file.

			// Write the new config.toml file
			if len(sectionKeyValuesToWrite) > 0 {
				err := OverwriteWithCustomConfig(configFilePath, sectionKeyValuesToWrite)
				if err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("config.toml is not writable. Cannot apply update. Please consider manually changing to the following: %v\n", recommendedConfigTomlValues)
		}
	}
	return nil
}

// overwriteAppTomlValues overwrites app.toml values. Returns error if app.toml does not exist
//
// Currently, overwrites:
// - arbitrage-min-gas-fee
// - max-gas-wanted-per-tx
//
// Also overwrites the respective viper config value.
//
// Silently handles and skips any error/panic due to write permission issues.
// No-op otherwise.
func overwriteAppTomlValues(serverCtx *server.Context) error {
	// Get paths to app.toml and config parent directory
	rootDir := serverCtx.Viper.GetString(tmcli.HomeFlag)

	configParentDirPath := filepath.Join(rootDir, "config")
	appFilePath := filepath.Join(configParentDirPath, "app.toml")

	fileInfo, err := os.Stat(appFilePath)
	if err != nil {
		// something besides a does not exist error
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read in %s: %w", appFilePath, err)
		}
	} else {
		// app.toml exists

		// Check if each key is already set to the recommended value
		// If it is, we don't need to overwrite it and can also skip the app.toml overwrite
		var sectionKeyValuesToWrite []SectionKeyValue

		for _, rec := range recommendedAppTomlValues {
			currentValue := serverCtx.Viper.Get(rec.Section + "." + rec.Key)
			if currentValue != rec.Value {
				// Current value in app.toml is not the recommended value
				// Set the value in viper to the recommended value
				// and add it to the list of key values we will overwrite in the app.toml
				serverCtx.Viper.Set(rec.Section+"."+rec.Key, rec.Value)
				sectionKeyValuesToWrite = append(sectionKeyValuesToWrite, rec)
			}
		}

		// Check if the file is writable
		if fileInfo.Mode()&os.FileMode(0o200) != 0 {
			// It will be re-read in server.InterceptConfigsPreRunHandler
			// this may panic for permissions issues. So we catch the panic.
			// Note that this exits with a non-zero exit code if fails to write the file.

			// Write the new app.toml file
			if len(sectionKeyValuesToWrite) > 0 {
				err := OverwriteWithCustomConfig(appFilePath, sectionKeyValuesToWrite)
				if err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("app.toml is not writable. Cannot apply update. Please consider manually changing to the following: %v\n", recommendedAppTomlValues)
		}
	}
	return nil
}

func getHomeEnvironment() string {
	envPath := filepath.Join(symphony.DefaultNodeHome, ".env")

	// Use default node home if can't get environment.
	// Overload must be used here in the event that the .env gets updated.
	err := godotenv.Overload(envPath)
	if err != nil {
		// Failed to load, using default home directory
		return EnvMainnet
	}
	val := os.Getenv(EnvVariable)
	return val
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	type SymphonyMempoolConfig struct {
		MaxGasWantedPerTx         string `mapstructure:"max-gas-wanted-per-tx"`
		MinGasPriceForArbitrageTx string `mapstructure:"arbitrage-min-gas-fee"`
		MinGasPriceForHighGasTx   string `mapstructure:"min-gas-price-for-high-gas-tx"`
		Mempool1559Enabled        string `mapstructure:"adaptive-fee-enabled"`
	}

	type CustomAppConfig struct {
		serverconfig.Config

		SymphonyMempoolConfig SymphonyMempoolConfig `mapstructure:"symphony-mempool"`

		WasmConfig wasmtypes.WasmConfig `mapstructure:"wasm"`
	}

	DefaultSymphonyMempoolConfig := SymphonyMempoolConfig{
		MaxGasWantedPerTx:         "60000000",
		MinGasPriceForArbitrageTx: ".1",
		MinGasPriceForHighGasTx:   ".0025",
		Mempool1559Enabled:        "true",
	}
	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	srvCfg.API.Enable = true
	srvCfg.MinGasPrices = "0note"

	// 128MB IAVL cache
	srvCfg.IAVLCacheSize = 781250

	memCfg := DefaultSymphonyMempoolConfig

	wasmCfg := wasmtypes.DefaultWasmConfig()

	SymphonyAppCfg := CustomAppConfig{Config: *srvCfg, SymphonyMempoolConfig: memCfg, WasmConfig: wasmCfg}

	SymphonyAppTemplate := serverconfig.DefaultConfigTemplate + `
###############################################################################
###                      Symphony Mempool Configuration                      ###
###############################################################################

[symphony-mempool]
# This is the max allowed gas any tx.
# This is only for local mempool purposes, and thus	is only ran on check tx.
max-gas-wanted-per-tx = "{{ .SymphonyMempoolConfig.MaxGasWantedPerTx }}"

# This is the minimum gas fee any arbitrage tx should have, denominated in uosmo per gas
# Default value of ".1" then means that a tx with 1 million gas costs (.1 uosmo/gas) * 1_000_000 gas = .1 osmo
arbitrage-min-gas-fee = "{{ .SymphonyMempoolConfig.MinGasPriceForArbitrageTx }}"

# This is the minimum gas fee any tx with high gas demand should have, denominated in uosmo per gas
# Default value of ".0025" then means that a tx with 1 million gas costs (.0025 uosmo/gas) * 1_000_000 gas = .0025 osmo
min-gas-price-for-high-gas-tx = "{{ .SymphonyMempoolConfig.MinGasPriceForHighGasTx }}"

# This parameter enables EIP-1559 like fee market logic in the mempool
adaptive-fee-enabled = "{{ .SymphonyMempoolConfig.Mempool1559Enabled }}"

###############################################################################
###                            Wasm Configuration                           ###
###############################################################################
` + wasmtypes.DefaultConfigTemplate()

	return SymphonyAppTemplate, SymphonyAppCfg
}

// initRootCmd initializes root commands when creating a new root command for simd.
func initRootCmd(rootCmd *cobra.Command, encodingConfig params.EncodingConfig, tempApp *symphony.SymphonyApp) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	debugCmd := debug.Cmd()
	debugCmd.AddCommand(ConvertBech32Cmd())
	debugCmd.AddCommand(DebugProtoMarshalledBytes())

	valOperAddressCodec := address.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())
	rootCmd.AddCommand(
		// genutilcli.InitCmd(tempApp.ModuleBasics, symphony.DefaultNodeHome),
		forceprune(),
		moduleHashByHeightQuery(newApp),
		InitCmd(tempApp.ModuleBasics, symphony.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, symphony.DefaultNodeHome, genutiltypes.DefaultMessageValidator, valOperAddressCodec),
		ExportDeriveBalancesCmd(),
		StakedToCSVCmd(),
		AddGenesisAccountCmd(symphony.DefaultNodeHome),
		genutilcli.GenTxCmd(tempApp.ModuleBasics, encodingConfig.TxConfig, banktypes.GenesisBalancesIterator{}, symphony.DefaultNodeHome, valOperAddressCodec),
		genutilcli.ValidateGenesisCmd(tempApp.ModuleBasics),
		PrepareGenesisCmd(symphony.DefaultNodeHome, tempApp.ModuleBasics),
		tmcli.NewCompletionCmd(rootCmd, true),
		testnetCmd(tempApp.ModuleBasics, banktypes.GenesisBalancesIterator{}),
		debugCmd,
		confixcmd.ConfigCommand(),
		ChangeEnvironmentCmd(),
		PrintEnvironmentCmd(),
		PrintAllEnvironmentCmd(),
		UpdateAssetListCmd(symphony.DefaultNodeHome, tempApp.ModuleBasics),
		snapshot.Cmd(newApp),
		pruning.Cmd(newApp, symphony.DefaultNodeHome),
	)

	server.AddCommands(rootCmd, symphony.DefaultNodeHome, newApp, createSymphonyAppAndExport, addModuleInitFlags)
	server.AddTestnetCreatorCommand(rootCmd, newTestnetApp, addModuleInitFlags)

	for i, cmd := range rootCmd.Commands() {
		if cmd.Name() == "start" {
			startRunE := cmd.RunE

			// Instrument start command pre run hook with custom logic
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				serverCtx := server.GetServerContextFromCmd(cmd)

				// Get flag value for rejecting config defaults
				rejectConfigDefaults := serverCtx.Viper.GetBool(FlagRejectConfigDefaults)

				// overwrite config.toml and app.toml values, if rejectConfigDefaults is false
				if !rejectConfigDefaults {
					// Add ctx logger line to indicate that config.toml and app.toml values are being overwritten
					serverCtx.Logger.Info("Overwriting config.toml and app.toml values with some recommended defaults. To prevent this, set the --reject-config-defaults flag to true.")

					err := overwriteConfigTomlValues(serverCtx)
					if err != nil {
						return err
					}

					err = overwriteAppTomlValues(serverCtx)
					if err != nil {
						return err
					}
				}

				return startRunE(cmd, args)
			}

			rootCmd.Commands()[i] = cmd
			break
		}
	}

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		queryCommand(),
		txCommand(tempApp.ModuleBasics),
		keys.Commands(),
	)
	rootCmd.AddCommand(CmdListQueries(rootCmd))
	// add rosetta
	rootCmd.AddCommand(rosettaCmd.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Marshaler))
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	wasm.AddModuleInitFlags(startCmd)
	startCmd.Flags().Bool(FlagRejectConfigDefaults, false, "Reject some select recommended default values from being automatically set in the config.toml and app.toml")
}

// CmdListQueries list all available modules' queries
func CmdListQueries(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-queries",
		Short: "listing all available modules' queries",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() != "query" {
					continue
				}
				for _, cmd := range cmd.Commands() {
					for _, cmd := range cmd.Commands() {
						fmt.Println(cmd.CommandPath())
					}
				}
			}
			return nil
		},
	}
	return cmd
}

func CmdModuleNameToAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-name-to-address [module-name]",
		Short: "module name to address",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			address := authtypes.NewModuleAddress(args[0])
			return clientCtx.PrintString(address.String())
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// queryCommand adds transaction and account querying commands.
func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
		CmdModuleNameToAddress(),
	)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// txCommand adds transaction signing, encoding / decoding, and broadcasting commands.
func txCommand(moduleBasics module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)

	moduleBasics.AddTxCommands(cmd)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// newApp initializes and returns a new Symphony app.
func newApp(logger log.Logger, db cosmosdb.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	var cache storetypes.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(err)
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := cosmosdb.NewGoLevelDB("metadata", snapshotDir, nil)
	if err != nil {
		panic(err)
	}
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(err)
	}
	snapshotOptions := snapshottypes.NewSnapshotOptions(
		cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval)),
		cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent)),
	)

	homeDir := cast.ToString(appOpts.Get(flags.FlagHome))
	chainID := cast.ToString(appOpts.Get(flags.FlagChainID))
	if chainID == "" {
		// fallback to genesis chain-id
		appGenesis, err := genutiltypes.AppGenesisFromFile(filepath.Join(homeDir, "config", "genesis.json"))
		if err != nil {
			panic(err)
		}

		chainID = appGenesis.ChainID
	}

	var wasmOpts []wasmkeeper.Option
	if cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}

	fastNodeModuleWhitelist := server.ParseModuleWhitelist(appOpts)

	baseAppOptions := []func(*baseapp.BaseApp){
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(cast.ToString(appOpts.Get(server.FlagMinGasPrices))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		baseapp.SetIAVLCacheSize(cast.ToInt(appOpts.Get(server.FlagIAVLCacheSize))),
		baseapp.SetIAVLDisableFastNode(cast.ToBool(appOpts.Get(server.FlagDisableIAVLFastNode))),
		baseapp.SetIAVLFastNodeModuleWhitelist(fastNodeModuleWhitelist),
		baseapp.SetChainID(chainID),
	}

	// If this is an in place testnet, set any new stores that may exist
	if cast.ToBool(appOpts.Get(server.KeyIsTestnet)) {
		baseAppOptions = append(baseAppOptions)
	}

	return symphony.NewSymphonyApp(
		logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		appOpts,
		wasmOpts,
		baseAppOptions...,
	)
}

// newTestnetApp starts by running the normal newApp method. From there, the app interface returned is modified in order
// for a testnet to be created from the provided app.
func newTestnetApp(logger log.Logger, db cosmosdb.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	// Create an app and type cast to an SymphonyApp
	app := newApp(logger, db, traceStore, appOpts)
	symphonyApp, ok := app.(*symphony.SymphonyApp)
	if !ok {
		panic("app created from newApp is not of type symphonyApp")
	}

	newValAddr, ok := appOpts.Get(server.KeyNewValAddr).(bytes.HexBytes)
	if !ok {
		panic("newValAddr is not of type bytes.HexBytes")
	}
	newValPubKey, ok := appOpts.Get(server.KeyUserPubKey).(crypto.PubKey)
	if !ok {
		panic("newValPubKey is not of type crypto.PubKey")
	}
	newOperatorAddress, ok := appOpts.Get(server.KeyNewOpAddr).(string)
	if !ok {
		panic("newOperatorAddress is not of type string")
	}
	upgradeToTrigger, ok := appOpts.Get(server.KeyTriggerTestnetUpgrade).(string)
	if !ok {
		panic("upgradeToTrigger is not of type string")
	}

	// Make modifications to the normal SymphonyApp required to run the network locally
	return symphony.InitSymphonyAppForTestnet(symphonyApp, newValAddr, newValPubKey, newOperatorAddress, upgradeToTrigger)
}

// createSymphonyAppAndExport creates and exports the new Symphony app, returns the state of the new Symphony app for a genesis file.
func createSymphonyAppAndExport(
	logger log.Logger, db cosmosdb.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
	appOpts servertypes.AppOptions, modulesToExport []string,
) (servertypes.ExportedApp, error) {
	encCfg := symphony.MakeEncodingConfig() // Ideally, we would reuse the one created by NewRootCmd.
	encCfg.Marshaler = codec.NewProtoCodec(encCfg.InterfaceRegistry)
	loadLatest := height == -1
	homeDir := cast.ToString(appOpts.Get(flags.FlagHome))
	app := symphony.NewSymphonyApp(logger, db, traceStore, loadLatest, map[int64]bool{}, homeDir, 0, appOpts, symphony.EmptyWasmOpts)

	if !loadLatest {
		if err := app.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return app.ExportAppStateAndValidators(forZeroHeight, jailWhiteList, modulesToExport)
}

func UpdateAssetListCmd(defaultNodeHome string, mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-asset-list [chain-id]",
		Short: "Updates asset list used by the CLI to replace ibc denoms with human readable names",
		Long: `Updates asset list used by the CLI to replace ibc denoms with human readable names.
Outputs:
	- symphonydHomeDir + /config/symphony-1-assetlist-manual.json for symphony-1
	- symphonydHomeDir + /config/osmo-test-5-assetlist-manual.json for osmo-test-5
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assetListURL := ""
			fileName := ""

			clientCtx := client.GetClientContextFromCmd(cmd)
			homeDir := clientCtx.HomeDir

			chainID := ""
			if len(args) > 0 {
				chainID = args[0]
			} else {
				fmt.Println("No chain ID provided, defaulting to mainnet")
				chainID = mainnetId
			}

			if chainID == mainnetId {
				assetListURL = "https://raw.githubusercontent.com/symphony-labs/assetlists/main/symphony-1/symphony-1.assetlist.json"
				fileName = filepath.Join(homeDir, "config", "symphony-1-assetlist-manual.json")
			} else if chainID == testnetId {
				assetListURL = "https://raw.githubusercontent.com/symphony-labs/assetlists/main/osmo-test-5/osmo-test-5.assetlist.json"
				fileName = filepath.Join(homeDir, "config", "osmo-test-5-assetlist-manual.json")
			} else {
				return nil
			}

			// Try to fetch the asset list from the URL
			resp, err := http.Get(assetListURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Create a new file
			out, err := os.Create(fileName)
			if err != nil {
				return err
			}
			defer out.Close()

			// Copy the response body to the new file
			_, err = io.Copy(out, resp.Body)
			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")

	return cmd
}

func genAutoCompleteCmd(rootCmd *cobra.Command) {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "enable-cli-autocomplete [bash|zsh|fish|powershell]",
		Short: "Generates cli completion scripts",
		Long: `To configure your shell to load completions for each session, add to your profile:

# bash example
echo '. <(symphonyd enable-cli-autocomplete bash)' >> ~/.bash_profile
source ~/.bash_profile

# zsh example
echo '. <(symphonyd enable-cli-autocomplete zsh)' >> ~/.zshrc
source ~/.zshrc
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				_ = cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	})
}

// transformCoinValueToBaseInt transforms a cli input that has been split into a number and a denom into it's base int value and base denom.
// i.e. 10.7osmo -> 10700000uosmo
// 12atom -> 12000000uatom
// 15000000uakt -> 15000000uakt (does nothing since it's already in base denom format)
func transformCoinValueToBaseInt(coinValue, coinDenom string, assetMap map[string]DenomUnitMap) (string, error) {
	// If the index has a length of 3 then it has a number and a denom (this is a coin object)
	// Note, index 0 is the entire string, index 1 is the number, and index 2 is the denom
	if denomUnitMap, ok := assetMap[coinDenom]; ok {
		// In this case, we just need to replace the denom with the base denom and retain the number
		if denomUnitMap.Exponent != 0 {
			coinDec, err := osmomath.NewDecFromStr(coinValue)
			if err != nil {
				return "", err
			}
			transformedCoinValue := coinDec.Mul(osmomath.MustNewDecFromStr("10").Power(denomUnitMap.Exponent))
			transformedCoinValueInt := transformedCoinValue.TruncateInt()
			transformedCoinValueStr := transformedCoinValueInt.String()
			return transformedCoinValueStr + assetMap[coinDenom].Base, nil
		} else {
			return coinValue + assetMap[coinDenom].Base, nil
		}
	}
	return "", fmt.Errorf("denom %s not found in asset map", coinDenom)
}

// OverwriteWithCustomConfig searches the respective config file for the given section and key and overwrites the current value with the given value.
func OverwriteWithCustomConfig(configFilePath string, sectionKeyValues []SectionKeyValue) error {
	// Open the file for reading and writing
	file, err := os.OpenFile(configFilePath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a map from the sectionKeyValues array
	// This map will be used to quickly look up the new values for each section and key
	configMap := make(map[string]map[string]string)
	for _, skv := range sectionKeyValues {
		// If the section does not exist in the map, create it
		if _, ok := configMap[skv.Section]; !ok {
			configMap[skv.Section] = make(map[string]string)
		}
		// Add the key and value to the section in the map
		// If the value is a string, add quotes around it
		switch v := skv.Value.(type) {
		case string:
			configMap[skv.Section][skv.Key] = "\"" + v + "\""
		default:
			configMap[skv.Section][skv.Key] = fmt.Sprintf("%v", v)
		}
	}

	// Read the file line by line
	var lines []string
	scanner := bufio.NewScanner(file)
	currentSection := ""
	for scanner.Scan() {
		line := scanner.Text()
		// If the line is a section header, update the current section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
		} else if configMap[currentSection] != nil {
			// If the line is in a section that needs to be overwritten, check each key
			for key, value := range configMap[currentSection] {
				// Split the line into key and value parts
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				// Trim spaces and compare the key part with the target key
				if strings.TrimSpace(parts[0]) == key {
					// If the keys match, overwrite the line with the new key-value pair
					line = key + " = " + value
					break
				}
			}
		}
		// Add the line to the lines slice, whether it was overwritten or not
		lines = append(lines, line)
	}

	// Check for errors from the scanner
	if err := scanner.Err(); err != nil {
		return err
	}

	// Seek to the beginning of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Truncate the file to remove the old content
	err = file.Truncate(0)
	if err != nil {
		return err
	}

	// Write the new lines to the file
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func autoCliOpts(initClientCtx client.Context, tempApp *symphony.SymphonyApp) autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range tempApp.ModuleManager().Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(tempApp.ModuleManager().Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
		ClientCtx:             initClientCtx,
	}
}
