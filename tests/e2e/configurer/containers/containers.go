package containers

import (
	"fmt"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/configurer/chain"
)

type Manager struct {
	ImageConfig
	Pool    *dockertest.Pool
	Network *dockertest.Network

	hermesResource *dockertest.Resource
	ValResources   map[string][]*dockertest.Resource
}

func NewManager(isUpgradeEnabled bool) (docker *Manager, err error) {
	docker = &Manager{
		ImageConfig:  NewImageConfig(isUpgradeEnabled),
		ValResources: make(map[string][]*dockertest.Resource),
	}
	docker.Pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, err
	}
	docker.Network, err = docker.Pool.CreateNetwork("osmosis-testnet")
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
			NetworkID:  m.Network.Network.ID,
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
				fmt.Sprintf("OSMO_A_E2E_VAL_HOST=%s", m.ValResources[chainAID][0].Container.Name[1:]),
				fmt.Sprintf("OSMO_B_E2E_VAL_HOST=%s", m.ValResources[chainBID][0].Container.Name[1:]),
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

func (m *Manager) ClearResources() error {
	if err := m.Pool.Purge(m.hermesResource); err != nil {
		return err
	}

	for _, vr := range m.ValResources {
		for _, r := range vr {
			if err := m.Pool.Purge(r); err != nil {
				return err
			}
		}
	}

	if err := m.Pool.RemoveNetwork(m.Network); err != nil {
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
