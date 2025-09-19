package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/jedrw/brain/internal/brain"
	"github.com/jedrw/brain/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// TODO: read from config?
func publicKeyAuth(keyPath string) (ssh.AuthMethod, error) {
	if keyPath == "" {
		keyPath = path.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
	}

	buf, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "New brainfile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("not implimented: prompt for name")
		}

		config, err := config.New(configPath, cmd.Flags())
		if err != nil {
			return err
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
		err = editorCmd.Run()
		if err != nil {
			return err
		}

		authMethod, err := publicKeyAuth(config.KeyPath)
		if err != nil {
			return err
		}

		conConfig := &ssh.ClientConfig{
			Auth: []ssh.AuthMethod{authMethod},
			// TODO: use ssh.FixedHostKey, read Hostkey from config
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		remote := net.JoinHostPort(config.Address, strconv.Itoa(config.Port))
		con, err := ssh.Dial("tcp", remote, conConfig)
		if err != nil {
			return err
		}

		sess, err := con.NewSession()
		if err != nil {
			return err
		}
		defer con.Close()

		f, err := os.Open(newFileName)
		if err != nil {
			return err
		}
		defer f.Close()

		sess.Stdin = f
		out, err := sess.CombinedOutput(fmt.Sprintf("%s %s", brain.UPLOAD, newFilePath))
		if err != nil {
			return err
		}

		fmt.Print(string(out))
		return nil
	},
}

func init() {

}
