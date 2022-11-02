#!/bin/bash

# 1) this script creates an empty directory in app/upgrades called "vX" where X is a previous version + 1 with an empty upgrade handler.
# 2) increases E2E_UPGRADE_VERSION in makefile by 1
# 3) adds new version to app.go

# Also insures that all the imports make use of a current module version from go mod:
# (see:    module=$(go mod edit -json | jq ".Module.Path")      in this script)
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
 version_create=$((latest_version+1))
 new_file=./app/upgrades/v${version_create}

 mkdir $new_file
 CONSTANTS_FILE=$new_file/constants.go
 UPGRADES_FILE=$new_file/upgrades.go
 touch $CONSTANTS_FILE
 touch $UPGRADES_FILE
 
 module=$(go mod edit -json | jq ".Module.Path")
 module=${module%?}
 path=${module%???}

 bracks='"'
 # set packages
 echo -e "package v${version_create}\n" >> $CONSTANTS_FILE
 echo -e "package v${version_create}\n" >> $UPGRADES_FILE

 # imports
 echo "import (" >> $CONSTANTS_FILE
 echo "import (" >> $UPGRADES_FILE

 # set imports for constants.go
 echo -e "\t$module/app/upgrades$bracks\n" >> $CONSTANTS_FILE
 echo -e '\tstore "github.com/cosmos/cosmos-sdk/store/types"' >> $CONSTANTS_FILE

 # set imports for upgrades.go
 echo -e '\tsdk "github.com/cosmos/cosmos-sdk/types"' >> $UPGRADES_FILE
 echo -e '\t"github.com/cosmos/cosmos-sdk/types/module"' >> $UPGRADES_FILE
 echo -e '\tupgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"\n' >> $UPGRADES_FILE
 echo -e "\t$module/app/keepers$bracks" >> $UPGRADES_FILE
 echo -e "\t$module/app/upgrades$bracks" >> $UPGRADES_FILE

 # close import
 echo ")" >> $UPGRADES_FILE
 echo -e ")\n" >> $CONSTANTS_FILE
 
 # constants.go logic
 echo "// UpgradeName defines the on-chain upgrade name for the Osmosis v$version_create upgrade." >> $CONSTANTS_FILE
 echo "const UpgradeName = ${bracks}v$version_create$bracks" >> $CONSTANTS_FILE
 echo "
var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}" >> $CONSTANTS_FILE
 
 # upgrades.go logic
 echo "
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bpm upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}" >> $UPGRADES_FILE

  # change app/app.go file
 app_file=./app/app.go
 UPGRADES_LINE=$(grep -F upgrades.Upgrade{ $app_file)
 UPGRADES_LINE="${UPGRADES_LINE%?}, v${version_create}.Upgrade}"
 sed -i "s|.*upgrades.Upgrade{.*|$UPGRADES_LINE|" $app_file 

 PREV_IMPORT="v$latest_version $module/app/upgrades/v$latest_version$bracks"
 NEW_IMPORT="v$version_create $module/app/upgrades/v$version_create$bracks"
 sed -i "s|.*$PREV_IMPORT.*|\t$PREV_IMPORT\n\t$NEW_IMPORT|" $app_file

 # change e2e version in makefile
 sed -i "s/E2E_UPGRADE_VERSION := ${bracks}v$latest_version$bracks/E2E_UPGRADE_VERSION := ${bracks}v$version_create$bracks/" ./Makefile