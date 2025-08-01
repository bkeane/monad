package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	v5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Git struct {
	Path     string
	BasePath string
	Branch   string
	Sha      string
	Origin   string
	Owner    string
	Repo     string
	Dirty    bool
}

func Parse(path string) (Git, error) {
	var err error
	g := Git{}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return g, fmt.Errorf("failed to get absolute path: %w", err)
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return g, fmt.Errorf("failed to stat path: %w", err)
	}

	if !fileInfo.IsDir() {
		absPath = filepath.Dir(absPath)
	}

	g.Path = absPath
	g.BasePath = filepath.Base(absPath)

	_, repo, err := find(absPath)
	if err != nil {
		return g, fmt.Errorf("failed to find git repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return g, fmt.Errorf("failed to get head: %w", err)
	}

	g.Sha = sha(head)
	g.Branch = branch(head)
	g.Dirty = dirty(repo)
	g.Origin, err = origin(repo)
	if err != nil {
		return g, fmt.Errorf("failed to get origin: %w", err)
	}

	g.Owner = ownerFromOrigin(g.Origin)
	g.Repo = repoFromOrigin(g.Origin)

	return g, nil
}

func find(path string) (root string, repo *v5.Repository, err error) {
	// Validate initial path exists
	if _, err := os.Stat(path); err != nil {
		return "", nil, fmt.Errorf("invalid path: %w", err)
	}

	// Get absolute path to handle relative paths correctly
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	path = absPath

	for {
		if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
			repo, err := v5.PlainOpen(path)
			if err != nil {
				return "", nil, fmt.Errorf("failed to open git repository: %w", err)
			}

			return path, repo, nil
		}

		parentDir := filepath.Dir(path)
		if parentDir == path {
			return "", nil, fmt.Errorf("path does not appear to be within a git repository")
		}
		path = parentDir
	}
}

func branch(head *plumbing.Reference) string {
	return head.Name().Short()
}

func sha(head *plumbing.Reference) string {
	return head.Hash().String()
}

func dirty(repo *v5.Repository) bool {
	wt, err := repo.Worktree()
	if err != nil {
		return false
	}

	status, err := wt.Status()
	if err != nil {
		return false
	}

	return !status.IsClean()
}

func origin(repo *v5.Repository) (string, error) {
	var origin string

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", err
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("no remote origin found")
	}

	if len(urls) > 1 {
		return "", fmt.Errorf("multiple remote origins found")
	}

	// drop protocol from origin
	origin = strings.TrimPrefix(urls[0], "git@")
	origin = strings.TrimPrefix(origin, "https://")
	origin = strings.TrimPrefix(origin, "http://")

	// drop .git suffix
	origin = strings.TrimSuffix(origin, ".git")

	// swap remaining : with /
	origin = strings.Replace(origin, ":", "/", 1)

	return origin, nil
}

func ownerFromOrigin(origin string) string {
	return strings.Split(origin, "/")[1]
}

func repoFromOrigin(origin string) string {
	return strings.Split(origin, "/")[2]
}
