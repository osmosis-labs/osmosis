# Generates a state export from Osmosis mainnet (through Chainlayer) with a modified chain-id, and uploads to transfer.sh

export CHAIN_ID="osmosis-mainnet-testnet-X"
export VERSION="v4.2.0"


sudo apt-get update
sudo apt-get upgrade -y
sudo apt-get install curl build-essential jq git wget liblz4-tool aria2 -y

curl https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash

export GOROOT=/root/.go
export PATH=$GOROOT/bin:$PATH
export GOPATH=/root/go
export PATH=$GOPATH/bin:$PATH
 
git clone https://github.com/osmosis-labs/osmosis
cd osmosis
git checkout $VERSION
make install

osmosisd init "testnet-validator" --chain-id=$CHAIN_ID

cd ~/.osmosisd/
FILENAME=`curl https://quicksync.io/osmosis.json|jq -r '.[] |select(.network=="pruned")|select (.mirror=="SanFrancisco")|.filename'`

#Switch which of these sections is commented out if disk space is insufficeient

#single threaded / half space
# wget -O - https://getsfo.quicksync.io/$FILENAME | lz4 -d | tar -xvf -

#multi threaded / double space
aria2c -x5 https://getsfo.quicksync.io/$FILENAME
lz4 -d $FILENAME | tar xf -
rm $FILENAME
#/multi threaded


cd ~
osmosisd export > genesis.json
curl --upload-file ./genesis.txt https://transfer.sh/genesis.json