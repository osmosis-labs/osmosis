package apptesting

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v7/app"
	"github.com/stretchr/testify/suite"
)

// TODO: Consider an embedded struct here rather than an interface
type SuiteI interface {
	GetSuite() *suite.Suite
	GetCtx() sdk.Context
	SetCtx(sdk.Context)
	GetApp() *app.OsmosisApp
}

func SetupValidator(suite SuiteI, bondStatus stakingtypes.BondStatus) sdk.ValAddress {
	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address())
	bondDenom := suite.GetApp().StakingKeeper.GetParams(suite.GetCtx()).BondDenom
	selfBond := sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(100), Denom: bondDenom})

	err := simapp.FundAccount(suite.GetApp().BankKeeper, suite.GetCtx(), sdk.AccAddress(valAddr), selfBond)
	suite.GetSuite().Require().NoError(err)
	sh := teststaking.NewHelper(suite.GetSuite().T(), suite.GetCtx(), *suite.GetApp().StakingKeeper)
	msg := sh.CreateValidatorMsg(valAddr, valPub, selfBond[0].Amount)
	sh.Handle(msg, true)
	val, found := suite.GetApp().StakingKeeper.GetValidator(suite.GetCtx(), valAddr)
	suite.GetSuite().Require().True(found)
	val = val.UpdateStatus(bondStatus)
	suite.GetApp().StakingKeeper.SetValidator(suite.GetCtx(), val)

	consAddr, err := val.GetConsAddr()
	suite.GetSuite().Require().NoError(err)
	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		suite.GetCtx().BlockHeight(),
		0,
		time.Unix(0, 0),
		false,
		0,
	)
	suite.GetApp().SlashingKeeper.SetValidatorSigningInfo(suite.GetCtx(), consAddr, signingInfo)

	return valAddr
}

func BeginNewBlock(suite SuiteI, executeNextEpoch bool) {
	var valAddr []byte
	validators := suite.GetApp().StakingKeeper.GetAllValidators(suite.GetCtx())
	if len(validators) >= 1 {
		valAddrFancy, err := validators[0].GetConsAddr()
		suite.GetSuite().Require().NoError(err)
		valAddr = valAddrFancy.Bytes()
	} else {
		valAddrFancy := SetupValidator(suite, stakingtypes.Bonded)
		validator, _ := suite.GetApp().StakingKeeper.GetValidator(suite.GetCtx(), valAddrFancy)
		valAddr2, _ := validator.GetConsAddr()
		valAddr = valAddr2.Bytes()
	}

	epochIdentifier := suite.GetApp().SuperfluidKeeper.GetEpochIdentifier(suite.GetCtx())
	epoch := suite.GetApp().EpochsKeeper.GetEpochInfo(suite.GetCtx(), epochIdentifier)
	newBlockTime := suite.GetCtx().BlockTime().Add(5 * time.Second)
	if executeNextEpoch {
		endEpochTime := epoch.CurrentEpochStartTime.Add(epoch.Duration)
		newBlockTime = endEpochTime.Add(time.Second)
	}
	// fmt.Println(executeNextEpoch, suite.ctx.BlockTime(), newBlockTime)
	header := tmproto.Header{Height: suite.GetCtx().BlockHeight() + 1, Time: newBlockTime}
	newCtx := suite.GetCtx().WithBlockTime(newBlockTime).WithBlockHeight(suite.GetCtx().BlockHeight() + 1)
	suite.SetCtx(newCtx)
	lastCommitInfo := abci.LastCommitInfo{
		Votes: []abci.VoteInfo{{
			Validator:       abci.Validator{Address: valAddr, Power: 1000},
			SignedLastBlock: true},
		},
	}
	reqBeginBlock := abci.RequestBeginBlock{Header: header, LastCommitInfo: lastCommitInfo}

	fmt.Println("beginning block ", suite.GetCtx().BlockHeight())
	suite.GetApp().BeginBlocker(suite.GetCtx(), reqBeginBlock)
}

func EndBlock(suite SuiteI) {
	reqEndBlock := abci.RequestEndBlock{Height: suite.GetCtx().BlockHeight()}
	suite.GetApp().EndBlocker(suite.GetCtx(), reqEndBlock)
}
