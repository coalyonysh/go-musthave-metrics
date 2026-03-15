// multichecker объединяет статические анализаторы для проверки качества кода.
//
// Анализаторы включены:
// 1. Стандартные анализаторы golang.org/x/tools/go/analysis/passes:
//   - inspect: базовый анализатор для работы с AST
//   - printf: проверка формата printf
//   - structtag: проверка тегов структур
//   - fieldalignment: проверка выравнивания полей структуры
//   - shadow: проверка затенения переменных
//
// 2. Собственный анализатор:
//   - noexit: запрещает использование os.Exit в функции main
//
// Для расширенного статического анализа рекомендуется также использовать:
//   - staticcheck.io (отдельный инструмент): go install honnef.co/go/tools/cmd/staticcheck@latest
//   - golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
//
// Использование:
//
//	go run ./cmd/staticlint/... ./...
package main

import (
	"github.com/coalyonysh/go-musthave-metrics/cmd/staticlint/noexit"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
)

func main() {
	// multichecker.Main ожидает слайс указателей на анализаторы
	// Используем multichecker с анализаторами:
	// 1. Стандартные анализаторы golang.org/x/tools/go/analysis/passes
	// 2. Собственный анализатор noexit
	multichecker.Main(
		// Стандартные анализаторы golang.org/x/tools/go/analysis/passes
		inspect.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		shadow.Analyzer,
		fieldalignment.Analyzer,
		// Собственный анализатор
		noexit.NoExitAnalyzer,
	)
}
