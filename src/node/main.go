package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jnsoft/gamma/src/pkg/common"
	"github.com/jnsoft/gamma/src/pkg/database"
)

func main() {
	datadir := flag.String("datadir", "./data", "Directory for storing node data")
	ip := flag.String("ip", "127.0.0.1", "IP address to listen on")
	port := flag.Int("p", 8080, "Port to listen on")
	useSSL := flag.Bool("ssl", false, "Enable SSL")
	verbose := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	datadirAbs, err := common.GetFullPath(*datadir)
	if err == nil {
		datadir = &datadirAbs
	}

	if *verbose {
		fmt.Printf("Listening on: %s:%d\n", *ip, *port)
		if *useSSL {
			fmt.Println("SSL is enabled.")
		}
		fmt.Println("Data directory:", *datadir)
	}

	// check all files, create if missing
	err = database.InitDataDirectory(*datadir)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	state, err := database.GetStateFromDisk(*datadir)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println(state.DbFile)

}
