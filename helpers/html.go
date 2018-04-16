package helpers

import (
	"path/filepath"
	"strings"

	"github.com/vlad-s/filekeep/config"
)

// StripRoot will strip the root path from the config from the provided path.
func StripRoot(path string) string {
	root := config.Get().Root
	if strings.HasSuffix(root, "/") {
		root = strings.TrimRight(root, "/")
	}
	if root != "." {
		return strings.TrimPrefix(path, root)
	}
	return path
}

// Breadcrumb is a struct for a bread crumb containing a path and a name.
type Breadcrumb struct {
	Path string
	Name string
}

// Breadcrumbs returns a slice of Breadcrumb from a string, by splitting with a separator.
func Breadcrumbs(text, separator string) (b []Breadcrumb) {
	text = StripRoot(text)
	b = append(b, Breadcrumb{"/", "~"})
	split := strings.Split(text, separator)
	var path string
	for _, v := range split {
		if v == "" || v == "." {
			continue
		}
		path += Href(v)
		b = append(b, Breadcrumb{path, v})
	}
	return
}

// Href returns the URL address for a path by stripping the root dir
func Href(path string) string {
	path = StripRoot(path)
	if len(path) == 0 || path == "." {
		return "/"
	}
	if filepath.IsAbs(path) {
		return path
	}
	return "/" + path
}
