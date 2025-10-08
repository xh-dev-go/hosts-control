package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xh-dev-go/hosts-control/transforms"
)

var comment string

var addCmd = &cobra.Command{
	Use:   "add [ip] [domain]",
	Short: "Add or update a domain in the hosts file",
	Long: `Adds a new domain-to-IP mapping. If the domain already exists, it will be updated to the new IP.
If the IP already exists, the domain will be added to that line.

Example:
  hosts-control add 127.0.0.1 my.local.app --comment "Local dev"`,
	Args: cobra.ExactArgs(2), // Ensures we get exactly two arguments
	RunE: func(cmd *cobra.Command, args []string) error {
		ip := args[0]
		domain := args[1]

		content, err := transforms.LoadFileToString(hostsFile)
		if err != nil {
			return fmt.Errorf("error reading hosts file: %w", err)
		}

		hf, err := transforms.LoadStringToHostsFile(content)
		if err != nil {
			return fmt.Errorf("error parsing hosts file: %w", err)
		}

		changed, err := hf.AddDomain(ip, domain, comment)
		if err != nil {
			return err
		}

		if changed {
			if err := transforms.SaveHostsFile(hostsFile, hf); err != nil {
				return fmt.Errorf("error writing to hosts file: %w", err)
			}
			fmt.Printf("Successfully added/updated domain '%s' for IP %s in %s\n", domain, ip, hostsFile)
		} else {
			fmt.Printf("No changes needed for domain '%s' with IP %s.\n", domain, ip)
		}
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&comment, "comment", "c", "", "Add a comment to the entry")
	rootCmd.AddCommand(addCmd)
}
