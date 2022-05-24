package docker

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
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

func (r *Resources) ExecValidator(chainId string, validatorIndex int, command []string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	containerId := r.Validators[chainId][validatorIndex].Container.ID

	exec, err := r.Pool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    containerId,
		User:         "root",
		Cmd:          command,
	})
	if err != nil {
		return nil, err
	}

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	err = r.Pool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute: %v\nstdout: %s\nstderr: %s\nerr: %w", command, outBuf.String(), errBuf.String(), err)
	}
	return errBuf.Bytes(), nil
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
