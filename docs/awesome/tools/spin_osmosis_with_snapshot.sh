#! /bin/bash -x  
set -e 

#PEERS=""
VERSION=""
MODE=""

while getopts "v:m:h" option
	do
	case $option in
	v) VERSION="$OPTARG"	;;
	m) MODE="$OPTARG"	;;
	h) echo "Spin Osmosis Full Node (with ChainLayer QuickSync Snapshot)
Usage: ./spin_osmosis_with_snapshot.sh -v vX.Y.Z -m ______

 -v  Input Osmosis version you want to install (default \"latest version\")
 -m  Choose (pruned | default | archive) (default \"pruned\")" ;;
	esac
	done


main() {
        basic_setup
        install_go
        osmosis
 	get_snapshot
        get_peers
        spin_up
}



basic_setup() {
       sudo apt-get update -y && sudo apt-get upgrade -y
       sudo apt-get install build-essential jq wget -y
}


install_go() {
        sudo rm -rf /usr/local/go
        wget https://dl.google.com/go/go1.17.1.linux-amd64.tar.gz
        tar -xvf go1.17.1.linux-amd64.tar.gz
        sudo mv go /usr/local
        GOROOT=/usr/local/go
        PATH=$GOROOT/bin:$PATH
}


osmosis(){
        git clone https://github.com/osmosis-labs/osmosis
        cd osmosis
	if [ -z "$VERSION" ]; then
                VERSION=$(git describe --tags `git rev-list --tags --max-count=1`)
	fi
	git fetch && git checkout $VERSION
	make build
        ./build/osmosisd init --chain-id osmosis-1 BestDEX
}


# get snapshot from ChainLayer QuickSync
get_snapshot() {
        sudo apt-get install aria2 liblz4-tool -y
        if [ -z "$MODE" ]; then
        	MODE="pruned"
	fi
	FILENAME=`curl https://quicksync.io/osmosis.json | jq -r --arg MODE "$MODE" '.[] | select(.network==$MODE)|select (.mirror=="Netherlands")|.filename'`
        cd /$HOME/.osmosisd
        if [[ ! -f $FILENAME || -f $FILENAME.aria2 ]]; then
		aria2c -x5 https://get.quicksync.io/$FILENAME
	fi
        lz4 -d $FILENAME | tar xf -
}


get_peers() {
        sed -i "s/persistent_peers = \"\"persistent_peers = \"$PEERS\"/g" $HOME/.osmosisd/config/config.toml
}


spin_up() {
        ./$HOME/osmosis/build/osmosisd start
}

main; exit

