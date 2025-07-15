package service_test

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/osmosis-labs/osmosis/v30/ingest/sqs/domain/mocks"
	sqsservice "github.com/osmosis-labs/osmosis/v30/ingest/sqs/service"
	"github.com/stretchr/testify/require"

	"google.golang.org/grpc/connectivity"
)

func newClientWithMockConn(initial connectivity.State, transitions []connectivity.State) (*sqsservice.GRPCClient, *mocks.ClientConn) {
	mock := &mocks.ClientConn{
		State:        initial,
		StateChanges: transitions,
	}
	client := &sqsservice.GRPCClient{}
	client.SetConn(mock)
	return client, mock
}

func newClientWithStateChan(initial connectivity.State, transitions []connectivity.State) (*sqsservice.GRPCClient, chan connectivity.State) {
	stateChan := make(chan connectivity.State)
	mock := &mocks.ClientConn{
		State:        initial,
		StateChanges: transitions,
	}
	client := &sqsservice.GRPCClient{}
	client.SetConn(mock)
	client.SetStateChan(stateChan)
	client.SetTimeAfterFunc(func(d time.Duration) <-chan time.Time {
		return time.After(time.Millisecond)
	})
	return client, stateChan
}

func shortTimeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 1*time.Millisecond)
}

func longTimeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 100*time.Millisecond)
}

func assertFinalState(t *testing.T, got, want connectivity.State) {
	t.Helper()
	if got != want {
		t.Errorf("expected final state %v, got %v", want, got)
	}
}

func TestConnect_StateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		initial     connectivity.State
		transitions []connectivity.State
		expected    connectivity.State
	}{
		{"FailureToReady", connectivity.TransientFailure, []connectivity.State{connectivity.Connecting, connectivity.Ready}, connectivity.Ready},
		{"IdleToReady", connectivity.Idle, []connectivity.State{connectivity.Connecting, connectivity.Ready}, connectivity.Ready},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := newClientWithMockConn(tt.initial, tt.transitions)
			client.SetTimeAfterFunc(func(d time.Duration) <-chan time.Time {
				return time.After(time.Millisecond)
			})
			ctx, cancel := shortTimeoutCtx()
			defer cancel()
			client.Connect(ctx)

			assertFinalState(t, mock.GetState(), tt.expected)
		})
	}
}

func TestConnect_HandlesContextDone(t *testing.T) {
	client, mock := newClientWithMockConn(connectivity.TransientFailure, []connectivity.State{
		connectivity.Connecting,
		connectivity.Ready,
	})
	client.SetConn(mock)

	ctx, cancel := shortTimeoutCtx()
	cancel()

	client.Connect(ctx)

	assertFinalState(t, mock.GetState(), connectivity.TransientFailure)
}

func TestConnect_StateTransitionMultipleReady(t *testing.T) {
	client, stateChanged := newClientWithStateChan(connectivity.Idle, []connectivity.State{
		connectivity.Connecting,
		connectivity.TransientFailure,
		connectivity.Ready,
	})

	// context with time out to allow for state changes
	ctx, cancel := longTimeoutCtx()
	defer cancel()

	go func() {
		client.Connect(ctx) // in background to allow for state changes
	}()

	var states []connectivity.State
loop:
	for {
		select {
		case state, ok := <-stateChanged:
			if !ok {
				break loop // exit the loop when the channel is closed
			}
			states = append(states, state)
		case <-time.After(1 * time.Second):
			t.Fatal("test timed out")
		}
	}
	states = slices.Compact(states) // Remove duplicates

	// Assert the states collected match the expected sequence
	require.Equal(t, states, []connectivity.State{
		connectivity.Idle,
		connectivity.Connecting,
		connectivity.TransientFailure,
		connectivity.Ready,
	}, "State transitions did not match the expected sequence")
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
