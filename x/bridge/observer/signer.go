package keeper

import (
	"fmt"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type Signer struct {
	eventObserver Observer
	stopChan      chan struct{}
}

func NewSigner(rpcUrl string) (Signer, error) {
	obs, err := NewObesrver(rpcUrl)
	if err != nil {
		return Signer{}, err
	}

	return Signer{
		eventObserver: obs,
		stopChan:      make(chan struct{}),
	}, nil
}

func (s *Signer) Start() error {
	query := fmt.Sprintf("tm.event = '%s'", proto.MessageName(&types.EventOutboundTransfer{}))
	err := s.eventObserver.Start(query)
	if err != nil {
		return err
	}

	go s.processEvents(s.eventObserver.GetEvents())

	return nil
}

func (s *Signer) Stop() error {
	close(s.stopChan)
	return s.eventObserver.Stop()
}

func (s *Signer) processEvents(ch <-chan coretypes.ResultEvent) {
	for {
		select {
		case <-s.stopChan:
			return
		case event := <-ch:
			if e, ok := event.Data.(types.EventOutboundTransfer); ok {
				fmt.Println("Got OutboundTransfer event: ", e)
			} else if e, ok := event.Data.(comettypes.EventDataNewBlock); ok {
				fmt.Println("Got NewBlock event: ", e)
			} else {
				fmt.Println("Got unknown event: ", event)
			}
		}
	}
}
