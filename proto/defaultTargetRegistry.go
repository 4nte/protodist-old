package proto

import (
	"github.com/4nte/protodist/core"
)

var DefaultTargetRegistry = make(TargetRegistry)

var globalIgnoreFiles = []string{".gitignore", "README.md", ".git", "go.mod", "build.gradle", "package.json", "package.lock", ".gradle"}

func globalIgnoreFilesWith(files ...string) []string {
	var ignoreFiles []string

	ignoreFiles = append(ignoreFiles, globalIgnoreFiles...)
	ignoreFiles = append(ignoreFiles, files...)
	return ignoreFiles
}

func init() {
	// Default target registry assumes that plugins live in $PATH
	DefaultTargetRegistry.RegisterTarget("go", []string{"--go_out=plugins=grpc:{{.OutDir}}"}, core.MountPath{Source: "{{.GitBasePath}}/{{.DistributionRepoName}}", Destination: "."}, globalIgnoreFilesWith("go.mod", "go.sum"))
	DefaultTargetRegistry.RegisterTarget("java", []string{"--grpc-java_out=:{{.OutDir}}", "--java_out=:{{.OutDir}}" /*"--plugin=protoc-gen-grpc-java=protoc-gen-grpc-java"*/}, core.MountPath{Source: ".", Destination: "src/main/java/"}, globalIgnoreFilesWith())
	DefaultTargetRegistry.RegisterTarget("js", []string{"--ts_out=service=grpc-web:{{.OutDir}}", "--js_out=import_style=commonjs,binary:{{.OutDir}}", "--plugin=protoc-gen-ts=/usr/bin/protoc-gen-ts"}, core.MountPath{Source: ".", Destination: "."}, globalIgnoreFilesWith())
}
