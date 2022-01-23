#! /bin/bash
set -e
PS1="\e[0;34m\u@\h Color Test \w> \e[m"
echo "
 ██████╗ ███████╗███╗   ███╗ ██████╗ ███████╗██╗███████╗
██╔═══██╗██╔════╝████╗ ████║██╔═══██╗██╔════╝██║██╔════╝
██║   ██║███████╗██╔████╔██║██║   ██║███████╗██║███████╗
██║   ██║╚════██║██║╚██╔╝██║██║   ██║╚════██║██║╚════██║
╚██████╔╝███████║██║ ╚═╝ ██║╚██████╔╝███████║██║███████║
 ╚═════╝ ╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚══════╝╚═╝╚══════╝
Welcome to the Osmosis node installer.
For more information visit docs.osmosis.zone
"
#TODO CHAIN TESTNET MAINNET

#Select sync method. There are 3 options available.
echo "Please choose a sync method:"
PS3='Enter choice:'
modes=("default" "pruned" "archive")
select mode in "${modes[@]}"; do
MODE=$mode
printf "\033c"
    case $mode in
        "default")
            modeNote="This is a full network sync which will require about 350GB of disk space."
             break
            ;;
        "pruned")
            modeNote="This will require about 150GB of disk space."
             break
            ;;
        "archive")
             modeNote="This will require more than 700GB of disk space.."
             break
            ;;
	      "Quit")
	           echo "User requested exit"
	           exit
	          ;;
        *) echo "invalid option $REPLY.";;
    esac
    #set selected mode as a variable

  ##
##tput reset
done

#Select region
echo "Please select a region closest to you:"
PS3='Enter choice:'
regions=("San Francisco" "Singapore" "Netherlands")
select region in "${regions[@]}"; do
REGION=$region
printf "\033c"
    case $region in
        "San Francisco")
            break;
            ;;
        "Singapore")
            break;
            ;;
        "Netherlands")
            break;
            ;;
	      "Quit")
	           echo "User requested exit"
	           exit
	          ;;
        *) echo "invalid option $REPLY";;
    esac
    #set selected region as a variable
    #clear screen
   ## printf "\033c"
done

read -p "Installing using the $MODE sync method, $modeNote. Downloading from $REGION. Press enter to continue"



#PEERS=""
VERSION=""

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
	FILENAME=`curl https://quicksync.io/osmosis.json | jq -r --arg MODE "$MODE" '.[] | select(.network==$MODE)|select (.mirror=="$REGION")|.filename'`
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

