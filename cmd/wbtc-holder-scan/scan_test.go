package main

import (
	"io"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v32/app/apptesting"
	incentivestypes "github.com/osmosis-labs/osmosis/v32/x/incentives/types"
)

const testDenom = "factory/osmo1z0qrq605sjgcqpylfl4aa6s90x738j7m58wyatt0tdzflg2ha26q67k743/wbtc"

type ScanTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestScanTestSuite(t *testing.T) {
	suite.Run(t, new(ScanTestSuite))
}

func (s *ScanTestSuite) SetupTest() {
	s.Setup()
}

// TestScanFindsUnindexedHoldersAndClassifies checks that the full scan finds a holder missing from
// the denom index, reconciles with supply, and classifies each kind of holder.
func (s *ScanTestSuite) TestScanFindsUnindexedHoldersAndClassifies() {
	user := s.userWithPubKey()
	noPubKey := apptesting.CreateRandomAccounts(1)[0]
	unindexed := s.userWithPubKey()
	s.FundAcc(user, sdk.NewCoins(sdk.NewInt64Coin(testDenom, 1_000)))
	s.FundAcc(noPubKey, sdk.NewCoins(sdk.NewInt64Coin(testDenom, 20)))
	s.FundAcc(unindexed, sdk.NewCoins(sdk.NewInt64Coin(testDenom, 300)))
	s.FundModuleAcc(incentivestypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin(testDenom, 4_000)))

	// An open transfer channel whose escrow holds the denom.
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, "channel-0", channeltypes.Channel{
		State: channeltypes.OPEN, Ordering: channeltypes.UNORDERED,
		Counterparty: channeltypes.NewCounterparty(transfertypes.PortID, "channel-141"), ConnectionHops: []string{"connection-0"}, Version: transfertypes.Version,
	})
	escrow := transfertypes.GetEscrowAddress(transfertypes.PortID, "channel-0")
	s.FundAcc(escrow, sdk.NewCoins(sdk.NewInt64Coin(testDenom, 50_000)))

	// Reproduce mainnet: a balance that the denom index does not list.
	err := s.App.BankKeeper.Balances.Indexes.Denom.Unreference(s.Ctx, collections.Join(unindexed, testDenom), nil)
	s.Require().NoError(err)

	r, err := Scan(s.Ctx, &s.App.AppKeepers, Config{Denom: testDenom}, io.Discard)
	s.Require().NoError(err)

	s.Require().True(r.SupplyMatches, "holders %s, supply %s", r.HoldersTotal, r.Supply)
	s.Require().Equal(5, r.Holders)
	s.Require().Equal(1, r.UnindexedHolders)
	s.Require().Equal(osmomath.NewInt(300), r.UnindexedAmount)
	s.Require().Empty(r.StaleIndexEntries)

	byAddr := map[string]Holder{}
	for _, h := range r.holders {
		byAddr[h.Address] = h
	}
	s.Require().False(byAddr[unindexed.String()].InDenomIndex)
	s.Require().Equal("user_pubkey", byAddr[user.String()].Category)
	s.Require().True(byAddr[user.String()].WouldMigrate)
	s.Require().True(byAddr[unindexed.String()].WouldMigrate)
	s.Require().Equal("no_pubkey_other", byAddr[noPubKey.String()].Category)
	s.Require().False(byAddr[noPubKey.String()].WouldMigrate)
	s.Require().Equal("ibc_escrow", byAddr[escrow.String()].Category)
	s.Require().False(byAddr[escrow.String()].WouldMigrate)
	moduleAddr := s.App.AccountKeeper.GetModuleAddress(incentivestypes.ModuleName)
	s.Require().Equal("module_account", byAddr[moduleAddr.String()].Category)

	s.Require().Equal(2, r.WouldMigrate.Accounts)
	s.Require().Equal(osmomath.NewInt(1_300), r.WouldMigrate.Amount)
	s.Require().Equal(CategoryTotal{Accounts: 2, Amount: math.NewInt(1_300)}, r.Categories["user_pubkey"])
}

// userWithPubKey creates an account that has signed a transaction, i.e. has a public key set.
func (s *ScanTestSuite) userWithPubKey() sdk.AccAddress {
	pub := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pub.Address())
	acc := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr)
	s.Require().NoError(acc.SetPubKey(pub))
	s.App.AccountKeeper.SetAccount(s.Ctx, acc)
	return addr
}
