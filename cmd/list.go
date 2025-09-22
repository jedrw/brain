package cmd

import (
	"fmt"

	"github.com/jedrw/brain/internal/brain"
	"github.com/jedrw/brain/internal/client"
	"github.com/jedrw/brain/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List brainfiles",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		client, err := client.NewSSHClient(brainConfig)
		if err != nil {
			return err
		}
		defer client.Close()

		out, err := client.RunCommand(brain.LIST, nil)
		if err != nil {
			return err
		}

		fmt.Print(out)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&address, config.AddressFlag, "a", config.AddressDefault, "Brain host address")
	listCmd.Flags().StringVarP(&keyPath, config.KeyPathFlag, "i", config.KeyPathDefault, "Key path")
}
