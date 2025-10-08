package brain

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
)

type Tree struct {
	mu    sync.RWMutex
	nodes []*Node
}

func (t *Tree) Find(path string) (*Node, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	parts := strings.Split(path, string(os.PathSeparator))
	currentLevel := t.nodes
	var found *Node
	for _, part := range parts {
		if part == "" {
			continue
		}

		found = nil
		for _, n := range currentLevel {
			base := filepath.Base(n.Path)
			if base == part {
				found = n
				break
			}
		}

		if found == nil {
			return nil, ErrNotExist
		}

		currentLevel = found.Children
	}

	if found.IsDir {
		return found, ErrNodeIsDir
	}

	return found, nil
}

func (t *Tree) getNodes(baseDir, currentPath string) ([]*Node, error) {
	entries, err := os.ReadDir(filepath.Join(baseDir, currentPath))
	if err != nil {
		return nil, err
	}

	var nodes []*Node
	for _, entry := range entries {
		relPath := filepath.Join(currentPath, entry.Name())
		var node Node
		if entry.IsDir() {
			node = Node{
				Title: strings.TrimSuffix(entry.Name(), ".md"),
				Path:  relPath,
				IsDir: entry.IsDir(),
			}

			children, err := t.getNodes(baseDir, relPath)
			if err != nil {
				return nil, err
			}

			node.Children = children
		} else {
			filePath := filepath.Join(baseDir, relPath)
			node, err = NewNodeFromFile(filePath)
			if err != nil {
				if errors.Is(err, ErrInvalidBrainNode) {
					log.Warn("invalid brain node", "path", relPath)
					continue
				}

				return nil, err
			}

			node.Path = relPath
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}
