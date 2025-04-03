package osmosisibctesting

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	smartaccounttypes "github.com/osmosis-labs/osmosis/v27/x/smart-account/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"github.com/osmosis-labs/osmosis/v27/app"
)

const SimAppChainID = "simulation-app"

type TestChain struct {
	*ibctesting.TestChain
}

// NOTE: we create a global variable here to deal with testing directories,
// this is necessary as there is now a lock file in the latest wasm that will panics our tests,
// this is a workaround, we should do something smarter that sets the home dir and removes the home dir
var TestingDirectories []string

func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	// TODO: find a better way to do this, the likely hood that this will collide is small but not 0
	dirName := fmt.Sprintf("%d", rand.Int())
	osmosisApp := app.SetupWithCustomHome(false, dirName)
	TestingDirectories = append(TestingDirectories, dirName)
	return osmosisApp, app.NewDefaultGenesisState()
}

// Copied from ibctesting because it's private
func (chain *TestChain) commitBlock(res *abci.ResponseFinalizeBlock) {
	_, err := chain.App.Commit()
	require.NoError(chain.TB, err)

	// set the last header to the current header
	// use nil trusted fields
	chain.LastHeader = chain.CurrentTMClientHeader()

	// val set changes returned from previous block get applied to the next validators
	// of this block. See tendermint spec for details.
	chain.Vals = chain.NextVals
	chain.NextVals = ibctesting.ApplyValSetChanges(chain, chain.Vals, res.ValidatorUpdates)

	// increment the current header
	chain.CurrentHeader = cmtproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.NextVals.Hash(),
		ProposerAddress:    chain.CurrentHeader.ProposerAddress,
	}
}

// SendMsgsNoCheck is an alternative to ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func (chain *TestChain) SendMsgsNoCheck(msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain.TestChain)

	// increment acc sequence regardless of success or failure tx execution
	defer func() {
		err := chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
		if err != nil {
			panic(err)
		}
	}()

	resp, err := SignAndDeliver(chain.TB, chain.TxConfig, chain.App.GetBaseApp(), msgs, chain.ChainID, []uint64{chain.SenderAccount.GetAccountNumber()}, []uint64{chain.SenderAccount.GetSequence()}, chain.CurrentHeader.GetTime(), chain.NextVals.Hash(), chain.SenderPrivKey)
	if err != nil {
		return nil, err
	}

	chain.commitBlock(resp)

	chain.Coordinator.IncrementTime()

	require.Len(chain.TB, resp.TxResults, 1)
	txResult := resp.TxResults[0]

	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}

	chain.Coordinator.IncrementTime()

	return txResult, nil
}

// SendMsgsNoCheck is an alternative to ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func (chain *TestChain) SendMsgsFromPrivKeys(privKeys []cryptotypes.PrivKey, msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain.TestChain)

	// extract account numbers and sequences from messages
	accountNumbers := make([]uint64, len(msgs))
	accountSequences := make([]uint64, len(msgs))
	seenSequence := make(map[string]uint64)
	for i, msg := range msgs {
		signers, _, err := chain.Codec.GetMsgV1Signers(msg)
		if err != nil {
			return nil, err
		}
		signer := signers[0]
		signerAcc := sdk.AccAddress(signer)
		account := chain.GetOsmosisApp().AccountKeeper.GetAccount(chain.GetContext(), signerAcc)
		accountNumbers[i] = account.GetAccountNumber()
		if sequence, ok := seenSequence[signerAcc.String()]; ok {
			accountSequences[i] = sequence + 1
		} else {
			accountSequences[i] = account.GetSequence()
		}
		seenSequence[signerAcc.String()] = accountSequences[i]
	}

	resp, err := SignAndDeliver(chain.TB, chain.TxConfig, chain.App.GetBaseApp(), msgs, chain.ChainID, accountNumbers, accountSequences, chain.CurrentHeader.GetTime(), chain.NextVals.Hash(), privKeys...)
	if err != nil {
		return nil, err
	}

	chain.commitBlock(resp)

	// increment sequences for successful transaction execution
	for _, msg := range msgs {
		signers, _, err := chain.Codec.GetMsgV1Signers(msg)
		if err != nil {
			return nil, err
		}
		signer := signers[0]
		account := chain.GetOsmosisApp().AccountKeeper.GetAccount(chain.GetContext(), signer)
		err = account.SetSequence(account.GetSequence() + 1)
		if err != nil {
			return nil, err
		}
	}

	require.Len(chain.TB, resp.TxResults, 1)
	txResult := resp.TxResults[0]

	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	return txResult, nil
}

