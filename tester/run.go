package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jnsoft/gamma/database"
	"github.com/jnsoft/gamma/util/hexutil"
)

const a1 string = "0x0000000000000000000000000000000000000001"
const a2 string = "0x0000000000000000000000000000000000000002"
const a3 string = "0x0000000000000000000000000000000000000003"

func main() {

	//var v1, v2 database.Address
	//fmt.Println("v1 error:", json.Unmarshal([]byte(`"0x01"`), &v1))
	//fmt.Println("v2 error:", json.Unmarshal([]byte(`"0x0102030405060708091011121314151617181920"`), &v2))
	//fmt.Println("v2:", v2)
	// Output:
	// v1 error: hex string has length 2, want 10 for MyType
	// v2 error: <nil>
	// v2: 0x0101010101

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

	database.NewSimpleTx(database.Address(hexutil.MustDecode(a1)), database.Address(hexutil.MustDecode(a1)), 1, "")

	block0 := database.NewSimpleBlock(
		database.Hash{},
		[]database.SimpleTx{
			database.NewSimpleTxStringAddress(a1, a2, 3, ""),
			database.NewSimpleTxStringAddress(a1, a2, 5, ""),
		},
	)

	state.AddSimpleBlock(block0)
	block0hash, _ := state.Persist()

	block1 := database.NewSimpleBlock(
		block0hash,
		[]database.SimpleTx{
			database.NewSimpleTxStringAddress(a1, a2, 2000, ""),
			database.NewSimpleTxStringAddress(a1, a1, 100, "mint"),
			database.NewSimpleTxStringAddress(a2, a1, 1, ""),
			database.NewSimpleTxStringAddress(a2, a3, 1000, ""),
			database.NewSimpleTxStringAddress(a2, a1, 50, ""),
			database.NewSimpleTxStringAddress(a1, a1, 100, "mint"),
		},
	)

	state.AddSimpleBlock(block1)
	state.Persist()

}
