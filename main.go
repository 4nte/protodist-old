package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/4nte/protodist/config"
	"github.com/4nte/protodist/core"
	"github.com/4nte/protodist/distribution"
	steps "github.com/4nte/protodist/git"
	"github.com/4nte/protodist/proto"
	"github.com/4nte/protodist/provider"
	"github.com/4nte/protodist/util"
	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

//var distributionConfigs = []distribution.Config{
//	{
//		Name:          "proto-internal",
//		ProtoPackages: []proto.PackageName{"domain", "deployment", "locator", "gateway"},
//		DistributionPackages: []distribution.Package{
//			distribution.NewPackage("go", "bitbucket.org/ag04/proto-internal-go/{{.ProtoPackage}}"),
//			distribution.NewPackage("js", ""),
//		},
//	},
//	{
//		Name:          "proto-public",
//		ProtoPackages: []proto.PackageName{"domain", "deployment"},
//		DistributionPackages: []distribution.Package{
//			distribution.NewPackage("java", "stream.locator.{{.ProtoPackage}}"),
//			distribution.NewPackage( "js", ""),
//		},
//	},
//}

var supportedTargets = []string{"go", "java", "js"}

func compileStrategy(strategy distribution.Strategy, protodistCfg config.ProtodistConfig, config2 config.Config, distOutDir string, protoRepoDir string) error {
	// Make a copy of proto repo
	tempProtoDir, err := ioutil.TempDir(os.TempDir(), "protodist-compile")
	if err != nil {
		return fmt.Errorf("failed to create a temporary proto root dir: %s", err)
	}
	if err := util.CopyDirContents(protoRepoDir, tempProtoDir); err != nil {
		return fmt.Errorf("failed to copy proto repo to temporary dir: %s", err)
	}

	// Load packages from proto dir
	packages := getProtoPackagesInDir(tempProtoDir)

	// ::: PRE-COMPILATION :::
	// Modify proto files (e.g add "package go_option")
	for _, protoPackage := range packages {
		// Pre-compilation, language specific modifications
		if strategy.Target == "go" {
			for _, file := range protoPackage.Files {
				f, err := os.OpenFile(file,
					os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}

				goPackage := fmt.Sprintf("%s/%s/%s-go/%s", "github.com", protodistCfg.Git.RepoOwner(), strategy.Name, protoPackage.Name)
				packageOptionLine := fmt.Sprintf("\noption %s_package = \"%s\";\n", strategy.Target, goPackage)
				if _, err := f.WriteString(packageOptionLine); err != nil {
					log.Println(err)
				}

				err = f.Close()
				if err != nil {
					log.Println(err)
				}

			}
		}
	}

	for _, packageName := range strategy.Packages {
		protoPackage, ok := packages.FindByName(packageName)
		if !ok {
			panic(fmt.Errorf("package %s not found", protoPackage.Name))
		}

		// Compile package to target
		if err := compileProtoPackage(protoPackage, strategy.Target, distOutDir, tempProtoDir, config2.ProtoImportPath, strategy.Plugins); err != nil {
			return fmt.Errorf("failed to compile proto package: %s", err)
		}
	}

	return nil
}

