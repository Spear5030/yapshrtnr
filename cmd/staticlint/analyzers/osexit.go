// Package analyzers содержит свои статические анализаторы
package analyzers

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer проверяет наличие вызова os.Exit в main функции
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "OsExitAnalyzer",
	Doc:  "Check for os.Exit in main func in main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.Package:
				if x.Name != "main" {
					return false
				}
			case *ast.FuncDecl:
				if x.Name.String() != "main" {
					return false
				}
			case *ast.Ident:
				if x.Name != "os" {
					return false
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "os.Exit in main func")
				}
			}
			return true
		})

	}
	return nil, nil
}
