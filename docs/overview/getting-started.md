# Quickstart
_(Note: This repository is under active development. Architecture and implementation may change without documentation)_

This is what you'd use to get a node up and running, fast. It assumes that it is starting on a blank ubuntu machine.  It eschews a systemd unit, allowing automation to be up to the user.  It assumes that installing Go is in-scope since Ubuntu's repositories aren't up to date and you'll be needing go to use osmosis.  It handles the Go environment variables because those are a common pain point.

**Install go**
```bash
wget -q -O - https://git.io/vQhTU | bash -s -- --version 1.17.2
```

Then exit and re-enter your shell.

**Install Osmosis and check that it is on $PATH**
```bash
git clone https://github.com/osmosis-labs/osmosis
cd osmosis
git checkout v3.1.0
make install
which osmosisd
```

**Launch Osmosis**
```bash
osmosisd init yourmonikerhere
wget -O ~/.osmosisd/config/genesis.json https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json
osmosisd start
```

More Nodes ==> More Network

More Network ==> Faster Sync

Faster Sync ==> Less Developer Friction

Less Developer Friction ==> More Osmosis

Thank you for supporting a healthy blockchain network and community by running an Osmosis node!
