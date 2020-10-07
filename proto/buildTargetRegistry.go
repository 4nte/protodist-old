package proto

import (
	"fmt"
	"github.com/4nte/protodist/core"
)

type PostBuildScript func(protoPackage string, target string, protocOutDir string, targetRepoHome string, gitBasePath string) error

type targetConfig struct {
	protocArgs      []string
	mountPaths      []core.MountPath
	ignoreRepoFiles []string
	PostBuildScript PostBuildScript
}

func (t targetConfig) GetMountPaths() []core.MountPath {
	return t.mountPaths
}

func (t targetConfig) GetIgnoreRepoFiles() []string {
	return t.ignoreRepoFiles
}

type TargetRegistry map[core.BuildTarget]targetConfig

func (t TargetRegistry) RegisterTarget(name string, protocArgs []string, copyOperation core.MountPath, ignoreRepoFiles []string) {
	if _, ok := t[core.BuildTarget(name)]; ok {
		panic(fmt.Errorf("target (%s) already registered", name))
	}

	//var postBuildScriptWrapper PostBuildScript = func(protoPackage string, target string, protocOutDir string, targetRepoHome string, gitBasePath string) error {
	//	fmt.Printf("Executing Post-Build script. Package: %s, target: %s\n", protoPackage, target)
	//	return postBuildScript(protoPackage, target, protocOutDir, targetRepoHome, gitBasePath)
	//}

	t[core.BuildTarget(name)] = targetConfig{
		protocArgs: protocArgs,
		//PostBuildScript: postBuildScriptWrapper,
		mountPaths:      []core.MountPath{copyOperation},
		ignoreRepoFiles: ignoreRepoFiles,
	}
}
func (t TargetRegistry) GetConfig(target core.BuildTarget) (targetConfig, bool) {
	targetConfig, ok := t[target]
	return targetConfig, ok

}
