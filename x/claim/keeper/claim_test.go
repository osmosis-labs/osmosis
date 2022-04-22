package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v7/x/claim/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestHookOfUnclaimableAccount() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())
	suite.App.AccountKeeper.SetAccount(suite.Ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	claim, err := suite.App.ClaimKeeper.GetClaimRecord(suite.Ctx, addr1)
	suite.NoError(err)
	suite.Equal(types.ClaimRecord{}, claim)

	suite.App.ClaimKeeper.AfterSwap(suite.Ctx, addr1)

	balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
	suite.Equal(sdk.Coins{}, balances)
}

func (suite *KeeperTestSuite) TestHookBeforeAirdropStart() {
	suite.SetupTest()

	airdropStartTime := time.Now().Add(time.Hour)

	err := suite.App.ClaimKeeper.SetParams(suite.Ctx, types.Params{
		AirdropStartTime:   airdropStartTime,
		DurationUntilDecay: time.Hour,
		DurationOfDecay:    time.Hour * 4,
	})
	suite.Require().NoError(err)

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}
	suite.App.AccountKeeper.SetAccount(suite.Ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	err = suite.App.ClaimKeeper.SetClaimRecords(suite.Ctx, claimRecords)
	suite.Require().NoError(err)

	coins, err := suite.App.ClaimKeeper.GetUserTotalClaimable(suite.Ctx, addr1)
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(coins.Empty())

	coins, err = suite.App.ClaimKeeper.GetClaimableAmountForAction(suite.Ctx, addr1, types.ActionSwap)
	suite.NoError(err)
	// Now, it is before starting air drop, so this value should return the empty coins
	suite.True(coins.Empty())

	suite.App.ClaimKeeper.AfterSwap(suite.Ctx, addr1)
	balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
	// Now, it is before starting air drop, so claim module should not send the balances to the user after swap.
	suite.True(balances.Empty())

	suite.App.ClaimKeeper.AfterSwap(suite.Ctx.WithBlockTime(airdropStartTime), addr1)
	balances = suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
	// Now, it is the time for air drop, so claim module should send the balances to the user after swap.
	suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)), balances.AmountOf(sdk.DefaultBondDenom))
}

func (suite *KeeperTestSuite) TestHookAfterAirdropEnd() {
	suite.SetupTest()

	// airdrop recipient address
	addr1, _ := sdk.AccAddressFromBech32("osmo122fypjdzwscz998aytrrnmvavtaaarjjt6223p")

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}
	suite.App.AccountKeeper.SetAccount(suite.Ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
	err := suite.App.ClaimKeeper.SetClaimRecords(suite.Ctx, claimRecords)
	suite.Require().NoError(err)

	params, err := suite.App.ClaimKeeper.GetParams(suite.Ctx)
	suite.Require().NoError(err)
	suite.Ctx = suite.Ctx.WithBlockTime(params.AirdropStartTime.Add(params.DurationUntilDecay).Add(params.DurationOfDecay))

	suite.App.ClaimKeeper.EndAirdrop(suite.Ctx)

	suite.Require().NotPanics(func() {
		suite.App.ClaimKeeper.AfterSwap(suite.Ctx, addr1)
	})
}

func (suite *KeeperTestSuite) TestDuplicatedActionNotWithdrawRepeatedly() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}
	suite.App.AccountKeeper.SetAccount(suite.Ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))

	err := suite.App.ClaimKeeper.SetClaimRecords(suite.Ctx, claimRecords)
	suite.Require().NoError(err)

	coins1, err := suite.App.ClaimKeeper.GetUserTotalClaimable(suite.Ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, claimRecords[0].InitialClaimableAmount)

	suite.App.ClaimKeeper.AfterSwap(suite.Ctx, addr1)
	claim, err := suite.App.ClaimKeeper.GetClaimRecord(suite.Ctx, addr1)
	suite.NoError(err)
	suite.True(claim.ActionCompleted[types.ActionSwap])
	claimedCoins := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(sdk.DefaultBondDenom), claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)))

	suite.App.ClaimKeeper.AfterSwap(suite.Ctx, addr1)
	claim, err = suite.App.ClaimKeeper.GetClaimRecord(suite.Ctx, addr1)
	suite.NoError(err)
	suite.True(claim.ActionCompleted[types.ActionSwap])
	claimedCoins = suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
	suite.Require().Equal(claimedCoins.AmountOf(sdk.DefaultBondDenom), claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)))
}

