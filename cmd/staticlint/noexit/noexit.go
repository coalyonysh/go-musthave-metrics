// Package noexit реализует анализатор NoExitInMain, запрещающий использование os.Exit в функции main.
package noexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const modulePrefix = "github.com/coalyonysh/go-musthave-metrics"

var NoExitAnalyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "запрещает использование os.Exit в функции main",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Фильтруем пакеты - анализируем только исходный код проекта
	pkgPath := pass.Pkg.Path()
	if !strings.HasPrefix(pkgPath, modulePrefix) || strings.Contains(pkgPath, "go-build") {
		return nil, nil
	}

	// Проверяем только пакет main
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if fn.Name.Name == "main" && fn.Body != nil {
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}

					sel, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					ident, ok := sel.X.(*ast.Ident)
					if !ok {
						return true
					}

					if ident.Name == "os" && sel.Sel.Name == "Exit" {
						pass.Reportf(call.Pos(), "использование os.Exit в функции main не рекомендуется; "+
							"вместо этого возвращайте ошибку или используйте log.Fatal")
					}

					return true
				})
			}
		}
	}

	return nil, nil
}
