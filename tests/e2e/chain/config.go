package chain

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	staketypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/viper"
	tmconfig "github.com/tendermint/tendermint/config"
	tmjson "github.com/tendermint/tendermint/libs/json"

	epochtypes "github.com/osmosis-labs/osmosis/v7/x/epochs/types"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/v7/x/mint/types"
	poolitypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v7/x/txfees/types"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/util"
)

type ValidatorConfig struct {
	Pruning            string // default, nothing, everything, or custom
	PruningKeepRecent  string // keep all of the last N states (only used with custom pruning)
	PruningInterval    string // delete old states from every Nth block (only used with custom pruning)
	SnapshotInterval   uint64 // statesync snapshot every Nth block (0 to disable)
	SnapshotKeepRecent uint32 // number of recent snapshots to keep and serve (0 to keep all)
}

const (
	// common
	OsmoDenom     = "uosmo"
	StakeDenom    = "stake"
	OsmoIBCDenom  = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
	StakeIBCDenom = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B7787"
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
	StakeAmountCoinA = sdk.NewCoin(OsmoDenom, StakeAmountIntA)
	StakeAmountIntB  = sdk.NewInt(StakeAmountB)
	StakeAmountCoinB = sdk.NewCoin(OsmoDenom, StakeAmountIntB)

	InitBalanceStrA = fmt.Sprintf("%d%s,%d%s", OsmoBalanceA, OsmoDenom, StakeBalanceA, StakeDenom)
	InitBalanceStrB = fmt.Sprintf("%d%s,%d%s", OsmoBalanceB, OsmoDenom, StakeBalanceB, StakeDenom)
	OsmoToken       = sdk.NewInt64Coin(OsmoDenom, IbcSendAmount)  // 3,300uosmo
	StakeToken      = sdk.NewInt64Coin(StakeDenom, IbcSendAmount) // 3,300ustake
	tenOsmo         = sdk.Coins{sdk.NewInt64Coin(OsmoDenom, 10_000_000)}
)

func addAccount(path, moniker, amountStr string, accAddr sdk.AccAddress, forkHeight int) error {
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

	// TODO: Make the SDK make it far cleaner to add an account to GenesisState
	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	genDoc.InitialHeight = int64(forkHeight)

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

//nolint:typecheck
func updateModuleGenesis[V proto.Message](appGenState map[string]json.RawMessage, moduleName string, protoVal V, updateGenesis func(V)) error {
	if err := util.Cdc.UnmarshalJSON(appGenState[moduleName], protoVal); err != nil {
		return err
	}
	updateGenesis(protoVal)
	newGenState := protoVal

	bz, err := util.Cdc.MarshalJSON(newGenState)
	if err != nil {
		return err
	}
	appGenState[moduleName] = bz
	return nil
}

func initGenesis(c *internalChain, votingPeriod time.Duration) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(c.validators[0].configDir())
	config.Moniker = c.validators[0].getMoniker()

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, banktypes.ModuleName, &banktypes.GenesisState{}, updateBankGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, staketypes.ModuleName, &staketypes.GenesisState{}, updateStakeGenesis)
	if err != nil {
		return err
	}

	// pool incentives
	err = updateModuleGenesis(appGenState, poolitypes.ModuleName, &poolitypes.GenesisState{}, updatePoolIncentiveGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, incentivestypes.ModuleName, &incentivestypes.GenesisState{}, updateIncentivesGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, minttypes.ModuleName, &minttypes.GenesisState{}, updateMintGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, txfeestypes.ModuleName, &txfeestypes.GenesisState{}, updateTxfeesGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, gammtypes.ModuleName, &gammtypes.GenesisState{}, updateGammGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, epochtypes.ModuleName, &epochtypes.GenesisState{}, updateEpochGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, crisistypes.ModuleName, &crisistypes.GenesisState{}, updateCrisisGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, govtypes.ModuleName, &govtypes.GenesisState{}, updateGovGenesis(votingPeriod))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, genutiltypes.ModuleName, &genutiltypes.GenesisState{}, updateGenUtilGenesis(c))
	if err != nil {
		return err
	}

	bz, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc.AppState = bz

	genesisJson, err := tmjson.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}

	// write the updated genesis file to each validator
	for _, val := range c.validators {
		if err := util.WriteFile(filepath.Join(val.configDir(), "config", "genesis.json"), genesisJson); err != nil {
			return err
		}
	}
	return nil
}

func updateBankGenesis(bankGenState *banktypes.GenesisState) {
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
	if len(bankGenState.SupplyOffsets) == 0 {
		bankGenState.SupplyOffsets = []banktypes.GenesisSupplyOffset{}
	}
}

