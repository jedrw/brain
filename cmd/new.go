package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/jedrw/brain/internal/brain"
	"github.com/jedrw/brain/internal/client"
	"github.com/jedrw/brain/internal/config"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "New brainfile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("not implimented: prompt for name")
		}

		// TODO: save to tempfile first, and use this for editing, cleanup after.
		newFilePath := args[0]
		newFileName := path.Base(newFilePath)
		editor, editorSet := os.LookupEnv("EDITOR")
		if !editorSet {
			editor = "vi"
		}

		editorCmd := exec.Command(editor, newFileName)
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		editorCmd.Stdin = os.Stdin
		err := editorCmd.Run()
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

		f, err := os.Open(newFileName)
		if err != nil {
			return err
		}
		defer f.Close()

		out, err := client.RunCommand(brain.NEW, f, newFilePath)
		if err != nil {
			return err
		}

		fmt.Print(string(out))
		return nil
	},
}
