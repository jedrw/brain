package brain

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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
