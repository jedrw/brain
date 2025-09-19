package brain

import (
	"bytes"
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

func NewNodeFromBytes(b []byte) (Node, error) {
	context := parser.NewContext()
	var buf bytes.Buffer
	err := markdownParser.Convert(
		b,
		&buf,
		parser.WithContext(context),
	)
	if err != nil {
		return Node{}, fmt.Errorf("%w: %s", ErrInvalidBrainNode, err)
	}

	node := Node{
		Raw:     b,
		Content: buf.Bytes(),
		IsDir:   false,
	}

	metadata := meta.Get(context)
	v, ok := metadata["Title"]
	if !ok {
		return node, fmt.Errorf("%w: brain node must contain Title frontmatter", ErrInvalidBrainNode)
	}

	node.Title, ok = v.(string)
	if !ok {
		return node, fmt.Errorf("%w: brain node Title frontmatter must be a string", ErrInvalidBrainNode)
	}

	tagsRaw, ok := metadata["Tags"]
	var tags []string
	if ok {
		if s, ok := tagsRaw.([]string); ok {
			tags = s
		} else if sIface, ok := tagsRaw.([]any); ok {
			for _, e := range sIface {
				if str, ok := e.(string); ok {
					tags = append(tags, str)
				}
			}
		}
	}

	for _, tag := range tags {
		node.Tags = append(node.Tags, strings.TrimSpace(tag))
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
