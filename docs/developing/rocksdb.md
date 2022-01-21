# install rocksdb on ubuntu focal fossa

update & install dependencies
```
sudo apt update
sudo apt install -yy libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev
```

clone into rocksdb repo
```
git clone https://github.com/facebook/rocksdb.git
cd rocksdb
```

checkout the specific version
```
git checkout v6.27.3
```

ignore GCC warnings
```
export CXXFLAGS='-Wno-error=deprecated-copy -Wno-error=pessimizing-move -Wno-error=class-memaccess'
```

build as a shared library
```
sudo make shared_lib
```

install shared library to /usr/lib/ and header files to /usr/include/rocksdb/:
```
sudo make install-shared INSTALL_PATH=/usr
```

export flags
```
echo "export LD_LIBRARY_PATH=/usr/local/lib" >> $HOME/.profile
source $HOME/.profile
```
```
echo "export LD_LIBRARY_PATH=/usr/local/lib" >> $HOME/.bashrc
source $HOME/.bashrc
```

# compile osmosis with rocksdb install tags
```
cd
git clone https://github.com/osmosis-labs/osmosis
cd osmosis
git checkout v6.1.0
BUILD_TAGS=rocksdb make install
```

# start osmosisd with rocksdb db_backend flag
```
osmosisd start --db_backend rocksdb
```
