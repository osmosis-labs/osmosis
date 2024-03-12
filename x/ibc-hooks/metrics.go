package ibc_hooks

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	packetsRoutedToContracts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "osmosis_ibc_hooks_packets_routed_to_contracts_total",
			Help: "Total number of IBC packets routed to contracts",
		},
		[]string{"channel_id", "port_id", "contract_address"},
	)

	packetsRoutedToContractsFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "osmosis_ibc_hooks_packets_routed_to_contracts_failures_total",
			Help: "Total number of IBC packets that failed to route to contracts",
		},
		[]string{"channel_id", "port_id", "contract_address"},
	)
)
