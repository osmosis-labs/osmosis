# Memory Requirements

For this upgrade we're expecting that nodes will need 32GB of total
memory available. This can be a combination of RAM and Swap, though RAM
is preferred for speed. If you cannot acquire 32GB of RAM, 16GB of RAM
and 16GB of Swap should suffice.

Short version swap setup instructions:

    sudo fallocate -l 16G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile

In depth swap setup instructions:
<https://www.digitalocean.com/community/tutorials/how-to-add-swap-space-on-ubuntu-20-04>

# GO Version

Please note, Osmosis `v5.0.0` requires go version `1.17`, if you run
`go version` and it says `1.16.x` you should uninstall go and update to
`1.17`. The easiest way to do this is with these
scripts<https://github.com/canha/golang-tools-install-script>

To uninstall run:

`curl https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash -s -- --remove`

Followed this to reinstall the latest `1.17.x` version:

`curl https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash`

# Install and setup Cosmovisor

We highly recommend validators use cosmovisor to run their nodes. This
will make low-downtime upgrades smoother, as validators don't have to
manually upgrade binaries during the upgrade, and instead can preinstall
new binaries, and cosmovisor will automatically update them based on
on-chain SoftwareUpgrade proposals.

You should review the docs for cosmovisor located here:
<https://docs.cosmos.network/master/run-node/cosmovisor.html>

If you choose to use cosmovisor, please continue with these
instructions:

To install Cosmovisor:

    git clone https://github.com/cosmos/cosmos-sdk
    cd cosmos-sdk
    git checkout v0.42.9
    make cosmovisor
    cp cosmovisor/cosmovisor $GOPATH/bin/cosmovisor
    cd $HOME

After this, you must make the necessary folders for cosmosvisor in your
daemon home directory (\~/.osmosisd).

``` {.sh}
mkdir -p ~/.osmosisd
mkdir -p ~/.osmosisd/cosmovisor
mkdir -p ~/.osmosisd/cosmovisor/genesis
mkdir -p ~/.osmosisd/cosmovisor/genesis/bin
mkdir -p ~/.osmosisd/cosmovisor/upgrades
```

Cosmovisor requires some ENVIRONMENT VARIABLES be set in order to
function properly. We recommend setting these in your `.profile` so it
is automatically set in every session.

For validators we recommmend setting

- `DAEMON_ALLOW_DOWNLOAD_BINARIES=false` for security reasons
- `DAEMON_LOG_BUFFER_SIZE=512` to avoid a bug with extra long log
    lines crashing the server.
- `DAEMON_RESTART_AFTER_UPGRADE=true` for unattended upgrades

```{=html}
<!-- -->
```

    echo "# Setup Cosmovisor" >> ~/.profile
    echo "export DAEMON_NAME=osmosisd" >> ~/.profile
    echo "export DAEMON_HOME=$HOME/.osmosisd" >> ~/.profile
    echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=false" >> ~/.profile
    echo "export DAEMON_LOG_BUFFER_SIZE=512" >> ~/.profile
    echo "export DAEMON_RESTART_AFTER_UPGRADE=true" >> ~/.profile
    echo "export UNSAFE_SKIP_BACKUP=true" >> ~/.profile
    source ~/.profile

You may leave out `UNSAFE_SKIP_BACKUP=true`, however the backup takes a
decent amount of time and public snapshots of old states are available.

Finally, you should copy the current osmosisd binary into the
cosmovisor/genesis folder.

    cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin

## Prepare for upgrade (v5)

To prepare for the upgrade, you need to create some folders, and build
and install the new binary.

    mkdir -p ~/.osmosisd/cosmovisor/upgrades/v5/bin
    git clone https://github.com/osmosis-labs/osmosis
    cd osmosis
    git checkout v5.0.0
    make build
    cp build/osmosisd ~/.osmosisd/cosmovisor/upgrades/v5/bin

Now cosmovisor will run with the current binary, and will automatically
upgrade to this new binary at the appropriate height if run with:

    cosmovisor start

Please note, this does not automatically update your
`$GOPATH/bin/osmosisd` binary, to do that after the upgrade, please run
`make install` in the osmosis source folder.
