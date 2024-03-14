package keeper

import (
	"encoding/json"
	"fmt"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	proto "github.com/cosmos/gogoproto/proto"

	bridge "github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

type Signer struct {
	eventObserver Observer
	stopChan      chan struct{}
	eventsOutChan chan abcitypes.Event
}

func NewSigner(rpcUrl string) (Signer, error) {
	eventsOutChan := make(chan abcitypes.Event)
	obs, err := NewObserver(rpcUrl, eventsOutChan)
	if err != nil {
		return Signer{}, err
	}

	return Signer{
		eventObserver: obs,
		stopChan:      make(chan struct{}),
		eventsOutChan: eventsOutChan,
	}, nil
}

func (s *Signer) Start() error {
	query := cmttypes.QueryForEvent(cmttypes.EventNewBlock)
	events := []string{proto.MessageName(&bridge.EventOutboundTransfer{})}

	err := s.eventObserver.Start(query.String(), events)
	if err != nil {
		return err
	}

	go s.processEvents()

	return nil
}

func (s *Signer) Stop() error {
	close(s.stopChan)
	return s.eventObserver.Stop()
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
			fmt.Println(string(js))
		}
	}
}
