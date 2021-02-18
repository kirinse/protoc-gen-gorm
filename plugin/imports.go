package plugin

import (
	"fmt"
	"strings"
)

var specialImports = map[string]struct{}{
	"github.com/edhaight/protoc-gen-gorm/types":              {},
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource": {},
	"github.com/golang/protobuf/ptypes/timestamp":            {},
}

type pkgImport struct {
	packagePath string
	alias       string
}

// Import takes a package and adds it to the list of packages to import
// It will generate a unique new alias using the last portion of the import path
// unless the package is already imported for this file. Either way, it returns
// the package alias
func (p *OrmPlugin) Import(packagePath string) string {
	subpath := packagePath[strings.LastIndex(packagePath, "/")+1:]
	// package will always be suffixed with an integer to prevent any collisions
	// with standard package imports
	for i := 1; ; i++ {
		newAlias := fmt.Sprintf("%s%d", strings.Replace(subpath, ".", "_", -1), i)
		if pkg, ok := p.GetFileImports().packages[newAlias]; ok {
			if packagePath == pkg.packagePath {
				return pkg.alias
			}
		} else {
			p.GetFileImports().packages[newAlias] = &pkgImport{packagePath: packagePath, alias: newAlias}
			return newAlias
		}
	}
	// Should never reach here
}

type fileImports struct {
	// wktPkgName string
	// typesToRegister []string
	// stdImports []string
	packages map[string]*pkgImport
}

func newFileImports() *fileImports {
	return &fileImports{packages: make(map[string]*pkgImport)}
}

func (p *OrmPlugin) GetFileImports() *fileImports {
	return p.fileImports[p.currentFile]
}
