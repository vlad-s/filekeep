package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/vlad-s/filekeep/assets/css"
	"github.com/vlad-s/filekeep/assets/templates"
	"github.com/vlad-s/filekeep/config"
	"github.com/vlad-s/filekeep/fs"
	"github.com/vlad-s/filekeep/helpers"
)

// NewServer returns a new http.Server after setting the routes.
func NewServer() *http.Server {
	r := httprouter.New()
	r.NotFound = http.HandlerFunc(notFoundHandler)
	r.GET("/*path", pathHandler)

	listen := config.Get().Listen
	return &http.Server{
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler:           r,
		Addr:              fmt.Sprintf("%s:%d", listen.Addr, listen.Port),
	}
}

type httpResponse struct {
	Error   bool   `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Raw     string `json:"raw,omitempty"`
}

func (r *httpResponse) String() string {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

func (r *httpResponse) JSON(c int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(c)
	fmt.Fprint(w, r.String())
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	t, err := template.New("404").Parse(templates.HTMLHeader + templates.HTMLFooter + templates.HTML404)
	if err != nil {
		res := httpResponse{true, "couldn't parse template", err.Error()}
		res.JSON(http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	if err := t.Execute(w, nil); err != nil {
		res := httpResponse{true, "couldn't execute template", err.Error()}
		res.JSON(http.StatusInternalServerError, w)
		return
	}
}

func listHandler(w http.ResponseWriter, n *fs.Node) {
	fm := template.FuncMap{
		"breadcrumbs": helpers.Breadcrumbs,
		"href":        helpers.Href,
	}

	tpls := templates.HTMLHeader + templates.HTMLFooter + templates.HTMLDirList
	t, err := template.New("dir").Funcs(fm).Parse(tpls)
	if err != nil {
		res := httpResponse{true, "couldn't parse template", err.Error()}
		res.JSON(http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err = t.Execute(w, n); err != nil {
		res := httpResponse{true, "couldn't execute template", err.Error()}
		res.JSON(http.StatusInternalServerError, w)
		return
	}
}

func handleAsset(w http.ResponseWriter, _ *http.Request, path string) bool {
	switch path {
	case string(os.PathSeparator) + filepath.Join("assets", "css", "hack.css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		fmt.Fprint(w, css.HackCSS)
		return true
	case string(os.PathSeparator) + filepath.Join("assets", "css", "custom.css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		fmt.Fprint(w, css.CustomCSS)
		return true
	case string(os.PathSeparator) + "about":
		aboutHandler(w)
		return true
	default:
		return false
	}
}

func aboutHandler(w http.ResponseWriter) {
	tpls := templates.HTMLHeader + templates.HTMLFooter + templates.HTMLAbout
	t, err := template.New("about").Parse(tpls)
	if err != nil {
		res := httpResponse{true, "couldn't parse template", err.Error()}
		res.JSON(http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err = t.Execute(w, nil); err != nil {
		res := httpResponse{true, "couldn't execute template", err.Error()}
		res.JSON(http.StatusInternalServerError, w)
		return
	}
}

func pathHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	path := filepath.Clean(ps.ByName("path"))

	if handleAsset(w, r, path) {
		return
	}

	if filepath.IsAbs(path) {
		path = path[1:]
	}

	path = filepath.Join(config.Get().Root, path)

	fd, err := fs.Read(path)
	if err != nil {
		notFoundHandler(w, nil)
		return
	}

	if fd.IsDir {
		listHandler(w, fd)
	} else {
		http.ServeFile(w, r, path)
	}
}