// SignAndDeliver signs and delivers a transaction without asserting the results. This overrides the function
// from ibctesting
func SignAndDeliver(
	tb testing.TB,
	txCfg client.TxConfig,
	app *baseapp.BaseApp,
	msgs []sdk.Msg,
	chainID string,
	accNums, accSeqs []uint64,
	blockTime time.Time,
	nextValHash []byte,
	priv ...cryptotypes.PrivKey,
) (*abci.ResponseFinalizeBlock, error) {
	tb.Helper()
	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)

	if err != nil {
		return nil, err
	}

	txBytes, err := txCfg.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	return app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Time:               blockTime,
		NextValidatorsHash: nextValHash,
		Txs:                [][]byte{txBytes},
	})
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
func (chain *TestChain) GetOsmosisApp() *app.SymphonyApp {
	v, _ := chain.App.(*app.SymphonyApp)
	return v
}

// SendMsgsNoCheck is an alternative to ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func (chain *TestChain) SendMsgsFromPrivKeysWithAuthenticator(
	signers, signatures []cryptotypes.PrivKey,
	selectedAuthenticators []uint64,
	msgs ...sdk.Msg,
) (*abci.ExecTxResult, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain.TestChain)

	// extract account numbers and sequences from messages
	accountNumbers := make([]uint64, len(msgs))
	accountSequences := make([]uint64, len(msgs))
	seenSequence := make(map[string]uint64)
	for i, msg := range msgs {
		signersFromMsg, _, err := chain.Codec.GetMsgV1Signers(msg)
		if err != nil {
			return nil, err
		}
		signer := signersFromMsg[0]
		signerAcc := sdk.AccAddress(signer)
		account := chain.GetOsmosisApp().AccountKeeper.GetAccount(chain.GetContext(), signerAcc)
		accountNumbers[i] = account.GetAccountNumber()
		if sequence, ok := seenSequence[signerAcc.String()]; ok {
			accountSequences[i] = sequence + 1
		} else {
			accountSequences[i] = account.GetSequence()
		}
		seenSequence[signerAcc.String()] = accountSequences[i]
	}

	resp, err := SignAndDeliverWithAuthenticator(
		chain.GetContext(),
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		accountNumbers,
		accountSequences,
		chain.GetContext().BlockTime(),
		chain.App.LastCommitID().Hash,
		signers,
		signatures,
		selectedAuthenticators,
	)
	if err != nil {
		return nil, err
	}

	chain.commitBlock(resp)

	require.Len(chain.TB, resp.TxResults, 1)
	txResult := resp.TxResults[0]

	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}

	// increment sequences for successful transaction execution
	for _, msg := range msgs {
		signers, _, err := chain.Codec.GetMsgV1Signers(msg)
		if err != nil {
			return nil, err
		}
		signer := signers[0]
		account := chain.GetOsmosisApp().AccountKeeper.GetAccount(chain.GetContext(), signer)
		err = account.SetSequence(account.GetSequence() + 1)
		if err != nil {
			return nil, err
		}
	}

	chain.Coordinator.IncrementTime()

	return txResult, nil
}

