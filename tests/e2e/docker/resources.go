package docker

import (
	"fmt"

	"github.com/ory/dockertest/v3"
)

type Resources struct {
	Pool       *dockertest.Pool
	Network    *dockertest.Network
	Hermes     *dockertest.Resource
	Validators map[string][]*dockertest.Resource
}

func NewResources() (*Resources, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}
	network, err := pool.CreateNetwork(fmt.Sprintf("osmosis-e2e-testnet"))
	if err != nil {
		return nil, err
	}

	return &Resources{
		Pool:       pool,
		Network:    network,
		Validators: make(map[string][]*dockertest.Resource),
	}, nil
}

func (r *Resources) Purge() error {
	if err := r.Pool.Purge(r.Hermes); err != nil {
		return err
	}

	for _, valResources := range r.Validators {
		for _, valResource := range valResources {
			if err := r.Pool.Purge(valResource); err != nil {
				return err
			}
		}
	}

	if err := r.Pool.RemoveNetwork(r.Network); err != nil {
		return err
	}

	return nil
}
