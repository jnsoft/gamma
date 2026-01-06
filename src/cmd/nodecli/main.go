package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jnsoft/gamma/src/pkg/database"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "init":
		cmdInit()
	default:
		usage()
	}
}

func usage() {
	fmt.Println("nodecli <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init     - initialize data directory")
}

func cmdInit() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dataDir := fs.String("d", "", "data directory")
	fs.Parse(os.Args[2:])
	if err := database.InitDataDirectory(*dataDir); err != nil {
		// handle
	}
}
