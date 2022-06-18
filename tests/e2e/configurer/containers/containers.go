package containers

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	chaininit "github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/chain"
)

type Manager struct {
	ImageConfig
	Pool    *dockertest.Pool
	network *dockertest.Network

	hermesResource *dockertest.Resource
	valResources   map[string][]*dockertest.Resource
}

func NewManager(isUpgradeEnabled bool) (docker *Manager, err error) {
	docker = &Manager{
		ImageConfig:  NewImageConfig(isUpgradeEnabled),
		valResources: make(map[string][]*dockertest.Resource),
	}
	docker.Pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, err
	}
	docker.network, err = docker.Pool.CreateNetwork("osmosis-testnet")
	if err != nil {
		return nil, err
	}
	return docker, nil
}

func (m *Manager) RunHermesResource(chainConfigA, chainConfigB *chain.Config, hermesCfgPath string) (*dockertest.Resource, error) {
	chainAID := chainConfigA.ChainId
	chainBID := chainConfigB.ChainId

	osmoAValMnemonic := chainConfigA.Chain.Validators[0].Mnemonic
	osmoBValMnemonic := chainConfigB.Chain.Validators[0].Mnemonic

	var err error
	m.hermesResource, err = m.Pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s-%s-relayer", chainAID, chainBID),
			Repository: m.RelayerRepository,
			Tag:        m.RelayerTag,
			NetworkID:  m.network.Network.ID,
			Cmd: []string{
				"start",
			},
			User: "root:root",
			Mounts: []string{
				fmt.Sprintf("%s/:/root/hermes", hermesCfgPath),
			},
			ExposedPorts: []string{
				"3031",
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("OSMO_A_E2E_CHAIN_ID=%s", chainAID),
				fmt.Sprintf("OSMO_B_E2E_CHAIN_ID=%s", chainBID),
				fmt.Sprintf("OSMO_A_E2E_VAL_MNEMONIC=%s", osmoAValMnemonic),
				fmt.Sprintf("OSMO_B_E2E_VAL_MNEMONIC=%s", osmoBValMnemonic),
				fmt.Sprintf("OSMO_A_E2E_VAL_HOST=%s", m.valResources[chainAID][0].Container.Name[1:]),
				fmt.Sprintf("OSMO_B_E2E_VAL_HOST=%s", m.valResources[chainBID][0].Container.Name[1:]),
			},
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/hermes/hermes_bootstrap.sh && /root/hermes/hermes_bootstrap.sh",
			},
		},
		noRestart,
	)
	if err != nil {
		return nil, err
	}
	return m.hermesResource, nil
}

func (m *Manager) GetHermesContainerID() string {
	return m.hermesResource.Container.ID
}

func (m *Manager) RunValidatorResource(chainId string, val *chaininit.Validator) (*dockertest.Resource, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	runOpts := &dockertest.RunOptions{
		Name:      val.Name,
		NetworkID: m.network.Network.ID,
		Mounts: []string{
			fmt.Sprintf("%s/:/osmosis/.osmosisd", val.ConfigDir),
			fmt.Sprintf("%s/scripts:/osmosis", pwd),
		},
		Repository: m.OsmosisRepository,
		Tag:        m.OsmosisTag,
		Cmd: []string{
			"start",
		},
	}

	resource, err := m.Pool.RunWithOptions(runOpts, noRestart)
	if err != nil {
		return nil, err
	}

	chainValidatorResources, exists := m.valResources[chainId]
	if !exists {
		chainValidatorResources = make([]*dockertest.Resource, 0)
	}
	m.valResources[chainId] = append(chainValidatorResources, resource)

	return resource, nil
}

func (m *Manager) RunChainInitResource(chainConfig *chain.Config, tmpDir string) error {
	validatorConfigBytes, err := json.Marshal(chainConfig.ValidatorConfig)
	if err != nil {
		return err
	}

	votingPeriodDuration := time.Duration(int(chainConfig.VotingPeriod) * 1000000000)

	initResource, err := m.Pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s", chainConfig.ChainId),
			Repository: m.ImageConfig.InitRepository,
			Tag:        m.ImageConfig.InitTag,
			NetworkID:  m.network.Network.ID,
			Cmd: []string{
				fmt.Sprintf("--data-dir=%s", tmpDir),
				fmt.Sprintf("--chain-id=%s", chainConfig.ChainId),
				fmt.Sprintf("--config=%s", validatorConfigBytes),
				fmt.Sprintf("--voting-period=%v", votingPeriodDuration),
			},
			User: "root:root",
			Mounts: []string{
				fmt.Sprintf("%s:%s", tmpDir, tmpDir),
			},
		},
		noRestart,
	)

	if err := m.Pool.Purge(initResource); err != nil {
		return err
	}

	return nil
}

func (m *Manager) GetValidatorResource(chainId string, validatorIndex int) (*dockertest.Resource, bool) {
	chainValidators, exists := m.valResources[chainId]
	if !exists || validatorIndex >= len(chainValidators) {
		return nil, false
	}
	return chainValidators[validatorIndex], true
}

func (m *Manager) RemoveValidatorResource(chainId string, validatorIndex int) (string, error) {
	chainValidators, exists := m.valResources[chainId]
	if !exists || validatorIndex >= len(chainValidators) {
		return "", fmt.Errorf("validator %d on chain %s does not exist", validatorIndex, chainId)
	}

	validatorResource := m.valResources[chainId][validatorIndex]
	containerName := validatorResource.Container.Name

	var opts docker.RemoveContainerOptions
	opts.ID = validatorResource.Container.ID
	opts.Force = true
	if err := m.Pool.Client.RemoveContainer(opts); err != nil {
		return "", err
	}
	m.valResources[chainId] = append(chainValidators[:validatorIndex], chainValidators[validatorIndex+1:]...)
	return containerName, nil
}

func (m *Manager) ClearResources() error {
	if err := m.Pool.Purge(m.hermesResource); err != nil {
		return err
	}

	for _, vr := range m.valResources {
		for _, r := range vr {
			if err := m.Pool.Purge(r); err != nil {
				return err
			}
		}
	}

	if err := m.Pool.RemoveNetwork(m.network); err != nil {
		return err
	}
	return nil
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}
