package cmd

import (
	"fmt"

	"github.com/jedrw/brain/internal/brain"
	"github.com/jedrw/brain/internal/client"
	"github.com/jedrw/brain/internal/config"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [path]",
	Short: "Delete a brainfile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := client.NewSSHClient(brainConfig)
		if err != nil {
			return err
		}
		defer client.Close()

		out, err := client.RunCommand(brain.DELETE, nil, args[0])
		if err != nil {
			return err
		}

		fmt.Print(out)
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&address, config.AddressFlag, "a", config.AddressDefault, "Brain host address")
	deleteCmd.Flags().StringVarP(&keyPath, config.KeyPathFlag, "i", config.KeyPathDefault, "Key path")
}
