package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xh-dev-go/hosts-control/transforms"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all IP-to-domain mappings in a tree format",
	Long:  `Reads the specified hosts file and prints a structured tree view of all active IP address to domain mappings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := transforms.LoadFileToString(hostsFile)
		if err != nil {
			return fmt.Errorf("error reading hosts file: %w", err)
		}

		hf, err := transforms.LoadStringToHostsFile(content)
		if err != nil {
			return fmt.Errorf("error parsing hosts file: %w", err)
		}

		fmt.Printf("Displaying entries from %s\n\n", hostsFile)
		for _, line := range hf.Lines {
			if line.IsEntry && len(line.Domains) > 0 {
				fmt.Println(line.IP)
				for _, domain := range line.Domains {
					fmt.Printf("  - %s\n", domain)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
