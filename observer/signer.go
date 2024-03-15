package observer

import (
	"context"
	"encoding/json"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	cmttypes "github.com/cometbft/cometbft/types"
	proto "github.com/cosmos/gogoproto/proto"

	bridge "github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

const ModuleNameSigner = "signer"

type Signer struct {
	logger        log.Logger
	eventObserver Observer
	stopChan      chan struct{}
}

func NewSigner(logger log.Logger, rpcUrl string) (Signer, error) {
	obs, err := NewObserver(logger, rpcUrl)
	if err != nil {
		return Signer{}, err
	}

	return Signer{
		logger:        logger.With("module", ModuleNameSigner),
		eventObserver: obs,
		stopChan:      make(chan struct{}),
	}, nil
}

func (s *Signer) Start(ctx context.Context) error {
	query := cmttypes.QueryForEvent(cmttypes.EventNewBlock)
	events := []string{proto.MessageName(&bridge.EventOutboundTransfer{})}

	err := s.eventObserver.Start(ctx, query.String(), events)
	if err != nil {
		return err
	}

	go s.processEvents(s.eventObserver.Events())

	return nil
}

func (s *Signer) Stop(ctx context.Context) error {
	return s.eventObserver.Stop(ctx)
}

func (s *Signer) processEvents(events <-chan abcitypes.Event) {
	for event := range events {
		// Start TSS process here
		js, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			panic(err)
		}
		s.logger.Debug("Observed event", string(js))
	}
}
