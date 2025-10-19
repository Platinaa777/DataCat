package httpapi_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAPIAdaptersUseApplicationOperations(t *testing.T) {
	err := filepath.WalkDir("..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, dependency := range file.Imports {
			name, err := strconv.Unquote(dependency.Path.Value)
			if err != nil {
				return err
			}
			for _, forbidden := range []string{"/internal/repository/", "/internal/extapi/", "/internal/pkg/cache"} {
				if strings.Contains(name, forbidden) {
					t.Errorf("%s imports infrastructure: %s", path, name)
				}
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if selector.Sel.Name == "Store" {
				t.Errorf("%s accesses application internals: %s", path, selector.Sel.Name)
			}
			if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Name == "usecases" {
				switch selector.Sel.Name {
				case "Read", "List", "Create", "Update":
					t.Errorf("%s bypasses application operations: %s", path, selector.Sel.Name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
