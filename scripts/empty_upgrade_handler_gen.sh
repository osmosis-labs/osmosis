#!/bin/bash

 latest_version=0
 for f in app/upgrades/*; do 
     s_f=(${f//// })
     version=${s_f[2]}
     num_version=${version//[!0-9]/}
     if [[ $num_version -gt $latest_version ]]; then
         latest_version=$num_version
     fi
 done

 version_create=$((latest_version+1))
 new_file=./app/upgrades/v${version_create}

 mkdir $new_file

 touch $new_file/constants.go
 touch $new_file/upgrades.go

 module=$(go mod edit -json | jq ".Module.Path")
 module=${module%?}
 path=${module%???}

 cp ./app/upgrades/v${latest_version}/constants.go $new_file/constants.go
 cp ./app/upgrades/v${latest_version}/upgrades.go $new_file/upgrades.go
#  SEDBRACKETS=
# if [[ "$OSTYPE" == "darwin"* ]]; then
#     SEDBRACKETS=''
# fi
 sed -i "s/v$latest_version/v$version_create/g" $new_file/constants.go
 sed -i "s/v$latest_version/v$version_create/g" $new_file/upgrades.go

 bracks='"'

 # change imports in case go mod changed
 sed -i "s|.*/app/upgrades.*|\t$module/app/upgrades$bracks|" $new_file/constants.go
 sed -i "s|.*/app/upgrades.*|\t$module/app/upgrades$bracks|" $new_file/upgrades.go
 sed -i "s|.*/app/keepers.*|\t$module/app/keepers$bracks|" $new_file/upgrades.go
 sed -i "s|.*/x/lockup/types.*|\tlockuptypes $module/x/lockup/types$bracks|" $new_file/upgrades.go

 # change e2e version in makefile
 sed -i "s/E2E_UPGRADE_VERSION := ${bracks}v$latest_version$bracks/E2E_UPGRADE_VERSION := ${bracks}v$version_create$bracks/" ./Makefile