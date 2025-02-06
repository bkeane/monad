package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	v5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
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

func (c Git) Parse(path string) (*Git, error) {
	log.Debug().Msgf("Parsing git repository at path: %s", path)
	var err error
	g := &Git{}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		path = filepath.Dir(path)
	}

	// Get absolute path before setting Base
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	g.Path = absPath
	// Use the last directory name from the absolute path instead of filepath.Base(path)
	g.BasePath = filepath.Base(absPath)

	_, repo, err := find(path)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	g.Sha = sha(head)
	g.Branch = branch(head)
	g.Dirty = dirty(repo)
	g.Origin, err = origin(repo)
	if err != nil {
		return nil, err
	}

	g.Owner = ownerFromOrigin(g.Origin)
	g.Repo = repoFromOrigin(g.Origin)

	return g, nil
}

func (g *Git) AsMap() map[string]string {
	return map[string]string{
		"BasePath": g.BasePath,
		"Sha":      g.Sha,
		"Branch":   g.Branch,
		"Origin":   g.Origin,
		"Owner":    g.Owner,
		"Repo":     g.Repo,
		"Dirty":    strconv.FormatBool(g.Dirty),
	}
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
