package fs

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"filekeep/config"
	"filekeep/helpers"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/sirupsen/logrus"
)

var (
	// ErrDirLimit gets returned by lsDir if it goes in recursion over dirLimit levels.
	ErrDirLimit = errors.New("hit directory limit")
	// ErrDirNotFound is the error if a requested directory is hidden by config, or calling os.Stat on it fails.
	ErrDirNotFound = errors.New("directory not found")
	// ErrFileNotFound is the error if a requested file is hidden by config, or calling os.Stat on it fails.
	ErrFileNotFound = errors.New("file not found")
)

const dirLimit = 2 // one directory and its children

// FileSize is a wrapper for the file size, providing a human readable form for the String result.
type FileSize int64

func (f FileSize) String() string {
	return datasize.ByteSize(f).HumanReadable()
}

// Node represents a file or a directory.
type Node struct {
	Name     string      `json:"name"`     // Name is the basename of the file or directory.
	Path     string      `json:"path"`     // Path is the relative path of the file or directory.
	Prev     string      `json:"prev"`     // Prev is the parent directory of the file or directory.
	ModTime  time.Time   `json:"mod_time"` // ModTime is the modification time.
	Mode     os.FileMode `json:"mode"`
	IsDir    bool        `json:"is_dir"` // IsDir is a flag if it's a file or directory.
	Size     FileSize    `json:"size"`   // Size is the file or directory size.
	Password string      `json:"-"`      // Password is the md5 sum of the file's password.

	// Files keeps all children files of a directory, or a slice of empty Nodes if it's a child directory
	// in order to show how many children the directory has.
	Files []*Node `json:"files,omitempty"`
	// FilesSize represents the summed size of all children files.
	FilesSize FileSize `json:"files_size,omitempty"`
	// Dirs keeps all children directories of a directory, or a slice of empty Nodes if it's a child directory
	// ir order to show how many children the directory has.
	Dirs []*Node `json:"dirs,omitempty"`
}

// JSON returns the node as a JSON encoded string.
func (n *Node) JSON() string {
	jn, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return ""
	}
	return string(jn)
}

// HasPassword returns if a node has the specified password set.
func (n *Node) HasPassword(s string) bool {
	sum := md5.Sum([]byte(s))
	return n.Password == fmt.Sprintf("%x", sum)
}

func newNode(path string, info os.FileInfo) *Node {
	n := &Node{
		Name:    filepath.Base(path),
		Path:    helpers.StripRoot(path),
		Prev:    filepath.Dir(helpers.StripRoot(path)),
		ModTime: info.ModTime(),
		Mode:    info.Mode(),
		Size:    FileSize(info.Size()),
		IsDir:   info.IsDir(),
		Files:   make([]*Node, 0),
		Dirs:    make([]*Node, 0),
	}

	if path == config.Get().Root {
		n.Name = "."
		n.Path = "."
	}

	passFile := filepath.Join(filepath.Dir(path), "."+filepath.Base(path))
	pass, err := ioutil.ReadFile(passFile)
	if err != nil {
		return n
	}

	n.Password = strings.Trim(string(pass), "\n")
	logrus.Debugf("read password %q for file %q", n.Password, n.Name)
	return n
}

func lsDir(path string, info os.FileInfo, count int) (*Node, error) {
	if count > dirLimit {
		return nil, ErrDirLimit
	}
	count++

	fd := newNode(path, info)

	ls, err := ioutil.ReadDir(path)
	if err != nil {
		logrus.Debugf("error calling ioutil.ReadDir on path %q, returning ErrDirNotFound", path)
		return nil, ErrDirNotFound
	}

	for _, fileInfo := range ls {
		filePath := filepath.Join(path, fileInfo.Name())

		if config.Get().IsHidden(filePath, fileInfo.Name()) {
			logrus.Debugf("path %q or %q is hidden, continuing lsDir loop", fileInfo.Name(), path)
			continue
		}

		file := newNode(filePath, fileInfo)
		if file.IsDir {
			subDir, err := lsDir(filePath, fileInfo, count)
			if err != nil {
				continue
			}
			fd.Dirs = append(fd.Dirs, subDir)
		} else {
			fd.FilesSize += file.Size
			fd.Files = append(fd.Files, file)
		}
	}

	return fd, nil
}

// Read checks if a path is hidden, and if not, will return its Node or an error if it fails.
func Read(path string) (fd *Node, err error) {
	if config.Get().IsHidden(path) {
		logrus.Debugf("path %q is hidden, returning ErrFileNotFound", path)
		return nil, ErrFileNotFound
	}

	info, err := os.Stat(path)
	if err != nil {
		logrus.Debugf("error calling os.Stat on path %q, returning ErrFileNotFound", path)
		return nil, ErrFileNotFound
	}

	if info.IsDir() {
		fd, err = lsDir(path, info, 0)
	} else {
		fd = newNode(path, info)
	}

	return
}
