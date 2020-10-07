package config

import (
	"errors"
	"github.com/4nte/protodist/distribution"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
)

type GitContext struct {
	Branch string
	Tag    string
}
type Config struct {
	SshAuth         ssh.AuthMethod
	HttpAuth        http.AuthMethod
	ProtoOutDir     string
	GitCloneDir     string
	ProtoRepoDir    string
	ProtoImportPath []string
	GitAccessToken  string
	GitContext      GitContext
}

type GitConfig struct {
	User         string
	Organization string
	Host         string
}

func (c GitConfig) GetRepoBase() string {
	return path.Join(c.Host, c.GetRepoOwner())
}

func (c GitConfig) GetRepoOwner() string {
	var repoOwner string
	if len(c.Organization) > 0 {
		repoOwner = c.Organization
	} else {
		repoOwner = c.User
	}
	return repoOwner
}

func (g GitConfig) RepoOwner() string {
	if len(g.Organization) > 0 {
		return g.Organization
	}
	return g.User
}

type ProtodistConfig struct {
	Name         string
	Git          GitConfig
	Distribution []distribution.Config
}

func ParseProtodistConfig(configFile string) (ProtodistConfig, error) {
	protodistConfig := ProtodistConfig{}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return protodistConfig, err
	}

	err = yaml.Unmarshal(data, &protodistConfig)
	if err != nil {
		return protodistConfig, err
	}

	if protodistConfig.Name == "" {
		return protodistConfig, errors.New("project name not set")
	}
	if protodistConfig.Git.User == "" && protodistConfig.Git.Organization == "" {
		return protodistConfig, errors.New("git.user or git.organization are both empty")
	}

	if protodistConfig.Git.Host != "github.com" {
		return protodistConfig, errors.New("github.com is the only supported git provider")
	}

	return protodistConfig, nil
}
