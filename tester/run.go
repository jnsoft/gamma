package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jnsoft/gamma/database"
)

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

	// database.NewSimpleTx(database.Address(hexutil.MustDecode(a1)), database.Address(hexutil.MustDecode(a1)), 1, "")

	block0 := database.NewSimpleBlock(
		database.Hash{},
		0,
		database.ToAddress(database.A0),
		[]database.SimpleTx{
			database.NewSimpleTxStringAddress(database.A1, database.A2, 3, ""),
			database.NewSimpleTxStringAddress(database.A1, database.A2, 5, ""),
		},
	)

	state.AddSimpleBlock(block0)
	block0hash, _ := state.Persist()

	block1 := database.NewSimpleBlock(
		block0hash,
		1,
		database.ToAddress(database.A0),
		[]database.SimpleTx{
			database.NewSimpleTxStringAddress(database.A1, database.A2, 2000, ""),
			database.NewSimpleTxStringAddress(database.A1, database.A1, 100, "mint"),
			database.NewSimpleTxStringAddress(database.A2, database.A1, 1, ""),
			database.NewSimpleTxStringAddress(database.A2, database.A3, 1000, ""),
			database.NewSimpleTxStringAddress(database.A2, database.A1, 50, ""),
			database.NewSimpleTxStringAddress(database.A1, database.A1, 100, "mint"),
		},
	)

	state.AddSimpleBlock(block1)
	state.Persist()

}
