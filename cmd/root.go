/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/4nte/protodist/config"
	"github.com/4nte/protodist/internal"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	envPrefix             = "PROTODIST"
	defaultConfigFilename = "protodist"
)

var (
	cfgFile         string
	gitBranch       string
	gitTag          string
	gitUser         string
	gitCloneDir     string
	gitAccessToken  string
	protoDir        string
	protoOutDir     string
	protoImportPath []string
	verbose         bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "protodist",
	Short: "Protobuf manager",
	Long:  `Compile, bundle and distribute protobuf packages..`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("pre run")
		// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
		return initializeConfig(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if gitAccessToken == "" {
			fmt.Println("warning: GIT_ACCESS_TOKEN is empty")
		}

		if gitUser == "" {
			panic("PROTODIST_GIT_USER must be set")
		}

		if gitBranch == "" {
			panic("PROTODIST_GIT_BRANCH must be set")
		}

		gitConfig := config.GitContext{
			Branch: gitBranch,
			Tag:    gitTag,
		}
		cfg := config.Config{
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
		}
		internal.Protodist(cfg)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

/*
	gitUser := viper.GetString("GIT_USER")
	//gitHost := viper.GetString("GIT_HOST")
	gitBranch := viper.GetString("GIT_BRANCH")
	gitTag := viper.GetString("GIT_TAG")
	gitCloneDir := viper.GetString("GIT_CLONE_DIR")
	gitAccessToken := viper.GetString("GIT_ACCESS_TOKEN")
	//gitOrganization := viper.GetString("GIT_ORGANIZATION")

	protoDir := viper.GetString("PROTO_DIR")
	protoOutDir := viper.GetString("PROTO_OUT_DIR")

	protoImportPath := viper.GetStringSlice("PROTO_IMPORT_PATH")
*/
func init() {
	//cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.protodist.yaml)")
	// TODO: refactor to use git_ref instead and then infer if the ref is branch or tag
	rootCmd.PersistentFlags().StringVar(&gitBranch, "git_branch", "", "git branch")
	rootCmd.PersistentFlags().StringVar(&gitUser, "git_user", "", "git user")
	rootCmd.PersistentFlags().StringVar(&gitTag, "git_tag", "", "git tag")
	rootCmd.PersistentFlags().StringVar(&gitAccessToken, "git_access_token", "", "git access token")
	rootCmd.PersistentFlags().StringVar(&protoOutDir, "proto_out_dir", "/tmp/protodist/out", "output directory")
	rootCmd.PersistentFlags().StringVar(&gitCloneDir, "git_clone_dir", "/tmp/protodist/clone", "git clone directory")
	rootCmd.PersistentFlags().StringSliceVar(&protoImportPath, "proto_import_path", nil, "proto import paths")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show verbose logs")
	//cobra.MarkFlagRequired()
	/*
		cfgFile string
		gitBranch string
		gitTag string
		gitCloneDir string
		gitAccessToken string
		protoDir string
		protoOutDir string
		protoImportPath []string
	*/
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		panic(err)
	}
}
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		fmt.Println("f", f)
		if strings.Contains(f.Name, "-") {
			envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
			err := v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix))
			if err != nil {
				panic(err)
			}
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			fmt.Println("setting", val)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				panic(err)
			}
		}
	})
}
func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(defaultConfigFilename)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(".")

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}
