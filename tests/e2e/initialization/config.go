package initialization

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	staketypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
	tmjson "github.com/tendermint/tendermint/libs/json"

	epochtypes "github.com/osmosis-labs/osmosis/v14/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v14/x/incentives/types"
	minttypes "github.com/osmosis-labs/osmosis/v14/x/mint/types"
	poolitypes "github.com/osmosis-labs/osmosis/v14/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
	twaptypes "github.com/osmosis-labs/osmosis/v14/x/twap/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v14/x/txfees/types"

	types1 "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/osmosis-labs/osmosis/v14/tests/e2e/util"
)

// NodeConfig is a confiuration for the node supplied from the test runner
// to initialization scripts. It should be backwards compatible with earlier
// versions. If this struct is updated, the change must be backported to earlier
// branches that might be used for upgrade testing.
type NodeConfig struct {
	Name               string // name of the config that will also be assigned to Docke container.
	Pruning            string // default, nothing, everything, or custom
	PruningKeepRecent  string // keep all of the last N states (only used with custom pruning)
	PruningInterval    string // delete old states from every Nth block (only used with custom pruning)
	SnapshotInterval   uint64 // statesync snapshot every Nth block (0 to disable)
	SnapshotKeepRecent uint32 // number of recent snapshots to keep and serve (0 to keep all)
	IsValidator        bool   // flag indicating whether a node should be a validator
}

const (
	// common
	OsmoDenom           = "uosmo"
	IonDenom            = "uion"
	StakeDenom          = "stake"
	AtomDenom           = "uatom"
	OsmoIBCDenom        = "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518"
	StakeIBCDenom       = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B7787"
	E2EFeeToken         = "e2e-default-feetoken"
	MinGasPrice         = "0.000"
	IbcSendAmount       = 3300000000
	ValidatorWalletName = "val"
	// chainA
	ChainAID      = "osmo-test-a"
	OsmoBalanceA  = 200000000000
	IonBalanceA   = 100000000000
	StakeBalanceA = 110000000000
	StakeAmountA  = 100000000000
	// chainB
	ChainBID          = "osmo-test-b"
	OsmoBalanceB      = 500000000000
	IonBalanceB       = 100000000000
	StakeBalanceB     = 440000000000
	StakeAmountB      = 400000000000
	GenesisFeeBalance = 100000000000
	WalletFeeBalance  = 100000000

	EpochDuration         = time.Second * 60
	TWAPPruningKeepPeriod = EpochDuration / 4
)

