package containers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

// Manager is a wrapper around Docker API. It provides utilities
// necessary to run and manage Docker containers for e2e testing.
type Manager struct {
	ImageConfig
	pool    *dockertest.Pool
	network *dockertest.Network

	hermesResource *dockertest.Resource
	valResources   map[string][]*dockertest.Resource
}

// NewManager creates a new Manager instance and initializes
// all Docker specific utilies. Returns an error if initialiation fails.
func NewManager(isUpgradeEnabled bool) (docker *Manager, err error) {
	docker = &Manager{
		ImageConfig:  NewImageConfig(isUpgradeEnabled),
		valResources: make(map[string][]*dockertest.Resource),
	}
	docker.pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, err
	}
	docker.network, err = docker.pool.CreateNetwork("osmosis-testnet")
	if err != nil {
		return nil, err
	}
	return docker, nil
}

// ExecCmd executes command on chainId by running it on the validator container (specified by validatorIndex)
// success is the output of the command that needs to be observed for the command to be deemed successful.
// returns container std out, container std err, and error if any.
// An error is returned if the command fails to execute or if the success string is not found in the output.
func (m *Manager) ExecCmd(t *testing.T, chainId string, validatorIndex int, command []string, success string) (bytes.Buffer, bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var containerId string
	if chainId == "" {
		containerId = m.hermesResource.Container.ID
	} else {
		containerId = m.valResources[chainId][validatorIndex].Container.ID
	}

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	// We use the `require.Eventually` function because it is only allowed to do one transaction per block without
	// sequence numbers. For simplicity, we avoid keeping track of the sequence number and just use the `require.Eventually`.
	require.Eventually(
		t,
		func() bool {
			exec, err := m.pool.Client.CreateExec(docker.CreateExecOptions{
				Context:      ctx,
				AttachStdout: true,
				AttachStderr: true,
				Container:    containerId,
				User:         "root",
				Cmd:          command,
			})
			require.NoError(t, err)

			err = m.pool.Client.StartExec(exec.ID, docker.StartExecOptions{
				Context:      ctx,
				Detach:       false,
				OutputStream: &outBuf,
				ErrorStream:  &errBuf,
			})
			if err != nil {
				return false
			}

			if success != "" {
				return strings.Contains(outBuf.String(), success) || strings.Contains(errBuf.String(), success)
			}

			return true
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code; stdout: %s, stderr: %s", outBuf.String(), errBuf.String(),
	)

	return outBuf, errBuf, nil
}

// RunHermesResource runs a Hermes container. Returns the container resource and error if any.
func (m *Manager) RunHermesResource(chainAID, osmoAValMnemonic, chainBID, osmoBValMnemonic string, hermesCfgPath string) (*dockertest.Resource, error) {
	var err error
	m.hermesResource, err = m.pool.RunWithOptions(
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

// GetHermesContainerId returns the Hermes container ID.
func (m *Manager) GetHermesContainerID() string {
	return m.hermesResource.Container.ID
}

// RunValidatorResource runs a validator container. Assings valContainerName to the container.
// Mounts the container on valConfigDir volume on the running host. Returns the container resource and error if any.
func (m *Manager) RunValidatorResource(chainId string, valContainerName, valCondifDir string) (*dockertest.Resource, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	runOpts := &dockertest.RunOptions{
		Name:       valContainerName,
		Repository: m.OsmosisRepository,
		Tag:        m.OsmosisTag,
		NetworkID:  m.network.Network.ID,
		User:       "root:root",
		Mounts: []string{
			fmt.Sprintf("%s/:/osmosis/.osmosisd", valCondifDir),
			fmt.Sprintf("%s/scripts:/osmosis", pwd),
		},
	}

	resource, err := m.pool.RunWithOptions(runOpts, noRestart)
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

// RunChainInitResource runs a chain init container to initialize genesis and configs for a chain with chainId.
// The chain is to be configured with chainVotingPeriod and validators deserialized from validatorConfigBytes.
// The genesis and configs are to be mounted on the init container as volume on mountDir path.
// Returns the container resource and error if any. This method does not Purge the container. The caller
// must deal with removing the resource.
func (m *Manager) RunChainInitResource(chainId string, chainVotingPeriod int, validatorConfigBytes []byte, mountDir string) (*dockertest.Resource, error) {
	votingPeriodDuration := time.Duration(chainVotingPeriod * 1000000000)

	initResource, err := m.pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s", chainId),
			Repository: m.ImageConfig.InitRepository,
			Tag:        m.ImageConfig.InitTag,
			NetworkID:  m.network.Network.ID,
			Cmd: []string{
				fmt.Sprintf("--data-dir=%s", mountDir),
				fmt.Sprintf("--chain-id=%s", chainId),
				fmt.Sprintf("--config=%s", validatorConfigBytes),
				fmt.Sprintf("--voting-period=%v", votingPeriodDuration),
			},
			User: "root:root",
			Mounts: []string{
				fmt.Sprintf("%s:%s", mountDir, mountDir),
			},
		},
		noRestart,
	)
	if err != nil {
		return nil, err
	}
	return initResource, nil
}

// PurgeResource purges the container resource and returns an error if any.
func (m *Manager) PurgeResource(resource *dockertest.Resource) error {
	return m.pool.Purge(resource)
}

// GetValidatorResource returns the validator resource at validatorIndex for the given chainId.
func (m *Manager) GetValidatorResource(chainId string, validatorIndex int) (*dockertest.Resource, bool) {
	chainValidators, exists := m.valResources[chainId]
	if !exists || validatorIndex >= len(chainValidators) {
		return nil, false
	}
	return chainValidators[validatorIndex], true
}

// GetValidatorHostPort returns the port-forwarding address of the running host
// necessary to connect to the validator's portId exposed inside the container.
// The validator container is determined by chainId and validatorIndex.
// Returns the host-port or error if any.
func (m *Manager) GetValidatorHostPort(chainId string, validatorIndex int, portId string) (string, error) {
	validatorResource, exists := m.GetValidatorResource(chainId, validatorIndex)
	if !exists {
		return "", fmt.Errorf("validator resource not found: chainId: %s, validatorIndex: %d", chainId, validatorIndex)
	}
	return validatorResource.GetHostPort(portId), nil
}

// RemoveValidatorResource removes a validator container specified by chainId and containerName.
// Returns error if any.
func (m *Manager) RemoveValidatorResource(chainId string, containerName string) error {
	chainValidators, exists := m.valResources[chainId]
	if !exists {
		return fmt.Errorf("no validators on chain %s", chainId)
	}

	for validatorIndex, validator := range chainValidators {
		if validator.Container.Name[1:] == containerName {
			var opts docker.RemoveContainerOptions
			opts.ID = validator.Container.ID
			opts.Force = true
			if err := m.pool.Client.RemoveContainer(opts); err != nil {
				return err
			}
			m.valResources[chainId] = append(chainValidators[:validatorIndex], chainValidators[validatorIndex+1:]...)
			return nil
		}
	}

	return fmt.Errorf("no validator container %s on chain %s", containerName, chainId)
}

// ClearResources removes all outstanding Docker resources created by the Manager.
func (m *Manager) ClearResources() error {
	if err := m.pool.Purge(m.hermesResource); err != nil {
		return err
	}

	for _, vr := range m.valResources {
		for _, r := range vr {
			if err := m.pool.Purge(r); err != nil {
				return err
			}
		}
	}

	if err := m.pool.RemoveNetwork(m.network); err != nil {
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
