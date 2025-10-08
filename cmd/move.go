package cmd

import (
	"fmt"

	"github.com/jedrw/brain/internal/brain"
	"github.com/jedrw/brain/internal/client"
	"github.com/jedrw/brain/internal/config"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:     "move [from-path] [to-path]",
	Aliases: []string{"mv"},
	Short:   "Move a brainfile",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		from := args[0]
		to := args[1]

		client, err := client.NewSSHClient(brainConfig)
		if err != nil {
			return err
		}
		defer client.Close()

		out, err := client.RunCommand(brain.MOVE, nil, from, to)
		if err != nil {
			return err
		}

		fmt.Print(out)
		return nil
	},
}

func init() {
	moveCmd.Flags().StringVarP(&address, config.AddressFlag, "a", config.AddressDefault, "Brain host address")
	moveCmd.Flags().StringVarP(&keyPath, config.KeyPathFlag, "i", config.KeyPathDefault, "Key path")
}
