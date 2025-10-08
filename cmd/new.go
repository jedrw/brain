package cmd

import (
	"bufio"
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

var file string

var newCmd = &cobra.Command{
	Use:   "new [path]",
	Short: "New brainfile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var filePath string
		if len(args) == 0 {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter name: ")
			name, _ := reader.ReadString('\n')
			filePath = name
		} else {
			filePath = args[0]
		}

		inputFile, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		var initialContent []byte
		if inputFile != "" {
			initialContent, err = os.ReadFile(inputFile)
			if err != nil {
				return err
			}
		} else {
			initialContent = brain.BrainfileTemplate
		}

		newBytes, tempFilePath, err := editor.New(filePath, initialContent)
		if err != nil {
			return err
		}

		config, err := config.New(configPath, cmd.Flags())
		if err != nil {
			return err
		}

		client, err := client.NewSSHClient(config)
		if err != nil {
			return err
		}
		defer client.Close()

		out, err := client.RunCommand(brain.NEW, bytes.NewReader(newBytes), filePath)
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
	newCmd.Flags().StringVarP(&address, config.AddressFlag, "a", config.AddressDefault, "Brain host address")
	newCmd.Flags().StringVarP(&keyPath, config.KeyPathFlag, "i", config.KeyPathDefault, "Key path")
	newCmd.Flags().StringVarP(&file, "file", "f", "", "File")
}
