package keeper_test

import (
	"bytes"
	"encoding/hex"
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting/assets"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	tokenfactorytypes "github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"

	"github.com/cometbft/cometbft/crypto/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v27/x/oracle/types"
)

var (
	pubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	Addrs = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
		sdk.AccAddress(pubKeys[3].Address()),
		sdk.AccAddress(pubKeys[4].Address()),
	}

	ValAddrs = []sdk.ValAddress{
		sdk.ValAddress(pubKeys[0].Address()),
		sdk.ValAddress(pubKeys[1].Address()),
		sdk.ValAddress(pubKeys[2].Address()),
		sdk.ValAddress(pubKeys[3].Address()),
		sdk.ValAddress(pubKeys[4].Address()),
	}

	InitTokens = sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	InitCoins  = sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens))

	OracleDecPrecision = 8

	FaucetAccountName = tokenfactorytypes.ModuleName
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	valPubKeys []cryptotypes.PubKey
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	// Set the bond denom to be note to make volume tracking tests more readable.
	skParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
	require.NoError(s.T(), err)
	skParams.BondDenom = appparams.BaseCoinUnit
	s.App.StakingKeeper.SetParams(s.Ctx, skParams)
	s.App.TxFeesKeeper.SetBaseDenom(s.Ctx, "note")

	totalSupply := sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens.MulRaw(int64(len(Addrs)*10))))
	s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, totalSupply)

	s.App.BankKeeper.SendCoinsFromModuleToModule(s.Ctx, FaucetAccountName, stakingtypes.NotBondedPoolName, sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, InitTokens.MulRaw(int64(len(Addrs))))))

	for _, addr := range Addrs {
		s.App.AccountKeeper.SetAccount(s.Ctx, authtypes.NewBaseAccountWithAddress(addr))
		err := s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, addr, InitCoins)
		s.Require().NoError(err)
	}

	defaults := types.DefaultParams()
	s.App.OracleKeeper.SetParams(s.Ctx, defaults)
	for _, denom := range defaults.Whitelist {
		s.App.OracleKeeper.SetTobinTax(s.Ctx, denom.Name, denom.TobinTax)
	}
	s.valPubKeys = CreateTestPubKeys(5)
}

// NewTestMsgCreateValidator test msg creator
func (s *KeeperTestSuite) NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amt osmomath.Int) *stakingtypes.MsgCreateValidator {
	commission := stakingtypes.NewCommissionRates(osmomath.ZeroDec(), osmomath.ZeroDec(), osmomath.ZeroDec())
	msg, _ := stakingtypes.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(appparams.BaseCoinUnit, amt),
		stakingtypes.Description{}, commission, osmomath.OneInt(),
	)

	return msg
}

// FundAccount is a utility function that funds an account by minting and
// sending the coins to the address. This should be used for testing purposes
// only!
func (s *KeeperTestSuite) FundAccount(addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := s.App.BankKeeper.MintCoins(s.Ctx, FaucetAccountName, amounts); err != nil {
		return err
	}

	return s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, FaucetAccountName, addr, amounts)
}

// CreateTestPubKeys returns a total of numPubKeys public keys in ascending order.
func CreateTestPubKeys(numPubKeys int) []cryptotypes.PubKey {
	var publicKeys []cryptotypes.PubKey
	var buffer bytes.Buffer

	// start at 10 to avoid changing 1 to 01, 2 to 02, etc
	for i := 100; i < (numPubKeys + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AF") // base pubkey string
		buffer.WriteString(numString)                                                       // adding on final two digits to make pubkeys unique
		publicKeys = append(publicKeys, NewPubKeyFromHex(buffer.String()))
		buffer.Reset()
	}

	return publicKeys
}

// NewPubKeyFromHex returns a PubKey from a hex string.
func NewPubKeyFromHex(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	if len(pkBytes) != ed25519.PubKeySize {
		panic(errors.Wrap(errors.ErrInvalidPubKey, "invalid pubkey size"))
	}
	return &ed25519.PubKey{Key: pkBytes}
}

