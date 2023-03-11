package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	genesis_path = "postgres"
	tx_db_path   = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
)

func main() {
	var tbbCmd = &cobra.Command{
		Use:   "gamma",
		Short: "The Gamma CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	tbbCmd.AddCommand(versionCmd)
	tbbCmd.AddCommand(balancesCmd())
	tbbCmd.AddCommand(txCmd())

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
