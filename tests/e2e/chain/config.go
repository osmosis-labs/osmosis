package chain

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/server"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/osmosis-labs/osmosis/v8/tests/e2e/util"
)

// NodeConfig is a confiuration for the node supplied from the test runner
// to initialization scripts. It should be backwards compatible with earlier
// versions. If this struct is updated, the change must be backported to earlier
// branches that might be used for upgrade testing.
type NodeConfig struct {
	Name               string
	Pruning            string // default, nothing, everything, or custom
	PruningKeepRecent  string // keep all of the last N states (only used with custom pruning)
	PruningInterval    string // delete old states from every Nth block (only used with custom pruning)
	SnapshotInterval   uint64 // statesync snapshot every Nth block (0 to disable)
	SnapshotKeepRecent uint32 // number of recent snapshots to keep and serve (0 to keep all)
	IsValidator        bool   // flag indicating whether a node should be a validator
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
	StakeAmountCoinA = sdk.NewCoin(StakeDenom, StakeAmountIntA)
	StakeAmountIntB  = sdk.NewInt(StakeAmountB)
	StakeAmountCoinB = sdk.NewCoin(StakeDenom, StakeAmountIntB)

	InitBalanceStrA = fmt.Sprintf("%d%s,%d%s", OsmoBalanceA, OsmoDenom, StakeBalanceA, StakeDenom)
	InitBalanceStrB = fmt.Sprintf("%d%s,%d%s", OsmoBalanceB, OsmoDenom, StakeBalanceB, StakeDenom)
	OsmoToken       = sdk.NewInt64Coin(OsmoDenom, IbcSendAmount)  // 3,300uosmo
	StakeToken      = sdk.NewInt64Coin(StakeDenom, IbcSendAmount) // 3,300ustake
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

func initGenesis(chain *internalChain, votingPeriod time.Duration) error {
	// initialize a genesis file
	configDir := chain.nodes[0].configDir()
	for _, val := range chain.nodes {
		if chain.chainMeta.Id == ChainAID {
			if err := addAccount(configDir, "", InitBalanceStrA, val.getKeyInfo().GetAddress()); err != nil {
				return err
			}
		} else if chain.chainMeta.Id == ChainBID {
			if err := addAccount(configDir, "", InitBalanceStrB, val.getKeyInfo().GetAddress()); err != nil {
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
	config.Moniker = chain.nodes[0].getMoniker()

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

	var govGenState govtypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState); err != nil {
		return err
	}

	govGenState.VotingParams = govtypes.VotingParams{
		VotingPeriod: votingPeriod,
	}

	gz, err := util.Cdc.MarshalJSON(&govGenState)
	if err != nil {
		return err
	}
	appGenState[govtypes.ModuleName] = gz

	var genUtilGenState genutiltypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState); err != nil {
		return err
	}

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(chain.nodes))
	for i, val := range chain.nodes {
		if !val.isValidator {
			continue
		}

		stakeAmountCoin := StakeAmountCoinA
		if chain.chainMeta.Id != ChainAID {
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
	for _, val := range chain.nodes {
		if err := util.WriteFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz); err != nil {
			return err
		}
	}
	return nil
}

func newNode(chain *internalChain, nodeConfig *NodeConfig) (*internalNode, error) {
	node := &internalNode{
		chain:       chain,
		moniker:     fmt.Sprintf("%s-node-%s", chain.chainMeta.Id, nodeConfig.Name),
		isValidator: nodeConfig.IsValidator,
	}

	// generate genesis files
	if err := node.init(); err != nil {
		return nil, err
	}
	// create keys
	if err := node.createKey("node"); err != nil {
		return nil, err
	}
	if err := node.createNodeKey(); err != nil {
		return nil, err
	}
	if err := node.createConsensusKey(); err != nil {
		return nil, err
	}

	node.createAppConfig(nodeConfig)

	return node, nil
}
