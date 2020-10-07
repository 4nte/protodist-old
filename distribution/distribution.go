package distribution

import (
	"fmt"
	"github.com/4nte/protodist/core"
	"github.com/4nte/protodist/proto"
)

// Encapsulates multiple targets for a set of packages
type Config struct {
	Name                 string              `yaml:"name"`
	ProtoPackages        []proto.PackageName `yaml:"protoPackages"`
	DistributionPackages []Package           `yaml:"distributionPackages"`
	Mount                string              `yaml:"mount"`
}

func NewPackage(target core.BuildTarget, packageOption string) Package {
	return Package{
		Target:             target,
		ProtoPackageOption: packageOption,
	}
}

type Package struct {
	Target             core.BuildTarget `yaml:"target"`
	ProtoPackageOption string           `yaml:"packageOption"`
	ProtoPlugins       []proto.Plugin   `yaml:"plugins"`
}

func (c Config) ToStrategies() []Strategy {
	var strategies []Strategy
	for _, distPackage := range c.DistributionPackages {
		// TODO: allow custom mount paths for every target per strategy or similar.
		targetConfig, ok := proto.DefaultTargetRegistry.GetConfig(distPackage.Target)
		if !ok {
			panic(fmt.Errorf("no target config found for: %s", distPackage.Target))
		}
		strategies = append(strategies, NewStrategy(c.Name, c.ProtoPackages, distPackage.Target, targetConfig.GetMountPaths(), targetConfig.GetIgnoreRepoFiles(), distPackage.ProtoPackageOption, distPackage.ProtoPlugins))
	}
	return strategies
}

func NewStrategy(name string, packages []proto.PackageName, target core.BuildTarget, mountPaths []core.MountPath, ignoreRepoFiles []string, packageOption string, plugins []proto.Plugin) Strategy {
	return Strategy{
		Name:            name,
		Packages:        packages,
		Target:          target,
		Mount:           mountPaths,
		IgnoreRepoFiles: ignoreRepoFiles,
		PackageOption:   packageOption,
		Plugins:         plugins,
	}
}

type Strategy struct {
	Name     string
	Packages []proto.PackageName
	Target   core.BuildTarget
	// Set of paths to mount to target repo (generated stubs)
	Mount           []core.MountPath
	IgnoreRepoFiles []string
	PackageOption   string
	Plugins         []proto.Plugin
}

// Name of the repository (e.g
type RepoName string

func (s Strategy) GetDistributionRepoName() RepoName {
	return RepoName(fmt.Sprintf("%s-%s", s.Name, s.Target))
}
