package cosmwasmpool_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
)

type WhitelistSuite struct {
	apptesting.KeeperTestHelper
}

func TestWhitelistSuite(t *testing.T) {
	suite.Run(t, new(WhitelistSuite))
}

// TestWhitelist tests basic whitelist functionality
// 1. Check that a code id is not whitelisted by default
// 2. Add a code id to the whitelist and check that it is now whitelisted
// 3. Add the same code id to the whitelist and check that only one entry exists in params
// 4. Add another code id to the whitelist and check that it is now whitelisted
// 5. Try to remove a code id that is not in the whitelist and check that the whitelist is unchanged
// 6. Remove the first code id from the whitelist and check that it is no longer whitelisted while second is still there.
func (s *WhitelistSuite) TestWhitelist() {
	s.Setup()

	const (
		defaultCodeId uint64 = 5
	)

	// Check that the pool is not whitelisted
	s.Require().False(s.App.CosmwasmPoolKeeper.IsWhitelisted(s.Ctx, defaultCodeId))

	// Add the pool to the whitelist and check that it is now whitelisted.
	s.App.CosmwasmPoolKeeper.WhitelistCodeId(s.Ctx, defaultCodeId)
	s.Require().True(s.App.CosmwasmPoolKeeper.IsWhitelisted(s.Ctx, defaultCodeId))

	// Whitelist the same code id and assert that only one entrye exists in params
	s.App.CosmwasmPoolKeeper.WhitelistCodeId(s.Ctx, defaultCodeId)
	whitelist := s.App.CosmwasmPoolKeeper.GetParams(s.Ctx).CodeIdWhitelist
	s.Require().Equal(1, len(whitelist))
	s.Require().Equal(defaultCodeId, whitelist[0])

	// Add another pool to the whitelist and check that it is now whitelisted.
	s.App.CosmwasmPoolKeeper.WhitelistCodeId(s.Ctx, defaultCodeId+1)
	s.Require().True(s.App.CosmwasmPoolKeeper.IsWhitelisted(s.Ctx, defaultCodeId+1))

	// Try to remove a code id that is not in the whitelist and check that the whitelist is unchanged.
	s.App.CosmwasmPoolKeeper.DeWhitelistCodeId(s.Ctx, defaultCodeId+2)
	s.Require().True(s.App.CosmwasmPoolKeeper.IsWhitelisted(s.Ctx, defaultCodeId))
	s.Require().True(s.App.CosmwasmPoolKeeper.IsWhitelisted(s.Ctx, defaultCodeId+1))

	// Remove the first pool from the whitelist and check that it is no longer whitelisted
	// while the other pool still is.
	s.App.CosmwasmPoolKeeper.DeWhitelistCodeId(s.Ctx, defaultCodeId)
	s.Require().False(s.App.CosmwasmPoolKeeper.IsWhitelisted(s.Ctx, defaultCodeId))
	s.Require().True(s.App.CosmwasmPoolKeeper.IsWhitelisted(s.Ctx, defaultCodeId+1))
}
