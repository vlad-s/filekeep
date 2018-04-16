package helpers

import (
	"testing"

	"github.com/vlad-s/filekeep/config"
)

func TestBreadcrumbs(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		sep      string
		expected []Breadcrumb
	}{
		{"correct", "foo", "/", []Breadcrumb{{Path: "/", Name: "~"}, {Path: "/foo", Name: "foo"}}},
		{"correct", "foo-bar", "-", []Breadcrumb{{Path: "/", Name: "~"}, {Path: "/foo", Name: "foo"}, {Path: "/foo/bar", Name: "bar"}}},
		{"incorrect", "foo_bar", "/", []Breadcrumb{{Path: "/", Name: "~"}, {Path: "/foo_bar", Name: "foo_bar"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := Breadcrumbs(test.text, test.sep)
			if len(b) != len(test.expected) {
				t.Error(test.name)
				return
			}
			correct := 0
			for _, v := range b {
				for _, e := range test.expected {
					if v.Name == e.Name && v.Path == e.Path {
						correct++
					}
				}
			}
			if correct != len(test.expected) {
				t.Error(test.name)
				return
			}
		})
	}
}

func TestHref(t *testing.T) {
	oldRoot := config.Get().Root
	config.Get().Root = "/var/www"

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"correct", ".", "/"},
		{"correct", "foo", "/foo"},
		{"correct", "foo/bar", "/foo/bar"},
		{"correct, stripped", "/var/www/foo/bar", "/foo/bar"},
		{"incorrect", "foo_bar", "/foo_bar"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if Href(test.path) != test.expected {
				t.Error(test.name, test.expected, Href(test.path))
			}
		})
	}

	config.Get().Root = oldRoot
}
