package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/v29/ingest/sqs/domain/mocks"
	sqsservice "github.com/osmosis-labs/osmosis/v29/ingest/sqs/service"

	"google.golang.org/grpc/connectivity"
)

func TestConnect_HandlesContextDone(t *testing.T) {
	mock := &mocks.ClientConn{
		State: connectivity.TransientFailure,
		StateChanges: []connectivity.State{
			connectivity.Connecting,
			connectivity.Ready,
		},
	}

	client := &sqsservice.GRPCClient{}
	client.SetConn(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	cancel()

	client.Connect(ctx)

	if mock.GetState() != connectivity.TransientFailure {
		t.Errorf("expected final state to be TransientFailure, got %v", mock.GetState())
	}
}

func TestConnect_StateTransitionFailureToReady(t *testing.T) {
	mock := &mocks.ClientConn{
		State: connectivity.TransientFailure,
		StateChanges: []connectivity.State{
			connectivity.Connecting,
			connectivity.Ready,
		},
	}

	client := &sqsservice.GRPCClient{}
	client.SetConn(mock)
	client.SetTimeAfterFunc(func(d time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	})

	// Simulate connection attempts for 10 milliseconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	client.Connect(ctx)

	// Final state should be Ready
	if mock.GetState() != connectivity.Ready {
		t.Errorf("expected final state to be Ready, got %v", mock.GetState())
	}
}

func TestConnect_StateTransitionMultipleReady(t *testing.T) {
	mock := &mocks.ClientConn{
		State: connectivity.Ready,
		StateChanges: []connectivity.State{
			connectivity.Connecting,
			connectivity.TransientFailure,
			connectivity.Ready,
		},
	}

	client := &sqsservice.GRPCClient{}
	client.SetConn(mock)
	client.SetTimeAfterFunc(func(d time.Duration) <-chan time.Time {
		ch := make(chan time.Time, 1)
		ch <- time.Now()
		return ch
	})

	// Simulate connection attempts for 10 milliseconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// Run in the background to allow for state changes
	go func() {
		defer wg.Done()
		client.Connect(ctx)
	}()

	time.Sleep(5 * time.Millisecond)
	mock.State = connectivity.TransientFailure

	wg.Wait()

	if mock.GetState() != connectivity.Ready {
		t.Errorf("expected final state to be Ready, got %v", mock.GetState())
	}
}

func TestConnect_TimeAfterCall(t *testing.T) {
	client := &sqsservice.GRPCClient{}
	client.SetConn(&mocks.ClientConn{
		State: connectivity.Ready,
	})

	ch := make(chan time.Time, 1)
	client.SetTimeAfterFunc(func(d time.Duration) <-chan time.Time {
		return ch
	})

	// Simulate persistent connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	var wg sync.WaitGroup
	wg.Add(1)

	// Run in the background to allow for state changes
	go func() {
		defer wg.Done()
		client.Connect(ctx)
	}()

	// Simulate time.After call
	ch <- time.Now()
	cancel()
	wg.Wait()
}
