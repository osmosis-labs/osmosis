package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/osmosis-labs/osmosis/v7/tests/e2e/chain"
)

func main() {
	var dataDir string
	var chainId string
	var m chain.Chain
	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.StringVar(&chainId, "chain-id", "", "chain ID")
	flag.Parse()

	if len(dataDir) == 0 {
		panic("data-dir is required")
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(err)
	}

	chain, err := chain.Init(chainId, dataDir)
	if err != nil {
		panic(err)
	}

	// enc := gob.NewEncoder(f)
	// if err := enc.Encode(chain); err != nil {
	// 	log.Fatal(err)
	// }

	// var b bytes.Buffer
	// e := gob.NewEncoder(&b)
	// if err := e.Encode(chain); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Encoded Struct ", b)

	// fmt.Println(chain)

	// var buf bytes.Buffer
	// enc := gob.NewEncoder(&buf)

	// if err := enc.Encode(chain); err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(buf.Bytes())
	fmt.Printf("test %+v", chain.Validators[0].KeyInfo)
	b, err := json.Marshal(chain)
	fmt.Println(b)
	fmt.Println(m)

	fileName := fmt.Sprintf("%v/%v-encode", dataDir, chainId)
	err2 := os.WriteFile(fileName, b, 0777)
	if err2 != nil {
		panic(err)
	}
	encJson, _ := os.ReadFile(fileName)

	err3 := json.Unmarshal(encJson, &m)
	fmt.Println(err3)
	fmt.Println(m)
	fmt.Printf("TEEEEEEEEEST %+v\n", m.Validators[0])
	fmt.Printf("TEEEEEEEEEST %+v\n", m.Validators[1])

}
