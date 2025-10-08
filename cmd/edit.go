package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/jedrw/brain/internal/brain"
	"github.com/jedrw/brain/internal/client"
	"github.com/jedrw/brain/internal/config"
	"github.com/jedrw/brain/internal/editor"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [path]",
	Short: "Edit a brainfile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		client, err := client.NewSSHClient(brainConfig)
		if err != nil {
			return err
		}
		defer client.Close()

		out, err := client.RunCommand(brain.EDIT, nil, filePath)
		if err != nil {
			return err
		}

		if strings.HasPrefix(out, "ERROR") {
			fmt.Println(out)
			return nil
		}

		editedBytes, tempFilePath, err := editor.New(filePath, []byte(out))
		if err != nil {
			return err
		}

		out, err = client.RunCommand(brain.NEW, bytes.NewReader(editedBytes), filePath)
		if err != nil {
			return err
		}

		fmt.Print(out)
		if strings.Contains(out, brain.ErrInvalidBrainNode.Error()) {
			fmt.Printf("temp brainfile saved at: %s\n", tempFilePath)
		} else {
			os.Remove(tempFilePath)
		}

		return nil
	},
}

func init() {
	editCmd.Flags().StringVarP(&address, config.AddressFlag, "a", config.AddressDefault, "Brain host address")
	editCmd.Flags().StringVarP(&keyPath, config.KeyPathFlag, "i", config.KeyPathDefault, "Key path")
}
