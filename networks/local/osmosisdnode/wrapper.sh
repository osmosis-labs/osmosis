#!/usr/bin/env sh

##
## Input parameters
##
BINARY=/osmosisd/${BINARY:-osmosisd}
ID=${ID:-0}
LOG=${LOG:-osmosisd.log}

##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'omsisisd' E.g.: -e BINARY=osmosisd_my_test_version"
	exit 1
fi
BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"
if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

##
## Run binary with all parameters
##
export OSMOSISDHOME="/osmosisd/node${ID}/osmosisd"

if [ -d "$(dirname "${OSMOSISDHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${OSMOSISDHOME}" "$@" | tee "${OSMOSISDHOME}/${LOG}"
else
  "${BINARY}" --home "${OSMOSISDHOME}" "$@"
fi