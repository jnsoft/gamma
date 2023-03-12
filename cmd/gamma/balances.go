package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/jnsoft/gamma/database"
	"github.com/spf13/cobra"
)

func balancesCmd() *cobra.Command {
	var balancesCmd = &cobra.Command{
		Use:   "balances",
		Short: "Interact with balances (list...).",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	balancesCmd.AddCommand(balancesListCmd)

	return balancesCmd
}

var balancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all balances.",
	Run: func(cmd *cobra.Command, args []string) {
		state, err := database.NewStateFromDisk(genesis_path, tx_db_path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer state.Close()

		keys := make([]database.Account, 0, len(state.Balances))
		for k := range state.Balances {
			keys = append(keys, k)
		}

		sort.SliceStable(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		fmt.Printf("Accounts balances at %x:\n", state.LatestSnapshot())
		fmt.Println("__________________")
		fmt.Println("")
		for _, account := range keys {
			fmt.Println(fmt.Sprintf("%s: %d", account, state.Balances[account]))
		}
	},
}
