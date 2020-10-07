package steps

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"os"
	"path"
)

func InitTargetRepo(remoteURL string, localRepoPath string, signature object.Signature, auth transport.AuthMethod) (*git.Repository, error) {
	// Init repo
	repository, err := git.PlainInit(localRepoPath, false)
	if err != nil {
		return nil, err
	}

	// Add a remote
	_, err = repository.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	})
	if err != nil {
		return nil, err
	}

	// Move misc files to the repo
	// e.g package.json and similar
	// TODO: move from a template dir
	_, err = os.Create(path.Join(localRepoPath, "README.md"))
	if err != nil {
		return nil, err
	}

	w, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	// Add files to the staging
	_, err = w.Add("README.md")
	if err != nil {
		return nil, err
	}

	_, err = w.Commit("Initial commit", &git.CommitOptions{
		All:       true,
		Author:    &signature,
		Committer: &signature,
	})
	if err != nil {
		return nil, err
	}

	refSpec := []config.RefSpec{ // Refspec defines what to push to origin, and how to map it to remote
		"+refs/heads/*:refs/heads/*",
	}
	err = repository.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   refSpec,
		Auth:       auth,
	})
	if err != nil {
		return nil, err
	}

	return repository, err
}
