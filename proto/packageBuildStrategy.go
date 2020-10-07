package proto

type ProtoPackageBuildStrategy struct {
	BuildTargets []string
	ProtoPackage string
	ProtoFiles   []string
}

func (p *ProtoPackageBuildStrategy) setBuildTargets(targets ...string) {
	p.BuildTargets = targets
}
