package brain

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
)

//go:embed template/brainfile.md
var BrainfileTemplate []byte

type Node struct {
	Title    string
	Tags     []string
	Raw      []byte
	Content  []byte
	Path     string
	IsDir    bool
	Children []*Node
}

func NewNodeFromBytes(data []byte) (Node, error) {
	context := parser.NewContext()
	var htmlBytes bytes.Buffer
	err := markdownParser.Convert(
		data,
		&htmlBytes,
		parser.WithContext(context),
	)
	if err != nil {
		return Node{}, fmt.Errorf("%w: %s", ErrInvalidBrainNode, err)
	}

	node := Node{
		Raw:     data,
		Content: htmlBytes.Bytes(),
		IsDir:   false,
	}

	metadata := meta.Get(context)
	v, ok := metadata["title"]
	if !ok {
		return node, fmt.Errorf("%w: brainfile must contain \"title\" frontmatter", ErrInvalidBrainNode)
	}

	node.Title, ok = v.(string)
	if !ok {
		return node, fmt.Errorf("%w: brainfile \"title\" must be a string", ErrInvalidBrainNode)
	}

	if node.Title == "" {
		return node, fmt.Errorf("%w: brainfile \"title\" must not be an empty string", ErrInvalidBrainNode)
	}

	return node, nil
}

func NewNodeFromFile(filePath string) (Node, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return Node{}, err
	}

	return NewNodeFromBytes(fileBytes)
}

// Make this a Tree method?
func getNodes(baseDir, currentPath string) ([]*Node, error) {
	entries, err := os.ReadDir(path.Join(baseDir, currentPath))
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

			children, err := getNodes(baseDir, relPath)
			if err != nil {
				return nil, err
			}

			node.Children = children
		} else {
			filePath := path.Join(baseDir, relPath)
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