func (s *KeeperTestSuite) TestExchangeRate() {
	cnyExchangeRate := osmomath.NewDecWithPrec(839, int64(OracleDecPrecision)).MulInt64(appparams.MicroUnit)
	gbpExchangeRate := osmomath.NewDecWithPrec(4995, int64(OracleDecPrecision)).MulInt64(appparams.MicroUnit)
	krwExchangeRate := osmomath.NewDecWithPrec(2838, int64(OracleDecPrecision)).MulInt64(appparams.MicroUnit)
	noteExchangeRate := osmomath.NewDecWithPrec(3282384, int64(OracleDecPrecision)).MulInt64(appparams.MicroUnit)

	// Set & get rates
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroCNYDenom, cnyExchangeRate)
	rate, err := s.App.OracleKeeper.GetMelodyExchangeRate(s.Ctx, assets.MicroCNYDenom)
	s.Require().NoError(err)
	s.Require().Equal(cnyExchangeRate, rate)

	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroGBPDenom, gbpExchangeRate)
	rate, err = s.App.OracleKeeper.GetMelodyExchangeRate(s.Ctx, assets.MicroGBPDenom)
	s.Require().NoError(err)
	s.Require().Equal(gbpExchangeRate, rate)

	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroKRWDenom, krwExchangeRate)
	rate, err = s.App.OracleKeeper.GetMelodyExchangeRate(s.Ctx, assets.MicroKRWDenom)
	s.Require().NoError(err)
	s.Require().Equal(krwExchangeRate, rate)

	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, appparams.BaseCoinUnit, noteExchangeRate)
	rate, _ = s.App.OracleKeeper.GetMelodyExchangeRate(s.Ctx, appparams.BaseCoinUnit)
	s.Require().Equal(osmomath.OneDec(), rate)

	s.App.OracleKeeper.DeleteMelodyExchangeRate(s.Ctx, assets.MicroKRWDenom)
	_, err = s.App.OracleKeeper.GetMelodyExchangeRate(s.Ctx, assets.MicroKRWDenom)
	s.Require().Error(err)

	numExchangeRates := 0
	handler := func(denom string, exchangeRate osmomath.Dec) (stop bool) {
		numExchangeRates++
		return false
	}
	s.App.OracleKeeper.IterateNoteExchangeRates(s.Ctx, handler)

	s.Require().True(numExchangeRates == 3)
}

func (s *KeeperTestSuite) TestIterateMelodyExchangeRates() {
	cnyExchangeRate := osmomath.NewDecWithPrec(839, int64(OracleDecPrecision)).MulInt64(appparams.MicroUnit)
	gbpExchangeRate := osmomath.NewDecWithPrec(4995, int64(OracleDecPrecision)).MulInt64(appparams.MicroUnit)
	krwExchangeRate := osmomath.NewDecWithPrec(2838, int64(OracleDecPrecision)).MulInt64(appparams.MicroUnit)
	melodyExchangeRate := osmomath.NewDecWithPrec(3282384, int64(OracleDecPrecision)).MulInt64(appparams.MicroUnit)

	// Set & get rates
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroCNYDenom, cnyExchangeRate)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroGBPDenom, gbpExchangeRate)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, assets.MicroKRWDenom, krwExchangeRate)
	s.App.OracleKeeper.SetMelodyExchangeRate(s.Ctx, appparams.BaseCoinUnit, melodyExchangeRate)

	s.App.OracleKeeper.IterateNoteExchangeRates(s.Ctx, func(denom string, rate osmomath.Dec) (stop bool) {
		switch denom {
		case assets.MicroCNYDenom:
			s.Require().Equal(cnyExchangeRate, rate)
		case assets.MicroGBPDenom:
			s.Require().Equal(gbpExchangeRate, rate)
		case assets.MicroKRWDenom:
			s.Require().Equal(krwExchangeRate, rate)
		case appparams.BaseCoinUnit:
			s.Require().Equal(melodyExchangeRate, rate)
		}
		return false
	})
}

func (s *KeeperTestSuite) TestRewardPool() {
	fees := sdk.NewCoins(sdk.NewCoin(assets.MicroSDRDenom, osmomath.NewInt(1000)))
	acc := s.App.AccountKeeper.GetModuleAccount(s.Ctx, types.ModuleName)
	err := s.FundAccount(acc.GetAddress(), fees)
	if err != nil {
		panic(err) // never occurs
	}

	KFees := s.App.OracleKeeper.GetRewardPool(s.Ctx, assets.MicroSDRDenom)
	s.Require().Equal(fees[0], KFees)
}

