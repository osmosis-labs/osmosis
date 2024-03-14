package keeper

import (
	"context"
	"encoding/json"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	cmttypes "github.com/cometbft/cometbft/types"
	proto "github.com/cosmos/gogoproto/proto"

	bridge "github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type Signer struct {
	logger        log.Logger
	eventObserver Observer
	stopChan      chan struct{}
	eventsOutChan chan abcitypes.Event
}

func NewSigner(logger log.Logger, rpcUrl string) (Signer, error) {
	eventsOutChan := make(chan abcitypes.Event)
	obs, err := NewObserver(logger, rpcUrl, eventsOutChan)
	if err != nil {
		return Signer{}, err
	}

	return Signer{
		logger:        logger,
		eventObserver: obs,
		stopChan:      make(chan struct{}),
		eventsOutChan: eventsOutChan,
	}, nil
}

func (s *Signer) Start(ctx context.Context) error {
	query := cmttypes.QueryForEvent(cmttypes.EventNewBlock)
	events := []string{proto.MessageName(&bridge.EventOutboundTransfer{})}

	err := s.eventObserver.Start(ctx, query.String(), events)
	if err != nil {
		return err
	}

	go s.processEvents()

	return nil
}

func (s *Signer) Stop(ctx context.Context) error {
	close(s.stopChan)
	close(s.eventsOutChan)
	return s.eventObserver.Stop(ctx)
}

func (s *Signer) processEvents() {
	for {
		select {
		case <-s.stopChan:
			return
		case event := <-s.eventsOutChan:
			js, err := json.MarshalIndent(event, "", "  ")
			if err != nil {
				panic(err)
			}
			s.logger.Debug("Observed event", string(js))
		}
	}
}
