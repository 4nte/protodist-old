package proto

import (
	"bytes"
	"fmt"
	"github.com/4nte/protodist/core"
	"log"
	"os/exec"
	"text/template"
)

type CompilerCommandBuilder struct {
	// Additional args (is this needed?)
	args []string
	// Compiler binary (protoc, protowrap, prototools?)
	compiler string
	// Where to save generated proto files
	outputDirectory string
	// protoc -I option
	protoPaths []string
	// Build targets (go, js, java, docs)
	targets []core.BuildTarget
	// Proto package to build
	protoPackage Package
	// Root path (-I param)
	rootPath string
	// Build targets
	plugins []Plugin
}

func NewProtocCmdBuilder() CompilerCommandBuilder {
	return CompilerCommandBuilder{
		compiler:        "protoc",
		outputDirectory: "./out",
		rootPath:        "",
	}
}
func (p *CompilerCommandBuilder) SetBuildTargets(targets ...core.BuildTarget) {
	p.targets = targets
}
func (p *CompilerCommandBuilder) SetOutputDir(dir string) {
	p.outputDirectory = dir
}
func (p *CompilerCommandBuilder) SetProtoPath(path ...string) {
	p.protoPaths = path
}
func (p *CompilerCommandBuilder) AddProtoPath(path ...string) {
	p.protoPaths = append(p.protoPaths, path...)
}
func (p *CompilerCommandBuilder) AddPlugin(path ...Plugin) {
	p.plugins = append(p.plugins, path...)
}
func (p *CompilerCommandBuilder) SetCompiler(compiler string) {
	p.compiler = compiler
}
func (p *CompilerCommandBuilder) SetProtoPackage(protoPackage Package) {
	p.protoPackage = protoPackage
}
func (p *CompilerCommandBuilder) SetProtoRootPath(rootPath string) {
	p.rootPath = rootPath
}
func (p *CompilerCommandBuilder) SetImportPaths([]string) {

}

func (p *CompilerCommandBuilder) Build() *exec.Cmd {
	type templateVariables struct {
		OutDir            string
		ProtocGenTsBinary string
	}
	vars := templateVariables{
		OutDir:            p.outputDirectory,
		ProtocGenTsBinary: "protoc-gen-ts",
	}

	var templatedArgs []string
	templatedArgs = append(templatedArgs, "-I"+p.rootPath)

	for _, protoPath := range p.protoPaths {
		templatedArgs = append(templatedArgs, "-I"+protoPath)
	}

	// Plugins
	for _, plugin := range p.plugins {
		for argName, argValue := range plugin.Args {
			argValueTemplate, err := template.New("arg").Parse(argValue)
			if err != nil {
				log.Fatal(fmt.Errorf("failed to template %s", argValue))
			}
			templatedArgValue := bytes.NewBufferString("")
			err = argValueTemplate.Execute(templatedArgValue, vars)
			if err != nil {
				log.Fatal(fmt.Errorf("failed to template arg value: %s: %w", templatedArgs, err))
			}
			arg := fmt.Sprintf("--%s=%s", argName, templatedArgValue)
			templatedArgs = append(templatedArgs, arg)

		}
	}

	if len(p.targets) == 0 {
		panic("no targets set")
	}

	// Add proto files to compile
	for _, protoFile := range p.protoPackage.Files {
		templatedArgs = append(templatedArgs, protoFile)
	}

	for _, target := range p.targets {
		targetConfig, ok := DefaultTargetRegistry[target]
		if !ok {
			log.Fatal(fmt.Errorf("protoc arg creator not found target %s", target))
		}

		argTemplateStrings := targetConfig.protocArgs
		for _, argTemplate := range argTemplateStrings {
			//fmt.Println("arg template", argTemplate)
			template, err := template.New("arg").Parse(argTemplate)
			if err != nil {
				log.Fatal(fmt.Errorf("failed to template %s", target))
			}
			argBuffer := bytes.NewBufferString("")
			err = template.Execute(argBuffer, vars)
			if err != nil {
				log.Fatal(fmt.Errorf("failed to template arg: %s: %w", templatedArgs, err))
			}
			templatedArgs = append(templatedArgs, argBuffer.String())

		}

		//fmt.Printf("target templatedArgs: %v", templatedArgs)
	}

	fmt.Println("templated args", templatedArgs)
	buildCmd := exec.Command(p.compiler, templatedArgs...)
	buildCmd.Dir = p.rootPath
	return buildCmd
}
