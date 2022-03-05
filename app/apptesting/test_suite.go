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

type KeeperTestHelper struct {
	Suite suite.Suite

	App *app.OsmosisApp
	Ctx sdk.Context
}

func (keeperTestHelper *KeeperTestHelper) SetupValidator(bondStatus stakingtypes.BondStatus) sdk.ValAddress {
	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address())
	bondDenom := keeperTestHelper.App.StakingKeeper.GetParams(keeperTestHelper.Ctx).BondDenom
	selfBond := sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(100), Denom: bondDenom})

	err := simapp.FundAccount(keeperTestHelper.App.BankKeeper, keeperTestHelper.Ctx, sdk.AccAddress(valAddr), selfBond)
	keeperTestHelper.Suite.Require().NoError(err)

	sh := teststaking.NewHelper(keeperTestHelper.Suite.T(), keeperTestHelper.Ctx, *keeperTestHelper.App.StakingKeeper)
	msg := sh.CreateValidatorMsg(valAddr, valPub, selfBond[0].Amount)
	sh.Handle(msg, true)
	val, found := keeperTestHelper.App.StakingKeeper.GetValidator(keeperTestHelper.Ctx, valAddr)
	keeperTestHelper.Suite.Require().True(found)
	val = val.UpdateStatus(bondStatus)
	keeperTestHelper.App.StakingKeeper.SetValidator(keeperTestHelper.Ctx, val)

	consAddr, err := val.GetConsAddr()
	keeperTestHelper.Suite.Require().NoError(err)
	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		keeperTestHelper.Ctx.BlockHeight(),
		0,
		time.Unix(0, 0),
		false,
		0,
	)
	keeperTestHelper.App.SlashingKeeper.SetValidatorSigningInfo(keeperTestHelper.Ctx, consAddr, signingInfo)

	return valAddr
}

func (keeperTestHelper *KeeperTestHelper) BeginNewBlock(executeNextEpoch bool) {
	valAddr := []byte(":^) at this distribution workaround")
	validators := keeperTestHelper.App.StakingKeeper.GetAllValidators(keeperTestHelper.Ctx)
	if len(validators) >= 1 {
		valAddrFancy, err := validators[0].GetConsAddr()
		keeperTestHelper.Suite.Require().NoError(err)
		valAddr = valAddrFancy.Bytes()
	} else {
		valAddrFancy := keeperTestHelper.SetupValidator(stakingtypes.Bonded)
		validator, _ := keeperTestHelper.App.StakingKeeper.GetValidator(keeperTestHelper.Ctx, valAddrFancy)
		valAddr2, _ := validator.GetConsAddr()
		valAddr = valAddr2.Bytes()
	}

	epochIdentifier := keeperTestHelper.App.SuperfluidKeeper.GetEpochIdentifier(keeperTestHelper.Ctx)
	epoch := keeperTestHelper.App.EpochsKeeper.GetEpochInfo(keeperTestHelper.Ctx, epochIdentifier)
	newBlockTime := keeperTestHelper.Ctx.BlockTime().Add(5 * time.Second)
	if executeNextEpoch {
		endEpochTime := epoch.CurrentEpochStartTime.Add(epoch.Duration)
		newBlockTime = endEpochTime.Add(time.Second)
	}
	// fmt.Println(executeNextEpoch, keeperTestHelper.Ctx.BlockTime(), newBlockTime)
	header := tmproto.Header{Height: keeperTestHelper.Ctx.BlockHeight() + 1, Time: newBlockTime}
	newCtx := keeperTestHelper.Ctx.WithBlockTime(newBlockTime).WithBlockHeight(keeperTestHelper.Ctx.BlockHeight() + 1)
	keeperTestHelper.Ctx = newCtx
	lastCommitInfo := abci.LastCommitInfo{
		Votes: []abci.VoteInfo{{
			Validator:       abci.Validator{Address: valAddr, Power: 1000},
			SignedLastBlock: true},
		},
	}
	reqBeginBlock := abci.RequestBeginBlock{Header: header, LastCommitInfo: lastCommitInfo}

	fmt.Println("beginning block ", keeperTestHelper.Ctx.BlockHeight())
	keeperTestHelper.App.BeginBlocker(keeperTestHelper.Ctx, reqBeginBlock)
}

func (keeperTestHelper *KeeperTestHelper) EndBlock() {
	reqEndBlock := abci.RequestEndBlock{Height: keeperTestHelper.Ctx.BlockHeight()}
	keeperTestHelper.App.EndBlocker(keeperTestHelper.Ctx, reqEndBlock)
}
