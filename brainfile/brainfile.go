package brainfile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

var markdownParser = goldmark.New(
	goldmark.WithExtensions(
		extension.Typographer,
		meta.New(),
	),
)

type Brainfile struct {
	Title    string
	Tags     []string
	Content  []byte
	Path     string
	IsDir    bool
	Children []*Brainfile
}

var ErrInvalidBrainfile = errors.New("invalid brainfile")

func NewFromFile(path string) (Brainfile, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return Brainfile{}, err
	}

	context := parser.NewContext()
	var buf bytes.Buffer
	err = markdownParser.Convert(
		fileBytes,
		&buf,
		parser.WithContext(context),
	)
	if err != nil {
		return Brainfile{}, fmt.Errorf("%w: %s", ErrInvalidBrainfile, err)
	}

	brainfile := Brainfile{
		Content: buf.Bytes(),
		IsDir:   false,
	}

	metadata := meta.Get(context)
	v, ok := metadata["Title"]
	if !ok {
		return brainfile, fmt.Errorf("%w: brainfile must contain Title frontmatter", ErrInvalidBrainfile)
	}

	brainfile.Title, ok = v.(string)
	if !ok {
		return brainfile, fmt.Errorf("%w: rainfile Title frontmatter must be a string", ErrInvalidBrainfile)
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
		brainfile.Tags = append(brainfile.Tags, strings.TrimSpace(tag))
	}

	return brainfile, nil
}
