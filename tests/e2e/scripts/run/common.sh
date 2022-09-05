#!/bin/bash

# N.B.: We match all tags but "Debug" and semantic version tags such as "V10". These are the only
# tags we support. As a result, the only remaining tag is the Git SHA tag.
LIST_DOCKER_IMAGE_HASHES=$(docker images osmosis --format "{{ title .Tag }}" | awk '!/Debug/ && !/V[0-9-]+/' | awk '{print tolower($0)}')