func (s *KeeperTestSuite) TestParams() {
	// Test default params setting
	s.App.OracleKeeper.SetParams(s.Ctx, types.DefaultParams())
	params := s.App.OracleKeeper.GetParams(s.Ctx)
	s.Require().NotNil(params)

	// Test custom params setting
	votePeriod := uint64(10)
	voteThreshold := osmomath.NewDecWithPrec(33, 2)
	oracleRewardBand := osmomath.NewDecWithPrec(1, 2)
	rewardDistributionWindow := uint64(10000000000000)
	slashFraction := osmomath.NewDecWithPrec(1, 2)
	slashWindow := uint64(1000)
	minValidPerWindow := osmomath.NewDecWithPrec(1, 4)
	whitelist := types.DenomList{
		{Name: assets.MicroSDRDenom, TobinTax: types.DefaultTobinTax},
		{Name: assets.MicroKRWDenom, TobinTax: types.DefaultTobinTax},
	}

	// Should really test validateParams, but skipping because obvious
	newParams := types.Params{
		VotePeriod:               votePeriod,
		VoteThreshold:            voteThreshold,
		RewardBand:               oracleRewardBand,
		RewardDistributionWindow: rewardDistributionWindow,
		Whitelist:                whitelist,
		SlashFraction:            slashFraction,
		SlashWindow:              slashWindow,
		MinValidPerWindow:        minValidPerWindow,
	}
	s.App.OracleKeeper.SetParams(s.Ctx, newParams)

	storedParams := s.App.OracleKeeper.GetParams(s.Ctx)
	s.Require().NotNil(storedParams)
	s.Require().Equal(storedParams, newParams)
}

func (s *KeeperTestSuite) TestFeederDelegation() {
	// Test default getters and setters
	delegate := s.App.OracleKeeper.GetFeederDelegation(s.Ctx, ValAddrs[0])
	s.Require().Equal(Addrs[0], delegate)

	s.App.OracleKeeper.SetFeederDelegation(s.Ctx, ValAddrs[0], Addrs[1])
	delegate = s.App.OracleKeeper.GetFeederDelegation(s.Ctx, ValAddrs[0])
	s.Require().Equal(Addrs[1], delegate)
}

func (s *KeeperTestSuite) TestIterateFeederDelegations() {
	// Test default getters and setters
	delegate := s.App.OracleKeeper.GetFeederDelegation(s.Ctx, ValAddrs[0])
	s.Require().Equal(Addrs[0], delegate)

	s.App.OracleKeeper.SetFeederDelegation(s.Ctx, ValAddrs[0], Addrs[1])

	var delegators []sdk.ValAddress
	var delegates []sdk.AccAddress
	s.App.OracleKeeper.IterateFeederDelegations(s.Ctx, func(delegator sdk.ValAddress, delegate sdk.AccAddress) (stop bool) {
		delegators = append(delegators, delegator)
		delegates = append(delegates, delegate)
		return false
	})

	s.Require().Equal(1, len(delegators))
	s.Require().Equal(1, len(delegates))
	s.Require().Equal(ValAddrs[0], delegators[0])
	s.Require().Equal(Addrs[1], delegates[0])
}

func (s *KeeperTestSuite) TestMissCounter() {
	// Test default getters and setters
	counter := s.App.OracleKeeper.GetMissCounter(s.Ctx, ValAddrs[0])
	s.Require().Equal(uint64(0), counter)

	missCounter := uint64(10)
	s.App.OracleKeeper.SetMissCounter(s.Ctx, ValAddrs[0], missCounter)
	counter = s.App.OracleKeeper.GetMissCounter(s.Ctx, ValAddrs[0])
	s.Require().Equal(missCounter, counter)

	s.App.OracleKeeper.DeleteMissCounter(s.Ctx, ValAddrs[0])
	counter = s.App.OracleKeeper.GetMissCounter(s.Ctx, ValAddrs[0])
	s.Require().Equal(uint64(0), counter)
}

func (s *KeeperTestSuite) TestIterateMissCounters() {
	// Test default getters and setters
	counter := s.App.OracleKeeper.GetMissCounter(s.Ctx, ValAddrs[0])
	s.Require().Equal(uint64(0), counter)

	missCounter := uint64(10)
	s.App.OracleKeeper.SetMissCounter(s.Ctx, ValAddrs[1], missCounter)

	var operators []sdk.ValAddress
	var missCounters []uint64
	s.App.OracleKeeper.IterateMissCounters(s.Ctx, func(delegator sdk.ValAddress, missCounter uint64) (stop bool) {
		operators = append(operators, delegator)
		missCounters = append(missCounters, missCounter)
		return false
	})

	s.Require().Equal(1, len(operators))
	s.Require().Equal(1, len(missCounters))
	s.Require().Equal(ValAddrs[1], operators[0])
	s.Require().Equal(missCounter, missCounters[0])
}

