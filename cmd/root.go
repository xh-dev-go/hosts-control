package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	hostsFile string
)

var rootCmd = &cobra.Command{
	Use:   "hosts-control",
	Short: "A CLI tool to manage your hosts file",
	Long: `hosts-control is a powerful and easy-to-use command-line interface
for managing entries in your /etc/hosts file.

You can easily list, add, or remove domain-to-IP mappings.
Note: Modifying the hosts file typically requires administrator privileges (e.g., run with 'sudo').`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports a persistent flag, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&hostsFile, "file", "f", "/etc/hosts", "path to the hosts file")
}
