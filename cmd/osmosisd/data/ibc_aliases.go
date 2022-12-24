package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type IBCAliases map[string]string

var DenomToIbcAlias IBCAliases

func GetIBCAliasesMap() IBCAliases {
	if DenomToIbcAlias == nil {
		return makeIBCAliasesMap()
	}
	return DenomToIbcAlias
}

func getAssetlist() {
	err := exec.Command("python3", "get_assetlist.py").Run()
	if err != nil {
		log.Fatal(err)
	}
}

func makeIBCAliasesMap() IBCAliases {
	// temporarily create an assetlist file
	getAssetlist()
	defer func() {
		err := os.Remove("assetlist.json")
		if err != nil {
			log.Fatal(err)
		}
	}()

	assetlistFile, err := os.Open("assetlist.json")

	if err != nil {
		panic(fmt.Sprintf("Could not open file: %s", err))
	}

	defer assetlistFile.Close()

	bz, err := ioutil.ReadAll(assetlistFile)
	if err != nil {
		panic(fmt.Sprintf("Could not read file: %s", err))
	}
	var result map[string]interface{}
	err = json.Unmarshal(bz, &result)
	if err != nil {
		fmt.Println(err)
	}

	DenomToIbcAlias = IBCAliases{}
	assets, _ := result["assets"].([]interface{})
	for i := 0; i < len(assets); i++ {
		var alias string

		asset, _ := assets[i].(map[string]interface{})

		ibcAlias, _ := asset["base"].(string)
		ibcField := asset["ibc"]

		if ibcField == nil {
			alias = ibcAlias
		} else {
			ibcMap, _ := ibcField.(map[string]interface{})
			alias, _ = ibcMap["source_denom"].(string)
		}

		DenomToIbcAlias[alias] = ibcAlias
	}

	return DenomToIbcAlias
}
