#!/bin/bash

# this script creates an empty directory in app/upgrades called "vX" where X is a previous version + 1 and
# copies upgrade handler from the previous release with increased version where needed. Also insures that all the imports 
# make use of a current MODULE version from go mod (see: 

# MODULE=$(go mod edit -json | jq ".MODULE.Path") in this script

# )
# Github workflow which calls this script can be found here: osmosis/.github/workflows/auto-update-upgrade.yml

 LATEST_VERSION=0
 for f in app/upgrades/*; do 
     s_f=(${f//// })
     version=${s_f[2]}
     num_version=${version//[!0-9]/}
     if [[ $num_version -gt $LATEST_VERSION ]]; then
        LATEST_FILE=$f
        LATEST_VERSION=$num_version
     fi
 done
 VERSION_CREATE=$((LATEST_VERSION+1))
 NEW_FILE=./app/upgrades/v${VERSION_CREATE}

 mkdir $NEW_FILE

 touch $NEW_FILE/constants.go
 touch $NEW_FILE/upgrades.go

 MODULE=$(go mod edit -json | jq ".MODULE.Path")
 MODULE=${MODULE%?}
 path=${MODULE%???}

 cp ./app/upgrades/v${LATEST_VERSION}/constants.go $NEW_FILE/constants.go
 cp ./app/upgrades/v${LATEST_VERSION}/upgrades.go $NEW_FILE/upgrades.go

 sed -i "s/v$LATEST_VERSION/v$VERSION_CREATE/g" $NEW_FILE/constants.go
 sed -i "s/v$LATEST_VERSION/v$VERSION_CREATE/g" $NEW_FILE/upgrades.go

 BRACKS='"'

 # change imports in case go mod changed
 sed -i "s|.*/app/upgrades.*|\t$MODULE/app/upgrades$BRACKS|" $NEW_FILE/constants.go
 sed -i "s|.*/app/upgrades.*|\t$MODULE/app/upgrades$BRACKS|" $NEW_FILE/upgrades.go
 sed -i "s|.*/app/keepers.*|\t$MODULE/app/keepers$BRACKS|" $NEW_FILE/upgrades.go
 sed -i "s|.*/x/lockup/types.*|\tlockuptypes $MODULE/x/lockup/types$BRACKS|" $NEW_FILE/upgrades.go

 # change app/app.go file
 APP_FILE=./app/app.go
 UPGRADES_LINE=$(grep -F upgrades.Upgrade{ $APP_FILE)
 UPGRADES_LINE="${UPGRADES_LINE%?}, v${VERSION_CREATE}.Upgrade}"
 sed -i "s|.*upgrades.Upgrade{.*|$UPGRADES_LINE|" $APP_FILE 

 PREV_IMPORT="v$LATEST_VERSION $MODULE/app/upgrades/v$LATEST_VERSION$BRACKS"
 NEW_IMPORT="v$VERSION_CREATE $MODULE/app/upgrades/v$VERSION_CREATE$BRACKS"
 sed -i "s|.*$PREV_IMPORT.*|\t$PREV_IMPORT\n\t$NEW_IMPORT|" $APP_FILE
 
 # change e2e version in makefile
 sed -i "s/E2E_UPGRADE_VERSION := ${BRACKS}v$LATEST_VERSION$BRACKS/E2E_UPGRADE_VERSION := ${BRACKS}v$VERSION_CREATE$BRACKS/" ./Makefile