// SignAndDeliver signs and delivers a transaction without asserting the results. This overrides the function
// from ibctesting
func SignAndDeliverWithAuthenticator(
	ctx sdk.Context,
	txCfg client.TxConfig,
	app *baseapp.BaseApp,
	header cmtproto.Header,
	msgs []sdk.Msg,
	chainID string,
	accNums,
	accSeqs []uint64,
	blockTime time.Time,
	nextValHash []byte,
	signers, signatures []cryptotypes.PrivKey,
	selectedAuthenticators []uint64,
) (*abci.ResponseFinalizeBlock, error) {
	tx, err := SignAuthenticatorMsg(
		ctx,
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 25000)},
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		signers,
		signatures,
		selectedAuthenticators,
	)
	if err != nil {
		return nil, err
	}

	txBytes, err := txCfg.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	return app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Time:               blockTime,
		NextValidatorsHash: nextValHash,
		Txs:                [][]byte{txBytes},
	})
}

// GenTx generates a signed mock transaction.
func SignAuthenticatorMsg(
	ctx sdk.Context,
	gen client.TxConfig,
	msgs []sdk.Msg,
	feeAmt sdk.Coins,
	gas uint64,
	chainID string,
	accNums, accSeqs []uint64,
	signers, signatures []cryptotypes.PrivKey,
	selectedAuthenticators []uint64,
) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(signers))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))
	signMode, err := authsigning.APISignModeToInternal(gen.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range signers {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	baseTxBuilder := gen.NewTxBuilder()

	txBuilder, ok := baseTxBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, fmt.Errorf("expected authtx.ExtensionOptionsTxBuilder, got %T", baseTxBuilder)
	}
	if len(selectedAuthenticators) > 0 {
		value, err := types.NewAnyWithValue(&smartaccounttypes.TxExtension{
			SelectedAuthenticators: selectedAuthenticators,
		})
		if err != nil {
			return nil, err
		}
		txBuilder.SetNonCriticalExtensionOptions(value)
	}

	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(feeAmt)
	txBuilder.SetGasLimit(gas)
	// TODO: set fee payer

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range signatures {
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := authsigning.GetSignBytesAdapter(
			ctx, gen.SignModeHandler(), signMode, signerData, txBuilder.GetTx())
		if err != nil {
			return nil, err
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			return nil, err
		}
		singleSigData, ok := sigs[i].Data.(*signing.SingleSignatureData)
		if !ok {
			return nil, fmt.Errorf("Error casting to SingleSignatureData")
		}
		singleSigData.Signature = sig

		err = txBuilder.SetSignatures(sigs...)
		if err != nil {
			return nil, err
		}
	}

	return txBuilder.GetTx(), nil
}

func (chain *TestChain) SendMsgsFromPrivKeysWithAuthenticatorAndCompoundSigs(
	signers []cryptotypes.PrivKey,
	signatures [][]cryptotypes.PrivKey, // Adjusted for compound signatures
	selectedAuthenticators []uint64,
	msgs ...sdk.Msg,
) (*abci.ExecTxResult, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain.TestChain)

	// extract account numbers and sequences from messages
	accountNumbers := make([]uint64, len(msgs))
	accountSequences := make([]uint64, len(msgs))
	seenSequence := make(map[string]uint64)
	for i, msg := range msgs {
		signers, _, err := chain.Codec.GetMsgV1Signers(msg)
		if err != nil {
			return nil, err
		}
		signer := signers[0]
		signerAcc := sdk.AccAddress(signer)
		account := chain.GetOsmosisApp().AccountKeeper.GetAccount(chain.GetContext(), signerAcc)
		accountNumbers[i] = account.GetAccountNumber()
		if sequence, ok := seenSequence[signerAcc.String()]; ok {
			accountSequences[i] = sequence + 1
		} else {
			accountSequences[i] = account.GetSequence()
		}
		seenSequence[signerAcc.String()] = accountSequences[i]
	}

	resp, err := SignAndDeliverWithAuthenticatorAndCompoundSigs(
		chain.GetContext(),
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		accountNumbers,
		accountSequences,
		chain.GetContext().BlockTime(),
		chain.App.LastCommitID().Hash,
		signers,
		signatures,
		selectedAuthenticators,
	)
	if err != nil {
		return nil, err
	}

	chain.commitBlock(resp)

	require.Len(chain.TB, resp.TxResults, 1)
	txResult := resp.TxResults[0]

	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}

	// increment sequences for successful transaction execution
	for _, msg := range msgs {
		signers, _, err := chain.Codec.GetMsgV1Signers(msg)
		if err != nil {
			return nil, err
		}
		signer := signers[0]
		account := chain.GetOsmosisApp().AccountKeeper.GetAccount(chain.GetContext(), signer)
		err = account.SetSequence(account.GetSequence() + 1)
		if err != nil {
			return nil, err
		}
	}

	chain.Coordinator.IncrementTime()

	return txResult, nil
}

