package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jnsoft/gamma/database"
)

func main() {
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
