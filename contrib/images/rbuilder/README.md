# Reproducible Build System

This image is meant to provide a minimal deterministic
buildsystem for Cosmos SDK applications.

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

This spins up an rbuilder container with a volume installed to the
root of the repository. This way, the builder has access to the `.build.sh`file and is
able to execute it.

Currently, only `linux/amd64'` is supported. We can add support for other
platforms by modifying TARGET_PLATFORMS environment variable. in `build-reproducible`
Makefile step. Adding more support is blocked by our dependency on wasmvm.
The support of some platforms are already added in new versions of wasmvm.
Follow the release log for more detaisl when updating our builder:
https://github.com/CosmWasm/wasmvm/releases
