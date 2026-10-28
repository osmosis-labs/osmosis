package main

import (
	"strings"
	"testing"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/apptesting"
)

const testDenom = "ibc/498A0751C798A0D9A389AA3691123DADA57DAA4FE165D5C75894505B876BA6E4"

type ScanTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestScanTestSuite(t *testing.T) {
	suite.Run(t, new(ScanTestSuite))
}

func (s *ScanTestSuite) SetupTest() {
	s.Setup()
}

// TestScanListsOnlyUnindexedHolders checks that a full balances scan finds a holder the denom
// index misses, omits indexed holders, and reconciles with supply.
func (s *ScanTestSuite) TestScanListsOnlyUnindexedHolders() {
	indexed := s.userWithPubKey()
	unindexed := s.userWithPubKey()
	s.FundAcc(indexed, sdk.NewCoins(sdk.NewInt64Coin(testDenom, 1_000)))
	s.FundAcc(unindexed, sdk.NewCoins(sdk.NewInt64Coin(testDenom, 300)))
	s.Require().NoError(s.App.BankKeeper.Balances.Indexes.Denom.Unreference(s.Ctx, collections.Join(unindexed, testDenom), nil))

	result, err := Scan(s.Ctx, &s.App.AppKeepers, ScanConfig{Denom: testDenom, MaxEntries: 1_000_000, MaxHolders: 100})
	s.Require().NoError(err)
	s.Require().Equal(2, result.Holders)
	s.Require().Equal([]sdk.AccAddress{unindexed}, result.Unindexed)
	s.Require().Equal(osmomath.NewInt(1_300), result.Supply)

	formatted := formatUnindexedList("osmosis-1", testDenom, 73270661, "ABCDEF", result.Unindexed)
	s.Require().Contains(formatted, "height 73270661")
	s.Require().Contains(formatted, "app hash ABCDEF")
	s.Require().Contains(formatted, "cmd/alloy-holder-scan")
	s.Require().True(strings.HasSuffix(formatted, unindexed.String()+"\n"))
	s.Require().NotContains(formatted, indexed.String())
}

// TestScanRejectsATruncatedHolderSet checks that a holder cap fails the scan instead of returning
// a short list.
func (s *ScanTestSuite) TestScanRejectsATruncatedHolderSet() {
	for range 3 {
		s.FundAcc(s.userWithPubKey(), sdk.NewCoins(sdk.NewInt64Coin(testDenom, 1)))
	}
	_, err := Scan(s.Ctx, &s.App.AppKeepers, ScanConfig{Denom: testDenom, MaxEntries: 1_000_000, MaxHolders: 2})
	s.Require().ErrorContains(err, "more than 2 holders")
}

// TestScanRejectsASupplyMismatch checks that a scan whose balances do not sum to supply fails.
func (s *ScanTestSuite) TestScanRejectsASupplyMismatch() {
	s.FundAcc(s.userWithPubKey(), sdk.NewCoins(sdk.NewInt64Coin(testDenom, 10)))
	s.Require().NoError(s.App.BankKeeper.Supply.Set(s.Ctx, testDenom, osmomath.NewInt(11)))

	_, err := Scan(s.Ctx, &s.App.AppKeepers, ScanConfig{Denom: testDenom, MaxEntries: 1_000_000, MaxHolders: 100})
	s.Require().ErrorContains(err, "supply is 11")
}

// userWithPubKey creates an account that has a public key set.
func (s *ScanTestSuite) userWithPubKey() sdk.AccAddress {
	pub := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pub.Address())
	acc := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr)
	s.Require().NoError(acc.SetPubKey(pub))
	s.App.AccountKeeper.SetAccount(s.Ctx, acc)
	return addr
}