func SignAndDeliverWithAuthenticatorAndCompoundSigs(
	ctx sdk.Context,
	txCfg client.TxConfig,
	app *baseapp.BaseApp,
	header cmtproto.Header,
	msgs []sdk.Msg,
	chainID string,
	accNums, accSeqs []uint64,
	blockTime time.Time,
	nextValHash []byte,
	signers []cryptotypes.PrivKey,
	signatures [][]cryptotypes.PrivKey, // Adjusted for compound signatures
	selectedAuthenticators []uint64,
) (*abci.ResponseFinalizeBlock, error) {
	// Now passing `signers` to the function
	tx, err := SignAuthenticatorMsgWithCompoundSigs(
		ctx,
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 25000)},
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		signers, // Correctly passing `signers` here
		signatures,
		selectedAuthenticators,
	)
	if err != nil {
		return nil, err
	}

	txBytes, err := txCfg.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	return app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Time:               blockTime,
		NextValidatorsHash: nextValHash,
		Txs:                [][]byte{txBytes},
	})
}

// SignAuthenticatorMsgWithCompoundSigs generates a transaction signed with compound signatures.
func SignAuthenticatorMsgWithCompoundSigs(
	ctx sdk.Context,
	gen client.TxConfig,
	msgs []sdk.Msg,
	feeAmt sdk.Coins,
	gas uint64,
	chainID string,
	accNums, accSeqs []uint64,
	signers []cryptotypes.PrivKey, // Reintroduced signers parameter
	signatures [][]cryptotypes.PrivKey, // Each inner slice are the privkeys for a message's signatures
	selectedAuthenticators []uint64,
) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(signers))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))
	signMode, err := authsigning.APISignModeToInternal(gen.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range signers {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	baseTxBuilder := gen.NewTxBuilder()

	txBuilder, ok := baseTxBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return nil, fmt.Errorf("expected authtx.ExtensionOptionsTxBuilder, got %T", baseTxBuilder)
	}
	if len(selectedAuthenticators) > 0 {
		value, err := types.NewAnyWithValue(&smartaccounttypes.TxExtension{
			SelectedAuthenticators: selectedAuthenticators,
		})
		if err != nil {
			return nil, err
		}
		txBuilder.SetNonCriticalExtensionOptions(value)
	}

	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(feeAmt)
	txBuilder.SetGasLimit(gas)

	// Differences start here
	for i, msgSigSet := range signatures {
		var compoundSignatures [][]byte
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
		}
		signBytes, err := authsigning.GetSignBytesAdapter(ctx, gen.SignModeHandler(), signMode, signerData, txBuilder.GetTx())
		if err != nil {
			return nil, err
		}

		for _, privKey := range msgSigSet {
			sig, err := privKey.Sign(signBytes)
			if err != nil {
				return nil, err
			}

			compoundSignatures = append(compoundSignatures, sig)
		}

		// Marshalling the array of SignatureV2 for compound signatures
		compoundSigData, err := json.Marshal(compoundSignatures)
		if err != nil {
			return nil, err
		}

		// Using the first signer's pubkey as a placeholder for the compound signature
		sigs[i] = signing.SignatureV2{
			PubKey: signers[i].PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: compoundSigData,
			},
			Sequence: accSeqs[i],
		}
	}

	// Finalize the transaction
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}

	return txBuilder.GetTx(), nil
}
