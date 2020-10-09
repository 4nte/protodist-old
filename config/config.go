package config

import (
	"errors"
	"github.com/4nte/protodist/distribution"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
)

type GitContext struct {
	Branch string
	Tag    string
}
type Config struct {
	//SshAuth         ssh.AuthMethod
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

//
//func LoadAppConfig() (Config, error) {
//	// Default values
//	viper.SetDefault("PROTO_DIR", ".")
//	viper.SetDefault("PROTO_OUT_DIR", "/tmp/protodist/out")
//	viper.SetDefault("GIT_CLONE_DIR", "/tmp/protodist/clone")
//
//	viper.AutomaticEnv()
//	viper.SetEnvPrefix("PROTODIST")
//	gitUser := viper.GetString("GIT_USER")
//	//gitHost := viper.GetString("GIT_HOST")
//	gitBranch := viper.GetString("GIT_BRANCH")
//	gitTag := viper.GetString("GIT_TAG")
//	gitCloneDir := viper.GetString("GIT_CLONE_DIR")
//	gitAccessToken := viper.GetString("GIT_ACCESS_TOKEN")
//	//gitOrganization := viper.GetString("GIT_ORGANIZATION")
//
//	protoDir := viper.GetString("PROTO_DIR")
//	protoOutDir := viper.GetString("PROTO_OUT_DIR")
//
//	protoImportPath := viper.GetStringSlice("PROTO_IMPORT_PATH")
//
//	if gitAccessToken == "" {
//		fmt.Println("warning: GIT_ACCESS_TOKEN is empty")
//	}
//
//	if gitUser == "" {
//		return Config{}, fmt.Errorf("PROTODIST_GIT_USER must be set")
//	}
//
//	if gitBranch == "" {
//		return Config{}, fmt.Errorf("PROTODIST_GIT_BRANCH must be set")
//	}
//
//	gitConfig := GitContext{
//		Branch: gitBranch,
//		Tag:    gitTag,
//	}
//	return Config{
//		GitContext:      gitConfig,
//		ProtoRepoDir:    protoDir,
//		ProtoOutDir:     protoOutDir,
//		ProtoImportPath: protoImportPath,
//		GitCloneDir:     gitCloneDir,
//		GitAccessToken:  gitAccessToken,
//		HttpAuth: &http.BasicAuth{
//			Username: gitUser,
//			Password: gitAccessToken,
//		},
//	}, nil
//}
