# Reproducible Build System

This image and scripts are meant to provide a minimal deterministic
build system for the Osmosis application.

It was created by referencing Tendermint's [implementation](https://github.com/tendermint/images/blob/cf0d1a9f3731e30540bbfa36a36d13e4dcccf5eb/rbuilder/README.md)

# Requirements And Usage

The Osmosis repository must include a`.build.sh` executable file 
in the root folder meant to drive the build process.

The build's outputs are produced in the top-level `artifacts` directory.

## Building the Image Locally

```
cd ./contrib/images

make rbuilder
```

This creates the `rbuilder` image. To run a container of this image locally and build the binaries:
```
cd <osmosis root>

make build-reproducible
```

This spins up an `rbuilder` container with a volume installed to the
root of the repository. This way, the builder has access to the `.build.sh`file and is
able to execute it.

## Wasmvm dependency

Currently, only `linux/amd64` is supported. Adding more support is blocked by our dependency on wasmvm.

The support of some platforms is already added in new versions of wasmvm.
Follow the release log for more detaisl when updating our builder:
https://github.com/CosmWasm/wasmvm/releases

Once wasmvm is upgraded, more platforms may be built by changing
`TARGET_PLATFORMS` environment variable in `build-reproducible`
Makefile step. 
