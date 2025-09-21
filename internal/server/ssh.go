package server

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
)

func NewSSHServer(hostKeyPath string, authorizedKeys []string, sshHandler func(s ssh.Handler) ssh.Handler) (*ssh.Server, error) {
	_, err := os.Stat(hostKeyPath)
	if err != nil {
		return nil, err
	}

	return wish.NewServer(
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(func(_ ssh.Context, key ssh.PublicKey) bool {
			for _, pubkey := range authorizedKeys {
				parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubkey))
				if err != nil {
					log.Warn("failed to parse authorized key", "key", pubkey, "err", err)
				}
				if ssh.KeysEqual(key, parsed) {
					return true
				}
			}

			return false
		}),
		wish.WithMiddleware(
			sshHandler,
			logging.Middleware(),
		),
	)
}
