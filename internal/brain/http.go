package brain

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

//go:embed templates/main.html
var pageTemplate string

func buildSidebar(b *Brain, currentPath string) string {
	b.tree.mu.RLock()
	defer b.tree.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString(`<aside class="menu is-flex-grow-0 pt-4" style="width: 220px;">`)
	sb.WriteString(`<ul class="menu-list">`)
	populateSidebar(&sb, b.tree.nodes, currentPath)
	sb.WriteString(`</ul>`)
	sb.WriteString(`</aside>`)
	return sb.String()
}

func populateSidebar(sb *strings.Builder, tree []*Node, currentPath string) {
	for _, n := range tree {
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

type HTTPHandler struct {
	brain *Brain
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

	file, err := NewNodeFromFile(path.Join(h.brain.config.ContentDir, strings.TrimPrefix(cleanPath, "/")+".md"))
	if err != nil {
		log.Error(err)
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	sidebar := buildSidebar(h.brain, cleanPath)
	html := fmt.Sprintf(pageTemplate, file.Title, sidebar, file.Content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func NewHttpServer(b *Brain) *http.Server {
	httpMux := http.NewServeMux()
	httpMux.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("./styles"))))
	httpMux.Handle("/", &HTTPHandler{brain: b})
	httpServer := &http.Server{
		Handler: httpMux,
	}

	return httpServer
}
