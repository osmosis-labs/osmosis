#!/bin/bash

if [ -d "$HOME/.osmosisd" ]; then
  echo "OSMOSISD_ENVIRONMENT=mainnet" > $HOME/.osmosisd/.env
fi