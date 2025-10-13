package app_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v31/app"
	"github.com/osmosis-labs/osmosis/v31/app/apptesting"

	cometabci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type MempoolCapacityTestSuite struct {
	apptesting.KeeperTestHelper

	encoding     client.TxConfig
	privKeys     []*secp256k1.PrivKey
	pubKeys      []cryptotypes.PubKey
	accAddresses []sdk.AccAddress
}

func TestMempoolCapacityTestSuite(t *testing.T) {
	suite.Run(t, new(MempoolCapacityTestSuite))
}

func (suite *MempoolCapacityTestSuite) SetupTest() {
	suite.Setup()

	// Setup encoding config
	suite.encoding = app.MakeEncodingConfig().TxConfig

	numAccounts := 3000
	suite.privKeys = make([]*secp256k1.PrivKey, numAccounts)
	suite.pubKeys = make([]cryptotypes.PubKey, numAccounts)
	suite.accAddresses = make([]sdk.AccAddress, numAccounts)

	for i := 0; i < numAccounts; i++ {
		suite.privKeys[i] = secp256k1.GenPrivKey()
		suite.pubKeys[i] = suite.privKeys[i].PubKey()
		suite.accAddresses[i] = sdk.AccAddress(suite.privKeys[i].PubKey().Address())

		acc := suite.App.AccountKeeper.NewAccountWithAddress(suite.Ctx, suite.accAddresses[i])
		suite.App.AccountKeeper.SetAccount(suite.Ctx, acc)

		// fund each account
		err := testutil.FundAccount(suite.Ctx, suite.App.BankKeeper, suite.accAddresses[i], sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000000000000000)))
		suite.Require().NoError(err)
	}

	suite.Commit()
}

// Ensure that the default lane is able to remove transaction after block has been finalized
// and not reach max capacity for the wrong reason
func (suite *MempoolCapacityTestSuite) TestDefaultLaneStaledTransactions() {
	// max capacity is 3000, we need to batch it due to gas limit
	batchSize := 101

	for n := 0; n <= 30; n++ {
		txs := make([]sdk.Tx, batchSize)
		for i := 0; i < batchSize; i++ {
			txs[i] = suite.createTestTx(i + n*batchSize)
		}

		txBytes := make([][]byte, batchSize)
		for _, tx := range txs {
			// force insert into mempool
			_, _, err := suite.App.SimCheck(suite.encoding.TxEncoder(), tx)
			require.NoError(suite.T(), err)

			encodedTx, err := suite.encoding.TxEncoder()(tx)
			require.NoError(suite.T(), err)

			txBytes = append(txBytes, encodedTx)
		}

		_, err := suite.App.FinalizeBlock(&cometabci.RequestFinalizeBlock{
			Height: suite.Ctx.BlockHeight(),
			Time:   suite.Ctx.BlockTime(),
			Txs:    txBytes,
		})
		require.NoError(suite.T(), err)
	}
}

func (suite *MempoolCapacityTestSuite) createTestTx(txIndex int) sdk.Tx {
	// Use modulo to cycle through accounts if we have more transactions than accounts
	accountIndex := txIndex % len(suite.accAddresses)

	// Create a simple bank transfer message
	msg := banktypes.NewMsgSend(
		suite.accAddresses[accountIndex],
		suite.accAddresses[accountIndex],
		sdk.NewCoins(sdk.NewInt64Coin("stake", 1)),
	)

	// Get the actual account number and sequence from the stored account
	account := suite.App.AccountKeeper.GetAccount(suite.Ctx, suite.accAddresses[accountIndex])

	// Use the GenTx function for proper transaction generation
	tx, err := GenTx(
		suite.Ctx,
		suite.encoding,
		[]sdk.Msg{msg},
		sdk.NewCoins(sdk.NewInt64Coin("stake", 20000)), // fee
		200000, // gas
		suite.Ctx.ChainID(),
		[]uint64{account.GetAccountNumber()}, // accNums
		[]uint64{account.GetSequence()},      // accSeqs
		[]cryptotypes.PrivKey{suite.privKeys[accountIndex]}, // signers
		[]cryptotypes.PrivKey{suite.privKeys[accountIndex]}, // signatures
		[]uint64{}, // selectedAuthenticators (empty for regular tx)
	)
	suite.Require().NoError(err)

	return tx
}

// GenTx generates a signed mock transaction.
func GenTx(
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

	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}

	txBuilder.SetFeeAmount(feeAmt)
	txBuilder.SetGasLimit(gas)

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
			panic(err)
		}

		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = txBuilder.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return txBuilder.GetTx(), nil
}
