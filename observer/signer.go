package observer

import (
	"context"
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	cmttypes "github.com/cometbft/cometbft/types"
	proto "github.com/cosmos/gogoproto/proto"

	bridge "github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

const ModuleNameSigner = "signer"

type Signer struct {
	logger   log.Logger
	observer Observer
}

// NewSigner returns new instance of `Signer` with `Observer` created
func NewSigner(logger log.Logger, rpcUrl string) (Signer, error) {
	obs, err := NewObserver(logger, rpcUrl)
	if err != nil {
		return Signer{}, errorsmod.Wrapf(err, "Failed to create Observer")
	}

	return Signer{
		logger:   logger.With("module", ModuleNameSigner),
		observer: obs,
	}, nil
}

// Start starts the observer listening to `NewBlock` events
// and filtering `EventOutboundTransfer` events
func (s *Signer) Start(ctx context.Context) error {
	query := cmttypes.QueryForEvent(cmttypes.EventNewBlock)
	events := []string{proto.MessageName(&bridge.EventOutboundTransfer{})}

	err := s.observer.Start(ctx, query.String(), events)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to start observer")
	}

	go s.processEvents()

	return nil
}

// Stop stops the observer
func (s *Signer) Stop(ctx context.Context) error {
	return s.observer.Stop(ctx)
}

func (s *Signer) processEvents() {
	for event := range s.observer.Events() {
		// Start TSS process here
		js, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			s.logger.Error("Failed to marshal event")
			continue
		}
		s.logger.Debug("Observed event", string(js))
	}
}
