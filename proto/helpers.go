package proto

import (
	"fmt"
)

// format: proto-{package}-{lang}
func BuildTargetProtoRepoName(packageName string, buildTarget string) string {
	repoPrefix := "proto"
	repoName := fmt.Sprintf("%s-%s-%s", repoPrefix, packageName, buildTarget)
	return repoName
}
