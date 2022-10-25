# this script checks if existing upgrade handler's version is not smaller than current release 

#!/bin/bash

VERSION=$1
latest_version=0
for f in app/upgrades/*; do 
    s_f=(${f//// })
    version=${s_f[2]}
    num_version=${version//[!0-9]/}
    if [[ $num_version -gt $latest_version ]]; then
        latest_version=$num_version
    fi
done

VERSION=${VERSION[@]:1}
VERSION_MAJOR=(${VERSION//./ })
VERSION_MAJOR=${VERSION_MAJOR[0]}
if [[ $VERSION_MAJOR -gt $latest_version ]]; then
    echo "MAJOR=1" >> $GITHUB_ENV
fi