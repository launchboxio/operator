package source

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"os"
)

type GitSource struct {
	Path   string
	URL    string
	Token  string
	Branch string
	Commit string
}

// Clone fetches a specific branch / commit from a remote git provider
func (gs *GitSource) Clone() error {
	cloneOpts := &git.CloneOptions{
		URL:      gs.URL,
		Progress: os.Stdout,
	}
	if gs.Token != "" {
		cloneOpts.Auth = &http.BasicAuth{
			Username: "launchboxhq",
			Password: gs.Token,
		}
	}

	if gs.Branch != "" {
		cloneOpts.ReferenceName = plumbing.NewNoteReferenceName(gs.Branch)
	}

	repo, err := git.PlainClone(gs.Path, false, cloneOpts)
	if err != nil {
		return err
	}

	if gs.Commit != "" {
		w, err := repo.Worktree()
		if err != nil {
			return err
		}
		return w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(gs.Commit),
		})
	}
	return nil
}

func (gs *GitSource) Remove() error {
	return os.RemoveAll(gs.Path)
}
