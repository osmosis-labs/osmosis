package containers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

const hermesContainerName = "hermes-relayer"

var errRegex = regexp.MustCompile(`(E|e)rror`)

// Manager is a wrapper around all Docker instances, and the Docker API.
// It provides utilities to run and interact with all Docker containers used within e2e testing.
type Manager struct {
	ImageConfig
	pool      *dockertest.Pool
	network   *dockertest.Network
	resources map[string]*dockertest.Resource
}

// NewManager creates a new Manager instance and initializes
// all Docker specific utilies. Returns an error if initialiation fails.
func NewManager(isUpgrade bool, isFork bool) (docker *Manager, err error) {
	docker = &Manager{
		ImageConfig: NewImageConfig(isUpgrade, isFork),
		resources:   make(map[string]*dockertest.Resource),
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

// ExecTxCmd Runs ExecCmd, with flags for txs added.
// namely adding flags `--chain-id={chain-id} -b=block --yes --keyring-backend=test "--log_format=json"`,
// and searching for `code: 0`
func (m *Manager) ExecTxCmd(t *testing.T, chainId string, containerName string, command []string) (bytes.Buffer, bytes.Buffer, error) {
	allTxArgs := []string{fmt.Sprintf("--chain-id=%s", chainId), "-b=block", "--yes", "--keyring-backend=test", "--log_format=json"}
	txCommand := append(command, allTxArgs...)
	successStr := "code: 0"
	return m.ExecCmd(t, containerName, txCommand, successStr)
}

// ExecHermesCmd executes command on the hermes relaer container.
func (m *Manager) ExecHermesCmd(t *testing.T, command []string, success string) (bytes.Buffer, bytes.Buffer, error) {
	return m.ExecCmd(t, hermesContainerName, command, success)
}

// ExecCmd executes command by running it on the node container (specified by containerName)
// success is the output of the command that needs to be observed for the command to be deemed successful.
// It is found by checking if stdout or stderr contains the success string anywhere within it.
// returns container std out, container std err, and error if any.
// An error is returned if the command fails to execute or if the success string is not found in the output.
func (m *Manager) ExecCmd(t *testing.T, containerName string, command []string, success string) (bytes.Buffer, bytes.Buffer, error) {
	if _, ok := m.resources[containerName]; !ok {
		return bytes.Buffer{}, bytes.Buffer{}, fmt.Errorf("no resource %s found", containerName)
	}
	containerId := m.resources[containerName].Container.ID

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
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

			errBufString := errBuf.String()
			// Note that this does not match all errors.
			// This only works if CLI outpurs "Error" or "error"
			// to stderr.
			if errRegex.MatchString(errBufString) {
				t.Log("Potential error in stderr:")
				t.Log(errBufString)
				// N.B: We should not be returning false here
				// because some applications such as Hermes might log
				// "error" to stderr when they function correctly,
				// causing test flakiness. This log is needed only for
				// debugging purposes.
			}

			if success != "" {
				return strings.Contains(outBuf.String(), success) || strings.Contains(errBufString, success)
			}

			return true
		},
		time.Minute,
		time.Second,
		"tx returned a non-zero code",
	)

	return outBuf, errBuf, nil
}

// RunHermesResource runs a Hermes container. Returns the container resource and error if any.
// the name of the hermes container is "<chain A id>-<chain B id>-relayer"
func (m *Manager) RunHermesResource(chainAID, osmoARelayerNodeName, osmoAValMnemonic, chainBID, osmoBRelayerNodeName, osmoBValMnemonic string, hermesCfgPath string) (*dockertest.Resource, error) {
	hermesResource, err := m.pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       hermesContainerName,
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
				fmt.Sprintf("OSMO_A_E2E_VAL_HOST=%s", osmoARelayerNodeName),
				fmt.Sprintf("OSMO_B_E2E_VAL_HOST=%s", osmoBRelayerNodeName),
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
	m.resources[hermesContainerName] = hermesResource
	return hermesResource, nil
}

// RunNodeResource runs a node container. Assings containerName to the container.
// Mounts the container on valConfigDir volume on the running host. Returns the container resource and error if any.
func (m *Manager) RunNodeResource(chainId string, containerName, valCondifDir string) (*dockertest.Resource, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	runOpts := &dockertest.RunOptions{
		Name:       containerName,
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

	m.resources[containerName] = resource

	return resource, nil
}

// RunChainInitResource runs a chain init container to initialize genesis and configs for a chain with chainId.
// The chain is to be configured with chainVotingPeriod and validators deserialized from validatorConfigBytes.
// The genesis and configs are to be mounted on the init container as volume on mountDir path.
// Returns the container resource and error if any. This method does not Purge the container. The caller
// must deal with removing the resource.
func (m *Manager) RunChainInitResource(chainId string, chainVotingPeriod int, validatorConfigBytes []byte, mountDir string, forkHeight int) (*dockertest.Resource, error) {
	votingPeriodDuration := time.Duration(chainVotingPeriod * 1000000000)

	initResource, err := m.pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       chainId,
			Repository: m.ImageConfig.InitRepository,
			Tag:        m.ImageConfig.InitTag,
			NetworkID:  m.network.Network.ID,
			Cmd: []string{
				fmt.Sprintf("--data-dir=%s", mountDir),
				fmt.Sprintf("--chain-id=%s", chainId),
				fmt.Sprintf("--config=%s", validatorConfigBytes),
				fmt.Sprintf("--voting-period=%v", votingPeriodDuration),
				fmt.Sprintf("--fork-height=%v", forkHeight),
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

// GetNodeResource returns the node resource for containerName.
func (m *Manager) GetNodeResource(containerName string) (*dockertest.Resource, error) {
	resource, exists := m.resources[containerName]
	if !exists {
		return nil, fmt.Errorf("node resource not found: container name: %s", containerName)
	}
	return resource, nil
}

// GetHostPort returns the port-forwarding address of the running host
// necessary to connect to the portId exposed inside the container.
// The container is determined by containerName.
// Returns the host-port or error if any.
func (m *Manager) GetHostPort(containerName string, portId string) (string, error) {
	resource, err := m.GetNodeResource(containerName)
	if err != nil {
		return "", err
	}
	return resource.GetHostPort(portId), nil
}

// RemoveNodeResource removes a node container specified by containerName.
// Returns error if any.
func (m *Manager) RemoveNodeResource(containerName string) error {
	resource, err := m.GetNodeResource(containerName)
	if err != nil {
		return err
	}
	var opts docker.RemoveContainerOptions
	opts.ID = resource.Container.ID
	opts.Force = true
	if err := m.pool.Client.RemoveContainer(opts); err != nil {
		return err
	}
	delete(m.resources, containerName)
	return nil
}

// ClearResources removes all outstanding Docker resources created by the Manager.
func (m *Manager) ClearResources() error {
	for _, resource := range m.resources {
		if err := m.pool.Purge(resource); err != nil {
			return err
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
