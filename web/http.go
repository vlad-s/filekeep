package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/vlad-s/filekeep/assets/css"
	"github.com/vlad-s/filekeep/assets/images"
	"github.com/vlad-s/filekeep/assets/templates"
	"github.com/vlad-s/filekeep/config"
	"github.com/vlad-s/filekeep/fs"
	"github.com/vlad-s/filekeep/helpers"
)

// NewServer returns a new http.Server after setting the routes.
func NewServer() *http.Server {
	r := httprouter.New()

	r.NotFound = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowed = http.HandlerFunc(notFoundHandler)
	r.PanicHandler = panicHandler

	r.GET("/*path", pathHandler)
	r.POST("/*path", pathHandler)

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
	Error   bool        `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Raw     interface{} `json:"raw,omitempty"`
}

// String returns the response as a JSON encoded string.
func (r *httpResponse) String() string {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// JSON sets the content type as application/json, writes the specified status code, and prints the response.
func (r *httpResponse) JSON(code int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprint(w, r.String())
}

func panicHandler(w http.ResponseWriter, r *http.Request, i interface{}) {
	fs.Log.Errorf("Caught panic on %s %q: %v", r.Method, r.URL.Path, i)
	res := httpResponse{true, "caught panic", i}
	res.JSON(http.StatusInternalServerError, w)
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

func handleAsset(w http.ResponseWriter, path string) bool {
	switch path {
	case "/assets/css/hack.css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		fmt.Fprint(w, css.HackCSS)
		return true
	case "/assets/css/custom.css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		fmt.Fprint(w, css.CustomCSS)
		return true
	case "/favicon.ico":
		ico, err := base64.StdEncoding.DecodeString(images.FaviconICO)
		if err != nil {
			return false
		}
		w.Header().Set("Content-Type", "image/x-icon; charset=binary")
		fmt.Fprint(w, string(ico))
		return true
	case "/about":
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

	if handleAsset(w, path) {
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

	if !checkPass(fd, w, r) {
		return
	}

	q := r.URL.Query()
	if _, ok := q["json"]; ok {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, fd.JSON())
		return
	}

	if fd.IsDir {
		listHandler(w, fd)
	} else {
		http.ServeFile(w, r, path)
	}
}

func passFormHandler(n *fs.Node, w http.ResponseWriter) {
	tpls := templates.HTMLHeader + templates.HTMLFooter + templates.HTMLPassForm
	t, err := template.New("pass").Parse(tpls)
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

func checkPass(n *fs.Node, w http.ResponseWriter, r *http.Request) bool {
	if n.Password == "" {
		return true
	}

	if r.Method == "GET" {
		passFormHandler(n, w)
		return false
	}

	if !n.HasPassword(r.FormValue("password")) {
		passFormHandler(n, w)
		return false
	}

	return true
}
