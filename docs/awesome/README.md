# awesome
List of public resources, documents, and tools for Osmosis.

### Reading lists
- [How to permissionlessly list tokens to Osmosis](./guides/token-listing.md)

### Block Explorers
- [Mintscan](https://mintscan.io/osmosis)
- [Big Dipper](https://osmosis.bigdipper.live)
- [Ping.pub](https://ping.pub/osmosis)

### Analytics, Data, and Visualization
- [Imperator Osmosis Dashboard](https://osmosis.imperator.co/)
- [Map of Zones](https://mapofzones.com)
- [Smartstake staking dashboard](https://osmosis.smartstake.io/)
- [Cosmissed by Blockpane](https://github.com/blockpane/cosmissed)

### Relayer infrastructure providers
- [Cephalopod Equipment](https://cephalopod.equipment/)
- [Vitwit](https://www.vitwit.com/)
- [Notional](https://github.com/faddat/notional)

### Validator snapshots & syncing
- [3Tekos Validator](https://3tekos.fr/#Archives)
- [Stake Systems](https://www.notion.so/Stake-Systems-LCD-RPC-gRPC-Instances-04a99a9a9aa14247a42944931eec7024) - Snapshots taken everyday at 06:00 UTC / 18:00 UTC

### Publicly available endpoints
- [DataHub](https://datahub.figment.io)
- [Stake Systems](https://www.notion.so/Stake-Systems-LCD-RPC-gRPC-Instances-04a99a9a9aa14247a42944931eec7024) - LCD(REST)/RPC/gRPC
- [Skynet Validator](http://202.61.192.186:26657/status) - RPC only

### Archive Nodes
- Notional's bus bar
  - 28f61c154c82f0122a841a12f8aa87703bd6ae1e@162.55.132.230:2000

You'd want to add the bus bar as a persistent peer. It can accept a vast number of connections, and can be used in a manner analagous to a seed node since it will do PEX with you.  It is not a seed node. 

- Seed nodes
  - `2308bed9e096a8b96d2aa343acc1147813c59ed2@3.225.38.25:26656`
  - `1b077d96ceeba7ef503fb048f343a538b2dcdf1b@136.243.218.244:26656` (Provided by Smartnodes)
  - `902bdfe51b6a97cc9369664a21c87ed61d471d2a@136.243.218.243:26656` (Provided by Smartnodes)
  - `f515a8599b40f0e84dfad935ba414674ab11a668@osmosis.blockpane.com:26656` (Provided by BlockPane)


### Documentation / Guides
- [Edge Validation](https://whimsical.com/validatron-PbUypC8tVMU8DxCFNLdDFu)
- [Relaying](https://github.com/faddat/notional)
- [Listing IBC tokens to Osmosis](./guides/token-listing.md)
- [Setting up a full node for Osmosis-1](https://catboss.medium.com/cat-boss-setting-up-a-fullnode-for-osmosis-osmosis-1-5f9752460f8f) / [Turning a full node into a validator node](https://catboss.medium.com/turning-a-full-node-in-to-a-validator-node-osmosis-1-36f3358f2412)

### Decentralization and resilience tools
- [Tradeberry](https://github.com/faddat/tradeberry) - Stateless rpi image that state syncs osmosis and allows access to the UI in a private, sovereign manner.
