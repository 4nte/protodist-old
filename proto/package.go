package proto

type PackageName string

type Package struct {
	Name  PackageName
	Files []string
}

type Packages []Package

func (p Packages) FindByName(name PackageName) (Package, bool) {
	for _, pkg := range p {
		if pkg.Name == name {
			return pkg, true
		}
	}
	return Package{}, false
}
