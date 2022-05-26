package e2e

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	configurer "github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer"
)

type chainConfig struct {
	// voting period is number of blocks it takes to deposit, 1.2 seconds per validator to vote on the prop, and a buffer.
	votingPeriod float32
	// upgrade proposal height for chain.
	propHeight int
	// Indexes of the validators to skip from running during initialization.
	// This is needed for testing functionality like state-sync where we would
	// like to start a node during tests post-initialization.
	skipRunValidatorIndexes map[int]struct{}
	chain                   *chain.Chain
}

const (
	// Environment variable name to skip the upgrade tests
	skipUpgradeEnv = "OSMOSIS_E2E_SKIP_UPGRADE"
	// Environment variable name to skip the IBC tests
	skipIBCEnv = "OSMOSIS_E2E_SKIP_IBC"
	// Environment variable name to skip cleaning up Docker resources in teardown.
	skipCleanupEnv = "OSMOSIS_E2E_SKIP_CLEANUP"
)

type IntegrationTestSuite struct {
	suite.Suite

	configurer configurer.Configurer

	skipUpgrade bool
	skipIBC     bool
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	// The e2e test flow is as follows:
	//
	// 1. Configure two chains - chan A and chain B.
	//   * For each chain, set up two validators
	//   * Initialize configs and genesis for all validators.
	// 2. Start both networks.
	// 3. Run IBC relayer betweeen the two chains.
	// 4. Execute various e2e tests, including IBC.
	var (
		err error
	)

	if str := os.Getenv(skipUpgradeEnv); len(str) > 0 {
		s.skipUpgrade, err = strconv.ParseBool(str)
		s.Require().NoError(err)
		s.T().Log(fmt.Sprintf("%s was true, skipping upgrafe tests", skipIBCEnv))
	}

	if str := os.Getenv(skipIBCEnv); len(str) > 0 {
		s.skipIBC, err = strconv.ParseBool(str)
		s.Require().NoError(err)
		s.T().Log(fmt.Sprintf("%s was true, skipping IBC tests", skipIBCEnv))
	}

	s.configurer, err = configurer.New(s.T(), !s.skipIBC, !s.skipUpgrade)
	s.Require().NoError(err)

	err = s.configurer.ConfigureChains()
	s.Require().NoError(err)

	err = s.configurer.RunSetup()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if str := os.Getenv(skipCleanupEnv); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
	}

	err := s.configurer.ClearResources()
	s.Require().NoError(err)
}
