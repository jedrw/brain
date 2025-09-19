package brain

import (
	"io"
	"os"
	"path"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
)

const (
	UPLOAD string = "upload"
)

func NewSSHServer(b *Brain) (*ssh.Server, error) {
	_, err := os.Stat(b.config.HostKeyPath)
	if err != nil {
		return nil, err
	}

	return wish.NewServer(
		wish.WithHostKeyPath(b.config.HostKeyPath),
		wish.WithPublicKeyAuth(func(_ ssh.Context, key ssh.PublicKey) bool {
			for _, pubkey := range b.config.AuthorizedKeys {
				parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubkey))
				if err != nil {
					log.Warn("failed to parse authorized key", "key", pubkey)
				}
				if ssh.KeysEqual(key, parsed) {
					return true
				}
			}

			return false
		}),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(s ssh.Session) {
					if len(s.Command()) > 0 {
						switch s.Command()[0] {
						case UPLOAD:
							relPath := s.Command()[1]
							data, err := io.ReadAll(s)
							if err != nil {
								wish.Println(s, "ERR:", err)
								return
							}

							node, err := NewNodeFromBytes(data)
							if err != nil {
								wish.Println(s, "ERR:", err)
								return
							}

							if err := os.WriteFile(path.Join(b.config.ContentDir, relPath), node.Raw, 0644); err != nil {
								wish.Println(s, "ERR:", err)
								return
							}

							b.updater <- struct{}{}
							wish.Println(s, "OK")
							return

						}
						// fallback...
					}
				}
			},
			logging.Middleware(),
		),
	)
}
