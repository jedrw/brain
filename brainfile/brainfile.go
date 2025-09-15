package brainfile

import (
	"os"
	"path/filepath"
	"strings"
)

type Brainfile struct {
	Title    string
	Name     string
	Path     string // relative path from rootDir, used for reading URLs/files
	IsDir    bool
	Children []*Brainfile
}

func GetAll(rootDir, currentPath string) ([]*Brainfile, error) {
	fullPath := filepath.Join(rootDir, currentPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var nodes []*Brainfile
	for _, entry := range entries {
		entryPath := filepath.Join(currentPath, entry.Name())
		name := strings.TrimSuffix(entry.Name(), ".md")
		node := &Brainfile{
			Name:  name,
			Path:  entryPath,
			IsDir: entry.IsDir(),
			Title: "", // TODO: parse file for title
		}

		if entry.IsDir() {
			children, err := GetAll(rootDir, entryPath)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
