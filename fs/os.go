package fs

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/c2h5oh/datasize"
	"github.com/vlad-s/filekeep/config"
	"github.com/vlad-s/filekeep/helpers"
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
	Name  string   `json:"name"`   // Name is the basename of the file or directory
	Path  string   `json:"path"`   // Path is the relative path of the file or directory
	Prev  string   `json:"prev"`   // Prev is the parent directory of the file or directory
	IsDir bool     `json:"is_dir"` // IsDir is a flag if it's a file or directory
	Size  FileSize `json:"size"`   // Size is the file or directory size

	// Files keeps all children files of a directory, or a slice of empty Nodes if it's a child directory
	// in order to show how many children the directory has.
	Files []*Node `json:"files"`
	// FilesSize represents the summed size of all children files.
	FilesSize FileSize `json:"files_size"`
	// Dirs keeps all children directories of a directory, or a slice of empty Nodes if it's a child directory
	// ir order to show how many children the directory has.
	Dirs []*Node `json:"dirs"`
}

func newNode(path string, info os.FileInfo) *Node {
	return &Node{
		Name:  filepath.Base(path),
		Path:  path,
		Prev:  filepath.Dir(helpers.StripRoot(path)),
		Size:  FileSize(info.Size()),
		IsDir: info.IsDir(),
		Files: make([]*Node, 0),
		Dirs:  make([]*Node, 0),
	}
}

func lsDir(path string, info os.FileInfo, count int) (*Node, error) {
	if count > dirLimit {
		return nil, ErrDirLimit
	}
	count++

	fd := newNode(path, info)

	ls, err := ioutil.ReadDir(path)
	if err != nil {
		Log.Debugf("Error calling ioutil.ReadDir on path %q, returning ErrDirNotFound", path)
		return nil, ErrDirNotFound
	}

	for _, fileInfo := range ls {
		filePath := filepath.Join(path, fileInfo.Name())

		if config.Get().IsHidden(filePath, fileInfo.Name()) {
			Log.Debugf("Path %q or %q is hidden, continuing lsDir loop", fileInfo.Name(), path)
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
		Log.Debugf("Path %q is hidden, returning ErrFileNotFound", path)
		return nil, ErrFileNotFound
	}

	info, err := os.Stat(path)
	if err != nil {
		Log.Debugf("Error calling os.Stat on path %q, returning ErrFileNotFound", path)
		return nil, ErrFileNotFound
	}

	if info.IsDir() {
		fd, err = lsDir(path, info, 0)
	} else {
		fd = newNode(path, info)
	}

	return
}
