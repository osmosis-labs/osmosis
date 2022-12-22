#!/bin/bash

if [ -d "$HOME/.osmosisd" ]; then
  if [ ! -f $HOME/.osmosisd/.env ]; then
    echo "OSMOSISD_ENVIRONMENT=mainnet" > $HOME/.osmosisd/.env
  fi
else
  echo "./osmosisd not exist"
fi

cat $HOME/.osmosisd/.env