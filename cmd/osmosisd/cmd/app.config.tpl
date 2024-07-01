###############################################################################
###                      Osmosis Mempool Configuration                      ###
###############################################################################

[osmosis-mempool]
# This is the max allowed gas any tx.
# This is only for local mempool purposes, and thus	is only ran on check tx.
max-gas-wanted-per-tx = "{{ .OsmosisMempoolConfig.MaxGasWantedPerTx }}"

# This is the minimum gas fee any arbitrage tx should have, denominated in uosmo per gas
# Default value of ".1" then means that a tx with 1 million gas costs (.1 uosmo/gas) * 1_000_000 gas = .1 osmo
arbitrage-min-gas-fee = "{{ .OsmosisMempoolConfig.MinGasPriceForArbitrageTx }}"

# This is the minimum gas fee any tx with high gas demand should have, denominated in uosmo per gas
# Default value of ".0025" then means that a tx with 1 million gas costs (.0025 uosmo/gas) * 1_000_000 gas = .0025 osmo
min-gas-price-for-high-gas-tx = "{{ .OsmosisMempoolConfig.MinGasPriceForHighGasTx }}"

# This parameter enables EIP-1559 like fee market logic in the mempool
adaptive-fee-enabled = "{{ .OsmosisMempoolConfig.Mempool1559Enabled }}"

###############################################################################
###              Osmosis Sidecar Query Server Configuration                 ###
###############################################################################

[osmosis-sqs]

# SQS service is disabled by default.
is-enabled = "{{ .SidecarQueryServerConfig.IsEnabled }}"

# The hostname of the GRPC sqs service
grpc-ingest-address = "{{ .SidecarQueryServerConfig.GRPCIngestAddress }}"
# The maximum size of the GRPC message that can be received by the sqs service in bytes.
grpc-ingest-max-call-size-bytes = "{{ .SidecarQueryServerConfig.GRPCIngestMaxCallSizeBytes }}"

###############################################################################
###              Osmosis Indexer Configuration                              ###
###############################################################################
[osmosis-indexer]

# The indexer service is disabled by default.
is-enabled = "{{ .IndexerConfig.IsEnabled }}"

# The GCP project id to use for the indexer service.
gcp-project-id = "{{ .IndexerConfig.GCPProjectId }}"

# The topic id to use for the publishing block data
block-topic-id = "{{ .IndexerConfig.BlockTopicId }}"

# The topic id to use for the publishing transaction data
transaction-topic-id = "{{ .IndexerConfig.TransactionTopicId }}"

# The topic id to use for the publishing pool data
pool-topic-id = "{{ .IndexerConfig.PoolTopicId }}"

# The topic id to use for the publishing token supply data
token-supply-topic-id = "{{ .IndexerConfig.TokenSupplyTopicId }}"

# The topic id to use for the publishing token supply offset data
token-supply-offset-topic-id = "{{ .IndexerConfig.TokenSupplyOffsetTopicId }}"

###############################################################################
###                            Wasm Configuration                           ###
###############################################################################