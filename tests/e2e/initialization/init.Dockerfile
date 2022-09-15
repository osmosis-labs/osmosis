# syntax=docker/dockerfile:1

ARG SETUP_VERSION

## Build Image
FROM osmosis-setup:${SETUP_VERSION} as builder

ARG E2E_SCRIPT_NAME

WORKDIR /osmosis
COPY . /osmosis

RUN BUILD_TAGS=muslc LINK_STATICALLY=true E2E_SCRIPT_NAME=${E2E_SCRIPT_NAME} make build-e2e-script

## Deploy image
FROM ubuntu

# Args only last for a single build stage - renew
ARG E2E_SCRIPT_NAME

COPY --from=builder /osmosis/build/${E2E_SCRIPT_NAME} /bin/${E2E_SCRIPT_NAME}

ENV HOME /osmosis
WORKDIR $HOME

# Docker ARGs are not expanded in ENTRYPOINT in the exec mode. At the same time,
# it is impossible to add CMD arguments when running a container in the shell mode.
# As a workaround, we create the entrypoint.sh script to bypass these issues.
RUN echo "#!/bin/bash\n${E2E_SCRIPT_NAME} \"\$@\"" >> entrypoint.sh && chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
