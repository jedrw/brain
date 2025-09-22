package brain

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

const (
	NEW  string = "upload"
	LIST string = "list"
)

func listNodes(sb *strings.Builder, nodes []*Node) {
	for _, node := range nodes {
		if node.IsDir {
			fmt.Fprintf(sb, "%s\n", node.Path)
			listNodes(sb, node.Children)
		} else {
			fmt.Fprintf(sb, "%s (%s)\n", node.Path, node.Title)
		}
	}
}

func (b *Brain) sshHandler(next ssh.Handler) ssh.Handler {
	return func(s ssh.Session) {
		if len(s.Command()) > 0 {
			switch s.Command()[0] {
			case NEW:
				relPath := s.Command()[1]
				data, err := io.ReadAll(s)
				if err != nil {
					wish.Printf(s, "ERROR: %s\n", err)
					return
				}

				// Validate the data
				_, err = NewNodeFromBytes(data)
				if err != nil {
					wish.Printf(s, "ERROR: %s\n", err)
					return
				}

				err = os.MkdirAll(path.Dir(path.Join(b.config.ContentDir, relPath)), 0770)
				if err != nil {
					wish.Printf(s, "ERROR: %s\n", err)
					return
				}

				if err := os.WriteFile(path.Join(b.config.ContentDir, relPath), data, 0644); err != nil {
					wish.Printf(s, "ERROR: %s\n", err)
					return
				}

				b.updater <- struct{}{}
				wish.Printf(s, "saved %s\n", relPath)
				return

			case LIST:
				sb := &strings.Builder{}
				listNodes(sb, b.tree.nodes)
				wish.Print(s, sb)
			}
		}
		next(s)
	}
}
