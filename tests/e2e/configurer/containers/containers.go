package containers

import (
	"github.com/ory/dockertest/v3"
)

type Manager struct {
	ImageConfig
	Pool    *dockertest.Pool
	Network *dockertest.Network

	HermesResource *dockertest.Resource
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
