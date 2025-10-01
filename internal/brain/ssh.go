package brain

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

const (
	NEW  string = "new"
	LIST string = "list"
	EDIT string = "edit"
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
				if !strings.HasSuffix(relPath, ".md") {
					wish.Println(s, "ERROR: brainfile path must end with \".md\"")
					return
				}

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

				newFilePath := path.Join(b.config.ContentDir, relPath)
				err = os.MkdirAll(path.Dir(newFilePath), 0770)
				if err != nil {
					wish.Printf(s, "ERROR: %s\n", err)
					return
				}

				err = os.WriteFile(newFilePath, data, 0644)
				if err != nil {
					wish.Printf(s, "ERROR: %s\n", err)
					return
				}

				b.updater <- struct{}{}
				wish.Printf(s, "OK: saved %s\n", relPath)
				return

			case LIST:
				sb := &strings.Builder{}
				listNodes(sb, b.tree.nodes)
				wish.Print(s, sb)

			case EDIT:
				relPath := filepath.Clean(s.Command()[1])
				node, err := b.tree.Find(relPath)
				if err != nil {
					wish.Printf(s, "ERROR: %s: %s", err, relPath)
					return
				}

				wish.Print(s, string(node.Raw))
				return
			}
		}
		next(s)
	}
}
