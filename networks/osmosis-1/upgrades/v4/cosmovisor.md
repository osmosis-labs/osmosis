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
    source ~/.profile

Finally, you should copy the current osmosisd binary into the
cosmovisor/genesis folder.

    cp $GOPATH/bin/osmosisd ~/.osmosisd/cosmovisor/genesis/bin

Prepare for upgrade (v4)
------------------------

To prepare for the upgrade, you need to create some folders, and build
and install the new binary.

    mkdir -p ~/.osmosisd/cosmovisor/upgrades/v4/bin
    git clone https://github.com/osmosis-labs/osmosis
    cd osmosis
    git checkout v4.0.0
    make build
    cp build/osmosisd ~/.osmosisd/cosmovisor/upgrades/v4/bin

Now cosmovisor will run with the current binary, and will automatically
upgrade to this new binary at the appropriate height if run with:

    cosmovisor start

Please note, this does not automatically update your
`$GOPATH/bin/osmosisd` binary, to do that after the upgrade, please run
`make install` in the osmosis source folder.
