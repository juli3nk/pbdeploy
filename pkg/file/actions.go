package file

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/juli3nk/go-utils/filedir"
	"github.com/thoas/go-funk"
)

type config struct {
	name        string
	description string
	srcPath     string
	dstPath     string
	ignores     []string
}

func newFile(name, description, srcPath, dstPath string) *config {
	return &config{
		name:        name,
		description: description,
		srcPath:     srcPath,
		dstPath:     dstPath,
		ignores:     []string{".git", "README.md"},
	}
}

func (c *config) addIgnoreFile(filename string) {
	c.ignores = append(c.ignores, filename)
}

func (c *config) deleteFiles() error {
	files, err := ioutil.ReadDir(c.dstPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if funk.Contains(c.ignores, f.Name()) {
			continue
		}

		if err := os.RemoveAll(path.Join(c.dstPath, f.Name())); err != nil {
			return err
		}
	}

	return nil
}

func (c *config) copyFiles() error {
	if err := filedir.CopyDir(c.srcPath, c.dstPath); err != nil {
		return err
	}

	return nil
}
