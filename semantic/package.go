package semantic

import (
	"github.com/pkg/errors"
)

const PackageMain = "main"

// PackageType represents the type and name of a package.
type PackageType struct {
	Name string
	Type PolyType
}

// CreatePackage constructs a Package from the node. The node must contain a Program node
// with a valid PackageClause
func CreatePackage(n Node, importer Importer) (PackageType, error) {
	v := new(packageCreateVisitor)
	Walk(v, n)
	if v.name == "" {
		return PackageType{}, errors.New("no package clause found")
	}
	ts, err := InferTypes(n, importer)
	if err != nil {
		return PackageType{}, err
	}

	// Create object type from all top level identifiers
	types := make(map[string]PolyType, len(v.body))
	names := make(LabelSet, 0, len(v.body))
	for _, s := range v.body {
		assignment, ok := s.(*NativeVariableAssignment)
		if !ok {
			continue
		}
		typ, err := ts.PolyTypeOf(assignment.Init)
		if err != nil {
			return PackageType{}, err
		}
		name := assignment.Identifier.Name
		types[name] = typ
		names = append(names, name)
	}
	return PackageType{
		Name: v.name,
		Type: NewObjectPolyType(
			types,
			nil,
			names,
		),
	}, nil
}

type packageCreateVisitor struct {
	name string
	body []Statement
}

func (p *packageCreateVisitor) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *File:
		p.body = n.Body
	case *PackageClause:
		if n.Name != nil {
			p.name = n.Name.Name
		}
		return nil
	}
	if p.body != nil && p.name != "" {
		// We found the package name and the body, stop walking
		return nil
	}
	return p
}

func (p *packageCreateVisitor) Done(node Node) {}