func (s *KeeperTestSuite) TestAggregatePrevoteAddDelete() {
	hash := types.GetAggregateVoteHash("salt", "100ukrw,1000uusd", sdk.ValAddress(Addrs[0]))
	aggregatePrevote := types.NewAggregateExchangeRatePrevote(hash, sdk.ValAddress(Addrs[0]), 0)
	s.App.OracleKeeper.SetAggregateExchangeRatePrevote(s.Ctx, sdk.ValAddress(Addrs[0]), aggregatePrevote)

	KPrevote, err := s.App.OracleKeeper.GetAggregateExchangeRatePrevote(s.Ctx, sdk.ValAddress(Addrs[0]))
	s.Require().NoError(err)
	s.Require().Equal(aggregatePrevote, KPrevote)

	s.App.OracleKeeper.DeleteAggregateExchangeRatePrevote(s.Ctx, sdk.ValAddress(Addrs[0]))
	_, err = s.App.OracleKeeper.GetAggregateExchangeRatePrevote(s.Ctx, sdk.ValAddress(Addrs[0]))
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestAggregatePrevoteIterate() {
	hash := types.GetAggregateVoteHash("salt", "100ukrw,1000uusd", sdk.ValAddress(Addrs[0]))
	aggregatePrevote1 := types.NewAggregateExchangeRatePrevote(hash, sdk.ValAddress(Addrs[0]), 0)
	s.App.OracleKeeper.SetAggregateExchangeRatePrevote(s.Ctx, sdk.ValAddress(Addrs[0]), aggregatePrevote1)

	hash2 := types.GetAggregateVoteHash("salt", "100ukrw,1000uusd", sdk.ValAddress(Addrs[1]))
	aggregatePrevote2 := types.NewAggregateExchangeRatePrevote(hash2, sdk.ValAddress(Addrs[1]), 0)
	s.App.OracleKeeper.SetAggregateExchangeRatePrevote(s.Ctx, sdk.ValAddress(Addrs[1]), aggregatePrevote2)

	i := 0
	bigger := bytes.Compare(Addrs[0], Addrs[1])
	s.App.OracleKeeper.IterateAggregateExchangeRatePrevotes(s.Ctx, func(voter sdk.ValAddress, p types.AggregateExchangeRatePrevote) (stop bool) {
		if (i == 0 && bigger == -1) || (i == 1 && bigger == 1) {
			s.Require().Equal(aggregatePrevote1, p)
			s.Require().Equal(voter.String(), p.Voter)
		} else {
			s.Require().Equal(aggregatePrevote2, p)
			s.Require().Equal(voter.String(), p.Voter)
		}

		i++
		return false
	})
}

func (s *KeeperTestSuite) TestAggregateVoteAddDelete() {
	aggregateVote := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
		{Denom: "foo", ExchangeRate: osmomath.NewDec(-1)},
		{Denom: "foo", ExchangeRate: osmomath.NewDec(0)},
		{Denom: "foo", ExchangeRate: osmomath.NewDec(1)},
	}, sdk.ValAddress(Addrs[0]))
	s.App.OracleKeeper.SetAggregateExchangeRateVote(s.Ctx, sdk.ValAddress(Addrs[0]), aggregateVote)

	KVote, err := s.App.OracleKeeper.GetAggregateExchangeRateVote(s.Ctx, sdk.ValAddress(Addrs[0]))
	s.Require().NoError(err)
	s.Require().Equal(aggregateVote, KVote)

	s.App.OracleKeeper.DeleteAggregateExchangeRateVote(s.Ctx, sdk.ValAddress(Addrs[0]))
	_, err = s.App.OracleKeeper.GetAggregateExchangeRateVote(s.Ctx, sdk.ValAddress(Addrs[0]))
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestAggregateVoteIterate() {
	aggregateVote1 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
		{Denom: "foo", ExchangeRate: osmomath.NewDec(-1)},
		{Denom: "foo", ExchangeRate: osmomath.NewDec(0)},
		{Denom: "foo", ExchangeRate: osmomath.NewDec(1)},
	}, sdk.ValAddress(Addrs[0]))
	s.App.OracleKeeper.SetAggregateExchangeRateVote(s.Ctx, sdk.ValAddress(Addrs[0]), aggregateVote1)

	aggregateVote2 := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{
		{Denom: "foo", ExchangeRate: osmomath.NewDec(-1)},
		{Denom: "foo", ExchangeRate: osmomath.NewDec(0)},
		{Denom: "foo", ExchangeRate: osmomath.NewDec(1)},
	}, sdk.ValAddress(Addrs[1]))
	s.App.OracleKeeper.SetAggregateExchangeRateVote(s.Ctx, sdk.ValAddress(Addrs[1]), aggregateVote2)

	i := 0
	bigger := bytes.Compare(address.MustLengthPrefix(Addrs[0]), address.MustLengthPrefix(Addrs[1]))
	s.App.OracleKeeper.IterateAggregateExchangeRateVotes(s.Ctx, func(voter sdk.ValAddress, p types.AggregateExchangeRateVote) (stop bool) {
		if (i == 0 && bigger == -1) || (i == 1 && bigger == 1) {
			s.Require().Equal(aggregateVote1, p)
			s.Require().Equal(voter.String(), p.Voter)
		} else {
			s.Require().Equal(aggregateVote2, p)
			s.Require().Equal(voter.String(), p.Voter)
		}

		i++
		return false
	})
}

