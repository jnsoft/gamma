package main

import (
	"fmt"
	"os"

	"github.com/jnsoft/gamma/util/fs"
	"github.com/spf13/cobra"
)

const (
	genesis_path = "/workspaces/gamma/database"
	tx_db_path   = "/workspaces/gamma/database"
)

const flagKeystoreFile = "keystore"
const flagDataDir = "datadir"
const flagMiner = "miner"
const flagSSLEmail = "ssl-email"
const flagDisableSSL = "disable-ssl"
const flagIP = "ip"
const flagPort = "port"
const flagBootstrapAcc = "bootstrap-account"
const flagBootstrapIp = "bootstrap-ip"
const flagBootstrapPort = "bootstrap-port"

func main() {
	var tbbCmd = &cobra.Command{
		Use:   "gamma",
		Short: "The Gamma Blockchain CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	tbbCmd.AddCommand(versionCmd)
	tbbCmd.AddCommand(balancesCmd())
	tbbCmd.AddCommand(walletCmd())
	// tbbCmd.AddCommand(runCmd())

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path to your node's data dir where the DB will be/is stored")
	cmd.MarkFlagRequired(flagDataDir)
}

func addKeystoreFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagKeystoreFile, "", "Absolute path to the encrypted keystore file")
	cmd.MarkFlagRequired(flagKeystoreFile)
}

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)

	return fs.ExpandPath(dataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
