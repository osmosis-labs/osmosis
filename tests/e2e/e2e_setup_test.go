package e2e

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	configurer "github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer"
)

const (
	// Environment variable name to skip the upgrade tests
	skipUpgradeEnv = "OSMOSIS_E2E_SKIP_UPGRADE"
	// Environment variable name to skip the IBC tests
	skipIBCEnv = "OSMOSIS_E2E_SKIP_IBC"
	// Environment variable name to determine if this upgrade is a fork
	forkHeightEnv = "OSMOSIS_E2E_FORK_HEIGHT"
	// Environment variable name to skip cleaning up Docker resources in teardown
	skipCleanupEnv = "OSMOSIS_E2E_SKIP_CLEANUP"
	// Environment variable name to determine what version we are upgrading to
	upgradeVersionEnv = "OSMOSIS_E2E_UPGRADE_VERSION"
)

type IntegrationTestSuite struct {
	suite.Suite

	configurer  configurer.Configurer
	skipUpgrade bool
	skipIBC     bool
	forkHeight  int
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")
	var (
		err             error
		upgradeSettings configurer.UpgradeSettings
	)

	// The e2e test flow is as follows:
	//
	// 1. Configure two chains - chan A and chain B.
	//   * For each chain, set up two validators
	//   * Initialize configs and genesis for all validators.
	// 2. Start both networks.
	// 3. Run IBC relayer betweeen the two chains.
	// 4. Execute various e2e tests, including IBC.
	if str := os.Getenv(skipUpgradeEnv); len(str) > 0 {
		s.skipUpgrade, err = strconv.ParseBool(str)
		s.Require().NoError(err)
		s.T().Log(fmt.Sprintf("%s was true, skipping upgrade tests", skipIBCEnv))
	}
	upgradeSettings.IsEnabled = !s.skipUpgrade

	if str := os.Getenv(forkHeightEnv); len(str) > 0 {
		upgradeSettings.ForkHeight, err = strconv.ParseInt(str, 0, 64)
		s.Require().NoError(err)

		s.T().Log(fmt.Sprintf("fork upgrade is enabled, %s was set to height %d", forkHeightEnv, upgradeSettings.ForkHeight))
	}

	if str := os.Getenv(skipIBCEnv); len(str) > 0 {
		s.skipIBC, err = strconv.ParseBool(str)
		s.Require().NoError(err)
		s.T().Log(fmt.Sprintf("%s was true, skipping IBC tests", skipIBCEnv))
	}

	if str := os.Getenv(upgradeVersionEnv); len(str) > 0 {
		upgradeSettings.Version = str
		s.T().Log(fmt.Sprintf("upgrade version set to %s", upgradeSettings.Version))
	}

	s.configurer, err = configurer.New(s.T(), !s.skipIBC, upgradeSettings)
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
			s.T().Log("skipping e2e resources clean up...")
			return
		}
	}

	err := s.configurer.ClearResources()
	s.Require().NoError(err)
}