func (s *KeeperTestSuite) TestTobinTaxGetSet() {
	tobinTaxes := map[string]osmomath.Dec{
		assets.MicroSDRDenom: osmomath.NewDec(1),
		assets.MicroUSDDenom: osmomath.NewDecWithPrec(1, 3),
		assets.StakeDenom:    osmomath.NewDec(1),
	}

	for denom, tobinTax := range tobinTaxes {
		s.App.OracleKeeper.SetTobinTax(s.Ctx, denom, tobinTax)
		factor, err := s.App.OracleKeeper.GetTobinTax(s.Ctx, denom)
		s.Require().NoError(err)
		s.Require().Equal(tobinTaxes[denom], factor)
	}

	s.App.OracleKeeper.IterateTobinTaxes(s.Ctx, func(denom string, tobinTax osmomath.Dec) (stop bool) {
		s.Require().Equal(tobinTaxes[denom], tobinTax)
		return false
	})

	s.App.OracleKeeper.ClearTobinTaxes(s.Ctx)
	for denom := range tobinTaxes {
		_, err := s.App.OracleKeeper.GetTobinTax(s.Ctx, denom)
		s.Require().Error(err)
	}
}

func (s *KeeperTestSuite) TestValidateFeeder() {
	// initial setup
	addr, val := ValAddrs[0], s.valPubKeys[0]
	addr1, val1 := ValAddrs[1], s.valPubKeys[1]
	amt := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	stakingMsgSvr := stakingkeeper.NewMsgServerImpl(s.App.StakingKeeper)
	ctx := s.Ctx

	// Validator created
	_, err := stakingMsgSvr.CreateValidator(ctx, s.NewTestMsgCreateValidator(addr, val, amt))
	s.Require().NoError(err)
	_, err = stakingMsgSvr.CreateValidator(ctx, s.NewTestMsgCreateValidator(addr1, val1, amt))
	s.Require().NoError(err)
	staking.EndBlocker(ctx, s.App.StakingKeeper)

	s.Require().Equal(
		s.App.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(s.App.StakingKeeper.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	s.Require().Equal(amt, s.App.StakingKeeper.Validator(ctx, addr).GetBondedTokens())
	s.Require().Equal(
		s.App.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr1)),
		sdk.NewCoins(sdk.NewCoin(s.App.StakingKeeper.GetParams(ctx).BondDenom, InitTokens.Sub(amt))),
	)
	s.Require().Equal(amt, s.App.StakingKeeper.Validator(ctx, addr1).GetBondedTokens())

	s.Require().NoError(s.App.OracleKeeper.ValidateFeeder(s.Ctx, sdk.AccAddress(addr), addr))
	s.Require().NoError(s.App.OracleKeeper.ValidateFeeder(s.Ctx, sdk.AccAddress(addr1), addr1))

	// delegate works
	s.App.OracleKeeper.SetFeederDelegation(s.Ctx, addr, sdk.AccAddress(addr1))
	s.Require().NoError(s.App.OracleKeeper.ValidateFeeder(s.Ctx, sdk.AccAddress(addr1), addr))
	s.Require().Error(s.App.OracleKeeper.ValidateFeeder(s.Ctx, Addrs[2], addr))

	// only active validators can do oracle votes
	validator, found := s.App.StakingKeeper.GetValidator(s.Ctx, addr)
	s.Require().True(found)
	validator.Status = stakingtypes.Unbonded
	s.App.StakingKeeper.SetValidator(s.Ctx, validator)
	s.Require().Error(s.App.OracleKeeper.ValidateFeeder(s.Ctx, sdk.AccAddress(addr1), addr))
}
