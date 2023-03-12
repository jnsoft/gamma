package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jnsoft/gamma/database"
	"github.com/jnsoft/gamma/node"
	"github.com/spf13/cobra"
)

func balancesCmd() *cobra.Command {
	var balancesCmd = &cobra.Command{
		Use:   "balances",
		Short: "Interacts with balances (list...).",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	balancesCmd.AddCommand(balancesListCmd())

	return balancesCmd
}

func balancesListCmd() *cobra.Command {
	var balancesListCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all balances.",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd), node.DefaultMiningDifficulty)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			// sort keys from hashset
			keys := make([]common.Address, 0, len(state.Balances))
			for k := range state.Balances {
				keys = append(keys, k)
			}
			sort.SliceStable(keys, func(i, j int) bool {
				return keys[i].Hex() < keys[j].Hex()
			})

			fmt.Printf("Accounts balances at %x:\n", state.LatestBlockHash())
			fmt.Println("__________________")
			fmt.Println("")
			//for account, balance := range state.Balances {
			//	fmt.Println(fmt.Sprintf("%s: %d", account.String(), balance))
			//}
			for _, account := range keys {
				fmt.Printf("%s: %d\n", account.String(), state.Balances[account])
			}
			fmt.Println("")
			fmt.Printf("Accounts nonces:")
			fmt.Println("")
			fmt.Println("__________________")
			fmt.Println("")
			for account, nonce := range state.Account2Nonce {
				fmt.Printf("%s: %d\n", account.String(), nonce)
			}
		},
	}

	addDefaultRequiredFlags(balancesListCmd)

	return balancesListCmd
}
