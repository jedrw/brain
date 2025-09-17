package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/jedrw/brain/server"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// TODO make this better, read from config?
func publicKeyAuth() (ssh.AuthMethod, error) {
	keyPath := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
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
			// TODO prompt for name
			return errors.New("not implimented")
		}

		newFileName := args[0]
		editor, editorSet := os.LookupEnv("EDITOR")
		if !editorSet {
			editor = "nano"
		}

		// TODO: save to tempfile first, and use this for editing, cleanup after.
		editorArgs := []string{}
		if editor == "code" {
			editorArgs = append(editorArgs, "--wait")
		}
		editorArgs = append(editorArgs, newFileName)
		editorCmd := exec.Command(editor, editorArgs...)

		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		editorCmd.Stdin = os.Stdin
		err := editorCmd.Run()
		if err != nil {
			return err
		}

		authMethod, err := publicKeyAuth()
		if err != nil {
			return err
		}

		config := &ssh.ClientConfig{
			Auth: []ssh.AuthMethod{authMethod},
			// TODO: use ssh.FixedHostKey, read Hostkey from config
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		remote := net.JoinHostPort(address, strconv.Itoa(port))
		con, err := ssh.Dial("tcp", remote, config)
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
		out, err := sess.CombinedOutput(fmt.Sprintf("%s %s", server.UPLOAD, newFileName))
		if err != nil {
			return err
		}

		fmt.Println(string(out))
		return nil
	},
}

func init() {

}
