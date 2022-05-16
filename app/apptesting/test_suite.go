package apptesting

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v8/app"
	"github.com/osmosis-labs/osmosis/v8/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v8/x/gamm/types"
	lockupkeeper "github.com/osmosis-labs/osmosis/v8/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type KeeperTestHelper struct {
	suite.Suite

	App *app.OsmosisApp
	Ctx sdk.Context
}

func (keeperTestHelper *KeeperTestHelper) SetupValidator(bondStatus stakingtypes.BondStatus) sdk.ValAddress {
	valPub := secp256k1.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(valPub.Address())
	bondDenom := keeperTestHelper.App.StakingKeeper.GetParams(keeperTestHelper.Ctx).BondDenom
	selfBond := sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(100), Denom: bondDenom})

	err := simapp.FundAccount(keeperTestHelper.App.BankKeeper, keeperTestHelper.Ctx, sdk.AccAddress(valAddr), selfBond)
	keeperTestHelper.Require().NoError(err)

	sh := teststaking.NewHelper(keeperTestHelper.Suite.T(), keeperTestHelper.Ctx, *keeperTestHelper.App.StakingKeeper)
	msg := sh.CreateValidatorMsg(valAddr, valPub, selfBond[0].Amount)
	sh.Handle(msg, true)
	val, found := keeperTestHelper.App.StakingKeeper.GetValidator(keeperTestHelper.Ctx, valAddr)
	keeperTestHelper.Require().True(found)
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
		keeperTestHelper.Require().NoError(err)
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

func (keeperTestHelper *KeeperTestHelper) AllocateRewardsToValidator(valAddr sdk.ValAddress, rewardAmt sdk.Int) {
	validator, found := keeperTestHelper.App.StakingKeeper.GetValidator(keeperTestHelper.Ctx, valAddr)
	keeperTestHelper.Require().True(found)

	// allocate reward tokens to distribution module
	coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, rewardAmt)}
	err := simapp.FundModuleAccount(keeperTestHelper.App.BankKeeper, keeperTestHelper.Ctx, distrtypes.ModuleName, coins)
	keeperTestHelper.Require().NoError(err)

	// allocate rewards to validator
	keeperTestHelper.Ctx = keeperTestHelper.Ctx.WithBlockHeight(keeperTestHelper.Ctx.BlockHeight() + 1)
	decTokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(20000)}}
	keeperTestHelper.App.DistrKeeper.AllocateTokensToValidator(keeperTestHelper.Ctx, validator, decTokens)
}

// SetupGammPoolsWithBondDenomMultiplier uses given multipliers to set initial pool supply of bond denom.
func (keeperTestHelper *KeeperTestHelper) SetupGammPoolsWithBondDenomMultiplier(multipliers []sdk.Dec) []gammtypes.PoolI {
	keeperTestHelper.App.GAMMKeeper.SetParams(keeperTestHelper.Ctx, gammtypes.Params{
		PoolCreationFee: sdk.Coins{},
	})

	bondDenom := keeperTestHelper.App.StakingKeeper.BondDenom(keeperTestHelper.Ctx)
	//TODO: use sdk crypto instead of tendermint to generate address
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

	//fund account with pool creation fee
	poolCreationFee := keeperTestHelper.App.GAMMKeeper.GetParams(keeperTestHelper.Ctx)
	err := simapp.FundAccount(keeperTestHelper.App.BankKeeper, keeperTestHelper.Ctx, acc1, poolCreationFee.PoolCreationFee)
	keeperTestHelper.Require().NoError(err)

	pools := []gammtypes.PoolI{}

	for index, multiplier := range multipliers {
		token := fmt.Sprintf("token%d", index)

		uosmoAmount := gammtypes.InitPoolSharesSupply.ToDec().Mul(multiplier).RoundInt()

		err := simapp.FundAccount(keeperTestHelper.App.BankKeeper, keeperTestHelper.Ctx, acc1, sdk.NewCoins(
			sdk.NewCoin(bondDenom, uosmoAmount.Mul(sdk.NewInt(10))),
			sdk.NewInt64Coin(token, 100000),
		))
		keeperTestHelper.NoError(err)

		var (
			defaultFutureGovernor = ""

			// pool assets
			defaultFooAsset gammtypes.PoolAsset = gammtypes.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(bondDenom, uosmoAmount),
			}
			defaultBarAsset gammtypes.PoolAsset = gammtypes.PoolAsset{
				Weight: sdk.NewInt(100),
				Token:  sdk.NewCoin(token, sdk.NewInt(10000)),
			}
			poolAssets []gammtypes.PoolAsset = []gammtypes.PoolAsset{defaultFooAsset, defaultBarAsset}
		)

		poolId, err := keeperTestHelper.App.GAMMKeeper.CreateBalancerPool(keeperTestHelper.Ctx, acc1, balancer.PoolParams{
			SwapFee: sdk.NewDecWithPrec(1, 2),
			ExitFee: sdk.NewDecWithPrec(1, 2),
		}, poolAssets, defaultFutureGovernor)
		keeperTestHelper.Require().NoError(err)

		pool, err := keeperTestHelper.App.GAMMKeeper.GetPool(keeperTestHelper.Ctx, poolId)
		keeperTestHelper.Require().NoError(err)

		pools = append(pools, pool)
	}
	return pools
}

// SwapAndSetSpotPrice runs a swap to set Spot price of a pool using arbitrary values
// returns spot price after the arbitrary swap
func (keeperTestHelper *KeeperTestHelper) SwapAndSetSpotPrice(poolId uint64, fromAsset gammtypes.PoolAsset, toAsset gammtypes.PoolAsset) sdk.Dec {
	// create a dummy account
	acc1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())

	// fund dummy account with tokens to swap
	coins := sdk.Coins{sdk.NewInt64Coin(fromAsset.Token.Denom, 100000000000000)}
	err := simapp.FundAccount(keeperTestHelper.App.BankKeeper, keeperTestHelper.Ctx, acc1, coins)
	keeperTestHelper.Require().NoError(err)

	_, _, err = keeperTestHelper.App.GAMMKeeper.SwapExactAmountOut(
		keeperTestHelper.Ctx, acc1,
		poolId, fromAsset.Token.Denom, fromAsset.Token.Amount,
		sdk.NewCoin(toAsset.Token.Denom, toAsset.Token.Amount.Quo(sdk.NewInt(4))))
	keeperTestHelper.Require().NoError(err)

	spotPrice, err := keeperTestHelper.App.GAMMKeeper.CalculateSpotPrice(keeperTestHelper.Ctx, poolId, toAsset.Token.Denom, fromAsset.Token.Denom)
	keeperTestHelper.Require().NoError(err)
	return spotPrice
}

func (keeperTestHelper *KeeperTestHelper) LockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockID uint64) {
	msgServer := lockupkeeper.NewMsgServerImpl(keeperTestHelper.App.LockupKeeper)
	err := simapp.FundAccount(keeperTestHelper.App.BankKeeper, keeperTestHelper.Ctx, addr, coins)
	keeperTestHelper.Require().NoError(err)
	msgResponse, err := msgServer.LockTokens(sdk.WrapSDKContext(keeperTestHelper.Ctx), lockuptypes.NewMsgLockTokens(addr, duration, coins))
	keeperTestHelper.Require().NoError(err)
	return msgResponse.ID
}
