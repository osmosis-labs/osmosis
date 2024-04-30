package osmosisibctesting

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	smartaccounttypes "github.com/osmosis-labs/osmosis/v25/x/smart-account/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/osmosis-labs/osmosis/v25/app"
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

// SendMsgsNoCheck is an alternative to ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func (chain *TestChain) SendMsgsFromPrivKeysWithAuthenticator(
	signers, signatures []cryptotypes.PrivKey,
	selectedAuthenticators []uint64,
	msgs ...sdk.Msg,
) (*sdk.Result, error) {
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

	_, r, err := SignAndDeliverWithAuthenticator(
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		accountNumbers,
		accountSequences,
		signers,
		signatures,
		selectedAuthenticators,
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
func SignAndDeliverWithAuthenticator(
	txCfg client.TxConfig,
	app *baseapp.BaseApp,
	header tmproto.Header,
	msgs []sdk.Msg,
	chainID string,
	accNums,
	accSeqs []uint64,
	signers, signatures []cryptotypes.PrivKey,
	selectedAuthenticators []uint64,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := SignAuthenticatorMsg(
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
		return sdk.GasInfo{}, nil, err
	}

	// Simulate a sending a transaction
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	return gInfo, res, err
}

// GenTx generates a signed mock transaction.
func SignAuthenticatorMsg(
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
	signMode := gen.SignModeHandler().DefaultMode()

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

	err := txBuilder.SetMsgs(msgs...)
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
		signBytes, err := gen.SignModeHandler().GetSignBytes(
			signMode,
			signerData,
			txBuilder.GetTx(),
		)
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
) (*sdk.Result, error) {
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

	_, r, err := SignAndDeliverWithAuthenticatorAndCompoundSigs(
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		accountNumbers,
		accountSequences,
		signers,
		signatures,
		selectedAuthenticators,
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

func SignAndDeliverWithAuthenticatorAndCompoundSigs(
	txCfg client.TxConfig,
	app *baseapp.BaseApp,
	header tmproto.Header,
	msgs []sdk.Msg,
	chainID string,
	accNums, accSeqs []uint64,
	signers []cryptotypes.PrivKey,
	signatures [][]cryptotypes.PrivKey, // Adjusted for compound signatures
	selectedAuthenticators []uint64,
) (sdk.GasInfo, *sdk.Result, error) {
	// Now passing `signers` to the function
	tx, err := SignAuthenticatorMsgWithCompoundSigs(
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
		return sdk.GasInfo{}, nil, err
	}

	// Simulate sending the transaction
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	return gInfo, res, err
}

// SignAuthenticatorMsgWithCompoundSigs generates a transaction signed with compound signatures.
func SignAuthenticatorMsgWithCompoundSigs(
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
	signMode := gen.SignModeHandler().DefaultMode()

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

	err := txBuilder.SetMsgs(msgs...)
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
		signBytes, err := gen.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
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
