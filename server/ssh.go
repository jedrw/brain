package server

import (
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
)

const (
	UPLOAD string = "upload"
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
				return func(s ssh.Session) {
					if len(s.Command()) > 0 && s.Command()[0] == UPLOAD {
						relPath := s.Command()[1]
						data, err := io.ReadAll(s)
						if err != nil {
							wish.Println(s, "ERR:", err)
							return
						}

						if err := os.WriteFile(path.Join(absContentDir, relPath), data, 0644); err != nil {
							wish.Println(s, "ERR:", err)
							return
						}

						wish.Println(s, "OK")
						return
					}
					// fallback...
				}
			},
			logging.Middleware(),
		),
	)
}
