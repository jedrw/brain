package brain

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
)

const (
	NEW    string = "new"
	LIST   string = "list"
	EDIT   string = "edit"
	MOVE   string = "move"
	DELETE string = "delete"
)

func listNodes(sb *strings.Builder, nodes []*Node) {
	for _, node := range nodes {
		if node.IsDir {
			listNodes(sb, node.Children)
		} else {
			fmt.Fprintf(sb, "%s\n", node.Path)
		}
	}
}

func isEmptyDir(name string) (bool, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return false, err
	}

	if !fi.IsDir() {
		return false, fmt.Errorf("%s is not a directory", name)
	}

	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err
}

func removeEmptyDirs(baseDir, relPath string) error {
	if relPath == baseDir {
		return nil
	}

	path := filepath.Join(baseDir, relPath)
	empty, err := isEmptyDir(path)
	if err != nil {
		return err
	}

	if empty {
		err = os.Remove(path)
		if err != nil {
			return err
		}

		parentPathRel := filepath.Dir(relPath)
		if parentPathRel == path {
			return nil
		}

		return removeEmptyDirs(baseDir, parentPathRel)
	}

	return nil
}

func (b *Brain) sshHandler(next ssh.Handler) ssh.Handler {
	return func(s ssh.Session) {
		if len(s.Command()) > 0 {
			switch s.Command()[0] {
			case NEW:
				relPath := s.Command()[1]
				if !strings.HasSuffix(relPath, ".md") {
					log.Error("brainfile path must end with \".md\"")
					wish.Println(s, "ERROR: brainfile path must end with \".md\"")
					break
				}

				data, err := io.ReadAll(s)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				// Validate the data
				_, err = NewNodeFromBytes(data)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				newFilePath := filepath.Join(b.config.ContentDir, relPath)
				err = os.MkdirAll(filepath.Dir(newFilePath), 0770)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				err = os.WriteFile(newFilePath, data, 0644)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				b.updater <- struct{}{}
				log.Infof("saved %s", relPath)
				wish.Printf(s, "OK: saved %s\n", relPath)

			case LIST:
				sb := &strings.Builder{}
				listNodes(sb, b.tree.nodes)
				log.Info("listed nodes")
				wish.Print(s, sb)

			case EDIT:
				relPath := filepath.Clean(s.Command()[1])
				node, err := b.tree.Find(relPath)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				log.Infof("sent %s for editing", relPath)
				wish.Print(s, string(node.Raw))

			case MOVE:
				fromPathRel := filepath.Clean(s.Command()[1])
				toPathRel := filepath.Clean(s.Command()[2])
				if fromPathRel == toPathRel {
					log.Infof("moved %s to %s", fromPathRel, toPathRel)
					wish.Printf(s, "OK: moved %s to %s\n", fromPathRel, toPathRel)
					break
				}

				fromPath := filepath.Join(b.config.ContentDir, fromPathRel)
				toPath := filepath.Join(b.config.ContentDir, toPathRel)

				_, err := os.Stat(toPath)
				if err == nil {
					log.Errorf("%s already exists, move must be non-destructive", toPathRel)
					wish.Printf(s, "ERROR: %s already exists, move must be non-destructive\n", toPathRel)
					b.updater <- struct{}{}
					break
				}

				if !os.IsNotExist(err) {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				fromNode, err := b.tree.Find(fromPathRel)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					b.updater <- struct{}{}
					break
				}

				err = os.MkdirAll(path.Dir(toPath), 0770)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				err = os.WriteFile(toPath, fromNode.Raw, 0644)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				err = os.Remove(fromPath)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				fromDirRel := filepath.Dir(fromPathRel)
				err = removeEmptyDirs(b.config.ContentDir, fromDirRel)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				b.updater <- struct{}{}
				log.Infof("moved %s to %s", fromPathRel, toPathRel)
				wish.Printf(s, "OK: moved %s to %s\n", fromPathRel, toPathRel)

			case DELETE:
				relPath := filepath.Clean(s.Command()[1])
				path := filepath.Join(b.config.ContentDir, relPath)
				err := os.Remove(path)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				fromDirRel := filepath.Dir(relPath)
				err = removeEmptyDirs(b.config.ContentDir, fromDirRel)
				if err != nil {
					log.Error(err)
					wish.Printf(s, "ERROR: %s\n", err)
					break
				}

				b.updater <- struct{}{}
				log.Infof("deleted %s", relPath)
				wish.Printf(s, "OK: deleted %s\n", relPath)
			}
		}
		next(s)
	}
}
