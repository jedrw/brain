package server

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/scp"
)

func NewSSHServer(contentDir string, hostKeyPath string, authorizedKeys []string) (*ssh.Server, error) {
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

	absContentDir, err := filepath.Abs(contentDir)
	if err != nil {
		return nil, err
	}

	handler := scp.NewFileSystemHandler(absContentDir)

	return wish.NewServer(
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(func(_ ssh.Context, key ssh.PublicKey) bool {
			for _, pubkey := range authorizedKeys {
				parsed, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(pubkey))
				if ssh.KeysEqual(key, parsed) {
					return true
				}
			}

			return false
		}),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					wish.Println(sess, "It works SSH!")
					next(sess)
				}
			},
			scp.Middleware(handler, handler),
			logging.Middleware(),
		),
	)
}

// func sshHandler(next ssh.Handler) ssh.Handler {

// }
