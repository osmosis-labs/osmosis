package ante_test

import (
	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v27/ante"

	oracletypes "github.com/osmosis-labs/osmosis/v27/x/oracle/types"
)

func (suite *AnteTestSuite) TestOracleSpamming() {
	suite.SetupTest(true) // setup
	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()

	priv1, _, addr1 := testdata.KeyTestPubAddr()
	priv2, _, addr2 := testdata.KeyTestPubAddr()

	spd := ante.NewSpammingPreventionDecorator(dummyOracleKeeper{
		feeders: map[string]string{
			sdk.ValAddress(addr1).String(): addr1.String(),
			sdk.ValAddress(addr2).String(): addr2.String(),
		},
	})
	antehandler := sdk.ChainAnteDecorators(spd)

	// Set IsCheckTx to true
	suite.ctx = suite.ctx.WithIsCheckTx(true)

	// normal so ok
	suite.ctx = suite.ctx.WithBlockHeight(100)
	suite.Require().NoError(suite.txBuilder.SetMsgs(
		oracletypes.NewMsgAggregateExchangeRatePrevote(oracletypes.AggregateVoteHash{}, addr1, sdk.ValAddress(addr1)),
		oracletypes.NewMsgAggregateExchangeRateVote("", "", addr1, sdk.ValAddress(addr1)),
	))
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{0, 0}
	tx, err := suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NoError(err)

	// do it again is blocked
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Error(err)

	// next block; can put oracletypes again
	suite.ctx = suite.ctx.WithBlockHeight(101)
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().NoError(err)

	// catch wrong feeder
	suite.Require().NoError(suite.txBuilder.SetMsgs(
		oracletypes.NewMsgAggregateExchangeRatePrevote(oracletypes.AggregateVoteHash{}, addr2, sdk.ValAddress(addr1)),
		oracletypes.NewMsgAggregateExchangeRateVote("", "", addr1, sdk.ValAddress(addr1)),
	))
	tx, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockHeight(102)
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Error(err)

	// catch wrong feeder; again
	suite.Require().NoError(suite.txBuilder.SetMsgs(
		oracletypes.NewMsgAggregateExchangeRatePrevote(oracletypes.AggregateVoteHash{}, addr1, sdk.ValAddress(addr1)),
		oracletypes.NewMsgAggregateExchangeRateVote("", "", addr2, sdk.ValAddress(addr1)),
	))
	tx, err = suite.CreateTestTx(privs, accNums, accSeqs, suite.ctx.ChainID())
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockHeight(103)
	_, err = antehandler(suite.ctx, tx, false)
	suite.Require().Error(err)
}

type dummyOracleKeeper struct {
	feeders map[string]string
}

func (ok dummyOracleKeeper) ValidateFeeder(_ sdk.Context, feederAddr sdk.AccAddress, validatorAddr sdk.ValAddress) error {
	if val, ok := ok.feeders[validatorAddr.String()]; ok && val == feederAddr.String() {
		return nil
	}

	return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "cannot ensure feeder right")
}

func (ok dummyOracleKeeper) GetMelodyExchangeRate(_ sdk.Context, _ string) (osmomath.Dec, error) {
	return osmomath.NewDecWithPrec(1, 1), nil
}
