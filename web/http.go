package web

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"filekeep/assets/css"
	"filekeep/assets/images"
	"filekeep/assets/templates"
	"filekeep/config"
	"filekeep/fs"
	"filekeep/helpers"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// NewServer returns a new http.Server after setting the routes.
func NewServer() *http.Server {
	r := httprouter.New()

	r.NotFound = http.HandlerFunc(notFoundHandler)
	r.MethodNotAllowed = http.HandlerFunc(notFoundHandler)
	r.PanicHandler = panicHandler

	r.GET("/*path", pathHandler)
	r.POST("/*path", pathHandler)

	web := config.Get().Web
	return &http.Server{
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler:           r,
		Addr:              web.String(),
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
	if _, err := fmt.Fprint(w, r.String()); err != nil {
		logrus.WithError(err).Error("couldn't print to response writer for json response")
	}
}

type staticData struct {
	CSS       template.CSS
	DarkTheme bool
}

var headerData = staticData{
	CSS: template.CSS(css.HackCSS + css.HackDarkCSS + css.CustomCSS),
}

var funcMap = template.FuncMap{
	"breadcrumbs": helpers.Breadcrumbs,
	"href":        helpers.Href,
}

func panicHandler(w http.ResponseWriter, r *http.Request, i interface{}) {
	logrus.Errorf("caught panic on %s %q: %v", r.Method, r.URL.Path, i)
	res := httpResponse{true, "caught panic", i}
	res.JSON(http.StatusInternalServerError, w)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	pageIncomplete := false
	headerDataCopy := &headerData
	headerDataCopy.DarkTheme = r.Context().Value("dark-theme").(bool)
	buffer := bytes.NewBufferString("")

	if err := headerTpl.Execute(buffer, headerDataCopy); err != nil {
		pageIncomplete = true
	}

	if err := notFoundTpl.Execute(buffer, nil); err != nil {
		pageIncomplete = true
	}

	if err := footerTpl.Execute(buffer, nil); err != nil {
		pageIncomplete = true
	}

	w.WriteHeader(http.StatusNotFound)
	if _, err := fmt.Fprint(w, buffer); err != nil {
		pageIncomplete = true
	}

	if pageIncomplete {
		res := httpResponse{Error: true, Message: "couldn't execute template"}
		res.JSON(http.StatusInternalServerError, w)
		return
	}
}

func handleAsset(w http.ResponseWriter, r *http.Request, path string) bool {
	switch path {
	case "/favicon.ico":
		ico, err := base64.StdEncoding.DecodeString(images.FaviconICO)
		if err != nil {
			return false
		}
		w.Header().Set("Content-Type", "image/x-icon; charset=binary")
		if _, err := fmt.Fprint(w, string(ico)); err != nil {
			return false
		}
		return true
	case "/about":
		templateHandler(w, r, aboutTpl, nil)
		return true
	case "/_toggleTheme":
		var darkTheme bool
		themeCookie, err := r.Cookie("dark-theme")
		if err == nil {
			darkTheme, _ = strconv.ParseBool(themeCookie.Value)
		}

		// toggle the theme
		darkTheme = !darkTheme

		cookie := &http.Cookie{
			Name:    "dark-theme",
			Value:   strconv.FormatBool(darkTheme),
			Path:    "/",
			Expires: time.Now().Add(7 * 24 * time.Hour),
		}

		http.SetCookie(w, cookie)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusFound)
		return true
	default:
		return false
	}
}

func pathHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	path := filepath.Clean(ps.ByName("path"))
	if handleAsset(w, r, path) {
		return
	}

	ctx := r.Context()
	cookie, err := r.Cookie("dark-theme")
	if err == nil {
		value, _ := strconv.ParseBool(cookie.Value)
		ctx = context.WithValue(ctx, cookie.Name, value)
	}

	r = r.WithContext(ctx)

	if filepath.IsAbs(path) {
		path = path[1:]
	}

	path = filepath.Join(config.Get().Root, path)

	fd, err := fs.Read(path)
	if err != nil {
		notFoundHandler(w, r)
		return
	}

	if !checkPass(fd, w, r) {
		return
	}

	q := r.URL.Query()
	if _, ok := q["json"]; ok {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if _, err := fmt.Fprint(w, fd.JSON()); err != nil {
			logrus.WithError(err).Error("couldn't print to response writer for json query")
		}
		return
	}

	if fd.IsDir {
		templateHandler(w, r, dirListTpl, fd)
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
