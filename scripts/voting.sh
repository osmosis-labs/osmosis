#!/bin/bash
while [ 1 ]
do
  symphonyd tx oracle aggregate-prevote 1234 1.25uusd,2.0stake symphonyvaloper1xzajl6g6ulpenuhxv0s7l7t9jgvyjf5t422ehy --from=validator1 --keyring-backend=test --broadcast-mode=sync --home=$HOME/.symphonyd --gas=70000 --chain-id=testing --yes --fees 175stake
  sleep 30s
  symphonyd tx oracle aggregate-vote 1234 1.25uusd,2.0stake  --keyring-backend=test --from=validator1 --home=$HOME/.symphonyd --chain-id=testing  --fees=200stake --broadcast-mode=sync --gas=75000 --yes
  sleep 5s
done
