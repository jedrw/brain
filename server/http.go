package server

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/jedrw/brain/brainfile"
)

//go:embed templates/main.html
var pageTemplate string

func buildSidebar(nodes []*brainfile.Brainfile, currentPath string) string {
	var sb strings.Builder
	sb.WriteString(`<aside class="menu is-flex-grow-0 pt-4" style="width: 220px;">`)
	sb.WriteString(`<ul class="menu-list">`)
	populateSidebar(&sb, nodes, currentPath)
	sb.WriteString(`</ul>`)
	sb.WriteString(`</aside>`)
	return sb.String()
}

func populateSidebar(sb *strings.Builder, nodes []*brainfile.Brainfile, currentPath string) {
	for _, n := range nodes {
		if n.IsDir {
			sb.WriteString(`<li>`)
			fmt.Fprintf(sb, "<span class=\"has-text-weight-bold\">%s</span>", n.Title)
			if len(n.Children) > 0 {
				sb.WriteString(`<ul>`)
				populateSidebar(sb, n.Children, currentPath)
				sb.WriteString(`</ul>`)
			}
			sb.WriteString(`</li>`)
		} else {
			href := strings.TrimSuffix(n.Path, ".md")
			class := ""
			if "/"+href == currentPath {
				class = ` class="is-active"`
			}
			fmt.Fprintf(sb, `<li><a href="/%s"%s>%s</a></li>`, href, class, n.Title)
		}
	}
}

func getBrainfiles(rootDir, currentPath string) ([]*brainfile.Brainfile, error) {
	fullPath := filepath.Join(rootDir, currentPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var nodes []*brainfile.Brainfile
	for _, entry := range entries {
		relPath := filepath.Join(currentPath, entry.Name())
		var node brainfile.Brainfile
		if entry.IsDir() {
			node = brainfile.Brainfile{
				Title: strings.TrimSuffix(entry.Name(), ".md"),
				Path:  relPath,
				IsDir: entry.IsDir(),
			}

			children, err := getBrainfiles(rootDir, relPath)
			if err != nil {
				return nil, err
			}

			node.Children = children
		} else {
			path := filepath.Join(rootDir, relPath)
			node, err = brainfile.NewFromFile(path)
			if err != nil {
				if errors.Is(err, brainfile.ErrInvalidBrainfile) {
					log.Warn("invalid brainfile", "path", path)
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

type HTTPHandler struct {
	Dir string
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cleanPath := filepath.Clean(r.URL.Path)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	// TODO: generate contents page
	if cleanPath == "/" {
		cleanPath = "/index"
		log.Info("TODO: generate contents page")
	}

	filePath := filepath.Join(h.Dir, strings.TrimPrefix(cleanPath, "/")+".md")
	file, err := brainfile.NewFromFile(filePath)
	if err != nil {
		log.Error(err)
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	files, err := getBrainfiles(h.Dir, "")
	if err != nil {
		log.Error(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	sidebar := buildSidebar(files, cleanPath)
	html := fmt.Sprintf(pageTemplate, file.Title, sidebar, file.Content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func NewHttpServer(contentDir string) *http.Server {
	httpMux := http.NewServeMux()
	httpMux.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("./styles"))))
	httpMux.Handle("/", &HTTPHandler{Dir: contentDir})
	httpServer := &http.Server{
		Handler: httpMux,
	}

	return httpServer
}
