package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xh-dev-go/hosts-control/transforms"
)

var removeCmd = &cobra.Command{
	Use:   "remove [domain]",
	Short: "Remove a domain from the hosts file",
	Long: `Finds and removes the specified domain from the hosts file.
If the line for that IP becomes empty, the entire line is removed.

Example:
  hosts-control remove my.local.app`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]

		content, err := transforms.LoadFileToString(hostsFile)
		if err != nil {
			return fmt.Errorf("error reading hosts file: %w", err)
		}

		hf, err := transforms.LoadStringToHostsFile(content)
		if err != nil {
			return fmt.Errorf("error parsing hosts file: %w", err)
		}

		if hf.RemoveDomain(domain) {
			if err := transforms.SaveHostsFile(hostsFile, hf); err != nil {
				return fmt.Errorf("error writing to hosts file: %w", err)
			}
			fmt.Printf("Successfully removed domain '%s' from %s\n", domain, hostsFile)
		} else {
			fmt.Printf("Domain '%s' not found in %s. No changes made.\n", domain, hostsFile)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
