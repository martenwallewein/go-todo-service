package git

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/martenwallewein/todo-service/pkg/cmdexec"
	"github.com/sirupsen/logrus"
)

type GitRepo struct {
	Path string
}

func Clone(url string, path string) (*GitRepo, error) {
	err, _, errStr := cmdexec.Exec("git", "clone", url, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to clone git repo: %s", errStr)
	}

	return loadFromPath(path)

}

func loadFromPath(path string) (*GitRepo, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	logrus.Warn("Files in ", path)
	logrus.Warn(files)
	for _, file := range files {
		if file.IsDir() {
			fileName := file.Name()
			if fileName == "." || fileName == ".." {
				continue
			}
			if fileName == ".git" {
				logrus.Warn("Found repo with path ")
				return &GitRepo{
					Path: path,
				}, nil
			}

			subdirs, err := ioutil.ReadDir(filepath.Dir(file.Name()))
			if err != nil {
				return nil, err
			}

			for _, subfile := range subdirs {
				if strings.Index(subfile.Name(), ".git") == 0 {
					logrus.Warn("Found repo with path ")
					return &GitRepo{
						Path: filepath.Dir(file.Name()),
					}, nil
				}
			}

		}

		// Try to find folder with .git inside, maybe the name of the cloned repo
		/*if file.IsDir() {
			repo, err := loadFromPath(filepath.Dir(file.Name()))
		}*/
		// fmt.Println(file.Name(), file.IsDir())//
	}

	return nil, fmt.Errorf("Could not find cloned repo in path")
}

func Load(repo string) (*GitRepo, error) {
	return loadFromPath(repo)
}

func (r *GitRepo) FetchAndRebase() error {

	err, _, errStr := cmdexec.ExecInFolder(r.Path, "git", "fetch")
	if err != nil {
		return fmt.Errorf("Failed to fetch git repo: %s", errStr)
	}
	err, _, errStr = cmdexec.ExecInFolder(r.Path, "git", "rebase")
	if err != nil {
		return fmt.Errorf("Failed to rebase git repo: %s", errStr)
	}

	return nil
}

func (r *GitRepo) CommitAll(message string) error {
	err, _, errStr := cmdexec.ExecInFolder(r.Path, "git", "add", ".")
	if err != nil {
		return fmt.Errorf("Failed to add files to git repo: %s", errStr)
	}
	err, _, errStr = cmdexec.ExecInFolder(r.Path, "git", "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("Failed to commit to git repo: %s", errStr)
	}

	return nil
}

func (r *GitRepo) Push() error {
	err, _, errStr := cmdexec.ExecInFolder(r.Path, "git", "push")
	if err != nil {
		return fmt.Errorf("Failed to push git repo: %s", errStr)
	}
	return nil
}
