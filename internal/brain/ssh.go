package brain

import (
	"io"
	"os"
	"path"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

const (
	UPLOAD string = "upload"
)

func (b *Brain) sshHandler(next ssh.Handler) ssh.Handler {
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
		}
		next(s)
	}
}