func (suite *KeeperTestSuite) TestDelegationAutoWithdrawAndDelegateMore() {
	suite.SetupTest()

	pub1 := secp256k1.GenPrivKey().PubKey()
	pub2 := secp256k1.GenPrivKey().PubKey()
	addrs := []sdk.AccAddress{sdk.AccAddress(pub1.Address()), sdk.AccAddress(pub2.Address())}
	claimRecords := []types.ClaimRecord{
		{
			Address:                addrs[0].String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
		{
			Address:                addrs[1].String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}

	// initialize accts
	for i := 0; i < len(addrs); i++ {
		suite.App.AccountKeeper.SetAccount(suite.Ctx, authtypes.NewBaseAccount(addrs[i], nil, 0, 0))
	}
	// initialize claim records
	err := suite.App.ClaimKeeper.SetClaimRecords(suite.Ctx, claimRecords)
	suite.Require().NoError(err)

	// test claim records set
	for i := 0; i < len(addrs); i++ {
		coins, err := suite.App.ClaimKeeper.GetUserTotalClaimable(suite.Ctx, addrs[i])
		suite.Require().NoError(err)
		suite.Require().Equal(coins, claimRecords[i].InitialClaimableAmount)
	}

	// set addr[0] as a validator
	validator, err := stakingtypes.NewValidator(sdk.ValAddress(addrs[0]), pub1, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator = stakingkeeper.TestingUpdateValidator(*suite.App.StakingKeeper, suite.Ctx, validator, true)
	suite.App.StakingKeeper.AfterValidatorCreated(suite.Ctx, validator.GetOperator())

	validator, _ = validator.AddTokensFromDel(sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction))
	delAmount := sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
	suite.FundAcc(addrs[1],
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, delAmount)))
	_, err = suite.App.StakingKeeper.Delegate(suite.Ctx, addrs[1], delAmount, stakingtypes.Unbonded, validator, true)
	suite.Require().NoError(err)

	// delegation should automatically call claim and withdraw balance
	actualClaimedCoins := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addrs[1])
	actualClaimedCoin := actualClaimedCoins.AmountOf(sdk.DefaultBondDenom)
	expectedClaimedCoin := claimRecords[1].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(int64(len(claimRecords[1].ActionCompleted))))
	suite.Require().Equal(expectedClaimedCoin.String(), actualClaimedCoin.String())

	_, err = suite.App.StakingKeeper.Delegate(suite.Ctx, addrs[1], actualClaimedCoin, stakingtypes.Unbonded, validator, true)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TestAirdropFlow() {
	suite.SetupTest()

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
			ActionCompleted:        []bool{false, false, false, false},
		},
		{
			Address:                addr2.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}

	err := suite.App.ClaimKeeper.SetClaimRecords(suite.Ctx, claimRecords)
	suite.Require().NoError(err)

	coins1, err := suite.App.ClaimKeeper.GetUserTotalClaimable(suite.Ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(coins1, claimRecords[0].InitialClaimableAmount, coins1.String())

	coins2, err := suite.App.ClaimKeeper.GetUserTotalClaimable(suite.Ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, claimRecords[1].InitialClaimableAmount)

	coins3, err := suite.App.ClaimKeeper.GetUserTotalClaimable(suite.Ctx, addr3)
	suite.Require().NoError(err)
	suite.Require().Equal(coins3, sdk.Coins{})

	// get rewards amount per action
	coins4, err := suite.App.ClaimKeeper.GetClaimableAmountForAction(suite.Ctx, addr1, types.ActionAddLiquidity)
	suite.Require().NoError(err)
	suite.Require().Equal(coins4.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 25)).String()) // 2 = 10.Quo(4)

	// get completed activities
	claimRecord, err := suite.App.ClaimKeeper.GetClaimRecord(suite.Ctx, addr1)
	suite.Require().NoError(err)
	for i := range types.Action_name {
		suite.Require().False(claimRecord.ActionCompleted[i])
	}

	// do half of actions
	suite.App.ClaimKeeper.AfterAddLiquidity(suite.Ctx, addr1)
	suite.App.ClaimKeeper.AfterSwap(suite.Ctx, addr1)

	// check that half are completed
	claimRecord, err = suite.App.ClaimKeeper.GetClaimRecord(suite.Ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().True(claimRecord.ActionCompleted[types.ActionAddLiquidity])
	suite.Require().True(claimRecord.ActionCompleted[types.ActionSwap])

	// get balance after 2 actions done
	coins1 = suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)).String())

	// check that claimable for completed activity is 0
	coins4, err = suite.App.ClaimKeeper.GetClaimableAmountForAction(suite.Ctx, addr1, types.ActionAddLiquidity)
	suite.Require().NoError(err)
	suite.Require().Equal(coins4.String(), sdk.Coins{}.String()) // 2 = 10.Quo(4)

	// do rest of actions
	suite.App.ClaimKeeper.AfterProposalVote(suite.Ctx, 1, addr1)
	suite.App.ClaimKeeper.AfterDelegationModified(suite.Ctx, addr1, sdk.ValAddress(addr1))

	// get balance after rest actions done
	coins1 = suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr1)
	suite.Require().Equal(coins1.String(), sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)).String())

	// get claimable after withdrawing all
	coins1, err = suite.App.ClaimKeeper.GetUserTotalClaimable(suite.Ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().True(coins1.Empty())

	err = suite.App.ClaimKeeper.EndAirdrop(suite.Ctx)
	suite.Require().NoError(err)

	moduleAccAddr := suite.App.AccountKeeper.GetModuleAddress(types.ModuleName)
	coins := suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAccAddr, sdk.DefaultBondDenom)
	suite.Require().Equal(coins, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))

	coins2, err = suite.App.ClaimKeeper.GetUserTotalClaimable(suite.Ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().Equal(coins2, sdk.Coins{})
}

