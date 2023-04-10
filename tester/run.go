package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jnsoft/gamma/database"
	"github.com/jnsoft/gamma/util/hexutil"
)

func main() {

	var genesisJson = `{
		"symbol": "TGL",
		"balances": {
		  "0x1": 1000000,
		  "0x2": 1
		},
		"fork_tip_1": 35
	  }`

	data, err := json.Marshal(genesisJson) //Not Required
	if err != nil {
		fmt.Println("Error with marchal JSON: " + err.Error())
	}
	fmt.Println("data ", data)

	var res database.Genesis

	err = json.Unmarshal(data, &res)
	if err != nil {
		fmt.Println("Error with marchal JSON: " + err.Error())
	} else {
		fmt.Printf("Read a message from %v     %v \n", res.Symbol, res.ForkTIP1)
	}

	t := time.Now()
	fmt.Println(t.Month())
	fmt.Println(t.Day())
	fmt.Println(t.Year())

	// TODO: hantera
	// signera transaktioner?
	//

	state, err := database.NewStateFromDisk("/tmp/gammadb", 1) // kommer skapa /tmp/gammadb/database/ och filer där
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	database.NewSimpleTx(database.Address(hexutil.MustDecode("0x01")), database.Address(hexutil.MustDecode("0x01")), 1, "")

	block0 := database.NewSimpleBlock(
		database.Hash{},
		[]database.SimpleTx{
			database.NewSimpleTxStringAddress("0x01", "0x02", 3, ""),
			database.NewSimpleTxStringAddress("0x01", "0x02", 5, ""),
		},
	)

	state.AddSimpleBlock(block0)
	block0hash, _ := state.Persist()

	block1 := database.NewSimpleBlock(
		block0hash,
		[]database.SimpleTx{
			database.NewSimpleTxStringAddress("0x01", "0x02", 2000, ""),
			database.NewSimpleTxStringAddress("0x01", "0x01", 100, "mint"),
			database.NewSimpleTxStringAddress("0x02", "0x01", 1, ""),
			database.NewSimpleTxStringAddress("0x02", "0x03", 1000, ""),
			database.NewSimpleTxStringAddress("0x02", "0x01", 50, ""),
			database.NewSimpleTxStringAddress("0x01", "0x01", 100, "mint"),
		},
	)

	state.AddSimpleBlock(block1)
	state.Persist()

}
