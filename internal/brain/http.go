package brain

import (
	_ "embed"
	"fmt"
	"net/http"
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
		log.Error("TODO: generate contents page")
		http.NotFound(w, r)
		return
	}

	node, err := h.brain.tree.Find(cleanPath + ".md")
	if err != nil {
		if err == ErrNotExist {
			log.Warn(err, "path", cleanPath)
			http.NotFound(w, r)
			return
		}

		log.Warn(err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	sidebar := buildSidebar(h.brain, cleanPath)
	html := fmt.Sprintf(pageTemplate, node.Title, sidebar, node.Content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
