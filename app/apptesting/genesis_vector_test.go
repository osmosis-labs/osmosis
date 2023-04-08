package apptesting

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (s *KeeperTestHelper) CreateGenesisFile(numAccounts int, path string) {
	defaultCoins := sdk.NewCoins(sdk.NewCoin("coinA", sdk.NewInt(1000000000000000000)), sdk.NewCoin("coinB", sdk.NewInt(100)))
	for i := 0; i < numAccounts; i++ {
		addrS := fmt.Sprintf("sampleAcct%d", i)
		addr := sdk.AccAddress(fmt.Sprintf("%020s", addrS))
		s.FundAcc(addr, defaultCoins)

		i64 := int64(i)
		lockCoins := sdk.Coins{sdk.NewInt64Coin("coinA", i64+1)}
		s.LockTokens(addr, lockCoins, time.Hour*time.Duration(i+1))
	}

	mm := s.App.ModuleManager()
	genState := mm.ExportGenesisForModules(s.Ctx, s.App.AppCodec(), nil)
	appState, err := json.Marshal(genState)
	s.Require().NoError(err)

	genesis := tmtypes.GenesisDoc{
		GenesisTime:     time.Now(),
		ChainID:         "test-chain",
		InitialHeight:   1,
		ConsensusParams: tmtypes.DefaultConsensusParams(),
		Validators:      nil,
		AppState:        appState,
	}
	err = marshalJsonToFile(genesis, path)
	s.Require().NoError(err)
}

func marshalJsonToFile(v any, path string) error {
	// Open a file for writing.
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Marshal the Person object to JSON and write it to the file.
	encoder := json.NewEncoder(file)
	err = encoder.Encode(v)
	if err != nil {
		return err
	}
	return nil
}

// func TestCreateGenesis(t *testing.T) {
// 	s := new(KeeperTestHelper)
// 	suite.Run(t, s)
// 	s.Setup()
// 	s.CreateGenesisFile(20000, "bench_genesis.json")
// }