var (
	StakeAmountIntA  = sdk.NewInt(StakeAmountA)
	StakeAmountCoinA = sdk.NewCoin(OsmoDenom, StakeAmountIntA)
	StakeAmountIntB  = sdk.NewInt(StakeAmountB)
	StakeAmountCoinB = sdk.NewCoin(OsmoDenom, StakeAmountIntB)

	InitBalanceStrA = fmt.Sprintf("%d%s,%d%s,%d%s", OsmoBalanceA, OsmoDenom, StakeBalanceA, StakeDenom, IonBalanceA, IonDenom)
	InitBalanceStrB = fmt.Sprintf("%d%s,%d%s,%d%s", OsmoBalanceB, OsmoDenom, StakeBalanceB, StakeDenom, IonBalanceB, IonDenom)
	OsmoToken       = sdk.NewInt64Coin(OsmoDenom, IbcSendAmount)  // 3,300uosmo
	StakeToken      = sdk.NewInt64Coin(StakeDenom, IbcSendAmount) // 3,300ustake
	tenOsmo         = sdk.Coins{sdk.NewInt64Coin(OsmoDenom, 10_000_000)}
	fiftyOsmo       = sdk.Coins{sdk.NewInt64Coin(OsmoDenom, 50_000_000)}
	WalletFeeTokens = sdk.NewCoin(E2EFeeToken, sdk.NewInt(WalletFeeBalance))
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
	coins = coins.Add(sdk.NewCoin(E2EFeeToken, sdk.NewInt(GenesisFeeBalance)))

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

func initGenesis(chain *internalChain, votingPeriod, expeditedVotingPeriod time.Duration, forkHeight int) error {
	// initialize a genesis file
	configDir := chain.nodes[0].configDir()
	for _, val := range chain.nodes {
		if chain.chainMeta.Id == ChainAID {
			if err := addAccount(configDir, "", InitBalanceStrA, val.keyInfo.GetAddress(), forkHeight); err != nil {
				return err
			}
		} else if chain.chainMeta.Id == ChainBID {
			if err := addAccount(configDir, "", InitBalanceStrB, val.keyInfo.GetAddress(), forkHeight); err != nil {
				return err
			}
		}
	}

	// copy the genesis file to the remaining validators
	for _, val := range chain.nodes[1:] {
		_, err := util.CopyFile(
			filepath.Join(configDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		if err != nil {
			return err
		}
	}

	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(chain.nodes[0].configDir())
	config.Moniker = chain.nodes[0].moniker

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

	err = updateModuleGenesis(appGenState, poolmanagertypes.ModuleName, &poolmanagertypes.GenesisState{}, updatePoolManagerGenesis(appGenState))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, epochtypes.ModuleName, &epochtypes.GenesisState{}, updateEpochGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, twaptypes.ModuleName, &twaptypes.GenesisState{}, updateTWAPGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, crisistypes.ModuleName, &crisistypes.GenesisState{}, updateCrisisGenesis)
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, govtypes.ModuleName, &govtypes.GenesisState{}, updateGovGenesis(votingPeriod, expeditedVotingPeriod))
	if err != nil {
		return err
	}

	err = updateModuleGenesis(appGenState, genutiltypes.ModuleName, &genutiltypes.GenesisState{}, updateGenUtilGenesis(chain))
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
	for _, val := range chain.nodes {
		if err := util.WriteFile(filepath.Join(val.configDir(), "config", "genesis.json"), genesisJson); err != nil {
			return err
		}
	}
	return nil
}

func updateBankGenesis(bankGenState *banktypes.GenesisState) {
	denomsToRegister := []string{StakeDenom, IonDenom, OsmoDenom, AtomDenom}
	for _, denom := range denomsToRegister {
		setDenomMetadata(bankGenState, denom)
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
	txfeesGenState.Feetokens = []txfeestypes.FeeToken{
		{Denom: E2EFeeToken, PoolID: 1},
	}
}

func updateGammGenesis(gammGenState *gammtypes.GenesisState) {
	gammGenState.Params.PoolCreationFee = tenOsmo
	// setup fee pool, between "e2e_default_fee_token" and "uosmo"
	feePoolParams := balancer.NewPoolParams(sdk.MustNewDecFromStr("0.01"), sdk.ZeroDec(), nil)
	feePoolAssets := []balancer.PoolAsset{
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("uosmo", sdk.NewInt(100000000000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin(E2EFeeToken, sdk.NewInt(100000000000)),
		},
	}
	pool1, err := balancer.NewBalancerPool(1, feePoolParams, feePoolAssets, "", time.Unix(0, 0))
	if err != nil {
		panic(err)
	}
	anyPool, err := types1.NewAnyWithValue(&pool1)
	if err != nil {
		panic(err)
	}
	gammGenState.Pools = []*types1.Any{anyPool}
	gammGenState.NextPoolNumber = 2
}

func updatePoolManagerGenesis(appGenState map[string]json.RawMessage) func(*poolmanagertypes.GenesisState) {
	return func(s *poolmanagertypes.GenesisState) {
		gammGenState := &gammtypes.GenesisState{}
		if err := util.Cdc.UnmarshalJSON(appGenState[gammtypes.ModuleName], gammGenState); err != nil {
			panic(err)
		}
		s.NextPoolId = gammGenState.NextPoolNumber
	}
}

func updateEpochGenesis(epochGenState *epochtypes.GenesisState) {
	epochGenState.Epochs = []epochtypes.EpochInfo{
		epochtypes.NewGenesisEpochInfo("week", time.Hour*24*7),
		// override day epochs which are in default integrations, to be 1min
		epochtypes.NewGenesisEpochInfo("day", time.Second*60),
	}
}

func updateTWAPGenesis(twapGenState *twaptypes.GenesisState) {
	// Lower keep period from defaults to allos us to test pruning.
	twapGenState.Params.RecordHistoryKeepPeriod = time.Second * 15
}

func updateCrisisGenesis(crisisGenState *crisistypes.GenesisState) {
	crisisGenState.ConstantFee.Denom = OsmoDenom
}

func updateGovGenesis(votingPeriod, expeditedVotingPeriod time.Duration) func(*govtypes.GenesisState) {
	return func(govGenState *govtypes.GenesisState) {
		govGenState.VotingParams.VotingPeriod = votingPeriod
		govGenState.VotingParams.ExpeditedVotingPeriod = expeditedVotingPeriod
		govGenState.DepositParams.MinDeposit = tenOsmo
		govGenState.DepositParams.MinExpeditedDeposit = fiftyOsmo
	}
}

func updateGenUtilGenesis(c *internalChain) func(*genutiltypes.GenesisState) {
	return func(genUtilGenState *genutiltypes.GenesisState) {
		// generate genesis txs
		genTxs := make([]json.RawMessage, 0, len(c.nodes))
		for _, node := range c.nodes {
			if !node.isValidator {
				continue
			}

			stakeAmountCoin := StakeAmountCoinA
			if c.chainMeta.Id != ChainAID {
				stakeAmountCoin = StakeAmountCoinB
			}
			createValmsg, err := node.buildCreateValidatorMsg(stakeAmountCoin)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}

			signedTx, err := node.signMsg(createValmsg)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}

			txRaw, err := util.Cdc.MarshalJSON(signedTx)
			if err != nil {
				panic("genutil genesis setup failed: " + err.Error())
			}
			genTxs = append(genTxs, txRaw)
		}
		genUtilGenState.GenTxs = genTxs
	}
}

func setDenomMetadata(genState *banktypes.GenesisState, denom string) {
	genState.DenomMetadata = append(genState.DenomMetadata, banktypes.Metadata{
		Description: fmt.Sprintf("Registered denom %s for e2e testing", denom),
		Display:     denom,
		Base:        denom,
		Symbol:      denom,
		Name:        denom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: 0,
			},
		},
	})
}