func (suite *KeeperTestSuite) TestClaimOfDecayed() {
	airdropStartTime := time.Now()
	durationUntilDecay := time.Hour
	durationOfDecay := time.Hour * 4

	pub1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pub1.Address())

	claimRecords := []types.ClaimRecord{
		{
			Address:                addr1.String(),
			InitialClaimableAmount: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)),
			ActionCompleted:        []bool{false, false, false, false},
		},
	}

	tests := []struct {
		fn func()
	}{
		{
			fn: func() {
				ctx := suite.Ctx.WithBlockTime(airdropStartTime)
				coins, err := suite.App.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionSwap)
				suite.NoError(err)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

				suite.App.ClaimKeeper.AfterSwap(ctx, addr1)
				coins = suite.App.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
			},
		},
		{
			fn: func() {
				ctx := suite.Ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay))
				coins, err := suite.App.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionSwap)
				suite.NoError(err)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

				suite.App.ClaimKeeper.AfterSwap(ctx, addr1)
				coins = suite.App.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(4)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
			},
		},
		{
			fn: func() {
				ctx := suite.Ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay / 2))
				coins, err := suite.App.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionSwap)
				suite.NoError(err)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(8)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())

				suite.App.ClaimKeeper.AfterSwap(ctx, addr1)
				coins = suite.App.BankKeeper.GetAllBalances(ctx, addr1)
				suite.Equal(claimRecords[0].InitialClaimableAmount.AmountOf(sdk.DefaultBondDenom).Quo(sdk.NewInt(8)).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
			},
		},
		{
			fn: func() {
				ctx := suite.Ctx.WithBlockTime(airdropStartTime.Add(durationUntilDecay).Add(durationOfDecay))
				coins, err := suite.App.ClaimKeeper.GetClaimableAmountForAction(ctx, addr1, types.ActionSwap)
				suite.NoError(err)
				suite.True(coins.Empty())

				suite.App.ClaimKeeper.AfterSwap(ctx, addr1)
				coins = suite.App.BankKeeper.GetAllBalances(ctx, addr1)
				suite.True(coins.Empty())
			},
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		err := suite.App.ClaimKeeper.SetParams(suite.Ctx, types.Params{
			AirdropStartTime:   airdropStartTime,
			DurationUntilDecay: durationUntilDecay,
			DurationOfDecay:    durationOfDecay,
		})
		suite.NoError(err)

		suite.App.AccountKeeper.SetAccount(suite.Ctx, authtypes.NewBaseAccount(addr1, nil, 0, 0))
		err = suite.App.ClaimKeeper.SetClaimRecords(suite.Ctx, claimRecords)
		suite.Require().NoError(err)

		test.fn()
	}
}

func (suite *KeeperTestSuite) TestClawbackAirdrop() {
	suite.SetupTest()

	tests := []struct {
		name           string
		address        string
		sequence       uint64
		expectClawback bool
	}{
		{
			name:           "airdrop address active",
			address:        "osmo122fypjdzwscz998aytrrnmvavtaaarjjt6223p",
			sequence:       1,
			expectClawback: false,
		},
		{
			name:           "airdrop address inactive",
			address:        "osmo122g3jv9que3zkxy25a2wt0wlgh68mudwptyvzw",
			sequence:       0,
			expectClawback: true,
		},
		{
			name:           "non airdrop address active",
			address:        sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			sequence:       1,
			expectClawback: false,
		},
		{
			name:           "non airdrop address inactive",
			address:        sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			sequence:       0,
			expectClawback: false,
		},
	}

	for _, tc := range tests {
		addr, err := sdk.AccAddressFromBech32(tc.address)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		acc := authtypes.NewBaseAccountWithAddress(addr)
		err = acc.SetSequence(tc.sequence)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		suite.App.AccountKeeper.SetAccount(suite.Ctx, acc)
		coins := sdk.NewCoins(
			sdk.NewInt64Coin("uosmo", 100), sdk.NewInt64Coin("uion", 100))
		simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, coins)
	}

	err := suite.App.ClaimKeeper.EndAirdrop(suite.Ctx)
	suite.Require().NoError(err, "err: %s", err)

	for _, tc := range tests {
		addr, err := sdk.AccAddressFromBech32(tc.address)
		suite.Require().NoError(err, "err: %s test: %s", err, tc.name)
		coins := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
		if tc.expectClawback {
			suite.Require().True(coins.IsEqual(sdk.NewCoins()),
				"balance incorrect. test: %s", tc.name)
		} else {
			suite.Require().True(coins.IsEqual(sdk.NewCoins(
				sdk.NewInt64Coin("uosmo", 100), sdk.NewInt64Coin("uion", 100),
			)), "balance incorrect. test: %s", tc.name)
		}
	}
}
