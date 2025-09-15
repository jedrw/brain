package server

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/jedrw/brain/brainfile"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
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
			fmt.Fprintf(sb, "<span class=\"has-text-weight-bold\">%s</span>", n.Name)
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
			fmt.Fprintf(sb, `<li><a href="/%s"%s>%s</a></li>`, href, class, n.Name)
		}
	}
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
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		log.Error("error reading file:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	md := goldmark.New(goldmark.WithExtensions(extension.Typographer))
	var buf bytes.Buffer
	if err := md.Convert(data, &buf); err != nil {
		log.Error("failed to convert markdown:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// TODO title stuff
	rawtitle := bytes.SplitN(data, []byte("\n"), 2)[0]
	title := bytes.TrimLeft(rawtitle, "# ")
	files, err := brainfile.GetAll(h.Dir, "")
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}

	sidebar := buildSidebar(files, cleanPath)
	html := fmt.Sprintf(pageTemplate, title, sidebar, buf.String())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func NewHttpServer() *http.Server {
	httpMux := http.NewServeMux()
	httpMux.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("./styles"))))
	httpMux.Handle("/", &HTTPHandler{Dir: "./content"})
	httpServer := &http.Server{
		Handler: httpMux,
	}

	return httpServer
}
