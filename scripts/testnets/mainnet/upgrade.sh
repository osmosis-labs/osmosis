#FIXME incomplete / WIP
# should be replaced with parameterized upgrade script based on branch / upgrade handler to be tested
# should dynamically get usable upgrade height from state export / live chain
# should use existing local validator key, created in modify/setup, rather than hardcoded seed

echo "kitchen comic flower drip sick prize account cheese truth income weekend nominee segment punch call satisfy captain earth ethics wasp clump tunnel orchard advance" | osmosisd keys add validator --recover --keyring-backend=test
yes | osmosisd tx gov submit-proposal software-upgrade v5 --upgrade-height=2292611\
    --from=validator\
    --keyring-backend=test\
    --title="Boron v5 upgrade"\
    --description="v5 upgrade - #changelog"\
    --upgrade-info="na"\
    --fees=100000uosmo\
yes | osmosisd tx gov deposit 89 500000000uosmo\
    --from=validator --keyring-backend=test --fees=100000uosmo
sleep 7
yes | osmosisd tx gov vote 89 yes\
    --from=validator --keyring-backend=test --fees=100000uosmo