func executeDistributionStrategy(strategy distribution.Strategy, protodistCfg config.ProtodistConfig, cfg config.Config) error {
	// (1) Fetch dist repo
	// (2) compile
	wg := sync.WaitGroup{}
	wg.Add(2)

	distRepoName := strategy.GetDistributionRepoName()
	distOutDir := path.Join(cfg.ProtoOutDir, string(distRepoName))
	distRepoDir := path.Join(cfg.GitCloneDir, string(distRepoName))

	err := os.MkdirAll(distOutDir, os.ModePerm)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	waitCh := make(chan struct{})

	// Send signal to waitCh once waitGroup is finished
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	// (0) Create git repo if it doesn't exist already
	gitClient := provider.NewGithubClient(cfg.GitAccessToken)

	_, res, err := gitClient.Repositories.Get(context.Background(), protodistCfg.Git.GetRepoOwner(), string(distRepoName))
	if err != nil && res.StatusCode != http.StatusNotFound {
		panic(errors.Wrap(err, "failed to get the repo"))
	}
	if res.StatusCode == http.StatusNotFound {
		fmt.Printf("git repository for package %s doesn't exist. Creating a new one.", string(distRepoName))
		// Repo doesn't exist, create it
		user, _, err := gitClient.Users.Get(context.Background(), "4nte")
		if err != nil {
			panic(fmt.Errorf("failed to get user: %w", err))
		}
		_, _, err = gitClient.Repositories.Create(context.Background(), protodistCfg.Git.Organization, &github.Repository{
			Owner:       user,
			Name:        github.String(string(distRepoName)),
			FullName:    github.String(string(distRepoName)),
			Description: github.String("Protodist target repo"),
		})
		if err != nil {
			panic(fmt.Errorf("failed to create repo: %w", err))
		}

	}

	// (1)
	go func() {
		err := steps.CloneRepo(protodistCfg.Git, distRepoName, cfg.GitContext.Branch, cfg.GitCloneDir, cfg.HttpAuth)
		if err != nil {
			errChan <- err
			return
		}

		fmt.Printf("Successfully cloned repo: %s\n", distRepoName)
		wg.Done()
	}()

	// (2) Compile
	go func() {
		err := compileStrategy(strategy, protodistCfg, cfg, distOutDir, cfg.ProtoRepoDir)
		if err != nil {
			errChan <- err
			return
		}

		fmt.Printf("Distribution strategy: %s Compiled packages %v to '%s' target\n", strategy.Name, strategy.Packages, strategy.Target)
		wg.Done()
	}()

	// Wait for wg to finish, or handle err
	select {
	case err := <-errChan:
		{
			return err
		}
	case <-waitCh:
		{
		}
	}

	// (3) move compiled files to dist repo
	type Vars struct {
		GitBasePath          string
		DistributionRepoName string
	}
	vars := Vars{
		GitBasePath:          protodistCfg.Git.GetRepoBase(),
		DistributionRepoName: string(distRepoName),
	}

	var templatedMountPaths []core.MountPath

	for _, copyOperation := range strategy.Mount {
		templatedSource := bytes.NewBufferString("")
		templatedDestination := bytes.NewBufferString("")

		t, err := template.New("source").Parse(copyOperation.Source)
		if err != nil {
			return fmt.Errorf("failed to parse source template: %s", err)
		}
		if err := t.Execute(templatedSource, vars); err != nil {
			return err
		}

		t, err = template.New("target").Parse(copyOperation.Destination)
		if err != nil {
			return fmt.Errorf("failed to parse destination template: %s", err)
		}
		if err := t.Execute(templatedDestination, vars); err != nil {
			return err
		}

		templatedMountPaths = append(templatedMountPaths, core.MountPath{Source: templatedSource.String(), Destination: templatedDestination.String()})

	}

	if err := moveGeneratedProtos(templatedMountPaths, distOutDir, distRepoDir, strategy.IgnoreRepoFiles); err != nil {
		return fmt.Errorf("failed to move generated protos: %s", err)
	}

	if err := postBuildTarget(distRepoDir, strategy); err != nil {
		return fmt.Errorf("failed executing postBuildTarget: %s", err)
	}

	// (4) add & commit changes
	if err := steps.AddAllFilesAndCommit(distRepoDir, cfg.GitContext.Tag); err != nil {
		return err
	}

	return nil

}

