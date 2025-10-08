package client

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jedrw/brain/internal/config"
	"golang.org/x/crypto/ssh"
)

type sshClient struct {
	con *ssh.Client
}

func publicKeyAuth(keyPath string) (ssh.AuthMethod, error) {
	if keyPath == "" {
		keyPath = filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519")
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

func NewSSHClient(config config.Config) (*sshClient, error) {
	authMethod, err := publicKeyAuth(config.KeyPath)
	if err != nil {
		return nil, err
	}

	conConfig := &ssh.ClientConfig{
		Auth: []ssh.AuthMethod{authMethod},
		// TODO: use ssh.FixedHostKey, read Hostkey from config
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	remote := net.JoinHostPort(config.Address, strconv.Itoa(config.Port))
	con, err := ssh.Dial("tcp", remote, conConfig)
	if err != nil {
		return nil, err
	}

	return &sshClient{
		con: con,
	}, nil
}

func (c *sshClient) Close() error {
	return c.con.Close()
}

func (c *sshClient) RunCommand(command string, in io.Reader, args ...string) (string, error) {
	sess, err := c.con.NewSession()
	if err != nil {
		return "", err
	}

	if in != nil {
		sess.Stdin = in
	}

	out, err := sess.Output(strings.Join(append([]string{command}, args...), " "))
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}
