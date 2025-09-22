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
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.New(configPath, cmd.Flags())
		if err != nil {
			return err
		}

		client, err := client.NewSSHClient(config)
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
