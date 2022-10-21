#!/bin/bash

# this script creates an empty directory in app/upgrades called "vX" where X is a previous version + 1 and
# copies upgrade handler from the previous release with increased version where needed. Also insures that all the imports 
# make use of a current module version from go mod (see: 

# module=$(go mod edit -json | jq ".Module.Path") in this script

# )
# Github workflow which calls this script can be found here: osmosis/.github/workflows/auto-update-upgrade.yml

 latest_version=0
 for f in app/upgrades/*; do 
     s_f=(${f//// })
     version=${s_f[2]}
     num_version=${version//[!0-9]/}
     if [[ $num_version -gt $latest_version ]]; then
        LATEST_FILE=$f
        latest_version=$num_version
     fi
 done
 VERSION_CREATE=$((latest_version+1))
 NEW_FILE=./app/upgrades/v${VERSION_CREATE}

 mkdir $NEW_FILE

 touch $NEW_FILE/constants.go
 touch $NEW_FILE/upgrades.go

 module=$(go mod edit -json | jq ".Module.Path")
 module=${module%?}
 path=${module%???}

 cp ./app/upgrades/v${latest_version}/constants.go $NEW_FILE/constants.go
 cp ./app/upgrades/v${latest_version}/upgrades.go $NEW_FILE/upgrades.go

 sed -i "s/v$latest_version/v$VERSION_CREATE/g" $NEW_FILE/constants.go
 sed -i "s/v$latest_version/v$VERSION_CREATE/g" $NEW_FILE/upgrades.go

 bracks='"'

 # change imports in case go mod changed
 sed -i "s|.*/app/upgrades.*|\t$module/app/upgrades$bracks|" $NEW_FILE/constants.go
 sed -i "s|.*/app/upgrades.*|\t$module/app/upgrades$bracks|" $NEW_FILE/upgrades.go
 sed -i "s|.*/app/keepers.*|\t$module/app/keepers$bracks|" $NEW_FILE/upgrades.go
 sed -i "s|.*/x/lockup/types.*|\tlockuptypes $module/x/lockup/types$bracks|" $NEW_FILE/upgrades.go

 # change app/app.go file
 app_file=./app/app.go
 UPGRADES_LINE=$(grep -F upgrades.Upgrade{ $app_file)
 UPGRADES_LINE="${UPGRADES_LINE%?}, v${VERSION_CREATE}.Upgrade}"
 sed -i "s|.*upgrades.Upgrade{.*|$UPGRADES_LINE|" $app_file 

 PREV_IMPORT="v$latest_version $module/app/upgrades/v$latest_version$bracks"
 NEW_IMPORT="v$VERSION_CREATE $module/app/upgrades/v$VERSION_CREATE$bracks"
 sed -i"s|.*$PREV_IMPORT.*|\t$PREV_IMPORT\n\t$NEW_IMPORT|" $app_file
 
 # change e2e version in makefile
 sed -i "s/E2E_UPGRADE_VERSION := ${bracks}v$latest_version$bracks/E2E_UPGRADE_VERSION := ${bracks}v$VERSION_CREATE$bracks/" ./Makefile