func postBuildTarget(targetDir string, strategy distribution.Strategy) error {
	switch strategy.Target {
	case "go":
		return postBuildTargetGo(targetDir, strategy)
	case "js":
		return nil
	}

	return fmt.Errorf("unknown target: %s", strategy.Target)
}
func postBuildTargetGo(targetDir string, strategy distribution.Strategy) error {
	// Go mod init
	cmd := exec.Command("go", "mod", "init", strategy.Name)
	cmd.Dir = targetDir
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	// Go mod tidy
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func moveGeneratedProtos(copyDirs []core.MountPath, buildDir string, RepoDir string, ignoreFiles []string) error {
	fmt.Println("move generated protos")
	for _, cp := range copyDirs {
		source := path.Join(buildDir, cp.Source)
		destination := path.Join(RepoDir, cp.Destination)

		// Clean previously generated pb files
		if err := util.RemoveDirContents(destination, ignoreFiles); err != nil {
			return fmt.Errorf("failed to clean destination path: %s", err)
		}

		// Copy new pb generated files to destination
		if err := util.CopyDirContents(source, destination); err != nil {
			return fmt.Errorf("failed to copy pb files to target repo: %s (source: %s) (destination: %s)", err, source, destination)
		}
	}

	return nil
}

func v2(distributionConfigs []distribution.Config, protodistConfig config.ProtodistConfig, cfg config.Config) {
	// Cleanup out dir
	if err := util.ClearDir(cfg.ProtoOutDir); err != nil {
		fmt.Printf("Unable to clear out dir %s\n", err)
	}

	if err := util.ClearDir(cfg.GitCloneDir); err != nil {
		fmt.Printf("Unable to clear clone dir %s\n", err)
	}

	wg := sync.WaitGroup{}

	// waitGroup setup
	for _, distConfig := range distributionConfigs {
		wg.Add(len(distConfig.ToStrategies()))
	}

	for _, distConfig := range distributionConfigs {
		strats := distConfig.ToStrategies()
		for _, strat := range strats {
			go func(strat distribution.Strategy) {
				err := executeDistributionStrategy(strat, protodistConfig, cfg)
				if err != nil {
					panic(fmt.Errorf("failed to exec strategy %s, target: %s, error: %s", strat.Name, strat.Target, err))
				}
				wg.Done()
			}(strat)
		}
	}

	wg.Wait()

	// Now that everything succeeded, we can push the changes
	for _, distConfig := range distributionConfigs {
		strats := distConfig.ToStrategies()
		for _, strat := range strats {
			strat.GetDistributionRepoName()
			distRepoDir := path.Join(cfg.GitCloneDir, string(strat.GetDistributionRepoName()))

			if err := steps.PushCommitsAndTags(distRepoDir, cfg.GitContext.Tag, cfg.HttpAuth); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func getProtoPackagesInDir(dir string) proto.Packages {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var packageDirs []os.FileInfo
	for _, f := range files {
		if f.IsDir() {
			packageDirs = append(packageDirs, f)
		}

	}

	var packages proto.Packages
	for _, packageDir := range packageDirs {
		// For each proto package, attempt to create a build strategy
		packageName := packageDir.Name()
		packageDirPath := path.Join(dir, packageName)
		protoFiles, err := filepath.Glob(path.Join(packageDirPath, "*.proto"))
		if err != nil {
			panic(err)
		}
		if len(protoFiles) == 0 {
			//fmt.Println("non-proto dir, skipping")
			continue
		}
		pkg := proto.Package{
			Name:  proto.PackageName(packageDir.Name()),
			Files: protoFiles,
		}

		packages = append(packages, pkg)
	}

	return packages
}

func main() {
	cfg, err := LoadAppConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load app config: %w", err))
	}

	protodistConfig, err := config.ParseProtodistConfig(path.Join(cfg.ProtoRepoDir, "protodist.yaml"))
	if err != nil {
		panic(fmt.Errorf("failed to load protodist config: %w", err))
	}

	// Cleanup out dirs
	for _, target := range supportedTargets {
		dirPath := path.Join(cfg.ProtoOutDir, target)
		if err = util.ClearDir(dirPath); err != nil {
			fmt.Printf("Unable to clear directory contents %s: %s\n", dirPath, err)
		}
	}

	v2(protodistConfig.Distribution, protodistConfig, cfg)
	return

}

func compileProtoPackage(protoPackage proto.Package, target core.BuildTarget, protoOutDir string, protoRepoDir string, protoPaths []string, plugins []proto.Plugin) error {
	// Create protowrap command
	protocBuilder := proto.NewProtocCmdBuilder()
	protocBuilder.AddPlugin(plugins...)
	protocBuilder.SetProtoPackage(protoPackage)
	//protocBuilder.SetProtoPath("./proto-repo-example")
	protocBuilder.AddProtoPath(protoPaths...)
	protocBuilder.SetOutputDir(protoOutDir)
	protocBuilder.SetProtoRootPath(protoRepoDir)
	protocBuilder.SetBuildTargets(target)
	// TODO: set plugins
	cmd := protocBuilder.Build()
	fmt.Printf("Compiling package %s: \"%s\"\n", protoPackage.Name, strings.Join(cmd.Args, " "))
	//cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("protoc run failed: %w", err)
	}

	return nil
}
