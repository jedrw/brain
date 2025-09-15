package server

import (
	"log"
	"os"
	"path"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

func NewSSHServer(hostKeyPath string) (*ssh.Server, error) {
	if hostKeyPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("could not find homedir")
		}

		hostKeyPath = path.Join(homeDir, ".ssh", "id_rsa")
	}

	_, err := os.Stat(hostKeyPath)
	if err != nil {
		return nil, err
	}

	return wish.NewServer(
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					wish.Println(sess, "It works SSH!")
					next(sess)
				}
			}),
	)
}
