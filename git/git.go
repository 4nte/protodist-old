package steps

import (
	"context"
	"errors"
	"fmt"
	config2 "github.com/4nte/protodist/config"
	"github.com/4nte/protodist/distribution"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"os"
	"os/exec"
	"path"
	"time"
)

func GetRepo(gitConfig config2.GitConfig, repoName distribution.RepoName, branch string, cloneDirPath string, auth transport.AuthMethod) (*git.Repository, error) {
	var repository *git.Repository
	//repoUrl := fmt.Sprintf("git@%s:%s/%s", gitConfig.Host, gitConfig.User, repoName)
	remoteUrl := fmt.Sprintf("https://%s/%s.git", gitConfig.GetRepoBase(), repoName)
	localRepoPath := path.Join(cloneDirPath, string(repoName))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	repository, err := git.PlainCloneContext(ctx, localRepoPath, false, &git.CloneOptions{
		Auth:     auth,
		URL:      remoteUrl,
		Progress: os.Stdout,
		Tags:     git.AllTags,
	})

	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		signature := &object.Signature{
			Name:  "Protodist",
			Email: "",
			When:  time.Now(),
		}

		repository, err = InitTargetRepo(remoteUrl, localRepoPath, *signature, auth)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to clone target repo (%s): %s", remoteUrl, err)
	}

	return repository, nil
}

func CloneRepo(gitConfig config2.GitConfig, repoName distribution.RepoName, branch string, cloneDirPath string, auth transport.AuthMethod) error {

	repoUrl := fmt.Sprintf("git@%s/%s", gitConfig.GetRepoBase(), repoName)

	fmt.Println("fetching", repoUrl)
	repo, err := GetRepo(gitConfig, repoName, branch, cloneDirPath, auth)
	if err != nil {
		return fmt.Errorf("failed to setup repo locally: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get Worktree: %s\n", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		Force:  true,
	})
	if err != nil {
		fmt.Printf("branch %s does not exist on origin\n", branch)
		//headRef, err := repo.Head()
		//if err != nil {
		//	return fmt.Errorf("failed to get head ref: %s", err)
		//}

		//headRef, err := repo.Reference(plumbing.HEAD, false)
		//if err != nil {
		//	return fmt.Errorf("unable to reference head: %s", err)
		//}
		//refName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
		//ref := plumbing.NewHashReference(refName, headRef.Hash())
		//
		//// The created reference is saved in the storage.
		//if err = repo.Storer.SetReference(ref); err != nil {
		//	return fmt.Errorf("failed to store reference: %s", err)
		//}

		branchRef := fmt.Sprintf("refs/heads/%s", branch)
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(branchRef),
			Create: true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch after it has been stored: %s", err)
		}
	}

	return nil
}

func DoesRepoHaveUnstagedChanges(repoDirectory string) (bool, error) {
	gitDiffCmd := exec.Command("git", "diff")
	gitDiffCmd.Dir = repoDirectory
	out, err := gitDiffCmd.Output()
	if err != nil {
		return false, err
	}

	//fmt.Printf("out: %s\n", out)
	if string(out) == "" {
		return false, nil
	}
	return true, nil
}

func AddAllFilesAndCommit(repoDir string, tag string) error {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repo: %s", err)
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get repo Worktree: %s", err)
	}

	if err := workTree.AddGlob("."); err != nil {
		return fmt.Errorf("failed to add changes to staging: %s", err)
	}

	authorSignature := &object.Signature{
		Name:  "Protodist",
		Email: "protodist@localhost",
		When:  time.Now(),
	}
	hash, err := workTree.Commit("generated files", &git.CommitOptions{
		Committer: authorSignature,
		Author:    authorSignature,
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %s", err)
	}

	if tag != "" {
		tagOptions := git.CreateTagOptions{
			Tagger:  authorSignature,
			Message: fmt.Sprintf("proto version %s", tag),
		}
		_, err := repo.CreateTag(tag, hash, &tagOptions)
		if err != nil {
			return fmt.Errorf("failed to create tag %s: %s", tag, err)
		}
	}

	return nil
}

func PushCommitsAndTags(repoDir string, tag string, authMethod transport.AuthMethod) error {
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repo: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel() //

	// Push commits
	refSpec := []config.RefSpec{ // Refspec defines what to push to origin, and how to map it to remote
		"+refs/heads/*:refs/heads/*",
	}
	err = repo.PushContext(ctx, &git.PushOptions{
		RefSpecs: refSpec,
		Auth:     authMethod,
	})
	if err != nil {
		return fmt.Errorf("failed to push commits: %s", err)
	}

	if tag != "" {
		refSpec = []config.RefSpec{ // Refspec defines what to push to origin, and how to map it to remote
			"refs/tags/*:refs/tags/*",
		}
		err = repo.PushContext(ctx, &git.PushOptions{
			RefSpecs: refSpec,
			Auth:     authMethod,
		})
		if err != nil {
			return fmt.Errorf("failed to push tags: %s", err)
		}
	}

	fmt.Printf("Successfully pushed changes to git repo\n")

	return nil
}