func updateStakeGenesis(stakeGenState *staketypes.GenesisState) {
	stakeGenState.Params = staketypes.Params{
		BondDenom:         OsmoDenom,
		MaxValidators:     100,
		MaxEntries:        7,
		HistoricalEntries: 10000,
		UnbondingTime:     240000000000,
		MinCommissionRate: sdk.ZeroDec(),
	}
}

func updatePoolIncentiveGenesis(pooliGenState *poolitypes.GenesisState) {
	pooliGenState.LockableDurations = []time.Duration{
		time.Second * 120,
		time.Second * 180,
		time.Second * 240,
	}
	pooliGenState.Params = poolitypes.Params{
		MintedDenom: OsmoDenom,
	}
}

func updateIncentivesGenesis(incentivesGenState *incentivestypes.GenesisState) {
	incentivesGenState.LockableDurations = []time.Duration{
		time.Second,
		time.Second * 120,
		time.Second * 180,
		time.Second * 240,
	}
	incentivesGenState.Params = incentivestypes.Params{
		DistrEpochIdentifier: "day",
	}
}

func updateMintGenesis(mintGenState *minttypes.GenesisState) {
	mintGenState.Params.MintDenom = OsmoDenom
	mintGenState.Params.EpochIdentifier = "day"
}

func updateTxfeesGenesis(txfeesGenState *txfeestypes.GenesisState) {
	txfeesGenState.Basedenom = OsmoDenom
}

func updateGammGenesis(gammGenState *gammtypes.GenesisState) {
	gammGenState.Params.PoolCreationFee = tenOsmo
}

func updateEpochGenesis(epochGenState *epochtypes.GenesisState) {
	epochGenState.Epochs = []epochtypes.EpochInfo{
		epochtypes.NewGenesisEpochInfo("week", time.Hour*24*7),
		// override day epochs which are in default integrations, to be 1min
		epochtypes.NewGenesisEpochInfo("day", time.Second*60),
	}
}

func updateCrisisGenesis(crisisGenState *crisistypes.GenesisState) {
	crisisGenState.ConstantFee.Denom = OsmoDenom
}

func updateGovGenesis(votingPeriod time.Duration) func(*govtypes.GenesisState) {
	return func(govGenState *govtypes.GenesisState) {
		govGenState.VotingParams = govtypes.VotingParams{
			VotingPeriod: votingPeriod,
		}
		govGenState.DepositParams.MinDeposit = tenOsmo
	}
}

func updateGenUtilGenesis(c *internalChain) func(*genutiltypes.GenesisState) {
	return func(genUtilGenState *genutiltypes.GenesisState) {
		// generate genesis txs
		genTxs := make([]json.RawMessage, len(c.validators))
		for i, val := range c.validators {
			stakeAmountCoin := StakeAmountCoinA
			if c.chainMeta.Id != ChainAID {
				stakeAmountCoin = StakeAmountCoinB
			}
			createValmsg, err := val.buildCreateValidatorMsg(stakeAmountCoin)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}

			signedTx, err := val.signMsg(createValmsg)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}

			txRaw, err := util.Cdc.MarshalJSON(signedTx)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}
			genTxs[i] = txRaw
		}
		genUtilGenState.GenTxs = genTxs
	}
}

func initNodes(c *internalChain, numVal, forkHeight int) error {
	if err := c.createAndInitValidators(numVal); err != nil {
		return err
	}

	// initialize a genesis file for the first validator
	val0ConfigDir := c.validators[0].configDir()
	for _, val := range c.validators {
		if c.chainMeta.Id == ChainAID {
			if err := addAccount(val0ConfigDir, "", InitBalanceStrA, val.getKeyInfo().GetAddress(), forkHeight); err != nil {
				return err
			}
		} else if c.chainMeta.Id == ChainBID {
			if err := addAccount(val0ConfigDir, "", InitBalanceStrB, val.getKeyInfo().GetAddress(), forkHeight); err != nil {
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

func initValidatorConfigs(c *internalChain, validatorConfigs []*ValidatorConfig) error {
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
			// skip adding a peer to yourself
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
		appConfig.BaseConfig.Pruning = validatorConfigs[i].Pruning
		appConfig.BaseConfig.PruningKeepRecent = validatorConfigs[i].PruningKeepRecent
		appConfig.BaseConfig.PruningInterval = validatorConfigs[i].PruningInterval
		appConfig.API.Enable = true
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", MinGasPrice, OsmoDenom)
		appConfig.StateSync.SnapshotInterval = validatorConfigs[i].SnapshotInterval
		appConfig.StateSync.SnapshotKeepRecent = validatorConfigs[i].SnapshotKeepRecent

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
	return nil
}
