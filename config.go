package main

import (
	"fmt"
	"github.com/4nte/protodist/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
)

func LoadAppConfig() (config.Config, error) {
	// Default values
	viper.SetDefault("PROTO_DIR", ".")
	viper.SetDefault("PROTO_OUT_DIR", "/tmp/protodist/out")
	viper.SetDefault("GIT_CLONE_DIR", "/tmp/protodist/clone")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("PROTODIST")
	gitUser := viper.GetString("GIT_USER")
	//gitHost := viper.GetString("GIT_HOST")
	gitSSHKeyFile := viper.GetString("GIT_SSH_KEY_FILE")
	gitBranch := viper.GetString("GIT_BRANCH")
	gitTag := viper.GetString("GIT_TAG")
	gitCloneDir := viper.GetString("GIT_CLONE_DIR")
	gitAccessToken := viper.GetString("GIT_ACCESS_TOKEN")
	//gitOrganization := viper.GetString("GIT_ORGANIZATION")

	protoDir := viper.GetString("PROTO_DIR")
	protoOutDir := viper.GetString("PROTO_OUT_DIR")

	protoImportPath := viper.GetStringSlice("PROTO_IMPORT_PATH")

	if gitUser == "" {
		return config.Config{}, fmt.Errorf("PROTODIST_GIT_USER must be set")
	}

	//if gitHost == "" {
	//	return config.Config{}, fmt.Errorf("PROTODIST_GIT_HOST must be set")
	//}

	if gitSSHKeyFile == "" {
		return config.Config{}, fmt.Errorf("PROTODIST_GIT_SSH_KEY_FILE must be set")
	}

	if gitBranch == "" {
		return config.Config{}, fmt.Errorf("PROTODIST_GIT_BRANCH must be set")
	}
	fmt.Println("git user:", gitUser)
	//fmt.Println("git host:", gitHost)
	fmt.Println("git ssh key file:", gitSSHKeyFile)

	pemData, err := ioutil.ReadFile(gitSSHKeyFile)
	if err != nil {
		return config.Config{}, err
	}

	publicKey, err := ssh.NewPublicKeys("git", pemData, "")
	if err != nil {
		return config.Config{}, err
	}

	gitConfig := config.GitContext{
		Branch: gitBranch,
		Tag:    gitTag,
	}
	return config.Config{
		SshAuth:         publicKey,
		GitContext:      gitConfig,
		ProtoRepoDir:    protoDir,
		ProtoOutDir:     protoOutDir,
		ProtoImportPath: protoImportPath,
		GitCloneDir:     gitCloneDir,
		GitAccessToken:  gitAccessToken,
		HttpAuth: &http.BasicAuth{
			Username: gitUser,
			Password: gitAccessToken,
		},
	}, nil
}
