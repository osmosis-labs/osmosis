
#!/bin/bash
# Based on the work of Joe (Chorus-One) for Microtick - https://github.com/microtick/bounties/tree/main/statesync
# You need config in two peers (avoid seed servers) this values in app.toml:
#     [state-sync]
#     snapshot-interval = 1000
#     snapshot-keep-recent = 10
# Pruning should be fine tuned also, for this testings is set to nothing
#     pruning = "nothing"

set -e

# Change for your custom chain
BINARY="https://github.com/osmosis-labs/osmosis/releases/download/v3.1.0/osmosisd-3.1.0-linux-amd64"
GENESIS="https://cloudflare-ipfs.com/ipfs/QmXRvBT3hgoXwwPqbK6a2sXUuArGM8wPyo1ybskyyUwUxs"
APP="OSMOSISD: ~/.osmosisd"


  # Osmosis State Sync client config.
  # rm -f osmosisd #deletes a previous downloaded binary
  # rm -rf $HOME/.osmosisd/ #deletes previous installation   
wget -nc $BINARY
mv osmosisd-3.1.0-linux-amd64 osmosisd
chmod +x osmosisd
./osmosisd init test 
wget -O $HOME/.osmosisd/config/genesis.json $GENESIS 
  
NODE1_IP="95.217.196.54"
RPC1="http://$NODE1_IP"
P2P_PORT1=2000
RPC_PORT1=2001

NODE2_IP="162.55.132.230"
RPC2="http://$NODE2_IP"
P2P_PORT2=2000
RPC_PORT2=2001

  #If you want to use a third StateSync Server... 
  #DOMAIN_3=seed1.bitcanna.io     # If you want to use domain names 
  #NODE3_IP=$(dig $DOMAIN_1 +short
  #RPC3="http://$NODE3_IP"
  #RPC_PORT3=26657
  #P2P_PORT3=26656

INTERVAL=1000

LATEST_HEIGHT=$(curl -s $RPC1:$RPC_PORT1/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL)) 
  


NODE1_ID=$(curl -s "$RPC1:$RPC_PORT1/status" | jq -r .result.node_info.id)
NODE2_ID=$(curl -s "$RPC2:$RPC_PORT2/status" | jq -r .result.node_info.id)
#NODE3_ID=$(curl -s "$RPC3:$RPC_PORT3/status" | jq -r .result.node_info.id)

echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"
echo "NODE ONE: $NODE1_ID@$NODE1_IP:$P2P_PORT1"
echo "NODE TWO: $NODE2_ID@$NODE2_IP:$P2P_PORT2"



sed -i.bak -E "s|^(enable[[:space:]]+=[[:space:]]+).*$|\1true| ; \
s|^(rpc_servers[[:space:]]+=[[:space:]]+).*$|\1\"http://$NODE1_IP:$RPC_PORT1,http://$NODE2_IP:$RPC_PORT2\"| ; \
s|^(trust_height[[:space:]]+=[[:space:]]+).*$|\1$BLOCK_HEIGHT| ; \
s|^(trust_hash[[:space:]]+=[[:space:]]+).*$|\1\"$TRUST_HASH\"| ; \
s|^(persistent_peers[[:space:]]+=[[:space:]]+).*$|\1\"${NODE1_ID}@${NODE1_IP}:${P2P_PORT1},${NODE2_ID}@${NODE2_IP}:${P2P_PORT2}\"|" $HOME/.osmosisd/config/config.toml

 

./osmosisd unsafe-reset-all
./osmosisd start
