package osmosisibctesting

import (
	"encoding/json"
	"math/rand"
	"time"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/osmosis-labs/osmosis/v21/app"
)

const SimAppChainID = "simulation-app"

type TestChain struct {
	*ibctesting.TestChain
}

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	osmosisApp := app.Setup(false)
	return osmosisApp, app.NewDefaultGenesisState()
}

// SendMsgsNoCheck is an alternative to ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func (chain *TestChain) SendMsgsNoCheck(msgs ...sdk.Msg) (*sdk.Result, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain.TestChain)

	_, r, err := SignAndDeliver(
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		chain.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	return r, nil
}

// SendMsgsNoCheck is an alternative to ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func (chain *TestChain) SendMsgsFromPrivKeys(privKeys []cryptotypes.PrivKey, msgs ...sdk.Msg) (*sdk.Result, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain.TestChain)

	// extract account numbers and sequences from messages
	accountNumbers := make([]uint64, len(msgs))
	accountSequences := make([]uint64, len(msgs))
	seenSequence := make(map[string]uint64)
	for i, msg := range msgs {
		signer := msg.GetSigners()[0]
		account := chain.GetOsmosisApp().AccountKeeper.GetAccount(chain.GetContext(), signer)
		accountNumbers[i] = account.GetAccountNumber()
		if sequence, ok := seenSequence[signer.String()]; ok {
			accountSequences[i] = sequence + 1
		} else {
			accountSequences[i] = account.GetSequence()
		}
		seenSequence[signer.String()] = accountSequences[i]
	}

	_, r, err := SignAndDeliver(
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		accountNumbers,
		accountSequences,
		privKeys...,
	)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequences for successful transaction execution
	for _, msg := range msgs {
		signer := msg.GetSigners()[0]
		account := chain.GetOsmosisApp().AccountKeeper.GetAccount(chain.GetContext(), signer)
		err = account.SetSequence(account.GetSequence() + 1)
		if err != nil {
			return nil, err
		}
	}

	chain.Coordinator.IncrementTime()

	return r, nil
}

// SignAndDeliver signs and delivers a transaction without asserting the results. This overrides the function
// from ibctesting
func SignAndDeliver(
	txCfg client.TxConfig, app *baseapp.BaseApp, header tmproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 25000)},
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	// Simulate a sending a transaction
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	return gInfo, res, err
}

// Move epochs to the future to avoid issues with minting
func (chain *TestChain) MoveEpochsToTheFuture() error {
	epochsKeeper := chain.GetOsmosisApp().EpochsKeeper
	ctx := chain.GetContext()
	for _, epoch := range epochsKeeper.AllEpochInfos(ctx) {
		epoch.StartTime = ctx.BlockTime().Add(time.Hour * 24 * 30)
		epochsKeeper.DeleteEpochInfo(chain.GetContext(), epoch.Identifier)
		err := epochsKeeper.AddEpochInfo(ctx, epoch)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetOsmosisApp returns the current chain's app as an OsmosisApp
func (chain *TestChain) GetOsmosisApp() *app.OsmosisApp {
	v, _ := chain.App.(*app.OsmosisApp)
	return v
}
